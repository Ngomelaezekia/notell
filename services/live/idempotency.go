package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IdempotencyRecord struct {
	ID uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;index"`
	Key string `gorm:"size:200;not null"`
	Operation string `gorm:"size:80;not null"`
	ResourceID string `gorm:"size:80"`
	StatusCode int `gorm:"not null"`
	ResponseBody []byte `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ensureIdempotencySchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&IdempotencyRecord{}); err != nil { return err }
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_live_idempotency_user_key ON idempotency_records (user_id, key)`).Error
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" || len(id) > 128 { id = randomKey() }
		c.Set("requestId", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func idempotencyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost { c.Next(); return }
		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" || len(key) > 200 { c.Next(); return }
		uid, ok := userID(c)
		if !ok || uid == 0 { c.Next(); return }
		op := c.Request.Method + " " + c.FullPath()
		record := IdempotencyRecord{UserID: uid, Key: key, Operation: op, StatusCode: 0}
		if err := db.Create(&record).Error; err != nil {
			var existing IdempotencyRecord
			if db.Where("user_id = ? AND key = ?", uid, key).First(&existing).Error != nil { c.AbortWithStatusJSON(409, gin.H{"message":"idempotency key conflict"}); return }
			if existing.StatusCode == 0 { c.AbortWithStatusJSON(409, gin.H{"message":"request with this idempotency key is already in progress"}); return }
			c.Data(existing.StatusCode, "application/json; charset=utf-8", existing.ResponseBody)
			return
		}
		w := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()
		body := append([]byte(nil), w.body...)
		if len(body) == 0 { body = []byte(`{"message":""}`) }
		if !json.Valid(body) { body = []byte(`{"message":"request completed"}`) }
		record.StatusCode = w.status
		record.ResponseBody = body
		_ = db.Model(&IdempotencyRecord{}).Where("id = ?", record.ID).Updates(map[string]interface{}{"status_code":record.StatusCode,"response_body":record.ResponseBody}).Error
	}
}

type captureWriter struct { gin.ResponseWriter; status int; body []byte }
func (w *captureWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }
func (w *captureWriter) Write(b []byte) (int,error) { if w.status == 0 { w.status = http.StatusOK }; w.body = append(w.body,b...); return w.ResponseWriter.Write(b) }
