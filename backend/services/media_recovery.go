package services

import (
	"context"
	"log"
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
		LastError: err.Error(),
		UpdatedAt: time.Now(),
	})
	log.Printf("media cleanup queued key=%s error=%v", key, err)
}

func RetryFailedMediaCleanup(ctx context.Context, storage MediaStorage) {
	for _, item := range failedMediaCleanupRetries {
		if err := storage.Delete(ctx, item.ObjectKey); err != nil {
			log.Printf("media cleanup retry failed key=%s error=%v", item.ObjectKey, err)
			continue
		}
		log.Printf("media cleanup retry completed key=%s", item.ObjectKey)
	}
}
