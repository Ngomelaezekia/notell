package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const maxRateLimitEntries = 10000

type rateLimitEntry struct {
	windowStart time.Time
	count       int
}

// RateLimit provides a small in-memory fixed-window limiter for single-instance deployments.
// Authenticated requests are keyed by user ID so changing source IPs cannot
// bypass the limit. Unauthenticated requests use Gin's trusted-proxy-aware IP.
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
		key := rateLimitKey(c)
		now := time.Now()

		mu.Lock()
		if now.Sub(lastCleanup) >= window {
			for client, candidate := range entries {
				if now.Sub(candidate.windowStart) >= window {
					delete(entries, client)
				}
			}
			lastCleanup = now
		}

		entry, exists := entries[key]
		if !exists || now.Sub(entry.windowStart) >= window {
			if !exists && len(entries) >= maxRateLimitEntries {
				mu.Unlock()
				c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(window)))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "too many requests"})
				return
			}
			entry = rateLimitEntry{windowStart: now}
		}
		entry.count++
		entries[key] = entry
		mu.Unlock()

		if entry.count > limit {
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(window)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "too many requests"})
			return
		}

		c.Next()
	}
}

func rateLimitKey(c *gin.Context) string {
	if userID, ok := c.Get("userId"); ok {
		if id, ok := userID.(uint); ok && id != 0 {
			return "user:" + strconv.FormatUint(uint64(id), 10)
		}
	}
	return "ip:" + c.ClientIP()
}

func retryAfterSeconds(window time.Duration) int {
	seconds := int(window.Seconds())
	if seconds < 1 {
		return 1
	}
	return seconds
}
