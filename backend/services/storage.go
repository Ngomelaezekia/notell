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
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.B2Region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.B2KeyID, cfg.B2ApplicationKey, "")))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize B2 credentials: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.B2Endpoint, "/"))
		o.UsePathStyle = true
	})
	return &s3MediaStorage{client: client, bucket: cfg.B2Bucket, signer: NewB2MediaSigner(client, cfg.B2Bucket)}, nil
}

func cleanLocalMediaPath(key string) (string, error) {
	c := filepath.Clean(filepath.FromSlash(key))
	if c == "." || filepath.IsAbs(c) || c == ".." || strings.HasPrefix(c, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid local media key")
	}
	return filepath.Join(".", c), nil
}

func (s *localMediaStorage) Put(_ context.Context, key, src, _ string) error {
	dst, err := cleanLocalMediaPath(key)
	if err != nil { return err }
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil { return fmt.Errorf("create local media directory: %w", err) }
	in, err := os.Open(src)
	if err != nil { return fmt.Errorf("open local media source: %w", err) }
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".media-*")
	if err != nil { return err }
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err = io.Copy(tmp, in); err != nil { tmp.Close(); return fmt.Errorf("copy local media: %w", err) }
	if err = tmp.Close(); err != nil { return err }
	if err = os.Chmod(tmpPath, 0644); err != nil { return err }
	if err = os.Rename(tmpPath, dst); err != nil { return fmt.Errorf("commit local media: %w", err) }
	return nil
}

func (s *localMediaStorage) Delete(_ context.Context, key string) error {
	p, err := cleanLocalMediaPath(key)
	if err != nil { return err }
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) { return fmt.Errorf("delete local media: %w", err) }
	return nil
}

func (s *localMediaStorage) Exists(_ context.Context, key string) (bool, error) {
	p, err := cleanLocalMediaPath(key)
	if err != nil { return false, err }
	_, err = os.Stat(p)
	if err == nil { return true, nil }
	if errors.Is(err, os.ErrNotExist) { return false, nil }
	return false, err
}

func (s *localMediaStorage) Open(_ context.Context, key, rng string) (io.ReadCloser, string, int64, string, error) {
	p, err := cleanLocalMediaPath(key)
	if err != nil { return nil, "", 0, "", err }
	f, err := os.Open(p)
	if err != nil { return nil, "", 0, "", err }
	info, err := f.Stat()
	if err != nil { f.Close(); return nil, "", 0, "", err }
	ct := mimeTypeForExtension(strings.ToLower(filepath.Ext(p)))
	start, end, ok := parseByteRange(rng, info.Size())
	if rng != "" && !ok { f.Close(); return nil, "", 0, "", errors.New("invalid byte range") }
	if ok {
		if _, err = f.Seek(start, io.SeekStart); err != nil { f.Close(); return nil, "", 0, "", err }
		return &limitedReadCloser{Reader: io.LimitReader(f, end-start+1), Closer: f}, ct, end-start+1, fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()), nil
	}
	return f, ct, info.Size(), "", nil
}

func mimeTypeForExtension(ext string) string {
	switch ext {
	case ".jpg", ".jpeg": return "image/jpeg"
	case ".png": return "image/png"
	case ".webp": return "image/webp"
	case ".mp3": return "audio/mpeg"
	case ".mp4": return "video/mp4"
	case ".mov": return "video/quicktime"
	case ".webm": return "video/webm"
	default: return "application/octet-stream"
	}
}

func (s *s3MediaStorage) Put(ctx context.Context, key, path, contentType string) error {
	f, err := os.Open(path)
	if err != nil { observability.MediaStorageFailed("put_open", err); return fmt.Errorf("open media for B2 upload: %w", err) }
	defer f.Close()
	info, err := f.Stat()
	if err != nil { return err }
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), Body: f, ContentType: aws.String(contentType), ContentLength: aws.Int64(info.Size()), CacheControl: aws.String("private, max-age=31536000, immutable")})
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
	if s3StatusCode(err) == 404 { return false, nil }
	observability.MediaStorageFailed("exists", err)
	return false, fmt.Errorf("check media in Backblaze B2: %w", err)
}

func s3StatusCode(err error) int {
	type statusCoder interface{ HTTPStatusCode() int }
	var sc statusCoder
	if errors.As(err, &sc) { return sc.HTTPStatusCode() }
	return 0
}

func (s *s3MediaStorage) Open(ctx context.Context, key, rng string) (io.ReadCloser, string, int64, string, error) {
	input := &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}
	if rng != "" {
		if !validOpenEndedByteRange(rng) { return nil, "", 0, "", errors.New("invalid byte range") }
		input.Range = aws.String(rng)
	}
	out, err := s.client.GetObject(ctx, input)
	if err != nil { observability.MediaStorageFailed("open", err); return nil, "", 0, "", fmt.Errorf("read media from Backblaze B2: %w", err) }
	contentType := "application/octet-stream"
	if out.ContentType != nil && strings.TrimSpace(*out.ContentType) != "" { contentType = *out.ContentType }
	var size int64
	if out.ContentLength != nil { size = *out.ContentLength }
	contentRange := ""
	if out.ContentRange != nil { contentRange = *out.ContentRange }
	return out.Body, contentType, size, contentRange, nil
}

func validOpenEndedByteRange(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "bytes=") { return false }
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 { return false }
	if parts[0] == "" {
		if parts[1] == "" { return false }
		n, err := strconv.ParseInt(parts[1], 10, 64)
		return err == nil && n > 0
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 { return false }
	if parts[1] == "" { return true }
	end, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil && end >= start
}

func parseByteRange(value string, size int64) (int64, int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "bytes=") { return 0, 0, false }
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 { return 0, 0, false }
	if parts[0] == "" {
		if parts[1] == "" || size <= 0 { return 0, 0, false }
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || n <= 0 { return 0, 0, false }
		if n > size { n = size }
		return size-n, size-1, true
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size { return 0, 0, false }
	end := size-1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start { return 0, 0, false }
		if end >= size { end = size-1 }
	}
	return start, end, true
}

func (s *s3MediaStorage) GeneratePlaybackURL(ctx context.Context, key string, d time.Duration) (string, error) {
	if s == nil || s.signer == nil { return "", errors.New("B2 playback signer is not initialized") }
	return s.signer.GeneratePlaybackURL(ctx, key, d)
}

func SetMediaStorage(storage MediaStorage) {
	mediaStorageRegistry.Lock()
	mediaStorageRegistry.storage = storage
	mediaStorageRegistry.Unlock()
}

func GenerateMediaPlaybackURL(ctx context.Context, key string, d time.Duration) (string, error) {
	mediaStorageRegistry.RLock()
	storage := mediaStorageRegistry.storage
	mediaStorageRegistry.RUnlock()
	if storage == nil { return "", errors.New("media storage is not registered") }
	signer, ok := storage.(MediaPlaybackSigner)
	if !ok { return "", errors.New("media storage does not support signed playback") }
	return signer.GeneratePlaybackURL(ctx, key, d)
}

func DeleteMediaObject(ctx context.Context, key string) error {
	mediaStorageRegistry.RLock()
	storage := mediaStorageRegistry.storage
	mediaStorageRegistry.RUnlock()
	if storage == nil { return errors.New("media storage is not registered") }
	return storage.Delete(ctx, key)
}

func CleanupStoredMedia(ctx context.Context, storage MediaStorage, key string) error {
	if storage == nil { return errors.New("media storage is not configured") }
	if !IsMediaObjectKeySafe(key) { return errors.New("invalid media object key") }
	return storage.Delete(ctx, key)
}

func IsMediaObjectKeySafe(key string) bool {
	key = strings.TrimSpace(key)
	return key != "" && filepath.Base(key) == key && key != "." && !strings.ContainsAny(key, `/\\`)
}

func RollbackStoredUpload(ctx context.Context, storage MediaStorage, filename string) error {
	return CleanupStoredMedia(ctx, storage, MediaObjectKey(filename))
}

func StartMediaReconciler(ctx context.Context, storage MediaStorage, db *gorm.DB) {
	SetMediaStorage(storage)
	go func() {
		reconcile := func() {
			ReconcileMediaState(db)
			if s, ok := storage.(*s3MediaStorage); ok {
				if err := s.reconcile(ctx, db); err != nil { observability.MediaStorageFailed("reconcile", err); log.Printf("media reconciliation failed: %v", err) }
				if err := ReconcileDatabaseMedia(ctx, storage, db); err != nil { observability.MediaStorageFailed("reconcile_db", err); log.Printf("media database availability reconciliation failed: %v", err) }
			}
		}
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
	if storage == nil || db == nil { return errors.New("media storage and database are required") }
	var uploads []models.Upload
	if err := db.Select("id, filename, post_id").Where("filename <> ''").Find(&uploads).Error; err != nil { return fmt.Errorf("list database media: %w", err) }
	for _, upload := range uploads {
		exists, err := storage.Exists(ctx, MediaObjectKey(upload.Filename))
		if err != nil { return err }
		if exists { continue }
		message := "physical media object is missing from durable storage"
		if upload.PostID == nil { message += "; upload must be replaced" }
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
			err = db.Where("filename = ?", filename).First(&upload).Error
			if err == nil { continue }
			if !errors.Is(err, gorm.ErrRecordNotFound) { return fmt.Errorf("find upload for %q: %w", filename, err) }
			created := time.Now()
			if object.LastModified != nil { created = *object.LastModified }
			if created.After(cutoff) { continue }
			if err := s.Delete(ctx, key); err != nil { return fmt.Errorf("delete orphan media %q: %w", key, err) }
		}
	}
	return nil
}

type limitedReadCloser struct { io.Reader; Closer io.Closer }
func (r *limitedReadCloser) Close() error { return r.Closer.Close() }
