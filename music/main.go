package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Track struct {
	ID              string   `json:"id"`
	Provider        string   `json:"provider"`
	ProviderTrackID string   `json:"providerTrackId,omitempty"`
	Title           string   `json:"title"`
	Artist          string   `json:"artist"`
	Album           string   `json:"album,omitempty"`
	ArtworkURL      string   `json:"artworkUrl,omitempty"`
	PreviewURL      string   `json:"previewUrl,omitempty"`
	DurationSec     float64  `json:"durationSec"`
	CanUseInPost    bool     `json:"canUseInPost"`
	Rights          Rights   `json:"rights"`
}

type Rights struct {
	Licensed       bool     `json:"licensed"`
	UGCUse         bool     `json:"ugcUse"`
	Streaming      bool     `json:"streaming"`
	Territories    []string `json:"territories,omitempty"`
	Attribution    bool     `json:"attributionRequired"`
	ProviderStatus string   `json:"providerStatus,omitempty"`
}

type SearchResponse struct {
	Query   string  `json:"query"`
	Tracks  []Track `json:"tracks"`
	HasMore bool    `json:"hasMore"`
}

type HealthResponse struct {
	Service  string `json:"service"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Time     string `json:"time"`
}

var catalog = []Track{
	{
		ID: "internal-demo-001", Provider: "internal", ProviderTrackID: "internal-demo-001",
		Title: "Notell Demo Sound", Artist: "Notell Library", Album: "Demo", DurationSec: 30,
		CanUseInPost: true,
		Rights: Rights{Licensed: true, UGCUse: true, Streaming: true, Territories: []string{"*"}, Attribution: false, ProviderStatus: "demo"},
	},
}

func configuredProvider() string {
	provider := strings.TrimSpace(strings.ToLower(os.Getenv("MUSIC_PROVIDER")))
	if provider == "" {
		return "internal"
	}
	return provider
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/api/music/health", health)
	mux.HandleFunc("/api/music/search", search)
	mux.HandleFunc("/api/music/trending", trending)
	mux.HandleFunc("/api/music/tracks/", track)
	mux.HandleFunc("/api/music/tracks/segment", segment)

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(withSecurityHeaders(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("music service listening on %s provider=%s", server.Addr, configuredProvider())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{
		Service: "music", Status: "ok", Provider: configuredProvider(), Time: time.Now().UTC().Format(time.RFC3339),
	})
}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, SearchResponse{Query: "", Tracks: catalog, HasMore: false})
		return
	}

	needle := strings.ToLower(q)
	results := make([]Track, 0, len(catalog))
	for _, item := range catalog {
		text := strings.ToLower(item.Title + " " + item.Artist + " " + item.Album)
		if strings.Contains(text, needle) {
			results = append(results, item)
		}
	}

	writeJSON(w, http.StatusOK, SearchResponse{Query: q, Tracks: results, HasMore: false})
}

func trending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, SearchResponse{Query: "", Tracks: catalog, HasMore: false})
}

func track(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/music/tracks/")
	id = strings.TrimSpace(id)
	for _, item := range catalog {
		if item.ID == id {
			writeJSON(w, http.StatusOK, item)
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"error": "track not found"})
}

func segment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	trackID := strings.TrimSpace(r.URL.Query().Get("trackId"))
	if trackID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "trackId is required"})
		return
	}

	var selected *Track
	for i := range catalog {
		if catalog[i].ID == trackID {
			selected = &catalog[i]
			break
		}
	}
	if selected == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "track not found"})
		return
	}

	start, err := queryFloat(r, "startSec", 0)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid startSec"})
		return
	}
	end, err := queryFloat(r, "endSec", selected.DurationSec)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endSec"})
		return
	}
	maxSegment := 60.0
	if configured := os.Getenv("MUSIC_MAX_SEGMENT_SECONDS"); configured != "" {
		if value, parseErr := strconv.ParseFloat(configured, 64); parseErr == nil && value > 0 {
			maxSegment = value
		}
	}
	if err := validateSegment(start, end, selected.DurationSec, maxSegment); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"trackId": trackID, "startSec": start, "endSec": end, "durationSec": selected.DurationSec,
		"provider": selected.Provider, "canUseInPost": selected.CanUseInPost,
	})
}

func queryFloat(r *http.Request, name string, fallback float64) (float64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(value, 64)
}

func validateSegment(start, end, duration, maxLength float64) error {
	if start < 0 || end <= start {
		return errInvalidSegment
	}
	if duration <= 0 || end > duration {
		return errSegmentOutsideTrack
	}
	if end-start > maxLength {
		return errSegmentTooLong
	}
	return nil
}

var (
	errInvalidSegment    = simpleError("music segment must satisfy 0 <= startSec < endSec")
	errSegmentOutsideTrack = simpleError("music segment exceeds track duration")
	errSegmentTooLong    = simpleError("music segment exceeds the configured maximum length")
)

type simpleError string
func (e simpleError) Error() string { return string(e) }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("json response error: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("FRONTEND_URL")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
