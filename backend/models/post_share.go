package models

import "time"

type PostShare struct {
	PostID    uint      `gorm:"primaryKey;uniqueIndex:idx_post_shares_post_user"` json:"postId"
	UserID    uint      `gorm:"primaryKey;uniqueIndex:idx_post_shares_post_user;index"` json:"userId"
	CreatedAt time.Time `json:"createdAt"`

	Post Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"-"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PostShare) TableName() string { return "post_shares" }
