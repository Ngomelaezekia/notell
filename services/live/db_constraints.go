package main

import "gorm.io/gorm"

// ensureLiveConstraints applies invariants that must exist in every environment.
// The partial unique index prevents two live sessions for the same channel even
// when multiple live-service instances race concurrently.
func ensureLiveConstraints(db *gorm.DB) error {
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS ux_live_stream_one_active_per_channel
		ON streams (channel_id)
		WHERE status = 'live'
	`).Error
}
