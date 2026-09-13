package main

import (
	"context"
	"log"
	"os"
	"strings"
)

// Until the main HTTP handlers are switched to a persistent catalog store,
// a licensed provider can bootstrap the in-memory catalog at service start.
// Credentials are never committed; without them the internal demo catalog remains active.
func init() {
	if strings.ToLower(strings.TrimSpace(os.Getenv("MUSIC_PROVIDER"))) != "massivemusic" && strings.ToLower(strings.TrimSpace(os.Getenv("MUSIC_PROVIDER"))) != "7digital" {
		return
	}
	provider := newMassiveProvider()
	if !provider.ready() {
		log.Printf("massivemusic selected but credentials are not configured; keeping internal catalog")
		return
	}
	result, err := provider.Search(context.Background(), "", provider.country, 50, 0)
	if err != nil {
		log.Printf("massivemusic catalog bootstrap failed; keeping internal catalog: %v", err)
		return
	}
	if len(result.Tracks) > 0 {
		catalog = result.Tracks
		log.Printf("massivemusic catalog bootstrapped: %d tracks", len(catalog))
	}
}
