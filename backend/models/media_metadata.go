package models

import "time"

const (
	MediaStatusPending    = "pending"
	MediaStatusUploading  = "uploading"
	MediaStatusProcessing = "processing"
	MediaStatusReady      = "ready"
	MediaStatusFailed     = "failed"
	MediaStatusDeleted    = "deleted"
)

// MediaMetadata stores processing and playback information for uploaded media.
type MediaMetadata struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	UploadID uint `gorm:"uniqueIndex;not null" json:"uploadId"`

	ThumbnailURL string  `json:"thumbnailUrl,omitempty"`
	Duration     float64 `json:"duration,omitempty"`
	Width        int     `json:"width,omitempty"`
	Height       int     `json:"height,omitempty"`
	FileSize     int64   `json:"fileSize,omitempty"`
	Codec        string  `json:"codec,omitempty"`

	Status          string `gorm:"size:32;not null;default:'pending'" json:"status"`
	ProcessingError string `gorm:"type:text" json:"processingError,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"-"`
}
