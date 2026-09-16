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
// instance restart. Each stale transition updates the job and metadata in one
// transaction so their lifecycle states cannot diverge.
func recoverStaleMediaJobs(db *gorm.DB) {
	cutoff := time.Now().Add(-mediaJobStaleAfter)

	var jobs []models.MediaJob
	if err := db.Where("status = ? AND locked_at IS NOT NULL AND locked_at < ?", "processing", cutoff).Limit(20).Find(&jobs).Error; err != nil {
		log.Printf("media worker stale-job recovery lookup failed: %v", err)
		return
	}

	for _, job := range jobs {
		if job.LockedAt == nil {
			continue
		}
		lease := *job.LockedAt

		if err := recoverStaleMediaJob(db, job, lease); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("media worker stale-job recovery failed job=%d: %v", job.ID, err)
		}
	}
}

func recoverStaleMediaJob(db *gorm.DB, job models.MediaJob, lease time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var current models.MediaJob
		if err := tx.Where("id = ? AND status = ? AND locked_at = ?", job.ID, "processing", lease).First(&current).Error; err != nil {
			return err
		}

		var metadata models.MediaMetadata
		metadataErr := tx.Where("upload_id = ?", current.UploadID).First(&metadata).Error

		if metadataErr == nil && metadata.Status == models.MediaStatusReady {
			result := tx.Model(&models.MediaJob{}).
				Where("id = ? AND status = ? AND locked_at = ?", current.ID, "processing", lease).
				Updates(map[string]any{"status": "completed", "locked_at": nil, "error": ""})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			return nil
		}

		if metadataErr != nil && !errors.Is(metadataErr, gorm.ErrRecordNotFound) {
			return metadataErr
		}

		if current.Attempts >= MaxMediaAttempts {
			message := "media processing interrupted after maximum attempts"
			if metadataErr != nil {
				message = metadataErr.Error()
			}
			result := tx.Model(&models.MediaJob{}).
				Where("id = ? AND status = ? AND locked_at = ?", current.ID, "processing", lease).
				Updates(map[string]any{"status": "failed", "locked_at": nil, "error": message})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			if metadataErr == nil {
				if err := tx.Model(&models.MediaMetadata{}).Where("upload_id = ?", current.UploadID).Updates(map[string]any{
					"status":           models.MediaStatusFailed,
					"processing_error": message,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		}

		result := tx.Model(&models.MediaJob{}).
			Where("id = ? AND status = ? AND locked_at = ?", current.ID, "processing", lease).
			Updates(map[string]any{"status": "pending", "locked_at": nil})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if metadataErr == nil {
			if err := tx.Model(&models.MediaMetadata{}).Where("upload_id = ?", current.UploadID).Updates(map[string]any{
				"status": models.MediaStatusProcessing,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func processNextMediaJob(ctx context.Context, db *gorm.DB, processor *mediaservice.Processor) {
	job, err := ClaimPendingMediaJob(db)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("media worker claim failed: %v", err)
		}
		return
	}
	if job.LockedAt == nil {
		log.Printf("media worker claimed job without lease job=%d", job.ID)
		return
	}
	lease := *job.LockedAt

	observability.MediaJobClaimed(job.ID)
	observability.MediaProcessingStarted(job.ID, job.UploadID)

	log.Printf("media worker processing job=%d upload=%d attempt started", job.ID, job.UploadID)

	var upload models.Upload
	if err := db.WithContext(ctx).First(&upload, job.UploadID).Error; err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker upload lookup failed job=%d: %v", job.ID, err)
		if failErr := FailMediaJob(db, job.ID, err, &lease); failErr != nil {
			log.Printf("media worker job failure update failed job=%d: %v", job.ID, failErr)
		}
		return
	}

	if err := processor.Process(ctx, &upload); err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker processing failed job=%d: %v", job.ID, err)
		if failErr := FailMediaJob(db, job.ID, err, &lease); failErr != nil {
			log.Printf("media worker job failure update failed job=%d: %v", job.ID, failErr)
		}
		return
	}

	if err := CompleteMediaJob(db, job.ID, &lease); err != nil {
		observability.MediaProcessingFailed(job.ID, err)
		log.Printf("media worker completion failed job=%d: %v", job.ID, err)
		return
	}

	observability.MediaProcessingCompleted(job.ID)
	log.Printf("media worker completed job=%d", job.ID)
}
