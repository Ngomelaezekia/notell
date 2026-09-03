package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(2, time.Minute))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d returned %d, want %d", i+1, rec.Code, http.StatusNoContent)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request returned %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After = %q, want %q", got, "60")
	}
}

func TestRateLimitSeparatesClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(1, time.Minute))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "192.0.2.1:1234"
	firstRecorder := httptest.NewRecorder()
	r.ServeHTTP(firstRecorder, first)
	if firstRecorder.Code != http.StatusNoContent {
		t.Fatalf("first client returned %d, want %d", firstRecorder.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "192.0.2.2:1234"
	secondRecorder := httptest.NewRecorder()
	r.ServeHTTP(secondRecorder, second)
	if secondRecorder.Code != http.StatusNoContent {
		t.Fatalf("second client returned %d, want %d", secondRecorder.Code, http.StatusNoContent)
	}
}
