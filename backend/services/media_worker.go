package services

import (
	"context"
	"errors"
	"log"
	"time"

	"notell/models"
	mediaservice "notell/services/media"

	"gorm.io/gorm"
)

const mediaWorkerInterval = 5 * time.Second

// StartMediaWorker starts the background media processor and stops cleanly when
// the parent context is cancelled.
func StartMediaWorker(ctx context.Context, db *gorm.DB) {
	processor := mediaservice.NewProcessor(db)

	go func() {
		log.Printf("media worker started")

		ticker := time.NewTicker(mediaWorkerInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("media worker stopped")
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

	log.Printf("media worker processing job=%d upload=%d attempt started", job.ID, job.UploadID)

	var upload models.Upload
	if err := db.WithContext(ctx).First(&upload, job.UploadID).Error; err != nil {
		log.Printf("media worker upload lookup failed job=%d: %v", job.ID, err)
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := processor.Process(ctx, &upload); err != nil {
		log.Printf("media worker processing failed job=%d: %v", job.ID, err)
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := CompleteMediaJob(db, job.ID); err != nil {
		log.Printf("media worker completion failed job=%d: %v", job.ID, err)
		return
	}

	log.Printf("media worker completed job=%d", job.ID)
}
