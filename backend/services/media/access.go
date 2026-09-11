package media

import (
	"errors"
	"notell/models"
)

var ErrMediaAccessDenied = errors.New("media access denied")

// CanAccessMedia centralizes media authorization rules.
// The current rule set allows owners and preserves a single expansion point
// for subscription, privacy and signed playback policies.
func CanAccessMedia(userID uint, upload *models.Upload) error {
	if upload == nil {
		return ErrMediaAccessDenied
	}
	if upload.UserID == userID {
		return nil
	}
	return ErrMediaAccessDenied
}
