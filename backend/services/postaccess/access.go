package postaccess

import (
	"errors"

	"notell/models"

	"gorm.io/gorm"
)

var (
	ErrAccessDenied = errors.New("post access denied")
	ErrPostNotFound = errors.New("post not found")
)

// CanViewPost enforces the current public/private visibility contract.
// Unknown or NULL visibility is never treated as public.
func CanViewPost(db *gorm.DB, postID, viewerID uint) error {
	if db == nil || postID == 0 {
		return ErrPostNotFound
	}

	var post models.Post
	if err := db.Select("id, user_id, visibility").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}
	if post.Visibility == "public" || post.UserID == viewerID {
		return nil
	}
	return ErrAccessDenied
}

// VisibleTo is the canonical SQL scope for list/search/profile queries.
func VisibleTo(db *gorm.DB, viewerID uint) *gorm.DB {
	return db.Where("(posts.visibility = ? OR posts.user_id = ?)", "public", viewerID)
}

func CanEditPost(db *gorm.DB, postID, viewerID uint) error {
	if db == nil || postID == 0 || viewerID == 0 {
		return ErrAccessDenied
	}
	var post models.Post
	if err := db.Select("id, user_id").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}
	if post.UserID != viewerID {
		return ErrAccessDenied
	}
	return nil
}

func CanDeletePost(db *gorm.DB, postID, viewerID uint) error {
	return CanEditPost(db, postID, viewerID)
}

func CanManageMusic(db *gorm.DB, postID, viewerID uint) error {
	return CanEditPost(db, postID, viewerID)
}

func CanManageMedia(db *gorm.DB, postID, viewerID uint) error {
	return CanEditPost(db, postID, viewerID)
}
