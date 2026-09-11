package services

import (
	"context"
	"errors"
	"log"
	"time"

	"notell/models"
	"notell/observability"
	mediaservice "notell/services/media"

	"gorm.io/gorm"
)

const mediaWorkerInterval = 5 * time.Second
const mediaJobStaleAfter = 10 * time.Minute

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
				recoverStaleMediaJobs(db)
				processNextMediaJob(ctx, db, processor)
			}
		}
	}()
}

// recoverStaleMediaJobs makes the worker self-healing after a process crash or
// instance restart. If processing already produced ready metadata, the job is
// safely completed; otherwise an interrupted attempt is returned to the queue.
func recoverStaleMediaJobs(db *gorm.DB) {
	cutoff := time.Now().Add(-mediaJobStaleAfter)

	var jobs []models.MediaJob
	if err := db.Where("status = ? AND locked_at IS NOT NULL AND locked_at < ?", "processing", cutoff).Limit(20).Find(&jobs).Error; err != nil {
		log.Printf("media worker stale-job recovery lookup failed: %v", err)
		return
	}

	for _, job := range jobs {
		var metadata models.MediaMetadata
		metadataErr := db.Where("upload_id = ?", job.UploadID).First(&metadata).Error

		if metadataErr == nil && metadata.Status == "ready" {
			if err := CompleteMediaJob(db, job.ID); err != nil {
				log.Printf("media worker stale-job completion failed job=%d: %v", job.ID, err)
			}
			continue
		}

		if job.Attempts >= MaxMediaAttempts {
			message := "media processing interrupted after maximum attempts"
			if metadataErr != nil && !errors.Is(metadataErr, gorm.ErrRecordNotFound) {
				message = metadataErr.Error()
			}
			if err := db.Model(&models.MediaJob{}).Where("id = ? AND status = ?", job.ID, "processing").Updates(map[string]any{
				"status":    "failed",
				"locked_at": nil,
				"error":     message,
			}).Error; err != nil {
				log.Printf("media worker stale-job failure update failed job=%d: %v", job.ID, err)
				continue
			}
			if metadataErr == nil {
				_ = db.Model(&models.MediaMetadata{}).Where("upload_id = ?", job.UploadID).Updates(map[string]any{
					"status":           "failed",
					"processing_error": message,
				}).Error
			}
			continue
		}

		if err := db.Model(&models.MediaJob{}).Where("id = ? AND status = ?", job.ID, "processing").Updates(map[string]any{
			"status":    "pending",
			"locked_at": nil,
		}).Error; err != nil {
			log.Printf("media worker stale-job requeue failed job=%d: %v", job.ID, err)
		}
	}
}

func processNextMediaJob(ctx context.Context, db *gorm.DB, processor *mediaservice.Processor) {
	job, err := ClaimPendingMediaJob(db)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("media worker claim failed: %v", err)
		}
		return
	}

	observability.MediaJobClaimed(job.ID)
	observability.MediaProcessingStarted(job.ID, job.UploadID)

	log.Printf("media worker processing job=%d upload=%d attempt started", job.ID, job.UploadID)

	var upload models.Upload
	if err := db.WithContext(ctx).First(&upload, job.UploadID).Error; err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker upload lookup failed job=%d: %v", job.ID, err)
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := processor.Process(ctx, &upload); err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker processing failed job=%d: %v", job.ID, err)
		_ = FailMediaJob(db, job.ID, err)
		return
	}

	if err := CompleteMediaJob(db, job.ID); err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker completion failed job=%d: %v", job.ID, err)
		return
	}

	observability.MediaProcessingCompleted(job.ID)
	log.Printf("media worker completed job=%d", job.ID)
}
