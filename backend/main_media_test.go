package main

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"notell/services"

	"github.com/gin-gonic/gin"
)

type mediaTestStorage struct{}

func (mediaTestStorage) Put(context.Context, string, string, string) error { return nil }
func (mediaTestStorage) Delete(context.Context, string) error { return nil }
func (mediaTestStorage) PublicURL(string) string { return "" }
func (mediaTestStorage) Open(context.Context, string) (io.ReadCloser, string, int64, error) {
	return io.NopCloser(errorReader{}), "image/jpeg", 0, nil
}

type trackingMediaStorage struct {
	opened bool
}

func (s *trackingMediaStorage) Put(context.Context, string, string, string) error { return nil }
func (s *trackingMediaStorage) Delete(context.Context, string) error { return nil }
func (s *trackingMediaStorage) PublicURL(string) string { return "" }
func (s *trackingMediaStorage) Open(context.Context, string) (io.ReadCloser, string, int64, error) {
	s.opened = true
	return io.NopCloser(errorReader{}), "image/jpeg", 0, nil
}

type errorReader struct{}
func (errorReader) Read([]byte) (int, error) { return 0, errors.New("unexpected read") }

var _ services.MediaStorage = mediaTestStorage{}
var _ services.MediaStorage = (*trackingMediaStorage)(nil)

func TestServeMediaRejectsUnclaimedMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage := &trackingMediaStorage{}
	r := gin.New()
	called := false
	r.GET("/uploads/:filename", serveMedia(storage, func(context.Context, string) (bool, error) {
		called = true
		return false, nil
	}))

	req := httptest.NewRequest("GET", "/uploads/private.jpg", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != 404 {
		t.Fatalf("expected 404 for unclaimed media, got %d", resp.Code)
	}
	if !called {
		t.Fatal("expected claim check to run")
	}
	if storage.opened {
		t.Fatal("storage must not be opened for unclaimed media")
	}
}

func TestServeMediaAllowsClaimedMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/uploads/:filename", serveMedia(mediaTestStorage{}, func(context.Context, string) (bool, error) {
		return true, nil
	}))

	req := httptest.NewRequest("GET", "/uploads/public.jpg", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Fatalf("expected 200 for claimed media, got %d", resp.Code)
	}
	if got := resp.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected cache policy: %q", got)
	}
}

func TestServeMediaReturnsServerErrorWhenClaimCheckFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/uploads/:filename", serveMedia(mediaTestStorage{}, func(context.Context, string) (bool, error) {
		return false, errors.New("database unavailable")
	}))

	req := httptest.NewRequest("GET", "/uploads/public.jpg", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != 500 {
		t.Fatalf("expected 500 when claim check fails, got %d", resp.Code)
	}
}

func TestServeMediaRejectsPathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/uploads/*filename", serveMedia(mediaTestStorage{}, func(context.Context, string) (bool, error) {
		t.Fatal("claim check must not run for invalid filename")
		return false, nil
	}))

	req := httptest.NewRequest("GET", "/uploads/../private.jpg", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != 400 {
		t.Fatalf("expected 400 for path traversal, got %d", resp.Code)
	}
}
