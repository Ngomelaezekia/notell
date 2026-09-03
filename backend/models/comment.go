package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"commentId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	PostID    uint      `gorm:"not null;index:idx_comments_post_parent_created,priority:1" json:"postId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *uint     `gorm:"index:idx_comments_post_parent_created,priority:2" json:"parentId,omitempty"`
	CreatedAt time.Time `gorm:"index:idx_comments_post_parent_created,priority:3" json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post    Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// AfterCreate emits a reply notification for nested comments. Top-level
// comments keep using PostHandler's existing comment notification producer,
// so this hook only handles the previously missing reply case.
func (cmt *Comment) AfterCreate(tx *gorm.DB) error {
	if cmt.ParentID == nil {
		return nil
	}

	var parent Comment
	if err := tx.Select("id, user_id, post_id").First(&parent, *cmt.ParentID).Error; err != nil {
		return nil
	}
	if parent.PostID != cmt.PostID || parent.UserID == 0 || parent.UserID == cmt.UserID {
		return nil
	}

	postID := cmt.PostID
	commentID := cmt.ID
	_ = tx.Create(&Notification{
		UserID:    parent.UserID,
		ActorID:   cmt.UserID,
		Type:      "reply",
		PostID:    &postID,
		CommentID: &commentID,
	}).Error
	return nil
}
