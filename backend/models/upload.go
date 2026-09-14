package models

import "time"

type Upload struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"userId"`
	Filename  string     `gorm:"uniqueIndex;not null;size:255" json:"filename"`
	Path      string     `gorm:"uniqueIndex;not null;size:255" json:"path"`
	MediaType string     `gorm:"not null;size:32" json:"mediaType"`
	// Multiple managed media objects may belong to the same post (for example
	// the primary image/video plus an attached music track). PostMusic is the
	// one-per-post relation for music; Upload.PostID must therefore be a normal
	// lookup index rather than a unique index.
	PostID    *uint      `gorm:"index" json:"postId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	ClaimedAt *time.Time `json:"claimedAt,omitempty"`

	User          User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Post          *Post          `gorm:"foreignKey:PostID;constraint:OnDelete:SET NULL" json:"-"`
	MediaMetadata *MediaMetadata `gorm:"foreignKey:UploadID;constraint:OnDelete:CASCADE" json:"mediaMetadata,omitempty"`
}
