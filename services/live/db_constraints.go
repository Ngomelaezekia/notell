package main

import "gorm.io/gorm"

// ensureLiveConstraints applies invariants that must exist in every environment.
func ensureLiveConstraints(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS ux_live_stream_one_active_per_channel
		ON streams (channel_id)
		WHERE status = 'live'
	`).Error; err != nil {
		return err
	}
	return ensureIdempotencySchema(db)
}
