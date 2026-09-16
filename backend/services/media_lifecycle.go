package services

import "fmt"

// MediaStateTransition defines the only valid media lifecycle movement.
// PENDING -> UPLOADING -> READY is the successful path.
// Any terminal state cannot become active again.
func ValidateMediaTransition(from, to string) error {
	if from == to {
		return nil
	}
	if from == "DELETED" || from == "FAILED" {
		return fmt.Errorf("media state %s cannot transition to %s", from, to)
	}
	if from == "READY" && to != "DELETED" {
		return fmt.Errorf("ready media cannot transition to %s", to)
	}
	if from == "PENDING" && to == "UPLOADING" {
		return nil
	}
	if from == "UPLOADING" && (to == "READY" || to == "FAILED") {
		return nil
	}
	return fmt.Errorf("invalid media transition %s -> %s", from, to)
}
