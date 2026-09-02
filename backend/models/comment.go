package models

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"commentId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	PostID    uint      `gorm:"not null;index:idx_comments_post_parent_created" json:"postId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *uint     `gorm:"index:idx_comments_post_parent_created" json:"parentId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post    Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}
