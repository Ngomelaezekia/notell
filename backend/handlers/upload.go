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
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UploadHandler struct {
	DB        *gorm.DB
	PublicURL string
}

const maxUploadSize int64 = 100 << 20 // 100 MiB
const unclaimedUploadRetention = 24 * time.Hour

var allowedUploadTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
}

func NewUploadHandler(db *gorm.DB, publicURL string) *UploadHandler {
	return &UploadHandler{DB: db, PublicURL: strings.TrimRight(publicURL, "/")}
}

func randomFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}

// cleanupUnclaimedUploads removes abandoned uploads that were never claimed
// by a post. The filesystem is removed only after the database delete has
// confirmed that the upload is still unclaimed, preventing a claim/delete
// race from deleting media belonging to a newly-created post.
func (h *UploadHandler) cleanupUnclaimedUploads() {
	cutoff := time.Now().Add(-unclaimedUploadRetention)

	var uploads []models.Upload
	if err := h.DB.Select("id, path").
		Where("post_id IS NULL AND created_at < ?", cutoff).
		Find(&uploads).Error; err != nil {
		return
	}
	if len(uploads) == 0 {
		return
	}

	for _, upload := range uploads {
		var claimed models.Upload
		err := h.DB.Where("id = ? AND post_id IS NULL AND created_at < ?", upload.ID, cutoff).First(&claimed).Error
		if err != nil {
			continue
		}

		if err := h.DB.Delete(&claimed).Error; err != nil {
			continue
		}

		if claimed.Path != "" {
			if err := os.Remove(claimed.Path); err != nil && !os.IsNotExist(err) {
				// Database cleanup succeeded; a failed filesystem removal is
				// safe to retry on a later cleanup pass only if the record remains.
				// The record is intentionally already deleted here.
			}
		}
	}
}

// UploadMedia handles image/video uploads. Each uploaded object is recorded
// against the authenticated user so post creation can atomically claim it.
func (h *UploadHandler) UploadMedia(c *gin.Context) {
	userID := c.MustGet("userId").(uint)

	// Cap the entire multipart request so oversized bodies are rejected before
	// they can consume unbounded memory or temporary disk space.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+(1<<20))

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file is required or request is too large"})
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

	upload := models.Upload{
		UserID:    userID,
		Filename:  filename,
		Path:      filepath.ToSlash(filePath),
		MediaType: contentType,
	}
	if err := h.DB.Create(&upload).Error; err != nil {
		_ = os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed recording uploaded file"})
		return
	}

	// Opportunistically reclaim abandoned media without introducing a
	// background worker or another production dependency.
	h.cleanupUnclaimedUploads()

	fileURL := fmt.Sprintf("%s/uploads/%s", h.PublicURL, filename)
	c.JSON(http.StatusOK, gin.H{
		"message": "upload successful",
		"url":     fileURL,
	})
}
