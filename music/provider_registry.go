package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type unavailableProvider struct{ name string }
func (p unavailableProvider) Name() string { return p.name }
func (p unavailableProvider) Ready() bool { return false }
func (p unavailableProvider) Search(context.Context,string,string,int,int)(SearchResult,error){ return SearchResult{}, errors.New(p.name+" provider is not configured") }
func (p unavailableProvider) GetTrack(context.Context,string,string)(Track,error){ return Track{}, errors.New(p.name+" provider is not configured") }
func (p unavailableProvider) PlaybackURL(context.Context,Track,string,string,float64,float64)(string,time.Time,error){ return "",time.Time{},errors.New(p.name+" provider is not configured") }

type entitledProvider struct { provider MusicProvider }
func (p entitledProvider) Name() string { return p.provider.Name() }
func (p entitledProvider) Ready() bool { return p.provider.Ready() }
func (p entitledProvider) Search(ctx context.Context, q, country string, limit, offset int) (SearchResult, error) {
	result, err := p.provider.Search(ctx, q, country, limit, offset)
	if err != nil { return SearchResult{}, err }
	for i := range result.Tracks {
		item, err := applyAuthoritativeRights(ctx, result.Tracks[i], country)
		if err != nil { return SearchResult{}, err }
		result.Tracks[i] = item
	}
	return result, nil
}
func (p entitledProvider) GetTrack(ctx context.Context, id, country string) (Track, error) {
	item, err := p.provider.GetTrack(ctx, id, country)
	if err != nil { return Track{}, err }
	return applyAuthoritativeRights(ctx, item, country)
}
func (p entitledProvider) PlaybackURL(ctx context.Context, track Track, country, userID string, start, end float64) (string,time.Time,error) {
	return p.provider.PlaybackURL(ctx, track, country, userID, start, end)
}

type providerRegistry struct { mu sync.RWMutex; providers map[string]MusicProvider }
func newProviderRegistry()*providerRegistry{
	r:=&providerRegistry{providers:map[string]MusicProvider{}}
	r.providers["internal"]=internalProvider{}
	r.providers["massivemusic"]=newMassiveProvider()
	r.providers["7digital"]=r.providers["massivemusic"]
	return r
}
func(r *providerRegistry)current()MusicProvider{
	name:=configuredProvider()
	r.mu.RLock();p,ok:=r.providers[name];r.mu.RUnlock()
	if !ok{return unavailableProvider{name:name}}
	return entitledProvider{provider:p}
}
