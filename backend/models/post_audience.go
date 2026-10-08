package models

import "time"

// PostAudience grants a post with visibility "selected" to a specific user.
// The composite unique key prevents duplicate audience entries.
type PostAudience struct {
	PostID    uint      `gorm:"primaryKey;uniqueIndex:idx_post_audiences_post_user" json:"postId"`
	UserID    uint      `gorm:"primaryKey;uniqueIndex:idx_post_audiences_post_user;index" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`

	Post Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PostAudience) TableName() string { return "post_audiences" }
