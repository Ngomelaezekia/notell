package main

import (
	"context"
	"errors"
	"strings"
)

type MusicProvider interface {
	Name() string
	Search(context.Context, string, string, int, int) (SearchResult, error)
	GetTrack(context.Context, string, string) (Track, error)
}

type SearchResult struct {
	Tracks  []Track
	HasMore bool
}

type internalProvider struct{}

func (internalProvider) Name() string { return "internal" }

func (internalProvider) Search(_ context.Context, q, _ string, limit, offset int) (SearchResult, error) {
	needle := strings.ToLower(strings.TrimSpace(q))
	matched := make([]Track, 0, len(catalog))
	for _, item := range catalog {
		text := strings.ToLower(item.Title + " " + item.Artist + " " + item.Album)
		if needle == "" || strings.Contains(text, needle) {
			matched = append(matched, item)
		}
	}
	if limit <= 0 { limit = 20 }
	if offset < 0 { offset = 0 }
	if offset >= len(matched) { return SearchResult{Tracks: []Track{}}, nil }
	end := offset + limit
	if end > len(matched) { end = len(matched) }
	return SearchResult{Tracks: matched[offset:end], HasMore: end < len(matched)}, nil
}

func (internalProvider) GetTrack(_ context.Context, id, _ string) (Track, error) {
	for _, item := range catalog {
		if item.ID == id { return item, nil }
	}
	return Track{}, errors.New("track not found")
}
