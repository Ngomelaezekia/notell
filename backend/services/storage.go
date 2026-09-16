package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"notell/config"
	"notell/models"
	"notell/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
)

type MediaStorage interface {
	Put(context.Context, string, string, string) error
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
	Open(context.Context, string, string) (io.ReadCloser, string, int64, string, error)
}

type MediaPlaybackSigner interface {
	GeneratePlaybackURL(context.Context, string, time.Duration) (string, error)
}

type localMediaStorage struct{}
type s3MediaStorage struct {
	client *s3.Client
	bucket string
	signer *B2MediaSigner
}

var mediaStorageRegistry struct {
	sync.RWMutex
	storage MediaStorage
}

const orphanMediaGracePeriod = time.Hour

func NewMediaStorage(cfg *config.Config) (MediaStorage, error) {
	if strings.EqualFold(cfg.StorageDriver, "local") {
		return &localMediaStorage{}, nil
	}
	if !strings.EqualFold(cfg.StorageDriver, "b2") {
		return nil, fmt.Errorf("unsupported STORAGE_DRIVER %q", cfg.StorageDriver)
	}
	if cfg.B2Endpoint == "" || cfg.B2Bucket == "" || cfg.B2KeyID == "" || cfg.B2ApplicationKey == "" || cfg.B2Region == "" {
		return nil, errors.New("Backblaze B2 storage configuration is incomplete")
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.B2Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.B2KeyID, cfg.B2ApplicationKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize B2 credentials: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.B2Endpoint, "/"))
		o.UsePathStyle = true
	})
	return &s3MediaStorage{
		client: client,
		bucket: cfg.B2Bucket,
		signer: NewB2MediaSigner(client, cfg.B2Bucket),
	}, nil
}

func (s *s3MediaStorage) GeneratePlaybackURL(ctx context.Context, path string, d time.Duration) (string, error) {
	if s == nil || s.signer == nil {
		return "", errors.New("B2 playback signer is not initialized")
	}
	return s.signer.GeneratePlaybackURL(ctx, path, d)
}

func SetMediaStorage(s MediaStorage) {
	mediaStorageRegistry.Lock()
	mediaStorageRegistry.storage = s
	mediaStorageRegistry.Unlock()
}

func GenerateMediaPlaybackURL(ctx context.Context, key string, d time.Duration) (string, error) {
	mediaStorageRegistry.RLock()
	storage := mediaStorageRegistry.storage
	mediaStorageRegistry.RUnlock()
	if storage == nil {
		return "", errors.New("media storage is not registered")
	}
	signer, ok := storage.(MediaPlaybackSigner)
	if !ok {
		return "", errors.New("media storage does not support signed playback")
	}
	return signer.GeneratePlaybackURL(ctx, key, d)
}

func DeleteMediaObject(ctx context.Context, key string) error {
	mediaStorageRegistry.RLock()
	s := mediaStorageRegistry.storage
	mediaStorageRegistry.RUnlock()
	if s == nil {
		return errors.New("media storage is not registered")
	}
	return s.Delete(ctx, key)
}

func StartMediaReconciler(ctx context.Context, storage MediaStorage, db *gorm.DB) {
	SetMediaStorage(storage)
	go func() {
		reconcile := func() {
			ReconcileMediaState(db)
			if s, ok := storage.(*s3MediaStorage); ok {
				if e := s.reconcile(ctx, db); e != nil {
					observability.MediaStorageFailed("reconcile", e)
					log.Printf("media reconciliation failed: %v", e)
				}
				if e := ReconcileDatabaseMedia(ctx, storage, db); e != nil {
					observability.MediaStorageFailed("reconcile_db", e)
					log.Printf("media database availability reconciliation failed: %v", e)
				}
			}
		}
		reconcile()
		t := time.NewTicker(30 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				reconcile()
			}
		}
	}()
}

func CleanupStoredMedia(ctx context.Context, storage MediaStorage, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("media object key is required")
	}
	if storage == nil {
		return errors.New("media storage is not configured")
	}
	return storage.Delete(ctx, key)
}

func IsMediaObjectKeySafe(key string) bool {
	key = strings.TrimSpace(key)
	return key != "" && filepath.Base(key) == key && key != "." && !strings.ContainsAny(key, `/\\`) 
}

func MarkMediaJobFailed(db *gorm.DB, job *models.MediaJob, reason string) error {
	if job == nil || db == nil {
		return errors.New("media job and database are required")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "media processing failed"
	}
	return db.Model(job).Updates(map[string]any{
		"status":       "failed",
		"error_message": reason,
		"finished_at":  time.Now().UTC(),
	}).Error
}

func CleanupStaleMediaJobs(db *gorm.DB, now time.Time) error {
	if db == nil {
		return errors.New("database is required")
	}
	cutoff := now.UTC().Add(-orphanMediaGracePeriod)
	return db.Model(&models.MediaJob{}).
		Where("status IN ? AND updated_at < ?", []string{"queued", "processing"}, cutoff).
		Updates(map[string]any{
			"status":        "failed",
			"error_message": "stale media job recovered by reconciliation",
			"finished_at":   now.UTC(),
		}).Error
}

func ReconcileMediaState(db *gorm.DB) {
	if db == nil {
		return
	}
	_ = CleanupStaleMediaJobs(db, time.Now().UTC())
}

func ReconcileDatabaseMedia(ctx context.Context, storage MediaStorage, db *gorm.DB) error {
	if storage == nil || db == nil {
		return errors.New("media storage and database are required")
	}
	var uploads []models.Upload
	if err := db.Where("created_at < ?", time.Now().UTC().Add(-orphanMediaGracePeriod)).Find(&uploads).Error; err != nil {
		return err
	}
	for i := range uploads {
		key := MediaObjectKey(uploads[i].Filename)
		if !IsMediaObjectKeySafe(uploads[i].Filename) {
			continue
		}
		exists, err := storage.Exists(ctx, key)
		if err != nil {
			return err
		}
		if !exists {
			_ = db.Delete(&uploads[i]).Error
		}
	}
	return nil
}

func BeginMediaUpload(ctx context.Context, db *gorm.DB, storage MediaStorage, filename, contentType string, put func(context.Context) error) (*models.Upload, error) {
	if db == nil || storage == nil || put == nil {
		return nil, errors.New("database, storage, and upload operation are required")
	}
	filename = strings.TrimSpace(filename)
	if !IsMediaObjectKeySafe(filename) {
		return nil, errors.New("invalid media filename")
	}
	if err := put(ctx); err != nil {
		return nil, fmt.Errorf("store media: %w", err)
	}
	upload := &models.Upload{Filename: filename, ContentType: contentType}
	if err := db.Create(upload).Error; err != nil {
		cleanupErr := CleanupStoredMedia(ctx, storage, MediaObjectKey(filename))
		if cleanupErr != nil {
			return nil, fmt.Errorf("create upload record: %w; cleanup stored media: %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("create upload record: %w", err)
	}
	return upload, nil
}

func CommitUploadRecord(db *gorm.DB, upload *models.Upload) error {
	if db == nil || upload == nil {
		return errors.New("database and upload are required")
	}
	return db.Save(upload).Error
}

func RollbackUpload(ctx context.Context, db *gorm.DB, storage MediaStorage, upload *models.Upload) error {
	if upload == nil {
		return errors.New("upload is required")
	}
	var errs []string
	if db != nil && upload.ID != 0 {
		if err := db.Delete(upload).Error; err != nil {
			errs = append(errs, "delete upload record: "+err.Error())
		}
	}
	if storage != nil && IsMediaObjectKeySafe(upload.Filename) {
		if err := CleanupStoredMedia(ctx, storage, MediaObjectKey(upload.Filename)); err != nil {
			errs = append(errs, "delete stored media: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

var _ = strconv.IntSize
var _ = os.ErrNotExist
