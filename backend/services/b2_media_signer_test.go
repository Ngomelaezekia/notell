package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestValidateB2MediaObjectPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid", "uploads/photo.jpg", true},
		{"valid double dot filename", "uploads/photo..jpg", true},
		{"empty", "", false},
		{"outside namespace", "media/photo.jpg", false},
		{"traversal", "uploads/../photo.jpg", false},
		{"nested", "uploads/user/photo.jpg", false},
		{"windows traversal", "uploads/..\\photo.jpg", false},
		{"namespace only", "uploads/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateB2MediaObjectPath(tt.path)
			if (err == nil) != tt.want {
				t.Fatalf("validateB2MediaObjectPath(%q) error=%v, wantValid=%v", tt.path, err, tt.want)
			}
		})
	}
}

func TestB2MediaSignerValidationRunsBeforePresigner(t *testing.T) {
	signer := &B2MediaSigner{}
	for _, objectPath := range []string{"", "media/file.jpg", "uploads/../file.jpg", "uploads/user/file.jpg"} {
		_, err := signer.GeneratePlaybackURL(context.Background(), objectPath, time.Minute)
		if err == nil {
			t.Fatalf("GeneratePlaybackURL(%q) unexpectedly succeeded", objectPath)
		}
	}
}

func TestB2MediaSignerExpiryConstants(t *testing.T) {
	if defaultPlaybackURLExpiry != 5*time.Minute {
		t.Fatalf("default expiry = %s, want 5m", defaultPlaybackURLExpiry)
	}
	if maxPlaybackURLExpiry != 15*time.Minute {
		t.Fatalf("max expiry = %s, want 15m", maxPlaybackURLExpiry)
	}
}

type rollbackTestStorage struct {
	deleteCalls int
	deleteErr   error
}

func (s *rollbackTestStorage) Delete(context.Context, string) error {
	s.deleteCalls++
	return s.deleteErr
}

func TestRollbackUploadedObject(t *testing.T) {
	storage := &rollbackTestStorage{}
	if err := RollbackUploadedObject(context.Background(), storage, "uploads/file.jpg"); err != nil {
		t.Fatalf("RollbackUploadedObject() error = %v", err)
	}
	if storage.deleteCalls != 1 {
		t.Fatalf("delete calls = %d, want 1", storage.deleteCalls)
	}

	storage.deleteErr = errors.New("delete failed")
	if err := RollbackUploadedObject(context.Background(), storage, "uploads/file.jpg"); !errors.Is(err, storage.deleteErr) {
		t.Fatalf("rollback error = %v, want %v", err, storage.deleteErr)
	}
	if storage.deleteCalls != 2 {
		t.Fatalf("delete calls after failed rollback = %d, want 2", storage.deleteCalls)
	}
}

func TestRollbackUploadedObjectRejectsInvalidInput(t *testing.T) {
	storage := &rollbackTestStorage{}
	if err := RollbackUploadedObject(context.Background(), nil, "uploads/file.jpg"); err == nil {
		t.Fatal("expected nil storage to fail")
	}
	if err := RollbackUploadedObject(context.Background(), storage, ""); err == nil {
		t.Fatal("expected empty key to fail")
	}
	if storage.deleteCalls != 0 {
		t.Fatalf("delete calls = %d, want 0", storage.deleteCalls)
	}
}
