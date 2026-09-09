package media

import (
	"context"
	"fmt"

	"notell/models"

	"gorm.io/gorm"
)

// Processor manages media preparation without coupling uploads to processing.
// It can later be moved behind a worker queue.
type Processor struct {
	DB *gorm.DB
}

func NewProcessor(db *gorm.DB) *Processor {
	return &Processor{DB: db}
}

func (p *Processor) CreateMetadata(ctx context.Context, upload *models.Upload) error {
	if upload == nil {
		return fmt.Errorf("upload is required")
	}

	metadata := models.MediaMetadata{
		UploadID: upload.ID,
		Status:   "uploaded",
	}

	return p.DB.WithContext(ctx).Create(&metadata).Error
}

func (p *Processor) MarkProcessing(ctx context.Context, uploadID uint) error {
	return p.DB.WithContext(ctx).
		Model(&models.MediaMetadata{}).
		Where("upload_id = ?", uploadID).
		Update("status", "processing").Error
}

func (p *Processor) MarkReady(ctx context.Context, uploadID uint) error {
	return p.DB.WithContext(ctx).
		Model(&models.MediaMetadata{}).
		Where("upload_id = ?", uploadID).
		Update("status", "ready").Error
}
