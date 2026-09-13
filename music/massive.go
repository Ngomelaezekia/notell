package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type massiveProvider struct {
	client *http.Client
	baseURL, consumerKey, consumerSecret, token, tokenSecret, country, usageTypes string
}

func newMassiveProvider() *massiveProvider {
	return &massiveProvider{client:&http.Client{Timeout:10*time.Second}, baseURL:envOr("MUSIC_MASSIVE_BASE_URL","https://api.7digital.com/1.2"), consumerKey:os.Getenv("MUSIC_MASSIVE_CONSUMER_KEY"), consumerSecret:os.Getenv("MUSIC_MASSIVE_CONSUMER_SECRET"), token:os.Getenv("MUSIC_MASSIVE_ACCESS_TOKEN"), tokenSecret:os.Getenv("MUSIC_MASSIVE_ACCESS_TOKEN_SECRET"), country:strings.ToUpper(envOr("MUSIC_COUNTRY","TZ")), usageTypes:envOr("MUSIC_MASSIVE_USAGE_TYPES","adsupportedstreaming")}
}
func (p *massiveProvider) Name() string { return "massivemusic" }
func (p *massiveProvider) ready() bool { return p.consumerKey!="" && p.consumerSecret!="" && p.token!="" && p.tokenSecret!="" }

func (p *massiveProvider) Search(ctx context.Context, q, country string, limit, offset int) (SearchResult,error) {
	if !p.ready() { return SearchResult{}, fmt.Errorf("massivemusic provider credentials are not configured") }
	if country=="" { country=p.country }
	if limit<=0 { limit=20 }
	page:=offset/limit+1
	params:=url.Values{"q":{q},"country":{country},"pagesize":{strconv.Itoa(limit)},"page":{strconv.Itoa(page)},"usageTypes":{p.usageTypes}}
	body,err:=p.get(ctx,"/track/search",params); if err!=nil{return SearchResult{},err}
	var raw struct { EstimatedTotalItems int `json:"estimatedTotalItems"`; Results []struct { ID int64 `json:"id"`; Title string `json:"title"`; Duration float64 `json:"duration"`; Release struct{Title string `json:"title"`} `json:"release"`; Artist struct{Name string `json:"name"`} `json:"artist"` } `json:"results"` }
	if err=json.Unmarshal(body,&raw); err!=nil{return SearchResult{},fmt.Errorf("decode MassiveMusic search: %w",err)}
	tracks:=make([]Track,0,len(raw.Results)); for _,r:=range raw.Results { tracks=append(tracks,Track{ID:"massivemusic-"+strconv.FormatInt(r.ID,10),Provider:p.Name(),ProviderTrackID:strconv.FormatInt(r.ID,10),Title:r.Title,Artist:r.Artist.Name,Album:r.Release.Title,DurationSec:r.Duration,CanUseInPost:false,Rights:Rights{ProviderStatus:"requires-license-clearance"}}) }
	return SearchResult{Tracks:tracks,HasMore:page*limit<raw.EstimatedTotalItems},nil
}

func (p *massiveProvider) GetTrack(ctx context.Context,id,country string)(Track,error){
	providerID:=strings.TrimPrefix(id,"massivemusic-")
	if country=="" {country=p.country}
	body,err:=p.get(ctx,"/track/details",url.Values{"trackId":{providerID},"country":{country}});if err!=nil{return Track{},err}
	var raw struct { Track struct { ID int64 `json:"id"`; Title string `json:"title"`; Duration float64 `json:"duration"`; Artist struct{Name string `json:"name"`} `json:"artist"`; Release struct{Title string `json:"title"`} `json:"release"` } `json:"track"` }
	if err=json.Unmarshal(body,&raw);err!=nil{return Track{},fmt.Errorf("decode MassiveMusic track: %w",err)}
	return Track{ID:"massivemusic-"+strconv.FormatInt(raw.Track.ID,10),Provider:p.Name(),ProviderTrackID:providerID,Title:raw.Track.Title,Artist:raw.Track.Artist.Name,Album:raw.Track.Release.Title,DurationSec:raw.Track.Duration,Rights:Rights{ProviderStatus:"requires-license-clearance"}},nil
}

func (p *massiveProvider) get(ctx context.Context,path string,params url.Values)([]byte,error){
	u,err:=url.Parse(strings.TrimRight(p.baseURL,"/")+path);if err!=nil{return nil,err};q:=u.Query();for k,v:=range params{if len(v)>0{q.Set(k,v[0])}}
	nonce:=strconv.FormatInt(time.Now().UnixNano(),10);timestamp:=strconv.FormatInt(time.Now().Unix(),10);q.Set("oauth_consumer_key",p.consumerKey);q.Set("oauth_nonce",nonce);q.Set("oauth_signature_method","HMAC-SHA1");q.Set("oauth_timestamp",timestamp);q.Set("oauth_version","1.0");q.Set("oauth_token",p.token)
	base:=u.Scheme+"://"+u.Host+u.Path;q.Set("oauth_signature",oauthSignature("GET",base,q,p.consumerSecret,p.tokenSecret));u.RawQuery=q.Encode()
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,u.String(),nil);if err!=nil{return nil,err};resp,err:=p.client.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();data,_:=io.ReadAll(io.LimitReader(resp.Body,2<<20));if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("MassiveMusic API status %d: %s",resp.StatusCode,strings.TrimSpace(string(data)))};return data,nil
}
func oauthSignature(method,base string,params url.Values,consumerSecret,tokenSecret string)string{keys:=make([]string,0,len(params));for k:=range params{keys=append(keys,k)};sort.Strings(keys);pairs:=make([]string,0,len(keys));for _,k:=range keys{for _,v:=range params[k]{pairs=append(pairs,url.QueryEscape(k)+"="+url.QueryEscape(v))}};normalized:=strings.Join(pairs,"&");sign:=url.QueryEscape(strings.ToUpper(method))+"&"+url.QueryEscape(base)+"&"+url.QueryEscape(normalized);mac:=hmac.New(sha1.New,[]byte(url.QueryEscape(consumerSecret)+"&"+url.QueryEscape(tokenSecret)));_,_=mac.Write([]byte(sign));return base64.StdEncoding.EncodeToString(mac.Sum(nil))}
func envOr(k,d string)string{if v:=strings.TrimSpace(os.Getenv(k));v!=""{return v};return d}
