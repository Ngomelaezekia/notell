package models

import "time"

// Error and lease state are internal-only and never serialized.
type MediaJob struct {
    ID uint `gorm:"primaryKey" json:"id"`
    UploadID uint `gorm:"not null;index" json:"uploadId"`
    Status string `json:"status"`
    Attempts int `json:"attempts"`
    LockedAt *time.Time `json:"-"`
    Error string `json:"-"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Upload Upload `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"-"`
}
