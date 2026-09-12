package media

import (
	"context"
	"fmt"

	"notell/models"

	"gorm.io/gorm"
)

type Processor struct { DB *gorm.DB }

func NewProcessor(db *gorm.DB) *Processor { return &Processor{DB: db} }

func (p *Processor) CreateMetadata(ctx context.Context, upload *models.Upload) error {
	if upload == nil { return fmt.Errorf("upload is required") }
	metadata := models.MediaMetadata{UploadID: upload.ID, FileSize: uploadFileSize(upload), Status: StatusUploaded}
	return p.DB.WithContext(ctx).Create(&metadata).Error
}

// Process owns the metadata lifecycle. The storage object has already been
// durably committed before this asynchronous worker runs.
func (p *Processor) Process(ctx context.Context, upload *models.Upload) error {
	if upload == nil { return fmt.Errorf("upload is required") }
	var metadata models.MediaMetadata
	if err := p.DB.WithContext(ctx).Where("upload_id = ?", upload.ID).First(&metadata).Error; err != nil {
		return fmt.Errorf("load media metadata: %w", err)
	}
	fail := func(err error) error {
		if err == nil { return nil }
		if markErr := p.MarkFailed(ctx, upload.ID, err.Error()); markErr != nil {
			return fmt.Errorf("%w (also failed to record processing error: %v)", err, markErr)
		}
		return err
	}
	if err := ValidateUploadMetadata(upload.MediaType, metadata.FileSize, upload.Filename); err != nil { return fail(err) }
	if err := p.MarkProcessing(ctx, upload.ID); err != nil { return err }

	// Media probing/transcoding is intentionally behind this processor boundary.
	// The current MVP validates and prepares the durable object; ffmpeg/probing
	// can be introduced here without changing upload or post APIs.
	if err := p.DB.WithContext(ctx).Model(&models.MediaMetadata{}).
		Where("upload_id = ? AND status = ?", upload.ID, StatusProcessing).
		Updates(map[string]any{"file_size": metadata.FileSize, "status": StatusReady, "processing_error": ""}).Error; err != nil {
		return fail(err)
	}
	return nil
}

func (p *Processor) MarkProcessing(ctx context.Context, uploadID uint) error {
	return p.transitionMetadata(ctx, uploadID, StatusProcessing, map[string]any{"processing_error": ""})
}

func (p *Processor) MarkReady(ctx context.Context, uploadID uint) error {
	return p.transitionMetadata(ctx, uploadID, StatusReady, nil)
}

func (p *Processor) MarkFailed(ctx context.Context, uploadID uint, message string) error {
	return p.transitionMetadata(ctx, uploadID, StatusFailed, map[string]any{"processing_error": message})
}

func (p *Processor) transitionMetadata(ctx context.Context, uploadID uint, next string, updates map[string]any) error {
	var metadata models.MediaMetadata
	if err := p.DB.WithContext(ctx).Select("id, status").Where("upload_id = ?", uploadID).First(&metadata).Error; err != nil { return err }
	if err := Transition(metadata.Status, next); err != nil { return err }
	if updates == nil { updates = map[string]any{} }
	updates["status"] = next
	result := p.DB.WithContext(ctx).Model(&models.MediaMetadata{}).Where("id = ? AND status = ?", metadata.ID, metadata.Status).Updates(updates)
	if result.Error != nil { return result.Error }
	if result.RowsAffected != 1 { return gorm.ErrRecordNotFound }
	return nil
}

func uploadFileSize(upload *models.Upload) int64 {
	metadata, err := ExtractMetadata(upload.Path, upload.MediaType)
	if err != nil { return 0 }
	return metadata.FileSize
}
