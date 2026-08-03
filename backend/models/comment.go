package models

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"commentId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	PostID    uint      `gorm:"not null;index" json:"postId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *uint     `gorm:"index" json:"parentId,omitempty"` // For nested replies
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Associations
	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post    Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}
