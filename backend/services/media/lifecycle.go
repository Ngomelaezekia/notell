package media

import (
	"errors"

	"gorm.io/gorm"
)

// Valid media processing states.
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// Transition validates media state changes before persistence.
// Keeping transitions centralized prevents workers and handlers from creating
// impossible lifecycle states.
func Transition(current, next string) error {
	switch current {
	case StatusPending:
		if next == StatusProcessing || next == StatusFailed {
			return nil
		}
	case StatusProcessing:
		if next == StatusCompleted || next == StatusFailed || next == StatusPending {
			return nil
		}
	case StatusCompleted:
		if next == StatusCompleted {
			return nil
		}
	case StatusFailed:
		if next == StatusPending || next == StatusFailed {
			return nil
		}
	}
	return errors.New("invalid media lifecycle transition")
}

// UpdateLifecycle safely updates a lifecycle state with transition validation.
func UpdateLifecycle(db *gorm.DB, table string, id uint, current, next string) error {
	if err := Transition(current, next); err != nil {
		return err
	}

	result := db.Table(table).
		Where("id = ? AND status = ?", id, current).
		Update("status", next)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
