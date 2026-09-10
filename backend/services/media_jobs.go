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
func ClaimPendingMediaJob(db *gorm.DB) (*models.MediaJob, error) {
	var job models.MediaJob

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("status = ? AND attempts < ?", "pending", MaxMediaAttempts).
			Order("created_at ASC").First(&job).Error; err != nil {
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
