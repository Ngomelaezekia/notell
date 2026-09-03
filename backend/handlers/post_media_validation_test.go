package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateManagedMedia(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name        string
		extension   string
		contentType string
		wantErr     bool
	}{
		{name: "jpeg image", extension: ".jpg", contentType: "image"},
		{name: "png image", extension: ".png", contentType: "image"},
		{name: "webp image", extension: ".webp", contentType: "image"},
		{name: "mp4 video", extension: ".mp4", contentType: "video"},
		{name: "mov video", extension: ".mov", contentType: "video"},
		{name: "video declared image", extension: ".mp4", contentType: "image", wantErr: true},
		{name: "image declared video", extension: ".jpg", contentType: "video", wantErr: true},
		{name: "unsupported extension", extension: ".txt", contentType: "image", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "media"+tc.extension)
			if err := os.WriteFile(path, []byte("test"), 0600); err != nil {
				t.Fatal(err)
			}
			err := validateManagedMedia(path, tc.contentType)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateManagedMedia() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateManagedMediaRejectsMissingFile(t *testing.T) {
	err := validateManagedMedia(filepath.Join(t.TempDir(), "missing.jpg"), "image")
	if err == nil || err.Error() != "uploaded media file not found" {
		t.Fatalf("expected missing-file error, got %v", err)
	}
}

func TestManagedMediaPathRejectsTraversal(t *testing.T) {
	h := NewPostHandler(nil, "https://example.test")
	if path, ok := h.managedMediaPath("https://example.test/uploads/../secret.jpg"); ok || path != "" {
		t.Fatalf("expected traversal path to be rejected, got %q", path)
	}
}
