package media

import (
	"context"
	"fmt"

	"notell/models"

	"gorm.io/gorm"
)

// Processor manages media preparation without coupling uploads to processing.
// Tool-specific work such as video probing and thumbnail generation can be added behind
// the processor boundaries without changing the upload API.
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
		FileSize: uploadFileSize(upload),
		Status:   "uploaded",
	}

	return p.DB.WithContext(ctx).Create(&metadata).Error
}

// Process performs the metadata-safe portion of media preparation. It validates the
// existing upload contract and records file information before a future transcoder
// handles dimensions, duration and thumbnails.
func (p *Processor) Process(ctx context.Context, upload *models.Upload) error {
	if upload == nil {
		return fmt.Errorf("upload is required")
	}
	if err := ValidateUploadMetadata(upload.MediaType, uploadFileSize(upload), upload.Filename); err != nil {
		return err
	}

	if err := p.MarkProcessing(ctx, upload.ID); err != nil {
		return err
	}

	metadata, err := ExtractMetadata(upload.Path, upload.MediaType)
	if err != nil {
		_ = p.MarkFailed(ctx, upload.ID, err.Error())
		return err
	}

	updates := map[string]any{
		"file_size": metadata.FileSize,
		"status":    "ready",
	}
	if err := p.DB.WithContext(ctx).
		Model(&models.MediaMetadata{}).
		Where("upload_id = ?", upload.ID).
		Updates(updates).Error; err != nil {
		_ = p.MarkFailed(ctx, upload.ID, err.Error())
		return err
	}
	return nil
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

func (p *Processor) MarkFailed(ctx context.Context, uploadID uint, message string) error {
	return p.DB.WithContext(ctx).
		Model(&models.MediaMetadata{}).
		Where("upload_id = ?", uploadID).
		Updates(map[string]any{
			"status":           "failed",
			"processing_error": message,
		}).Error
}

// Upload does not currently carry a file-size field, so the processor obtains it
// from the stored path when available. CreateMetadata remains backward compatible.
func uploadFileSize(upload *models.Upload) int64 {
	metadata, err := ExtractMetadata(upload.Path, upload.MediaType)
	if err != nil {
		return 0
	}
	return metadata.FileSize
}
