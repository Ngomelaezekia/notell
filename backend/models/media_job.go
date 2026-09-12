package models

import "time"

// MediaJob tracks asynchronous media processing tasks.
type MediaJob struct {
	ID uint `gorm:"primaryKey"`

	UploadID uint `gorm:"not null;index"`

	Status string

	Attempts int

	LockedAt *time.Time

	Error string

	CreatedAt time.Time
	UpdatedAt time.Time

	Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"-"`
}
