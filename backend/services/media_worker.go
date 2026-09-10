package services

import (
	"context"
	"errors"
	"log"
	"time"

	mediaservice "notell/services/media"
	"notell/models"

	"gorm.io/gorm"
)

// StartMediaWorker starts the asynchronous media processing loop.
// It consumes MediaJob records and delegates actual processing to the media processor.
func StartMediaWorker(ctx context.Context, db *gorm.DB) {
	processor := mediaservice.NewProcessor(db)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processNextMediaJob(ctx, db, processor)
			}
		}
	}()
}

func processNextMediaJob(ctx context.Context, db *gorm.DB, processor *mediaservice.Processor) {
	job, err := ClaimPendingMediaJob(db)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("media worker claim failed: %v", err)
		}
		return
	}

	var upload models.Upload
	if err := db.WithContext(ctx).First(&upload, job.UploadID).Error; err != nil {
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := processor.Process(ctx, &upload); err != nil {
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := CompleteMediaJob(db, job.ID); err != nil {
		log.Printf("media worker completion failed: %v", err)
	}
}
