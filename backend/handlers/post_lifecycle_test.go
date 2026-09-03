package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"notell/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPostLifecycleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Skipf("sqlite test driver unavailable: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Upload{}, &models.Post{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPostBeforeCreateRejectsUnownedUpload(t *testing.T) {
	db := newPostLifecycleTestDB(t)

	upload := models.Upload{UserID: 1, Filename: "owned.jpg", Path: "uploads/owned.jpg", MediaType: "image/jpeg"}
	if err := db.Create(&upload).Error; err != nil {
		t.Fatal(err)
	}

	post := models.Post{UserID: 2, ContentType: "image", ContentURL: "https://example.test/uploads/owned.jpg"}
	err := db.Create(&post).Error
	if err == nil || !containsError(err, "not owned") {
		t.Fatalf("expected ownership validation error, got %v", err)
	}
}

func TestPostBeforeCreateRejectsMediaTypeMismatch(t *testing.T) {
	db := newPostLifecycleTestDB(t)

	upload := models.Upload{UserID: 1, Filename: "video.mp4", Path: "uploads/video.mp4", MediaType: "video/mp4"}
	if err := db.Create(&upload).Error; err != nil {
		t.Fatal(err)
	}

	post := models.Post{UserID: 1, ContentType: "image", ContentURL: "https://example.test/uploads/video.mp4"}
	err := db.Create(&post).Error
	if err == nil || !containsError(err, "does not match image") {
		t.Fatalf("expected media type validation error, got %v", err)
	}
}

func TestPostBeforeCreateAcceptsMatchingOwnedUpload(t *testing.T) {
	db := newPostLifecycleTestDB(t)

	upload := models.Upload{UserID: 1, Filename: "image.jpg", Path: "uploads/image.jpg", MediaType: "image/jpeg"}
	if err := db.Create(&upload).Error; err != nil {
		t.Fatal(err)
	}

	post := models.Post{UserID: 1, ContentType: "image", ContentURL: "https://example.test/uploads/image.jpg"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("expected valid post creation, got %v", err)
	}
}

func TestManagedMediaPathRejectsTraversal(t *testing.T) {
	h := NewPostHandler(nil, "https://example.test")
	if path, ok := h.managedMediaPath("https://example.test/uploads/../secret.jpg"); ok || path != "" {
		t.Fatalf("expected traversal path to be rejected, got %q", path)
	}
}

func TestValidateManagedMediaRejectsMissingFile(t *testing.T) {
	err := validateManagedMedia(filepath.Join(t.TempDir(), "missing.jpg"), "image")
	if !errors.Is(err, os.ErrNotExist) && (err == nil || err.Error() != "uploaded media file not found") {
		t.Fatalf("expected missing-file validation error, got %v", err)
	}
}

func containsError(err error, want string) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), want)
}

func contains(value, want string) bool {
	for i := 0; i+len(want) <= len(value); i++ {
		if value[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
