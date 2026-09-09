package models

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID          uint      `gorm:"primaryKey" json:"postId"`
	UserID      uint      `gorm:"not null;index:idx_posts_user_created,priority:1" json:"userId"`
	UploadID    *uint     `gorm:"uniqueIndex" json:"uploadId,omitempty"`
	ContentType string    `gorm:"not null" json:"contentType"`
	ContentURL  string    `gorm:"not null" json:"contentUrl"`
	Caption     string    `gorm:"type:text" json:"caption"`
	ViewCount   int64     `gorm:"not null;default:0;index:idx_posts_views" json:"viewCount"`
	CreatedAt   time.Time `gorm:"index:idx_posts_user_created,priority:2" json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	LikeCount    int  `gorm:"-" json:"likeCount"`
	CommentCount int  `gorm:"-" json:"commentCount"`
	Liked        bool `gorm:"-" json:"liked"`

	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Upload   *Upload   `gorm:"foreignKey:UploadID;constraint:OnDelete:SET NULL" json:"upload,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}

// BeforeCreate enforces the persisted upload/content-type contract at the
// database boundary. The HTTP handler performs the same validation earlier,
// but this prevents another Post creation path from bypassing it.
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	candidate, err := url.Parse(strings.TrimSpace(p.ContentURL))
	if err != nil || candidate.Scheme == "" || candidate.Host == "" || candidate.Path == "" {
		return errors.New("invalid post media URL")
	}
	if candidate.RawQuery != "" || candidate.Fragment != "" || !strings.HasPrefix(candidate.Path, "/uploads/") {
		return errors.New("invalid post media URL")
	}

	relative := strings.TrimPrefix(candidate.Path, "/uploads/")
	decodedRelative, err := url.PathUnescape(relative)
	if err != nil || decodedRelative == "" || decodedRelative != relative {
		return errors.New("invalid post media path")
	}
	filename := filepath.Base(filepath.FromSlash(decodedRelative))
	if filename != decodedRelative || filename == "." || filename == string(filepath.Separator) || filename == "" {
		return errors.New("invalid post media filename")
	}

	var upload Upload
	if err := tx.Select("id, user_id, media_type").Where("filename = ?", filename).First(&upload).Error; err != nil {
		return err
	}
	if upload.UserID != p.UserID {
		return errors.New("uploaded media is not owned by the post author")
	}
	if p.UploadID != nil && *p.UploadID != upload.ID {
		return errors.New("post upload does not match content URL")
	}
	p.UploadID = &upload.ID

	isImage := strings.HasPrefix(strings.ToLower(upload.MediaType), "image/")
	isVideo := strings.HasPrefix(strings.ToLower(upload.MediaType), "video/")
	switch p.ContentType {
	case "image":
		if !isImage {
			return errors.New("uploaded media type does not match image content type")
		}
	case "video":
		if !isVideo {
			return errors.New("uploaded media type does not match video content type")
		}
	default:
		return errors.New("unsupported post content type")
	}

	return nil
}
