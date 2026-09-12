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
	Public bool
	Private bool
}

// Authorize resolves the post that owns an upload and applies the media policy.
// Keeping this query here prevents HTTP handlers from growing their own rules.
func Authorize(db *gorm.DB, filename string, userID uint) (AccessPolicy, error) {
	if db == nil || filename == "" {
		return AccessPolicy{}, ErrMediaAccessDenied
	}
	var post models.Post
	err := db.Select("posts.user_id, posts.visibility").
		Joins("JOIN uploads ON uploads.post_id = posts.id").
		Where("uploads.filename = ? AND uploads.post_id IS NOT NULL", filename).
		First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccessPolicy{}, ErrMediaAccessDenied
	}
	if err != nil {
		return AccessPolicy{}, err
	}
	if post.Visibility == "private" {
		if userID == 0 || userID != post.UserID {
			return AccessPolicy{Private: true}, ErrMediaAccessDenied
		}
		return AccessPolicy{Private: true}, nil
	}
	return AccessPolicy{Public: true}, nil
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
