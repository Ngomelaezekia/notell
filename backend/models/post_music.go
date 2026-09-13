package models

import "time"

// PostMusic is a separate first-class record connecting one post to one audio upload.
// The audio remains in the common Upload/media lifecycle and is referenced here by UploadID.
type PostMusic struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	PostID   uint    `gorm:"uniqueIndex;not null;index" json:"postId"`
	UploadID uint    `gorm:"not null;index" json:"uploadId"`
	StartSec float64 `gorm:"not null;default:0" json:"startSec"`
	EndSec   float64 `gorm:"not null;default:0" json:"endSec"`
	Volume   float64 `gorm:"not null;default:1" json:"volume"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Post   Post   `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
	Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"upload,omitempty"`
}
