package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Track struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album,omitempty"`
	ArtworkURL  string  `json:"artworkUrl,omitempty"`
	PreviewURL  string  `json:"previewUrl,omitempty"`
	DurationSec float64 `json:"durationSec"`
	CanUseInPost bool   `json:"canUseInPost"`
}

type SearchResponse struct {
	Query   string  `json:"query"`
	Tracks  []Track `json:"tracks"`
	HasMore bool    `json:"hasMore"`
}

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

var catalog = []Track{
	{
		ID: "internal-demo-001", Provider: "internal", Title: "Notell Demo Sound",
		Artist: "Notell Library", Album: "Demo", DurationSec: 30,
		CanUseInPost: true,
	},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/api/music/health", health)
	mux.HandleFunc("/api/music/search", search)
	mux.HandleFunc("/api/music/trending", trending)
	mux.HandleFunc("/api/music/tracks/", track)

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

	log.Printf("music service listening on %s", server.Addr)
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
		Service: "music",
		Status:  "ok",
		Time:    time.Now().UTC().Format(time.RFC3339),
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
