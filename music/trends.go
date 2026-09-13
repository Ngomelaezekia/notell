package main

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type MusicEvent struct { TrackID string `json:"trackId"`; Type string `json:"type"`; Country string `json:"country,omitempty"`; City string `json:"city,omitempty"`; UserID string `json:"userId,omitempty"`; At time.Time `json:"at"` }
type TrendItem struct { TrackID string `json:"trackId"`; Score float64 `json:"score"`; Plays int64 `json:"plays"`; Uses int64 `json:"uses"`; Saves int64 `json:"saves"`; Shares int64 `json:"shares"`; Growth24h float64 `json:"growth24h"`; Country string `json:"country,omitempty"`; City string `json:"city,omitempty"` }
type trendStore struct { mu sync.RWMutex; events []MusicEvent }
var trends = &trendStore{}
var eventWeights = map[string]float64{"play":1,"complete":2,"replay":2,"use":8,"save":5,"share":6,"search":0.25}
func (s *trendStore) Add(e MusicEvent){if e.At.IsZero(){e.At=time.Now().UTC()};e.Type=strings.ToLower(strings.TrimSpace(e.Type));if _,ok:=eventWeights[e.Type];!ok{return};s.mu.Lock();defer s.mu.Unlock();s.events=append(s.events,e);if len(s.events)>100000{s.events=s.events[len(s.events)-100000:]}}
func (s *trendStore) Rank(country,city string,window time.Duration)[]TrendItem{now:=time.Now().UTC();cutoff:=now.Add(-window);previous:=cutoff.Add(-window);type aggregate struct{score,previous float64;plays,uses,saves,shares int64};totals:=map[string]*aggregate{};for _,e:=range s.snapshot(){if e.At.Before(previous)||e.At.After(now){continue};if country!=""&&!strings.EqualFold(e.Country,country){continue};if city!=""&&!strings.EqualFold(e.City,city){continue};a:=totals[e.TrackID];if a==nil{a=&aggregate{};totals[e.TrackID]=a};weight:=eventWeights[e.Type];if e.At.Before(cutoff){a.previous+=weight;continue};a.score+=weight;switch e.Type{case "play","complete","replay":a.plays++;case "use":a.uses++;case "save":a.saves++;case "share":a.shares++}};result:=make([]TrendItem,0,len(totals));for id,a:=range totals{growth:=0.0;if a.previous>0{growth=(a.score-a.previous)/a.previous*100};result=append(result,TrendItem{TrackID:id,Score:a.score,Plays:a.plays,Uses:a.uses,Saves:a.saves,Shares:a.shares,Growth24h:growth,Country:country,City:city})};sort.Slice(result,func(i,j int)bool{return result[i].Score>result[j].Score});if len(result)>50{result=result[:50]};return result}
func(s *trendStore)snapshot()[]MusicEvent{s.mu.RLock();defer s.mu.RUnlock();out:=make([]MusicEvent,len(s.events));copy(out,s.events);return out}
