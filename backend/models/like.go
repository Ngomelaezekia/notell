package models

import (
	"time"

	"gorm.io/gorm"
)

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"likeId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_post_like" json:"userId"`
	PostID    uint      `gorm:"not null;index:idx_likes_post;uniqueIndex:idx_user_post_like" json:"postId"`
	CreatedAt time.Time `json:"createdAt"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
}

// AfterCreate emits a notification when a like is successfully persisted.
// Notification failures are intentionally best-effort and must not roll back
// the like itself.
func (l *Like) AfterCreate(tx *gorm.DB) error {
	var post Post
	if err := tx.Select("id, user_id").First(&post, l.PostID).Error; err != nil {
		return nil
	}
	if post.UserID == 0 || post.UserID == l.UserID {
		return nil
	}

	_ = tx.Create(&Notification{
		UserID:  post.UserID,
		ActorID: l.UserID,
		Type:    "like",
		PostID:  &l.PostID,
	}).Error
	return nil
}
