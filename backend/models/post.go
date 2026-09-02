package models

import "time"

type Post struct {
	ID          uint      `gorm:"primaryKey" json:"postId"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	ContentType string    `gorm:"not null" json:"contentType"`
	ContentURL  string    `gorm:"not null" json:"contentUrl"`
	Caption     string    `gorm:"type:text" json:"caption,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Computed engagement fields. These are populated by feed/post queries
	// and are not persisted as database columns.
	LikeCount int  `gorm:"-" json:"likeCount"`
	Liked     bool `gorm:"-" json:"liked"`

	// Associations
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}
