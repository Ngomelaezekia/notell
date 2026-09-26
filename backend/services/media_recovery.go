package services

import (
	"context"
	"log"

	"notell/observability"
	"time"
)

// MediaCleanupRetry records cleanup attempts without blocking the upload path.
type MediaCleanupRetry struct {
	ObjectKey string
	Attempts  int
	LastError string
	UpdatedAt time.Time
}

var failedMediaCleanupRetries []MediaCleanupRetry

func TrackMediaCleanupFailure(key string, err error) {
	failedMediaCleanupRetries = append(failedMediaCleanupRetries, MediaCleanupRetry{
		ObjectKey: key,
		Attempts: 1,
		LastError: observability.HashError(err),
		UpdatedAt: time.Now(),
	})
	log.Printf("media cleanup queued key_hash=%s error_hash=%s", observability.HashSensitive(key), observability.HashError(err))
}

func RetryFailedMediaCleanup(ctx context.Context, storage MediaStorage) {
	for _, item := range failedMediaCleanupRetries {
		if err := storage.Delete(ctx, item.ObjectKey); err != nil {
			log.Printf("media cleanup retry failed key_hash=%s error_hash=%s", observability.HashSensitive(item.ObjectKey), observability.HashError(err))
			continue
		}
		log.Printf("media cleanup retry completed key_hash=%s", observability.HashSensitive(item.ObjectKey))
	}
}
