package services

import "time"

// MediaSigner defines the contract for generating playback URLs.
// Implementations can provide B2, S3, R2, or CDN based signing.
type MediaSigner interface {
	GeneratePlaybackURL(objectPath string, expiry time.Duration) (string, error)
}
