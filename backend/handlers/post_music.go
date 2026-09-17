package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postMusicInput struct {
	UploadID uint    `json:"uploadId"`
	Source   string  `json:"source"`
	TrackID  string  `json:"trackId"`
	Provider string  `json:"provider"`
	StartSec float64 `json:"startSec"`
	EndSec   float64 `json:"endSec"`
	Volume   float64 `json:"volume"`
}

type musicTrackContract struct {
	ID           string  `json:"id"`
	Provider     string  `json:"provider"`
	Title        string  `json:"title"`
	Artist       string  `json:"artist"`
	ArtworkURL   string  `json:"artworkUrl"`
	PreviewURL   string  `json:"previewUrl"`
	DurationSec  float64 `json:"durationSec"`
	CanUseInPost bool    `json:"canUseInPost"`
	Rights       struct {
		Licensed  bool `json:"licensed"`
		UGCUse    bool `json:"ugcUse"`
		Streaming bool `json:"streaming"`
	} `json:"rights"`
}

type musicSegmentContract struct {
	TrackID      string  `json:"trackId"`
	StartSec     float64 `json:"startSec"`
	EndSec       float64 `json:"endSec"`
	DurationSec  float64 `json:"durationSec"`
	Provider     string  `json:"provider"`
	CanUseInPost bool    `json:"canUseInPost"`
}

const maxPostMusicDuration = 600.0

func validatePostMusicWindow(input postMusicInput, duration float64) error {
	if math.IsNaN(input.StartSec) || math.IsInf(input.StartSec, 0) ||
		math.IsNaN(input.EndSec) || math.IsInf(input.EndSec, 0) ||
		math.IsNaN(input.Volume) || math.IsInf(input.Volume, 0) ||
		math.IsNaN(duration) || math.IsInf(duration, 0) {
		return errors.New("music timing and volume values must be finite")
	}
	if input.StartSec < 0 || input.EndSec < 0 {
		return errors.New("music start and end times cannot be negative")
	}
	if input.EndSec > 0 && input.EndSec <= input.StartSec {
		return errors.New("music end time must be greater than start time")
	}
	if input.EndSec > 0 && input.EndSec-input.StartSec > maxPostMusicDuration {
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
	if input.Volume < 0 || input.Volume > 1 {
		return errors.New("music volume must be between 0 and 1")
	}
	return nil
}

func musicServiceURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("MUSIC_SERVICE_URL")), "/")
}

func fetchClearedMusic(trackID string, startSec, endSec float64) (musicTrackContract, musicSegmentContract, error) {
	base := musicServiceURL()
	if base == "" {
		return musicTrackContract{}, musicSegmentContract{}, errors.New("MUSIC_SERVICE_URL is not configured")
	}

	client := &http.Client{Timeout: 8 * time.Second}
	segmentURL := fmt.Sprintf(
		"%s/api/music/tracks/segment?trackId=%s&startSec=%s&endSec=%s",
		base,
		url.QueryEscape(trackID),
		strconv.FormatFloat(startSec, 'f', -1, 64),
		strconv.FormatFloat(endSec, 'f', -1, 64),
	)
	req, err := http.NewRequest(http.MethodGet, segmentURL, nil)
	if err != nil {
		return musicTrackContract{}, musicSegmentContract{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return musicTrackContract{}, musicSegmentContract{}, fmt.Errorf("music service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return musicTrackContract{}, musicSegmentContract{}, fmt.Errorf("music service rejected track: %s", resp.Status)
	}
	var segment musicSegmentContract
	if err := json.NewDecoder(resp.Body).Decode(&segment); err != nil {
		return musicTrackContract{}, musicSegmentContract{}, err
	}

	trackURL := fmt.Sprintf("%s/api/music/tracks/%s", base, url.PathEscape(trackID))
	req, err = http.NewRequest(http.MethodGet, trackURL, nil)
	if err != nil {
		return musicTrackContract{}, musicSegmentContract{}, err
	}
	resp, err = client.Do(req)
	if err != nil {
		return musicTrackContract{}, musicSegmentContract{}, fmt.Errorf("music track lookup failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return musicTrackContract{}, musicSegmentContract{}, errors.New("music track not found")
	}
	var track musicTrackContract
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return musicTrackContract{}, musicSegmentContract{}, err
	}
	if track.ID != trackID || !track.CanUseInPost || !track.Rights.Licensed || !track.Rights.UGCUse || !track.Rights.Streaming || !segment.CanUseInPost {
		return musicTrackContract{}, musicSegmentContract{}, errors.New("track is not cleared for post use and streaming")
	}
	return track, segment, nil
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
	input.Source = strings.ToLower(strings.TrimSpace(input.Source))
	if input.Source == "" {
		if input.TrackID != "" {
			input.Source = "cloud"
		} else {
			input.Source = "local"
		}
	}
	if input.Source != "cloud" && input.Source != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "music source must be local or cloud"})
		return
	}
	if input.Source == "cloud" && strings.TrimSpace(input.TrackID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "trackId is required for cloud music"})
		return
	}
	if input.Source == "local" && input.UploadID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "uploadId is required for device music"})
		return
	}

	var cloudTrack musicTrackContract
	var cloudSegment musicSegmentContract
	if input.Source == "cloud" {
		cloudTrack, cloudSegment, err = fetchClearedMusic(strings.TrimSpace(input.TrackID), input.StartSec, input.EndSec)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		input.Provider = cloudTrack.Provider
		input.StartSec = cloudSegment.StartSec
		input.EndSec = cloudSegment.EndSec
		if err := validatePostMusicWindow(input, cloudTrack.DurationSec); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	var music models.PostMusic
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id,user_id").First(&post, uint(postID)).Error; err != nil {
			return err
		}
		if post.UserID != userID {
			return gorm.ErrRecordNotFound
		}

		var next models.PostMusic
		if input.Source == "cloud" {
			next = models.PostMusic{
				PostID:      post.ID,
				Source:      "cloud",
				TrackID:     cloudTrack.ID,
				Provider:    cloudTrack.Provider,
				Title:       cloudTrack.Title,
				Artist:      cloudTrack.Artist,
				ArtworkURL:  cloudTrack.ArtworkURL,
				PreviewURL:  cloudTrack.PreviewURL,
				DurationSec: cloudTrack.DurationSec,
				StartSec:    input.StartSec,
				EndSec:      input.EndSec,
				Volume:      input.Volume,
			}
		} else {
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
			next = models.PostMusic{
				PostID:      post.ID,
				UploadID:    upload.ID,
				Source:      "local",
				DurationSec: metadata.Duration,
				StartSec:    input.StartSec,
				EndSec:      input.EndSec,
				Volume:      input.Volume,
			}
		}

		var previous []models.PostMusic
		if err := tx.Where("post_id = ?", post.ID).Find(&previous).Error; err != nil {
			return err
		}
		for _, item := range previous {
			if item.UploadID != 0 {
				if err := tx.Model(&models.Upload{}).Where("id = ? AND post_id = ?", item.UploadID, post.ID).Update("post_id", nil).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.PostMusic{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&next).Error; err != nil {
			return err
		}
		if input.Source == "local" {
			if err := tx.Model(&models.Upload{}).Where("id = ? AND user_id = ? AND post_id IS NULL", input.UploadID, userID).Update("post_id", post.ID).Error; err != nil {
				return err
			}
		}
		music = next
		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"message": "post music changed concurrently; please retry"})
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
		var post models.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id,user_id").First(&post, uint(postID)).Error; err != nil {
			return err
		}
		if post.UserID != userID {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("post_id = ?", post.ID).First(&music).Error; err != nil {
			return err
		}
		if music.UploadID != 0 {
			if err := tx.Model(&models.Upload{}).Where("id = ? AND post_id = ?", music.UploadID, post.ID).Update("post_id", nil).Error; err != nil {
				return err
			}
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
