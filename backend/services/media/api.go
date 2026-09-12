package media

import (
	"errors"

	"notell/models"

	"gorm.io/gorm"
)

var ErrMediaNotFound = errors.New("media not found")

// Metadata returns the processing record only when the caller owns the upload.
// Public playback remains governed by Authorize; metadata is intentionally
// private so internal processing details are not exposed to other users.
func Metadata(db *gorm.DB, filename string, userID uint) (models.MediaMetadata, error) {
	if db == nil || filename == "" || userID == 0 {
		return models.MediaMetadata{}, ErrMediaNotFound
	}
	var metadata models.MediaMetadata
	err := db.Joins("JOIN uploads ON uploads.id = media_metadata.upload_id").
		Where("uploads.filename = ? AND uploads.user_id = ?", filename, userID).
		First(&metadata).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.MediaMetadata{}, ErrMediaNotFound
	}
	if err != nil {
		return models.MediaMetadata{}, err
	}
	return metadata, nil
}
