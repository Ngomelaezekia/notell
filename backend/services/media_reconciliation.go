package services

import (
	"log"

	"notell/models"

	"gorm.io/gorm"
)

// ReconcileMediaState removes database records that cannot complete their
// storage lifecycle. This is intentionally conservative: it only marks broken
// media as failed and leaves physical object cleanup to the storage layer.
func ReconcileMediaState(db *gorm.DB) {
	var uploads []models.Upload
	if err := db.Where("status IN ?", []string{"pending", "processing"}).Find(&uploads).Error; err != nil {
		log.Printf("media reconciliation lookup failed: %v", err)
		return
	}

	for _, upload := range uploads {
		var job models.MediaJob
		if err := db.Where("upload_id = ?", upload.ID).First(&job).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				_ = db.Model(&models.Upload{}).Where("id = ?", upload.ID).Update("status", "failed").Error
			}
		}
	}
}
