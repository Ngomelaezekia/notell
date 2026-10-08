package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// runDatabaseMigrations applies explicit PostgreSQL migrations that GORM
// AutoMigrate cannot safely express (constraints and schema evolution).
// Each migration is recorded so startup is idempotent across deployments.
func runDatabaseMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	if err := db.Exec("SELECT pg_advisory_lock(?)", int64(78123941)).Error; err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer db.Exec("SELECT pg_advisory_unlock(?)", int64(78123941))

	migrations := []struct {
		version int64
		name    string
		run     func(*gorm.DB) error
	}{
		{
			1,
			"post_visibility_relationship_constraints_and_post_audiences",
			func(tx *gorm.DB) error {
				// Refuse to silently reinterpret corrupted authorization data.
				var invalidVisibility int64
				if err := tx.Raw(`
					SELECT COUNT(*)
					FROM posts
					WHERE visibility IS NULL OR visibility NOT IN ('public','private','followers','subscribers','selected')
				`).Scan(&invalidVisibility).Error; err != nil {
					return fmt.Errorf("check post visibility data: %w", err)
				}
				if invalidVisibility != 0 {
					return fmt.Errorf("cannot add post visibility constraint: %d invalid rows exist", invalidVisibility)
				}

				var invalidRelationshipStatus int64
				if err := tx.Raw(`
					SELECT COUNT(*)
					FROM user_relationships
					WHERE status IS NULL OR status NOT IN ('accepted')
				`).Scan(&invalidRelationshipStatus).Error; err != nil {
					return fmt.Errorf("check relationship status data: %w", err)
				}
				if invalidRelationshipStatus != 0 {
					return fmt.Errorf("cannot add relationship status constraint: %d invalid rows exist", invalidRelationshipStatus)
				}

				var visibilityConstraintCount int64
				if err := tx.Raw(`SELECT COUNT(*) FROM pg_constraint WHERE conname = 'chk_posts_visibility' AND conrelid = 'posts'::regclass`).Scan(&visibilityConstraintCount).Error; err != nil {
					return fmt.Errorf("check post visibility constraint: %w", err)
				}
				if visibilityConstraintCount == 0 {
					if err := tx.Exec(`ALTER TABLE posts ADD CONSTRAINT chk_posts_visibility CHECK (visibility IN ('public','private','followers','subscribers','selected'))`).Error; err != nil {
						return fmt.Errorf("add post visibility constraint: %w", err)
					}
				}

				var relationshipConstraintCount int64
				if err := tx.Raw(`SELECT COUNT(*) FROM pg_constraint WHERE conname = 'chk_relationships_status' AND conrelid = 'user_relationships'::regclass`).Scan(&relationshipConstraintCount).Error; err != nil {
					return fmt.Errorf("check relationship status constraint: %w", err)
				}
				if relationshipConstraintCount == 0 {
					if err := tx.Exec(`ALTER TABLE user_relationships ADD CONSTRAINT chk_relationships_status CHECK (status IN ('accepted'))`).Error; err != nil {
						return fmt.Errorf("add relationship status constraint: %w", err)
					}
				}

				if err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS post_audiences (
						post_id BIGINT NOT NULL,
						user_id BIGINT NOT NULL,
						created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
						PRIMARY KEY (post_id, user_id),
						CONSTRAINT fk_post_audiences_post
							FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
						CONSTRAINT fk_post_audiences_user
							FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					)
				`).Error; err != nil {
					return fmt.Errorf("create post_audiences: %w", err)
				}
				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_post_audiences_user_post
					ON post_audiences (user_id, post_id)
				`).Error; err != nil {
					return fmt.Errorf("index post_audiences: %w", err)
				}

				// Upload.PostID is intentionally non-unique because a post may have
				// multiple managed uploads. Recreate it as a normal lookup index.
				if err := tx.Exec(`DROP INDEX IF EXISTS idx_uploads_post_id`).Error; err != nil {
					return fmt.Errorf("drop legacy upload post index: %w", err)
				}
				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_uploads_post_id
					ON uploads (post_id)
				`).Error; err != nil {
					return fmt.Errorf("create upload post index: %w", err)
				}

				return nil
			},
		},
		{
			2,
			"ensure_post_audience_table_and_indexes",
			func(tx *gorm.DB) error {
				if err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS post_audiences (
						post_id BIGINT NOT NULL,
						user_id BIGINT NOT NULL,
						created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
						PRIMARY KEY (post_id, user_id),
						CONSTRAINT fk_post_audiences_post
							FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
						CONSTRAINT fk_post_audiences_user
							FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					)
				`).Error; err != nil {
					return fmt.Errorf("ensure post_audiences: %w", err)
				}
				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_post_audiences_user_post
					ON post_audiences (user_id, post_id)
				`).Error; err != nil {
					return fmt.Errorf("ensure post audience user index: %w", err)
				}
				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_uploads_post_id
					ON uploads (post_id)
				`).Error; err != nil {
					return fmt.Errorf("ensure upload post index: %w", err)
				}
				return nil
			},
		},
	}

	for _, migration := range migrations {
		var applied int64
		if err := db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.version).Scan(&applied).Error; err != nil {
			return fmt.Errorf("check migration %d (%s): %w", migration.version, migration.name, err)
		}
		if applied != 0 {
			continue
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.run(tx); err != nil {
				return err
			}
			return tx.Exec(
				"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
				migration.version,
				time.Now().UTC(),
			).Error
		})
		if err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", migration.version, migration.name, err)
		}
	}

	return nil
}
