package models

import "time"

// PostMusic supports both user-uploaded audio and cleared tracks from the independent Notell Music service.
type PostMusic struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	PostID   uint    `gorm:"uniqueIndex;not null;index" json:"postId"`
	UploadID uint    `gorm:"index" json:"uploadId,omitempty"`
	Source   string  `gorm:"not null;default:local;size:16;index" json:"source"`
	TrackID  string  `gorm:"size:191;index" json:"trackId,omitempty"`
	Provider string  `gorm:"size:64" json:"provider,omitempty"`
	Title    string  `gorm:"size:255" json:"title,omitempty"`
	Artist   string  `gorm:"size:255" json:"artist,omitempty"`
	ArtworkURL string `gorm:"size:1000" json:"artworkUrl,omitempty"`
	PreviewURL string `gorm:"size:2000" json:"previewUrl,omitempty"`
	DurationSec float64 `gorm:"not null;default:0" json:"durationSec"`
	StartSec float64 `gorm:"not null;default:0" json:"startSec"`
	EndSec   float64 `gorm:"not null;default:0" json:"endSec"`
	Volume   float64 `gorm:"not null;default:1" json:"volume"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Post   Post   `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
	Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:SET NULL" json:"upload,omitempty"`
}
