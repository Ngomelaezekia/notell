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

type providerRegistry struct { mu sync.RWMutex; providers map[string]MusicProvider }
func newProviderRegistry()*providerRegistry{r:=&providerRegistry{providers:map[string]MusicProvider{}};r.providers["internal"]=internalProvider{};r.providers["massivemusic"]=newMassiveProvider();r.providers["7digital"]=r.providers["massivemusic"];return r}
func(r *providerRegistry)current()MusicProvider{name:=configuredProvider();r.mu.RLock();p,ok:=r.providers[name];r.mu.RUnlock();if !ok{return unavailableProvider{name:name}};return p}
