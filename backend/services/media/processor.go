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

// Process performs the metadata-safe portion of media preparation. The upload is
// stored in external media storage before the worker runs, so processing must not
// depend on upload.Path still existing on the local filesystem.
//
// Tool-specific work such as video probing, dimensions, duration and thumbnails
// can be added later using the storage layer as the source of the media bytes.
func (p *Processor) Process(ctx context.Context, upload *models.Upload) error {
	if upload == nil {
		return fmt.Errorf("upload is required")
	}

	var metadata models.MediaMetadata
	if err := p.DB.WithContext(ctx).
		Where("upload_id = ?", upload.ID).
		First(&metadata).Error; err != nil {
		return fmt.Errorf("load media metadata: %w", err)
	}

	if err := ValidateUploadMetadata(upload.MediaType, metadata.FileSize, upload.Filename); err != nil {
		return err
	}

	if err := p.MarkProcessing(ctx, upload.ID); err != nil {
		return err
	}

	// The current processor only guarantees metadata already recorded at upload
	// time. It deliberately does not stat upload.Path because B2 uploads remove
	// the temporary local file before the asynchronous worker executes.
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
		Updates(map[string]any{
			"status":           "failed",
			"processing_error": message,
		}).Error
}

// Upload does not currently carry a file-size field. Keep this helper for the
// backward-compatible CreateMetadata path; the asynchronous Process path uses
// the file size persisted in MediaMetadata during upload.
func uploadFileSize(upload *models.Upload) int64 {
	metadata, err := ExtractMetadata(upload.Path, upload.MediaType)
	if err != nil {
		return 0
	}
	return metadata.FileSize
}
