package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
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
}

type localMediaStorage struct{}

type r2MediaStorage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

const orphanMediaGracePeriod = time.Hour

func NewMediaStorage(cfg *config.Config) (MediaStorage, error) {
	if strings.EqualFold(cfg.StorageDriver, "local") {
		return &localMediaStorage{}, nil
	}
	if !strings.EqualFold(cfg.StorageDriver, "r2") {
		return nil, fmt.Errorf("unsupported STORAGE_DRIVER %q", cfg.StorageDriver)
	}
	if cfg.R2Endpoint == "" || cfg.R2Bucket == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		return nil, errors.New("R2 storage configuration is incomplete")
	}
	if err := validateMediaPublicURL(cfg.MediaPublicURL); err != nil {
		return nil, fmt.Errorf("MEDIA_PUBLIC_URL: %w", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.R2Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 credentials: %w", err)
	}

	endpoint := strings.TrimRight(cfg.R2Endpoint, "/")
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
	})

	return &r2MediaStorage{client: client, bucket: cfg.R2Bucket, publicURL: strings.TrimRight(cfg.MediaPublicURL, "/")}, nil
}

func (s *localMediaStorage) Put(context.Context, string, string, string) error { return nil }
func (s *localMediaStorage) Delete(context.Context, string) error                { return nil }

func (s *r2MediaStorage) Put(ctx context.Context, key, localPath, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open media for R2 upload: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat media for R2 upload: %w", err)
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
		return fmt.Errorf("upload media to R2: %w", err)
	}
	return nil
}

func (s *r2MediaStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete media from R2: %w", err)
	}
	return nil
}

func (s *r2MediaStorage) PublicURL(key string) string {
	return s.publicURL + "/" + strings.TrimLeft(key, "/")
}

func StartMediaReconciler(ctx context.Context, storage MediaStorage, db *gorm.DB) {
	r2, ok := storage.(*r2MediaStorage)
	if !ok {
		return
	}

	reconcile := func() {
		if err := r2.reconcile(ctx, db); err != nil {
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

func (s *r2MediaStorage) reconcile(ctx context.Context, db *gorm.DB) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String("uploads/"),
	})
	cutoff := time.Now().Add(-orphanMediaGracePeriod)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list R2 media: %w", err)
		}
		for _, object := range page.Contents {
			if object.Key == nil || !strings.HasPrefix(*object.Key, "uploads/") {
				continue
			}
			if object.LastModified != nil && object.LastModified.After(cutoff) {
				// A newly-created object may be between PutObject and the Upload
				// database insert. Give that transaction window a safe grace period.
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
				return fmt.Errorf("check R2 media ownership for %q: %w", filename, err)
			}

			if err := s.Delete(ctx, *object.Key); err != nil {
				log.Printf("failed to delete orphaned R2 media %q: %v", *object.Key, err)
			}
		}
	}
	return nil
}

func MediaObjectKey(filename string) string {
	return "uploads/" + filepath.Base(filename)
}

func MediaPublicURL(baseURL, filename string) string {
	return strings.TrimRight(baseURL, "/") + "/uploads/" + filepath.Base(filename)
}

func validateMediaPublicURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("must be an absolute http or https URL")
	}
	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must be a public origin without path, query, fragment, or user information")
	}
	return nil
}
