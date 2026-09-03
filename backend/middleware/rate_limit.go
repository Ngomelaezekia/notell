package middleware

import (
	"net/http"
	"net"
	"strconv"
	"strings"
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
// bypass the limit. Unauthenticated requests use the trusted client IP.
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
				evictOldest(entries)
			}
			entry = rateLimitEntry{windowStart: now}
		}
		entry.count++
		entries[key] = entry
		mu.Unlock()

		if entry.count > limit {
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(window, entry.windowStart, now)))
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
	return "ip:" + clientIP(c)
}

func clientIP(c *gin.Context) string {
	// Render routes public traffic through Cloudflare. Cloudflare documents
	// CF-Connecting-IP as the single, original client IP sent to the origin.
	// Prefer it over X-Forwarded-For, which can contain a client-supplied chain.
	if value := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); value != "" {
		if parsed := net.ParseIP(value); parsed != nil {
			return parsed.String()
		}
	}
	return c.ClientIP()
}

func evictOldest(entries map[string]rateLimitEntry) {
	var oldestKey string
	var oldest time.Time
	for key, entry := range entries {
		if oldestKey == "" || entry.windowStart.Before(oldest) {
			oldestKey = key
			oldest = entry.windowStart
		}
	}
	if oldestKey != "" {
		delete(entries, oldestKey)
	}
}

func retryAfterSeconds(window time.Duration, windowStart, now time.Time) int {
	remaining := window - now.Sub(windowStart)
	seconds := int(remaining / time.Second)
	if remaining%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		return 1
	}
	return seconds
}
