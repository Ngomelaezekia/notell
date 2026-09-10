package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// B2MediaSigner generates temporary playback URLs using the S3-compatible
// presigning API supported by Backblaze B2.
type B2MediaSigner struct {
	presigner *s3.PresignClient
	bucket    string
}

func NewB2MediaSigner(client *s3.Client, bucket string) *B2MediaSigner {
	return &B2MediaSigner{
		presigner: s3.NewPresignClient(client),
		bucket:    bucket,
	}
}

func (b *B2MediaSigner) GeneratePlaybackURL(objectPath string, expiry time.Duration) (string, error) {
	if b.presigner == nil {
		return "", fmt.Errorf("B2 presigner is not initialized")
	}

	result, err := b.presigner.PresignGetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(objectPath),
		},
		s3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", fmt.Errorf("generate B2 playback URL: %w", err)
	}

	return result.URL, nil
}
