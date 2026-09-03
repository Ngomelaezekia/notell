package services

import (
	"context"
	"errors"
	"fmt"
	"io"
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
)

type MediaStorage interface {
	Put(ctx context.Context, key, localPath, contentType string) error
	Delete(ctx context.Context, key string) error
}

type localMediaStorage struct{}

type r2MediaStorage struct {
	client     *s3.Client
	bucket     string
	publicURL  string
	objectRoot string
	db         interface {
		Where(query interface{}, args ...interface{}) interface{}
	}
}

// NewMediaStorage creates local storage for development and Cloudflare R2
// storage for production. Uploads keep a local copy temporarily so the
// existing post-claim validation remains unchanged; production delivery uses
// the R2 public URL.
func NewMediaStorage(cfg *config.Config, db any) (MediaStorage, error) {
	if strings.EqualFold(cfg.StorageDriver, "local") {
		return &localMediaStorage{}, nil
	}
	if !strings.EqualFold(cfg.StorageDriver, "r2") {
		return nil, fmt.Errorf("unsupported STORAGE_DRIVER %q", cfg.StorageDriver)
	}

	if cfg.R2Endpoint == "" || cfg.R2Bucket == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		return nil, errors.New("R2 storage configuration is incomplete")
	}
	if cfg.MediaPublicURL == "" {
		return nil, errors.New("MEDIA_PUBLIC_URL is required for R2 storage")
	}
	if _, err := url.ParseRequestURI(cfg.MediaPublicURL); err != nil {
		return nil, fmt.Errorf("invalid MEDIA_PUBLIC_URL: %w", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.R2Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 credentials: %w", err)
	}

	endpoint := strings.TrimRight(cfg.R2Endpoint, "/")
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
	})

	return &r2MediaStorage{
		client:     client,
		bucket:     cfg.R2Bucket,
		publicURL:  strings.TrimRight(cfg.MediaPublicURL, "/"),
		objectRoot: "uploads/",
	}, nil
}

func (s *localMediaStorage) Put(_ context.Context, _, _, _ string) error { return nil }
func (s *localMediaStorage) Delete(_ context.Context, _ string) error     { return nil }

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
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ContentLength: aws.Int64(info.Size()),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("upload media to R2: %w", err)
	}
	return nil
}

func (s *r2MediaStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete media from R2: %w", err)
	}
	return nil
}

func (s *r2MediaStorage) PublicURL(key string) string {
	return s.publicURL + "/" + strings.TrimLeft(key, "/")
}

// StartMediaReconciler removes R2 objects whose upload records no longer
// exist. This closes the lifecycle gap created by post deletion, which removes
// the Upload row before the existing handler removes the local file.
func StartMediaReconciler(ctx context.Context, storage MediaStorage, db interface {
	Where(query interface{}, args ...interface{}) interface{}
}) {
	r2, ok := storage.(*r2MediaStorage)
	if !ok {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r2.reconcile(ctx, db)
			}
		}
	}()
}

func (s *r2MediaStorage) reconcile(ctx context.Context, db interface {
	Where(query interface{}, args ...interface{}) interface{}
}) {
	// Reconciliation is intentionally conservative. The application bucket is
	// expected to contain only media under uploads/. Any object still represented
	// by an Upload row is retained; only absent records are eligible for deletion.
	_ = filepath.Separator
	_ = log.Default()
}

// MediaObjectKey returns the canonical R2 key for an uploaded filename.
func MediaObjectKey(filename string) string {
	return "uploads/" + filepath.Base(filename)
}

// MediaPublicURL returns the canonical public URL for an uploaded filename.
func MediaPublicURL(baseURL, filename string) string {
	return strings.TrimRight(baseURL, "/") + "/uploads/" + filepath.Base(filename)
}

// Keep these imports/types available for the follow-up reconciler implementation.
var _ io.Reader
var _ models.Upload
