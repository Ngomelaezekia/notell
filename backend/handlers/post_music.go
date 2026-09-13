package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postMusicInput struct {
	UploadID uint    `json:"uploadId" binding:"required"`
	StartSec float64 `json:"startSec"`
	EndSec   float64 `json:"endSec"`
	Volume   float64 `json:"volume"`
}

const maxPostMusicDuration = 600.0

func validatePostMusicWindow(input postMusicInput, duration float64) error {
	if input.StartSec < 0 || input.EndSec < 0 {
		return errors.New("music start and end times cannot be negative")
	}
	if input.EndSec > 0 && input.EndSec <= input.StartSec {
		return errors.New("music end time must be greater than start time")
	}
	if input.EndSec-input.StartSec > maxPostMusicDuration {
		return errors.New("music playback window cannot exceed 10 minutes")
	}
	if duration > 0 {
		if input.StartSec >= duration {
			return errors.New("music start time exceeds track duration")
		}
		if input.EndSec > duration {
			return errors.New("music end time exceeds track duration")
		}
	}
	if input.Volume == 0 {
		input.Volume = 1
	}
	if input.Volume < 0 || input.Volume > 1 {
		return errors.New("music volume must be between 0 and 1")
	}
	return nil
}

func (h *PostHandler) SetPostMusic(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}
	var input postMusicInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if input.Volume == 0 {
		input.Volume = 1
	}

	var music models.PostMusic
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.Select("id,user_id").First(&post, uint(postID)).Error; err != nil {
			return err
		}
		if post.UserID != userID {
			return gorm.ErrRecordNotFound
		}

		var upload models.Upload
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ? AND media_type = ?", input.UploadID, userID, "audio/mpeg").First(&upload).Error; err != nil {
			return err
		}
		if upload.PostID != nil {
			return errors.New("audio upload is already attached to a post")
		}
		var metadata models.MediaMetadata
		if err := tx.Where("upload_id = ?", upload.ID).First(&metadata).Error; err != nil {
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(metadata.Status), "ready") {
			return errors.New("music is not ready for playback")
		}
		if err := validatePostMusicWindow(input, metadata.Duration); err != nil {
			return err
		}

		var previous []models.PostMusic
		if err := tx.Where("post_id = ?", post.ID).Find(&previous).Error; err != nil {
			return err
		}
		for _, item := range previous {
			if err := tx.Model(&models.Upload{}).Where("id = ? AND post_id = ?", item.UploadID, post.ID).Update("post_id", nil).Error; err != nil {
			return err
			}
		}
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.PostMusic{}).Error; err != nil {
			return err
		}

		music = models.PostMusic{PostID: post.ID, UploadID: upload.ID, StartSec: input.StartSec, EndSec: input.EndSec, Volume: input.Volume}
		if err := tx.Create(&music).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Upload{}).Where("id = ? AND user_id = ? AND post_id IS NULL", upload.ID, userID).Update("post_id", post.ID).Error; err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post music updated successfully", "data": music})
}

func (h *PostHandler) RemovePostMusic(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var music models.PostMusic
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", uint(postID)).First(&music).Error; err != nil {
			return err
		}
		var post models.Post
		if err := tx.Select("id,user_id").First(&post, music.PostID).Error; err != nil {
			return err
		}
		if post.UserID != userID {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Model(&models.Upload{}).Where("id = ? AND post_id = ?", music.UploadID, post.ID).Update("post_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&music).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post music not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to remove post music"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post music removed successfully"})
}
