package services

import (
	"errors"
	"time"

	"notell/models"

	"gorm.io/gorm"
)

const MaxMediaAttempts = 3

// CreateMediaJob creates a pending asynchronous media processing task.
func CreateMediaJob(db *gorm.DB, uploadID uint) error {
	job := models.MediaJob{
		UploadID: uploadID,
		Status:   "pending",
		Attempts: 0,
	}

	return db.Create(&job).Error
}

// ClaimPendingMediaJob atomically claims one pending job for processing and
// moves its metadata to processing in the same transaction.
func ClaimPendingMediaJob(db *gorm.DB) (*models.MediaJob, error) {
	var job models.MediaJob

	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx.Raw(`
			SELECT *
			FROM media_jobs
			WHERE status = ?
			AND attempts < ?
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`, "pending", MaxMediaAttempts)

		if err := query.Scan(&job).Error; err != nil {
			return err
		}

		if job.ID == 0 {
			return gorm.ErrRecordNotFound
		}

		now := time.Now()
		job.Status = "processing"
		job.LockedAt = &now
		job.Attempts++

		result := tx.Model(&models.MediaJob{}).
			Where("id = ? AND status = ?", job.ID, "pending").
			Updates(map[string]any{
				"status":    job.Status,
				"locked_at": job.LockedAt,
				"attempts":  job.Attempts,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		metadataResult := tx.Model(&models.MediaMetadata{}).
			Where("upload_id = ? AND status IN ?", job.UploadID, []string{"uploaded", models.MediaStatusPending}).
			Updates(map[string]any{
				"status":           models.MediaStatusProcessing,
				"processing_error": "",
			})
		if metadataResult.Error != nil {
			return metadataResult.Error
		}
		if metadataResult.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// CompleteMediaJob completes only the specific processing lease supplied by
// the worker. This prevents a stale worker from completing a newer retry.
func CompleteMediaJob(db *gorm.DB, jobID uint, expectedLockedAt *time.Time) error {
	if expectedLockedAt == nil {
		return errors.New("media job lease is missing")
	}

	result := db.Model(&models.MediaJob{}).
		Where("id = ? AND status = ? AND locked_at = ?", jobID, "processing", *expectedLockedAt).
		Updates(map[string]any{
			"status":    "completed",
			"locked_at": nil,
			"error":     "",
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FailMediaJob fails or requeues only the specific processing lease supplied
// by the worker. The job and metadata transitions are committed atomically.
func FailMediaJob(db *gorm.DB, jobID uint, err error, expectedLockedAt *time.Time) error {
	if err == nil {
		err = errors.New("unknown media processing failure")
	}
	if expectedLockedAt == nil {
		return errors.New("media job lease is missing")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var job models.MediaJob
		if findErr := tx.Where("id = ? AND status = ? AND locked_at = ?", jobID, "processing", *expectedLockedAt).First(&job).Error; findErr != nil {
			return findErr
		}

		status := "pending"
		if job.Attempts >= MaxMediaAttempts {
			status = "failed"
		}

		result := tx.Model(&models.MediaJob{}).
			Where("id = ? AND status = ? AND locked_at = ?", jobID, "processing", *expectedLockedAt).
			Updates(map[string]any{
				"status":    status,
				"locked_at": nil,
				"error":     err.Error(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		metadataStatus := "processing"
		if status == "failed" {
			metadataStatus = models.MediaStatusFailed
		}
		if updateErr := tx.Model(&models.MediaMetadata{}).
			Where("upload_id = ?", job.UploadID).
			Updates(map[string]any{
				"status":           metadataStatus,
				"processing_error": err.Error(),
			}).Error; updateErr != nil {
			return updateErr
		}

		return nil
	})
}
