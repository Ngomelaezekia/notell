package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notell/config"
	"notell/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
)

type MediaStorage interface {
	Put(ctx context.Context, key, localPath, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type localMediaStorage struct{}

type s3MediaStorage struct {
	client    *s3.Client
	bucket    string
	publicURL string
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

	publicURL := strings.TrimRight(cfg.MediaPublicURL, "/")
	return &s3MediaStorage{client: client, bucket: cfg.B2Bucket, publicURL: publicURL}, nil
}

func (s *localMediaStorage) Put(context.Context, string, string, string) error { return nil }
func (s *localMediaStorage) Delete(context.Context, string) error                { return nil }
func (s *localMediaStorage) PublicURL(key string) string                         { return MediaPublicURL("", key) }

func (s *s3MediaStorage) Put(ctx context.Context, key, localPath, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open media for B2 upload: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat media for B2 upload: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(info.Size()),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("upload media to Backblaze B2: %w", err)
	}
	return nil
}

func (s *s3MediaStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete media from Backblaze B2: %w", err)
	}
	return nil
}

func (s *s3MediaStorage) PublicURL(key string) string {
	return MediaPublicURL(s.publicURL, key)
}

func StartMediaReconciler(ctx context.Context, storage MediaStorage, db *gorm.DB) {
	s3Store, ok := storage.(*s3MediaStorage)
	if !ok {
		return
	}

	reconcile := func() {
		if err := s3Store.reconcile(ctx, db); err != nil {
			log.Printf("media reconciliation failed: %v", err)
		}
	}

	go func() {
		reconcile()
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reconcile()
			}
		}
	}()
}

func (s *s3MediaStorage) reconcile(ctx context.Context, db *gorm.DB) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String("uploads/"),
	})
	cutoff := time.Now().Add(-orphanMediaGracePeriod)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list Backblaze B2 media: %w", err)
		}
		for _, object := range page.Contents {
			if object.Key == nil || !strings.HasPrefix(*object.Key, "uploads/") {
				continue
			}
			if object.LastModified != nil && object.LastModified.After(cutoff) {
				continue
			}
			filename := filepath.Base(strings.TrimPrefix(*object.Key, "uploads/"))
			if filename == "." || filename == "" {
				continue
			}

			var upload models.Upload
			err := db.Select("id").Where("filename = ?", filename).First(&upload).Error
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("check B2 media ownership for %q: %w", filename, err)
			}

			if err := s.Delete(ctx, *object.Key); err != nil {
				log.Printf("failed to delete orphaned B2 media %q: %v", *object.Key, err)
			}
		}
	}
	return nil
}

func MediaObjectKey(filename string) string {
	return "uploads/" + filepath.Base(filename)
}

func MediaPublicURL(baseURL, key string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(key, "/")
}
