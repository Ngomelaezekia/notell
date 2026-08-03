package models

import "time"

type Relationship struct {
	FollowerID  uint      `gorm:"primaryKey" json:"followerId"`
	FollowingID uint      `gorm:"primaryKey" json:"followingId"`
	Status      string    `gorm:"type:varchar(20);default:'accepted';index" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`

	// Associations
	Follower  User `gorm:"foreignKey:FollowerID;references:ID;constraint:OnDelete:CASCADE" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID;references:ID;constraint:OnDelete:CASCADE" json:"following,omitempty"`
}

func (Relationship) TableName() string {
	return "user_relationships"
}
