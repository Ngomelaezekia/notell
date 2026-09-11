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
	Put(ctx context.Context, key, localPath, contentType string) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Open(ctx context.Context, key, byteRange string) (io.ReadCloser, string, int64, string, error)
}

type mediaPlaybackSigner interface {
	GeneratePlaybackURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
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

	endpoint := strings.TrimRight(cfg.B2Endpoint, "/")
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})

	return &s3MediaStorage{
		client: client,
		bucket: cfg.B2Bucket,
		signer: NewB2MediaSigner(client, cfg.B2Bucket),
	}, nil
}

func (s *localMediaStorage) Put(context.Context, string, string, string) error { return nil }
func (s *localMediaStorage) Delete(_ context.Context, key string) error {
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	if cleanKey == "." || filepath.IsAbs(cleanKey) || cleanKey == ".." || strings.HasPrefix(cleanKey, ".."+string(filepath.Separator)) {
		return errors.New("invalid local media key")
	}
	path := filepath.Join(".", cleanKey)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete local media: %w", err)
	}
	return nil
}
func (s *localMediaStorage) Exists(_ context.Context, key string) (bool, error) {
	path := filepath.Join(".", filepath.FromSlash(key))
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if errors.Is(err, os.ErrNotExist) { return false, nil }
	return false, fmt.Errorf("stat local media: %w", err)
}
func (s *localMediaStorage) Open(_ context.Context, key, byteRange string) (io.ReadCloser, string, int64, string, error) {
	path := filepath.Join(".", filepath.FromSlash(key))
	file, err := os.Open(path)
	if err != nil { return nil, "", 0, "", err }
	info, err := file.Stat()
	if err != nil { file.Close(); return nil, "", 0, "", err }
	start, end, ok := parseByteRange(byteRange, info.Size())
	if byteRange != "" && !ok { file.Close(); return nil, "", 0, "", fmt.Errorf("invalid byte range") }
	if ok {
		if _, err := file.Seek(start, io.SeekStart); err != nil { file.Close(); return nil, "", 0, "", err }
		return &limitedReadCloser{Reader: io.LimitReader(file, end-start+1), Closer: file}, "", end - start + 1, fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()), nil
	}
	return file, "", info.Size(), "", nil
}

func (s *s3MediaStorage) Put(ctx context.Context, key, localPath, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil { observability.MediaStorageFailed("put_open", err); return fmt.Errorf("open media for B2 upload: %w", err) }
	defer file.Close()
	info, err := file.Stat()
	if err != nil { observability.MediaStorageFailed("put_stat", err); return fmt.Errorf("stat media for B2 upload: %w", err) }
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), Body: file, ContentType: aws.String(contentType), ContentLength: aws.Int64(info.Size()), CacheControl: aws.String("private, max-age=31536000, immutable")})
	if err != nil { observability.MediaStorageFailed("put", err); return fmt.Errorf("upload media to Backblaze B2: %w", err) }
	return nil
}

func (s *s3MediaStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil { observability.MediaStorageFailed("delete", err); return fmt.Errorf("delete media from Backblaze B2: %w", err) }
	return nil
}

func (s *s3MediaStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err == nil { return true, nil }
	if s3StatusCode(err) == httpStatusNotFound { return false, nil }
	observability.MediaStorageFailed("exists", err)
	return false, fmt.Errorf("check media in Backblaze B2: %w", err)
}

const httpStatusNotFound = 404

func s3StatusCode(err error) int {
	type statusCoder interface{ HTTPStatusCode() int }
	var sc statusCoder
	if errors.As(err, &sc) { return sc.HTTPStatusCode() }
	return 0
}

func (s *s3MediaStorage) Open(ctx context.Context, key, byteRange string) (io.ReadCloser, string, int64, string, error) {
	input := &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}
	if byteRange != "" {
		if !validOpenEndedByteRange(byteRange) {
			return nil, "", 0, "", fmt.Errorf("invalid byte range")
		}
		input.Range = aws.String(byteRange)
	}
	output, err := s.client.GetObject(ctx, input)
	if err != nil { observability.MediaStorageFailed("open", err); return nil, "", 0, "", fmt.Errorf("read media from Backblaze B2: %w", err) }
	contentType := "application/octet-stream"
	if output.ContentType != nil && strings.TrimSpace(*output.ContentType) != "" { contentType = *output.ContentType }
	contentLength := int64(0)
	if output.ContentLength != nil { contentLength = *output.ContentLength }
	contentRange := ""
	if output.ContentRange != nil { contentRange = *output.ContentRange }
	return output.Body, contentType, contentLength, contentRange, nil
}

func validOpenEndedByteRange(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "bytes=") { return false }
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 { return false }
	if parts[0] == "" {
		if parts[1] == "" { return false }
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		return err == nil && suffix > 0
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 { return false }
	if parts[1] == "" { return true }
	end, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil && end >= start
}

func (s *s3MediaStorage) GeneratePlaybackURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	if s == nil || s.signer == nil { return "", errors.New("B2 playback signer is not initialized") }
	return s.signer.GeneratePlaybackURL(ctx, objectPath, expiry)
}

func SetMediaStorage(storage MediaStorage) { mediaStorageRegistry.Lock(); mediaStorageRegistry.storage = storage; mediaStorageRegistry.Unlock() }

func DeleteMediaObject(ctx context.Context, key string) error {
	mediaStorageRegistry.RLock(); storage := mediaStorageRegistry.storage; mediaStorageRegistry.RUnlock()
	if storage == nil { return errors.New("media storage is not registered") }
	return storage.Delete(ctx, key)
}

type limitedReadCloser struct { io.Reader; Closer io.Closer }
func (r *limitedReadCloser) Close() error { return r.Closer.Close() }

func parseByteRange(value string, size int64) (int64, int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "bytes=") { return 0, 0, false }
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 { return 0, 0, false }
	if parts[0] == "" {
		if parts[1] == "" || size < 0 { return 0, 0, false }
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 || size <= 0 { return 0, 0, false }
		if suffix > size { suffix = size }
		return size - suffix, size - 1, true
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 { return 0, 0, false }
	var end int64
	if parts[1] == "" {
		if size < 0 { return 0, 0, false }
		end = size - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start { return 0, 0, false }
	}
	if size >= 0 {
		if start >= size { return 0, 0, false }
		if end >= size { end = size - 1 }
	}
	return start, end, true
}

func StartMediaReconciler(ctx context.Context, storage MediaStorage, db *gorm.DB) {
	SetMediaStorage(storage)
	reconcile := func() {
		ReconcileMediaState(db)
		s3Store, ok := storage.(*s3MediaStorage)
		if !ok { return }
		if err := s3Store.reconcile(ctx, db); err != nil { observability.MediaStorageFailed("reconcile", err); log.Printf("media reconciliation failed: %v", err) }
		if err := ReconcileDatabaseMedia(ctx, storage, db); err != nil { observability.MediaStorageFailed("reconcile_db", err); log.Printf("media database availability reconciliation failed: %v", err) }
	}
	go func() {
		reconcile()
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done(): return
			case <-ticker.C: reconcile()
			}
		}
	}()
}

func ReconcileDatabaseMedia(ctx context.Context, storage MediaStorage, db *gorm.DB) error {
	var uploads []models.Upload
	if err := db.Select("id, filename, post_id").Where("filename <> ''").Find(&uploads).Error; err != nil { return fmt.Errorf("list database media: %w", err) }
	for _, upload := range uploads {
		exists, err := storage.Exists(ctx, MediaObjectKey(upload.Filename))
		if err != nil { return err }
		if exists { continue }
		message := "physical media object is missing from durable storage"
		if upload.PostID == nil { message = "physical media object is missing from durable storage; upload must be replaced" }
		if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", upload.ID).Updates(map[string]any{"status": "failed", "processing_error": message}).Error; err != nil { return fmt.Errorf("mark missing media upload=%d: %w", upload.ID, err) }
	}
	return nil
}

func (s *s3MediaStorage) reconcile(ctx context.Context, db *gorm.DB) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket), Prefix: aws.String("uploads/")})
	cutoff := time.Now().Add(-orphanMediaGracePeriod)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil { return fmt.Errorf("list Backblaze B2 media: %w", err) }
		for _, object := range page.Contents {
			if object.Key == nil || !strings.HasPrefix(*object.Key, "uploads/") { continue }
			key := *object.Key
			filename := strings.TrimPrefix(key, "uploads/")
			var upload models.Upload
			err := db.Where("filename = ?", filename).First(&upload).Error
			if err == nil { continue }
			if !errors.Is(err, gorm.ErrRecordNotFound) { return fmt.Errorf("find upload for %q: %w", filename, err) }
			createdAt := time.Now()
			if object.LastModified != nil { createdAt = *object.LastModified }
			if createdAt.After(cutoff) { continue }
			if err := s.Delete(ctx, key); err != nil { return fmt.Errorf("delete orphan media %q: %w", key, err) }
		}
	}
	return nil
}
