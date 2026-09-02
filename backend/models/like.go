package models

import "time"

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"likeId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_post_like" json:"userId"`
	PostID    uint      `gorm:"not null;index:idx_likes_post;uniqueIndex:idx_user_post_like" json:"postId"`
	CreatedAt time.Time `json:"createdAt"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
}
