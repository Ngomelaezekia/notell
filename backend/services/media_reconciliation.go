package services

import (
	"log"

	"notell/models"

	"gorm.io/gorm"
)

// ReconcileMediaState repairs database-side media processing state when an
// upload has no corresponding job. Physical object cleanup is handled by the
// storage reconciler.
func ReconcileMediaState(db *gorm.DB) {
	var metadata []models.MediaMetadata
	activeStatuses := []string{
		models.MediaStatusPending,
		"processing",
		"uploaded",
	}
	if err := db.Where("status IN ?", activeStatuses).Find(&metadata).Error; err != nil {
		log.Printf("media reconciliation lookup failed: %v", err)
		return
	}

	for _, item := range metadata {
		var job models.MediaJob
		err := db.Where("upload_id = ?", item.UploadID).First(&job).Error
		if err == nil {
			if job.Status == "failed" {
				updates := map[string]any{
					"status": models.MediaStatusFailed,
				}
				if job.Error != "" {
					updates["processing_error"] = job.Error
				}
				if updateErr := db.Model(&models.MediaMetadata{}).
					Where("upload_id = ?", item.UploadID).
					Updates(updates).Error; updateErr != nil {
					log.Printf("media reconciliation failed-job update failed upload=%d: %v", item.UploadID, updateErr)
				}
			}
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Printf("media reconciliation job lookup failed upload=%d: %v", item.UploadID, err)
			continue
		}

		if err := db.Model(&models.MediaMetadata{}).
			Where("upload_id = ? AND status IN ?", item.UploadID, activeStatuses).
			Updates(map[string]any{
				"status":           models.MediaStatusFailed,
				"processing_error": "media processing job is missing",
			}).Error; err != nil {
			log.Printf("media reconciliation failure update failed upload=%d: %v", item.UploadID, err)
		}
	}
}
