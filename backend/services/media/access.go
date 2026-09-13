package media

import (
	"errors"

	"notell/models"

	"gorm.io/gorm"
)

var ErrMediaAccessDenied = errors.New("media access denied")

// AccessPolicy is the single authorization boundary for media playback.
// The current MVP policy permits public media to everyone and private media
// only to its owner. Future follower/subscriber/channel rules belong here.
type AccessPolicy struct {
	Public  bool
	Private bool
}

type mediaOwnerPolicy struct {
	UserID     uint
	Visibility string
}

// Authorize resolves both primary post media and optional post music uploads,
// then applies the post visibility policy in one place.
func Authorize(db *gorm.DB, filename string, userID uint) (AccessPolicy, error) {
	if db == nil || filename == "" {
		return AccessPolicy{}, ErrMediaAccessDenied
	}

	var owner mediaOwnerPolicy
	err := db.Table("uploads").
		Select("COALESCE(primary_posts.user_id, music_posts.user_id) AS user_id, COALESCE(primary_posts.visibility, music_posts.visibility) AS visibility").
		Joins("LEFT JOIN posts AS primary_posts ON primary_posts.id = uploads.post_id").
		Joins("LEFT JOIN post_musics ON post_musics.upload_id = uploads.id").
		Joins("LEFT JOIN posts AS music_posts ON music_posts.id = post_musics.post_id").
		Where("uploads.filename = ? AND (uploads.post_id IS NOT NULL OR post_musics.post_id IS NOT NULL)", filename).
		Limit(1).
		Scan(&owner).Error
	if err != nil {
		return AccessPolicy{}, err
	}
	if owner.UserID == 0 || owner.Visibility == "" {
		return AccessPolicy{}, ErrMediaAccessDenied
	}

	if owner.Visibility == "private" {
		if userID == 0 || userID != owner.UserID {
			return AccessPolicy{Private: true}, ErrMediaAccessDenied
		}
		return AccessPolicy{Private: true}, nil
	}
	if owner.Visibility == "public" {
		return AccessPolicy{Public: true}, nil
	}
	return AccessPolicy{}, ErrMediaAccessDenied
}

// CanAccessMedia remains a lightweight owner check for service callers that
// already loaded the upload. HTTP playback should use Authorize above so the
// post visibility policy is applied consistently.
func CanAccessMedia(userID uint, upload *models.Upload) error {
	if upload == nil || upload.UserID != userID {
		return ErrMediaAccessDenied
	}
	return nil
}
