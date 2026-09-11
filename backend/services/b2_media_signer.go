package services

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultPlaybackURLExpiry = 5 * time.Minute
const maxPlaybackURLExpiry = 15 * time.Minute

// validateB2MediaObjectPath keeps playback scoped to the application's upload
// namespace. MediaObjectKey always produces this shape, so accepting nested or
// traversal paths here would create an unnecessary storage escape hatch.
func validateB2MediaObjectPath(objectPath string) error {
	objectPath = strings.TrimSpace(objectPath)
	if objectPath == "" || !strings.HasPrefix(objectPath, "uploads/") {
		return fmt.Errorf("invalid B2 media object path")
	}
	if path.Clean(objectPath) != objectPath || strings.Contains(objectPath, "\\") {
		return fmt.Errorf("invalid B2 media object path")
	}
	name := strings.TrimPrefix(objectPath, "uploads/")
	if name == "" || strings.Contains(name, "/") || name == "." || name == ".." {
		return fmt.Errorf("invalid B2 media object path")
	}
	return nil
}

// B2MediaSigner generates short-lived playback URLs using Backblaze B2's
// S3-compatible presigning API.
type B2MediaSigner struct {
	presigner *s3.PresignClient
	bucket    string
}

func NewB2MediaSigner(client *s3.Client, bucket string) *B2MediaSigner {
	if client == nil || strings.TrimSpace(bucket) == "" {
		return &B2MediaSigner{bucket: bucket}
	}
	return &B2MediaSigner{
		presigner: s3.NewPresignClient(client),
		bucket:    strings.TrimSpace(bucket),
	}
}

func (b *B2MediaSigner) GeneratePlaybackURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	if b == nil || b.presigner == nil {
		return "", fmt.Errorf("B2 presigner is not initialized")
	}
	if err := validateB2MediaObjectPath(objectPath); err != nil {
		return "", err
	}
	objectPath = strings.TrimSpace(objectPath)
	if expiry <= 0 {
		expiry = defaultPlaybackURLExpiry
	}
	if expiry > maxPlaybackURLExpiry {
		expiry = maxPlaybackURLExpiry
	}

	result, err := b.presigner.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(objectPath),
		},
		s3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", fmt.Errorf("generate B2 playback URL: %w", err)
	}
	if strings.TrimSpace(result.URL) == "" {
		return "", fmt.Errorf("generate B2 playback URL: empty URL")
	}
	return result.URL, nil
}
