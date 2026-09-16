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
