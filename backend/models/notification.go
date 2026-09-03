package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"notificationId"`
	UserID    uint      `gorm:"not null;index:idx_notifications_user_read_created,priority:1" json:"userId"`
	ActorID   uint      `gorm:"not null;index" json:"actorId"`
	Type      string    `gorm:"type:varchar(30);not null" json:"type"`
	PostID    *uint     `gorm:"index" json:"postId,omitempty"`
	CommentID *uint     `gorm:"index" json:"commentId,omitempty"`
	Read      bool      `gorm:"not null;default:false;index:idx_notifications_user_read_created,priority:2" json:"read"`
	CreatedAt time.Time `gorm:"index:idx_notifications_user_read_created,priority:3" json:"createdAt"`

	User    User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Actor   User     `gorm:"foreignKey:ActorID;constraint:OnDelete:CASCADE" json:"actor,omitempty"`
	Post    *Post    `gorm:"foreignKey:PostID;constraint:OnDelete:SET NULL" json:"-"`
	Comment *Comment `gorm:"foreignKey:CommentID;constraint:OnDelete:SET NULL" json:"-"`
}

// BeforeCreate upgrades the existing comment notification emitted by the
// comment handler into a reply notification when the comment targets another
// comment. This keeps reply notifications single-fire without changing the
// existing comment creation path.
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.Type != "comment" || n.CommentID == nil {
		return nil
	}

	var comment Comment
	if err := tx.Select("id, user_id, post_id, parent_id").First(&comment, *n.CommentID).Error; err != nil || comment.ParentID == nil {
		return nil
	}

	var parent Comment
	if err := tx.Select("id, user_id, post_id").First(&parent, *comment.ParentID).Error; err != nil {
		return nil
	}
	if parent.PostID != comment.PostID || parent.UserID == 0 || parent.UserID == n.ActorID {
		return gorm.ErrInvalidData
	}

	n.UserID = parent.UserID
	n.Type = "reply"
	n.PostID = &comment.PostID
	return nil
}
