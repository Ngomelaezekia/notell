package media

import (
	"errors"

	"gorm.io/gorm"
)

// Media metadata states. Uploads enter uploaded, then move through processing
// to ready. Failed is terminal until an explicit retry returns it to processing.
const (
	StatusUploaded   = "uploaded"
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

func Transition(current, next string) error {
	switch current {
	case StatusUploaded:
		if next == StatusProcessing || next == StatusFailed { return nil }
	case StatusPending:
		if next == StatusProcessing || next == StatusFailed { return nil }
	case StatusProcessing:
		if next == StatusReady || next == StatusCompleted || next == StatusFailed || next == StatusPending { return nil }
	case StatusReady:
		if next == StatusReady || next == StatusCompleted { return nil }
	case StatusCompleted:
		if next == StatusCompleted { return nil }
	case StatusFailed:
		if next == StatusPending || next == StatusProcessing || next == StatusFailed { return nil }
	}
	return errors.New("invalid media lifecycle transition")
}

func UpdateLifecycle(db *gorm.DB, table string, id uint, current, next string) error {
	if err := Transition(current, next); err != nil { return err }
	result := db.Table(table).Where("id = ? AND status = ?", id, current).Update("status", next)
	if result.Error != nil { return result.Error }
	if result.RowsAffected != 1 { return gorm.ErrRecordNotFound }
	return nil
}
