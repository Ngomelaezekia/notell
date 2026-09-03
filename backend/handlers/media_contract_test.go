package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateManagedMediaAcceptsMatchingImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.webp")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test media: %v", err)
	}

	if err := validateManagedMedia(path, "image"); err != nil {
		t.Fatalf("expected matching image to validate: %v", err)
	}
}

func TestValidateManagedMediaRejectsImageAsVideo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.webp")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test media: %v", err)
	}

	if err := validateManagedMedia(path, "video"); err == nil {
		t.Fatal("expected image submitted as video to be rejected")
	}
}

func TestValidateManagedMediaRejectsVideoAsImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test media: %v", err)
	}

	if err := validateManagedMedia(path, "image"); err == nil {
		t.Fatal("expected video submitted as image to be rejected")
	}
}

func TestValidateManagedMediaRejectsMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.jpg")

	if err := validateManagedMedia(path, "image"); err == nil {
		t.Fatal("expected missing media file to be rejected")
	}
}
