package models

import "time"

type User struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	Username       string  `gorm:"uniqueIndex;not null" json:"username"`
	Email          string  `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   *string `json:"-"`
	GoogleID       *string `gorm:"uniqueIndex" json:"-"`
	Country        *string `json:"country,omitempty"`
	City           *string `json:"city,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	CoverPicture   *string `json:"coverPicture,omitempty"`
	Status         string  `gorm:"default:free" json:"status"`

	AllowFollowers bool `gorm:"default:true" json:"allowFollowers"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Followers []User `gorm:"many2many:user_relationships;foreignKey:ID;joinForeignKey:FollowingID;references:ID;joinReferences:FollowerID" json:"followers,omitempty"`
	Following []User `gorm:"many2many:user_relationships;foreignKey:ID;joinForeignKey:FollowerID;references:ID;joinReferences:FollowingID" json:"following,omitempty"`

	Posts    []Post    `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Likes    []Like    `gorm:"foreignKey:UserID" json:"likes,omitempty"`
	Channels []Channel `gorm:"foreignKey:OwnerID" json:"channels,omitempty"`
}
