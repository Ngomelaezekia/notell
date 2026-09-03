package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	windowStart time.Time
	count       int
}

// RateLimit provides a small in-memory fixed-window limiter for single-instance deployments.
// For multi-instance production deployments, replace this with a shared store such as Redis.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	var mu sync.Mutex
	entries := make(map[string]rateLimitEntry)
	lastCleanup := time.Now()

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()

		mu.Lock()
		entry, exists := entries[key]
		if !exists || now.Sub(entry.windowStart) >= window {
			entry = rateLimitEntry{windowStart: now, count: 0}
		}
		entry.count++
		entries[key] = entry

		if now.Sub(lastCleanup) >= window {
			for client, candidate := range entries {
				if now.Sub(candidate.windowStart) >= window {
					delete(entries, client)
				}
			}
			lastCleanup = now
		}
		mu.Unlock()

		if entry.count > limit {
			retryAfter := int(window.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", string(rune('0'+retryAfter)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "too many requests"})
			return
		}

		c.Next()
	}
}
