package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func multipartRequest(t *testing.T, filename, contentType string, body []byte) *http.Request {
	t.Helper()
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if contentType != "" {
		_ = contentType
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadMediaRejectsUnsupportedContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()
	r := gin.New()
	r.POST("/upload", h.UploadMedia)

	req := multipartRequest(t, "image.jpg", "image/jpeg", []byte("not an image"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRandomFilenameUsesSafeExtension(t *testing.T) {
	filename, err := randomFilename(".jpg")
	if err != nil {
		t.Fatalf("randomFilename() error = %v", err)
	}
	if filepath.Ext(filename) != ".jpg" {
		t.Fatalf("extension = %q, want .jpg", filepath.Ext(filename))
	}
	if filename == ".jpg" {
		t.Fatal("randomFilename() returned an empty base name")
	}
}

func TestUploadMediaRejectsEmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()
	r := gin.New()
	r.POST("/upload", h.UploadMedia)

	req := multipartRequest(t, "empty.jpg", "image/jpeg", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUploadDirectoryIsNotCreatedByRejectedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()
	r := gin.New()
	r.POST("/upload", h.UploadMedia)

	req := multipartRequest(t, "payload.bin", "application/octet-stream", []byte("payload"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if _, err := os.Stat(filepath.Join(".", "uploads")); err == nil {
		t.Fatal("rejected upload created the uploads directory")
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat() error = %v", err)
	}
}
