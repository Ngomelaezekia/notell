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

		return tx.Model(&models.MediaJob{}).
			Where("id = ? AND status = ?", job.ID, "pending").
			Updates(map[string]any{
				"status":    job.Status,
				"locked_at": job.LockedAt,
				"attempts": job.Attempts,
			}).Error
	})

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func CompleteMediaJob(db *gorm.DB, jobID uint) error {
	return db.Model(&models.MediaJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":    "completed",
			"locked_at": nil,
			"error":     "",
		}).Error
}

func FailMediaJob(db *gorm.DB, jobID uint, err error) error {
	if err == nil {
		err = errors.New("unknown media processing failure")
	}

	var job models.MediaJob
	if findErr := db.First(&job, jobID).Error; findErr != nil {
		return findErr
	}

	status := "pending"
	if job.Attempts >= MaxMediaAttempts {
		status = "failed"
	}

	return db.Model(&models.MediaJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":    status,
			"locked_at": nil,
			"error":     err.Error(),
		}).Error
}
