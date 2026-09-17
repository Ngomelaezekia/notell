package main

import (
	"context"
	"errors"
	"strings"
	"time"
)

type MusicProvider interface {
	Name() string
	Ready() bool
	Search(context.Context, string, string, int, int) (SearchResult, error)
	GetTrack(context.Context, string, string) (Track, error)
	PlaybackURL(context.Context, Track, string, string, float64, float64) (string, time.Time, error)
}

type SearchResult struct { Tracks []Track; HasMore bool }

var errPlaybackUnavailable = errors.New("licensed playback is not available from the active music provider")

type internalProvider struct{}
func (internalProvider) Name() string { return "internal" }
func (internalProvider) Ready() bool { return true }
func (internalProvider) Search(_ context.Context, q, _ string, limit, offset int) (SearchResult, error) {
	needle := strings.ToLower(strings.TrimSpace(q)); matched := make([]Track, 0, len(catalog))
	for _, item := range catalog { text:=strings.ToLower(item.Title+" "+item.Artist+" "+item.Album); if needle==""||strings.Contains(text,needle){matched=append(matched,item)} }
	if limit<=0 {limit=20}; if offset<0 {offset=0}; if offset>=len(matched){return SearchResult{Tracks:[]Track{},HasMore:false},nil}
	end:=offset+limit; if end>len(matched){end=len(matched)}
	return SearchResult{Tracks:matched[offset:end],HasMore:end<len(matched)},nil
}
func (internalProvider) GetTrack(_ context.Context,id,_ string)(Track,error){for _,item:=range catalog{if item.ID==id{return item,nil}};return Track{},errors.New("track not found")}
func (internalProvider) PlaybackURL(_ context.Context, _ Track, _ string, _ string, _ float64, _ float64) (string, time.Time, error) {
	return "", time.Time{}, errPlaybackUnavailable
}
