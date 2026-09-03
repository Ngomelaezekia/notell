package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func multipartRequest(t *testing.T, filename string, body []byte) *http.Request {
	t.Helper()
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
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

func newUploadTestRouter(h *UploadHandler) *gin.Engine {
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) {
		c.Set("userId", uint(1))
		h.UploadMedia(c)
	})
	return r
}

func TestUploadMediaRejectsUnsupportedContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler(nil, "http://localhost:8080")
	r := newUploadTestRouter(h)

	req := multipartRequest(t, "image.jpg", []byte("not an image"))
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
	h := NewUploadHandler(nil, "http://localhost:8080")
	r := newUploadTestRouter(h)

	req := multipartRequest(t, "empty.jpg", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
