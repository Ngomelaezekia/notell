package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

const maxUploadSize int64 = 100 << 20 // 100 MiB

var allowedUploadTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func randomFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}

// UploadMedia handles image/video uploads.
func (h *UploadHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file is required"})
		return
	}

	if file.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file is empty"})
		return
	}
	if file.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"message": "file exceeds the 100 MiB limit"})
		return
	}

	input, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to inspect uploaded file"})
		return
	}
	defer input.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(input, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to inspect uploaded file"})
		return
	}
	contentType := http.DetectContentType(header[:n])
	ext, allowed := allowedUploadTypes[contentType]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported file type"})
		return
	}

	if _, err := input.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read uploaded file"})
		return
	}

	uploadDir := filepath.Join(".", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed creating upload directory"})
		return
	}

	filename, err := randomFilename(ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed generating upload name"})
		return
	}

	filePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed uploading file"})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	fileURL := fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Request.Host, filename)

	c.JSON(http.StatusOK, gin.H{
		"message": "upload successful",
		"url":     fileURL,
		"type":    strings.TrimPrefix(contentType, ""),
	})
}
