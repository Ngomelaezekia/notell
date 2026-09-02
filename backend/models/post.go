package models

import "time"

type Post struct {
	ID          uint      `gorm:"primaryKey" json:"postId"`
	UserID      uint      `gorm:"not null;index:idx_posts_user_created" json:"userId"`
	ContentType string    `gorm:"not null" json:"contentType"`
	ContentURL  string    `gorm:"not null" json:"contentUrl"`
	Caption     string    `gorm:"type:text" json:"caption,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	LikeCount int  `gorm:"-" json:"likeCount"`
	Liked     bool `gorm:"-" json:"liked"`

	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}
