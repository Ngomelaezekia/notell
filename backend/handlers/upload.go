package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadMedia handles image/video uploads
func (h *UploadHandler) UploadMedia(c *gin.Context) {

	file, err := c.FormFile("file")

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "file is required",
			},
		)

		return
	}


	// Validate file type

	allowedTypes := map[string]bool{

		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,

		"video/mp4":       true,
		"video/quicktime": true,
	}


	if !allowedTypes[file.Header.Get("Content-Type")] {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "unsupported file type",
			},
		)

		return
	}



	// Create upload directory

	uploadDir := "./uploads/"


	if err := os.MkdirAll(
		uploadDir,
		0755,
	); err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "failed creating upload directory",
			},
		)

		return
	}



	// Generate unique filename

	filename :=
		time.Now().
			Format("20060102150405") +
			"_" +
			file.Filename



	filePath :=
		filepath.Join(
			uploadDir,
			filename,
		)



	// Save file

	if err :=
		c.SaveUploadedFile(
			file,
			filePath,
		); err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "failed uploading file",
			},
		)

		return
	}




	// Generate public URL

	scheme := "http"

	if c.Request.TLS != nil {
		scheme = "https"
	}


	fileURL :=
		scheme +
			"://" +
			c.Request.Host +
			"/uploads/" +
			filename




	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "upload successful",
			"url":     fileURL,
		},
	)

}