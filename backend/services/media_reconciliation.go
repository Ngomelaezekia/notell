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
	if err := db.Where("status IN ?", []string{"uploaded", "pending", "processing"}).Find(&metadata).Error; err != nil {
		log.Printf("media reconciliation lookup failed: %v", err)
		return
	}

	for _, item := range metadata {
		var job models.MediaJob
		err := db.Where("upload_id = ?", item.UploadID).First(&job).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Printf("media reconciliation job lookup failed upload=%d: %v", item.UploadID, err)
			continue
		}

		if err := db.Model(&models.MediaMetadata{}).
			Where("upload_id = ? AND status IN ?", item.UploadID, []string{"uploaded", "pending", "processing"}).
			Updates(map[string]any{
				"status":           "failed",
				"processing_error": "media processing job is missing",
			}).Error; err != nil {
			log.Printf("media reconciliation failure update failed upload=%d: %v", item.UploadID, err)
		}
	}
}
