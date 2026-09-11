package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultPlaybackURLExpiry = 5 * time.Minute

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
	objectPath = strings.TrimSpace(objectPath)
	if objectPath == "" || !strings.HasPrefix(objectPath, "uploads/") || strings.Contains(objectPath, "..") {
		return "", fmt.Errorf("invalid B2 media object path")
	}
	if expiry <= 0 {
		expiry = defaultPlaybackURLExpiry
	}
	if expiry > 15*time.Minute {
		expiry = 15 * time.Minute
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
