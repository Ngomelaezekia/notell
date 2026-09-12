package models

import "time"

// PostMusic stores the selected audio track and playback window for a post.
// The audio itself is an Upload object in the same protected media system.
type PostMusic struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `gorm:"uniqueIndex;not null" json:"postId"`
	UploadID   uint      `gorm:"not null;index" json:"uploadId"`
	StartSec   float64   `gorm:"not null;default:0" json:"startSec"`
	EndSec     float64   `gorm:"not null;default:0" json:"endSec"`
	Volume     float64   `gorm:"not null;default:1" json:"volume"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"upload,omitempty"`
}
