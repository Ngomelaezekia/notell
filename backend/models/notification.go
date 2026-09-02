package models

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"notificationId"`
	UserID    uint      `gorm:"not null;index:idx_notifications_user_read_created,priority:1" json:"userId"`
	ActorID   uint      `gorm:"not null;index" json:"actorId"`
	Type      string    `gorm:"type:varchar(30);not null" json:"type"`
	PostID    *uint     `gorm:"index" json:"postId,omitempty"`
	CommentID *uint     `gorm:"index" json:"commentId,omitempty"`
	Read      bool      `gorm:"not null;default:false;index:idx_notifications_user_read_created,priority:2" json:"read"`
	CreatedAt time.Time `json:"createdAt"`

	Actor User `gorm:"foreignKey:ActorID;constraint:OnDelete:CASCADE" json:"actor,omitempty"`
}
