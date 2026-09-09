package models

import "time"

// PostView records a unique authenticated viewer impression for a post.
// The unique user/post pair prevents repeated feed visibility from inflating views.
type PostView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;uniqueIndex:idx_post_views_user_post;index:idx_post_views_post" json:"postId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_post_views_user_post;index" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
