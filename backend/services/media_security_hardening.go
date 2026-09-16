package services

import "errors"

var ErrMediaNotReadyForPlayback = errors.New("media is not ready for playback")
var ErrMediaAccessDenied = errors.New("media access denied")

// CanGeneratePlayback prevents invalid lifecycle states from reaching playback signing.
func CanGeneratePlayback(status string) bool {
	return status == "ready"
}

// ValidateMediaOwner keeps private media access ownership checks centralized.
func ValidateMediaOwner(ownerID uint, requesterID uint) error {
	if ownerID == 0 || ownerID != requesterID {
		return ErrMediaAccessDenied
	}
	return nil
}
