package services

import (
	"errors"
	"time"

	"notell/models"

	"gorm.io/gorm"
)

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
func ClaimPendingMediaJob(db *gorm.DB) (*models.MediaJob, error) {
	var job models.MediaJob

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("status = ?", "pending").Order("created_at ASC").First(&job).Error; err != nil {
			return err
		}

		now := time.Now()
		job.Status = "processing"
		job.LockedAt = &now
		job.Attempts++

		return tx.Save(&job).Error
	})

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// CompleteMediaJob marks a job as completed.
func CompleteMediaJob(db *gorm.DB, jobID uint) error {
	return db.Model(&models.MediaJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":    "completed",
			"locked_at": nil,
			"error":     "",
		}).Error
}

// FailMediaJob records a failed processing attempt.
func FailMediaJob(db *gorm.DB, jobID uint, err error) error {
	if err == nil {
		err = errors.New("unknown media processing failure")
	}

	return db.Model(&models.MediaJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":    "failed",
			"locked_at": nil,
			"error":     err.Error(),
		}).Error
}
