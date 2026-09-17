package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type musicRightsStore interface { Resolve(context.Context,string,string,string)(Rights,bool,error); Close() }
type cachedRights struct { rights Rights; found bool; expiresAt time.Time }
type postgresMusicRightsStore struct { pool *pgxpool.Pool; mu sync.RWMutex; cache map[string]cachedRights; ttl time.Duration }

func newMusicRightsStore(ctx context.Context)(musicRightsStore,error){
	dsn:=strings.TrimSpace(envOr("MUSIC_DATABASE_URL","")); if dsn==""{dsn=strings.TrimSpace(envOr("DATABASE_URL",""))}; if dsn==""{return nil,nil}; if ctx==nil{ctx=context.Background()}
	cfg,err:=pgxpool.ParseConfig(dsn);if err!=nil{return nil,err};cfg.MaxConns=10;cfg.MinConns=1;cfg.MaxConnIdleTime=5*time.Minute
	pool,err:=pgxpool.NewWithConfig(ctx,cfg);if err!=nil{return nil,err};pingCtx,cancel:=context.WithTimeout(ctx,5*time.Second);defer cancel();if err:=pool.Ping(pingCtx);err!=nil{pool.Close();return nil,err};if err:=ensureMusicRightsSchema(ctx,pool);err!=nil{pool.Close();return nil,err};if err:=ensureMusicRightsSyncRunsSchema(ctx,pool);err!=nil{pool.Close();return nil,err}
	ttlMinutes:=positiveEnvInt("MUSIC_RIGHTS_CACHE_MINUTES",5);return &postgresMusicRightsStore{pool:pool,cache:make(map[string]cachedRights),ttl:time.Duration(ttlMinutes)*time.Minute},nil
}
func ensureMusicRightsSchema(ctx context.Context,pool *pgxpool.Pool)error{
	if _,err:=pool.Exec(ctx,`CREATE TABLE IF NOT EXISTS music_rights_entitlements (provider TEXT NOT NULL,provider_track_id TEXT NOT NULL,territory TEXT NOT NULL,licensed BOOLEAN NOT NULL DEFAULT FALSE,ugc_use BOOLEAN NOT NULL DEFAULT FALSE,streaming BOOLEAN NOT NULL DEFAULT FALSE,can_use_in_post BOOLEAN NOT NULL DEFAULT FALSE,attribution_required BOOLEAN NOT NULL DEFAULT FALSE,status TEXT NOT NULL DEFAULT 'pending',source TEXT NOT NULL DEFAULT 'manual',valid_from TIMESTAMPTZ,valid_until TIMESTAMPTZ,updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),last_sync_id TEXT,last_seen_at TIMESTAMPTZ,PRIMARY KEY(provider,provider_track_id,territory)); ALTER TABLE music_rights_entitlements ADD COLUMN IF NOT EXISTS last_sync_id TEXT; ALTER TABLE music_rights_entitlements ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ; CREATE INDEX IF NOT EXISTS idx_music_rights_lookup ON music_rights_entitlements(provider,provider_track_id,territory); CREATE INDEX IF NOT EXISTS idx_music_rights_expiry ON music_rights_entitlements(valid_until); CREATE INDEX IF NOT EXISTS idx_music_rights_sync_seen ON music_rights_entitlements(provider,last_sync_id,last_seen_at);`);err!=nil{return err}
	for _,item:=range catalog{if item.Provider!="internal"||item.ProviderTrackID==""{continue};if _,err:=pool.Exec(ctx,`INSERT INTO music_rights_entitlements(provider,provider_track_id,territory,licensed,ugc_use,streaming,can_use_in_post,attribution_required,status,source) VALUES($1,$2,'*',$3,$4,$5,$6,$7,$8,'catalog') ON CONFLICT(provider,provider_track_id,territory) DO NOTHING`,item.Provider,item.ProviderTrackID,item.Rights.Licensed,item.Rights.UGCUse,item.Rights.Streaming,item.CanUseInPost,item.Rights.Attribution,item.Rights.ProviderStatus);err!=nil{return err}}
	return nil
}
func(s *postgresMusicRightsStore)Resolve(ctx context.Context,provider,providerTrackID,territory string)(Rights,bool,error){
	provider=strings.ToLower(strings.TrimSpace(provider));providerTrackID=strings.TrimSpace(providerTrackID);territory=strings.ToUpper(strings.TrimSpace(territory));key:=provider+"|"+providerTrackID+"|"+territory;now:=time.Now().UTC();s.mu.RLock();cached,ok:=s.cache[key];s.mu.RUnlock();if ok&&now.Before(cached.expiresAt){return cached.rights,cached.found,nil}
	var rights Rights;var canUseInPost bool;err:=s.pool.QueryRow(ctx,`SELECT licensed,ugc_use,streaming,can_use_in_post,attribution_required,status FROM music_rights_entitlements WHERE provider=$1 AND provider_track_id=$2 AND territory IN ($3,'*') AND (valid_from IS NULL OR valid_from <= $4) AND (valid_until IS NULL OR valid_until > $4) AND status NOT IN ('revoked','stale') ORDER BY CASE WHEN territory=$3 THEN 0 ELSE 1 END LIMIT 1`,provider,providerTrackID,territory,now).Scan(&rights.Licensed,&rights.UGCUse,&rights.Streaming,&canUseInPost,&rights.Attribution,&rights.ProviderStatus)
	if err!=nil&&!errors.Is(err,pgx.ErrNoRows){return Rights{},false,err};found:=err==nil;if found{rights.Territories=[]string{territory};if !canUseInPost{rights.Licensed=false}}
	s.mu.Lock();s.cache[key]=cachedRights{rights:rights,found:found,expiresAt:now.Add(s.ttl)};s.mu.Unlock();return rights,found,nil
}
func(s *postgresMusicRightsStore)Close(){s.pool.Close()}
var(musicRightsMu sync.Mutex;musicRights musicRightsStore;musicRightsError error)
func currentMusicRightsStore()(musicRightsStore,error){
	musicRightsMu.Lock();defer musicRightsMu.Unlock()
	if musicRights!=nil{return musicRights,nil}
	store,err:=newMusicRightsStore(context.Background())
	if err!=nil{musicRightsError=err;return nil,err}
	musicRights=store;musicRightsError=nil;return musicRights,nil
}
func applyAuthoritativeRights(ctx context.Context,item Track,country string)(Track,error){store,err:=currentMusicRightsStore();if err!=nil{return item,err};if store==nil{if item.Provider=="internal"{return item,nil};item.CanUseInPost=false;item.Rights.Licensed=false;item.Rights.UGCUse=false;item.Rights.Streaming=false;item.Rights.Territories=nil;item.Rights.ProviderStatus="rights-store-unavailable";return item,nil};rights,found,err:=store.Resolve(ctx,item.Provider,item.ProviderTrackID,country);if err!=nil{return item,err};if !found{item.CanUseInPost=false;item.Rights.Licensed=false;item.Rights.UGCUse=false;item.Rights.Streaming=false;item.Rights.Territories=nil;item.Rights.ProviderStatus="not-cleared";return item,nil};item.Rights=rights;item.CanUseInPost=rights.Licensed&&rights.UGCUse&&rights.Streaming;return item,nil}
