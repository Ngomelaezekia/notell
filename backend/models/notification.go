package models

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"notificationId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	ActorID   uint      `gorm:"not null;index" json:"actorId"`
	Type      string    `gorm:"type:varchar(30);not null;index" json:"type"`
	PostID    *uint     `gorm:"index" json:"postId,omitempty"`
	CommentID *uint     `gorm:"index" json:"commentId,omitempty"`
	Read      bool      `gorm:"not null;default:false;index" json:"read"`
	CreatedAt time.Time `json:"createdAt"`

	Actor User `gorm:"foreignKey:ActorID;constraint:OnDelete:CASCADE" json:"actor,omitempty"`
}
