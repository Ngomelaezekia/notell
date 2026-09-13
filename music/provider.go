package main

import (
	"context"
	"errors"
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
	needle := lower(q)
	matched := make([]Track, 0, len(catalog))
	for _, item := range catalog {
		if needle == "" || contains(lower(item.Title+" "+item.Artist+" "+item.Album), needle) {
			matched = append(matched, item)
		}
	}
	if offset >= len(matched) {
		return SearchResult{Tracks: []Track{}, HasMore: false}, nil
	}
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

func lower(v string) string { return stringsToLower(v) }
func contains(s, sub string) bool { return stringsContains(s, sub) }
