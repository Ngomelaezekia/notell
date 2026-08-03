package models

import "time"

type Post struct {
	ID          uint      `gorm:"primaryKey" json:"postId"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	ContentType string    `gorm:"not null" json:"contentType"` // e.g., "image" or "video"
	ContentURL  string    `gorm:"not null" json:"contentUrl"`
	Caption     string    `gorm:"type:text" json:"caption,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Associations
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}
