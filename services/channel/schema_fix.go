package main

import "gorm.io/gorm"

// repairMembershipIndex keeps follower and paid subscription membership
// independent for the same user/channel pair. It is intentionally explicit
// because the first service migration must be safe for both fresh and
// existing databases.
func repairMembershipIndex(db *gorm.DB) error {
	if db.Migrator().HasIndex(&ChannelMember{}, "idx_channel_member") {
		if err := db.Migrator().DropIndex(&ChannelMember{}, "idx_channel_member"); err != nil { return err }
	}
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_member_type ON channel_members (channel_id, user_id, membership_type)`).Error
}
