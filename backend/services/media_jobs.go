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

// ClaimPendingMediaJob atomically claims one pending job for processing.
// The row lock prevents multiple workers from claiming the same job.
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
				"attempts": job.Attempts,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
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
// by the worker. This prevents a stale worker from overwriting a newer retry.
func FailMediaJob(db *gorm.DB, jobID uint, err error, expectedLockedAt *time.Time) error {
	if err == nil {
		err = errors.New("unknown media processing failure")
	}
	if expectedLockedAt == nil {
		return errors.New("media job lease is missing")
	}

	var job models.MediaJob
	if findErr := db.Where("id = ? AND status = ? AND locked_at = ?", jobID, "processing", *expectedLockedAt).First(&job).Error; findErr != nil {
		return findErr
	}

	status := "pending"
	if job.Attempts >= MaxMediaAttempts {
		status = "failed"
	}

	result := db.Model(&models.MediaJob{}).
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
	return nil
}
