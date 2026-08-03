package models

import "time"

type Channel struct {
	ID          uint      `gorm:"primaryKey" json:"channelId"`
	OwnerID     uint      `gorm:"not null;index" json:"ownerId"`
	Owner       User      `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	BannerImage string    `json:"bannerImage,omitempty"`
	IsPremium   bool      `gorm:"default:true" json:"isPremium"` // Premium channel flag
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Members []User `gorm:"many2many:channel_members;" json:"members,omitempty"`
}
