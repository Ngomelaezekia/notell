package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type rightsSyncRecord struct { Provider string `json:"provider"`; ProviderTrackID string `json:"providerTrackId"`; Territory string `json:"territory"`; Licensed bool `json:"licensed"`; UGCUse bool `json:"ugcUse"`; Streaming bool `json:"streaming"`; CanUseInPost bool `json:"canUseInPost"`; AttributionRequired bool `json:"attributionRequired"`; Status string `json:"status"`; Source string `json:"source"`; ValidFrom *time.Time `json:"validFrom,omitempty"`; ValidUntil *time.Time `json:"validUntil,omitempty"` }
type rightsSyncPayload struct { Records []rightsSyncRecord `json:"records"`; Full bool `json:"full,omitempty"` }
func newRightsSnapshotID() string { var b [16]byte; if _,err:=rand.Read(b[:]); err!=nil{return time.Now().UTC().Format("20060102T150405.000000000Z07:00")}; return hex.EncodeToString(b[:]) }
func syncRightsFromJSON(ctx context.Context, body []byte) (count int, syncErr error) {
	store,err:=currentMusicRightsStore(); if err!=nil{return 0,err}; postgres,ok:=store.(*postgresMusicRightsStore); if !ok||postgres==nil{return 0,errors.New("postgres rights store is required for synchronization")}
	var payload rightsSyncPayload; decoder:=json.NewDecoder(strings.NewReader(string(body))); decoder.DisallowUnknownFields(); if err:=decoder.Decode(&payload);err!=nil{return 0,err}; if len(payload.Records)==0||len(payload.Records)>10000{return 0,errors.New("rights sync must contain between 1 and 10000 records")}
	provider:=""; mode:="incremental"; if payload.Full { mode="full"; provider=strings.ToLower(strings.TrimSpace(payload.Records[0].Provider)); if provider==""{return 0,errors.New("full rights snapshot requires a provider")}; for _,record:=range payload.Records {if strings.ToLower(strings.TrimSpace(record.Provider))!=provider{return 0,errors.New("full rights snapshot must contain one provider")}} }
	runID,err:=beginRightsSyncRun(ctx,postgres.pool,providerOrFirstRecord(payload.Records,provider),mode,"provider-sync"); if err!=nil{return 0,err}; finishFailed:=true; defer func(){if finishFailed{finishRightsSyncRun(context.Background(),postgres.pool,runID,"failed",count,syncErr)}}()
	tx,err:=postgres.pool.Begin(ctx); if err!=nil{return 0,err}; defer tx.Rollback(ctx); if _,err:=tx.Exec(ctx,`ALTER TABLE music_rights_entitlements ADD COLUMN IF NOT EXISTS last_sync_id TEXT; ALTER TABLE music_rights_entitlements ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;`);err!=nil{return 0,err}
	snapshotID,now:=newRightsSnapshotID(),time.Now().UTC()
	for _,record:=range payload.Records { record.Provider=strings.ToLower(strings.TrimSpace(record.Provider)); record.ProviderTrackID=strings.TrimSpace(record.ProviderTrackID); record.Territory=strings.ToUpper(strings.TrimSpace(record.Territory)); record.Status=strings.TrimSpace(record.Status); record.Source=strings.TrimSpace(record.Source); if record.Provider==""||record.ProviderTrackID==""||record.Territory==""{return 0,errors.New("provider, providerTrackId and territory are required")}; if record.Status==""{record.Status="active"}; if record.Source==""{record.Source="sync"}; if payload.Full{record.Source="provider-sync"}; _,err=tx.Exec(ctx,`INSERT INTO music_rights_entitlements(provider,provider_track_id,territory,licensed,ugc_use,streaming,can_use_in_post,attribution_required,status,source,valid_from,valid_until,updated_at,last_sync_id,last_seen_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT(provider,provider_track_id,territory) DO UPDATE SET licensed=EXCLUDED.licensed,ugc_use=EXCLUDED.ugc_use,streaming=EXCLUDED.streaming,can_use_in_post=EXCLUDED.can_use_in_post,attribution_required=EXCLUDED.attribution_required,status=EXCLUDED.status,source=EXCLUDED.source,valid_from=EXCLUDED.valid_from,valid_until=EXCLUDED.valid_until,updated_at=EXCLUDED.updated_at,last_sync_id=EXCLUDED.last_sync_id,last_seen_at=EXCLUDED.last_seen_at`,record.Provider,record.ProviderTrackID,record.Territory,record.Licensed,record.UGCUse,record.Streaming,record.CanUseInPost,record.AttributionRequired,record.Status,record.Source,record.ValidFrom,record.ValidUntil,now,snapshotID,now); if err!=nil{return 0,err} }
	if payload.Full { _,err=tx.Exec(ctx,`UPDATE music_rights_entitlements SET licensed=FALSE,ugc_use=FALSE,streaming=FALSE,can_use_in_post=FALSE,status='stale',updated_at=$2 WHERE provider=$1 AND source='provider-sync' AND (last_sync_id IS NULL OR last_sync_id<>$3)`,provider,now,snapshotID); if err!=nil{return 0,err} }
	if err:=tx.Commit(ctx);err!=nil{return 0,err}; postgres.mu.Lock(); postgres.cache=make(map[string]cachedRights); postgres.mu.Unlock(); count=len(payload.Records); finishRightsSyncRun(context.Background(),postgres.pool,runID,"succeeded",count,nil); finishFailed=false; return count,nil
}
func providerOrFirstRecord(records []rightsSyncRecord, provider string) string { if strings.TrimSpace(provider)!="" { return strings.ToLower(strings.TrimSpace(provider)) }; if len(records)>0{return strings.ToLower(strings.TrimSpace(records[0].Provider))}; return "unknown" }
func syncRightsHTTP(w http.ResponseWriter,r *http.Request) { if r.Method!=http.MethodPost{http.Error(w,"method not allowed",http.StatusMethodNotAllowed);return}; secret,provided:=strings.TrimSpace(os.Getenv("MUSIC_RIGHTS_SYNC_KEY")),strings.TrimSpace(r.Header.Get("X-Music-Rights-Key")); if secret==""||subtle.ConstantTimeCompare([]byte(secret),[]byte(provided))!=1{http.Error(w,"unauthorized",http.StatusUnauthorized);return}; if r.ContentLength>5<<20{http.Error(w,"payload too large",http.StatusRequestEntityTooLarge);return}; body,err:=io.ReadAll(http.MaxBytesReader(w,r.Body,5<<20));if err!=nil{http.Error(w,"invalid payload",http.StatusBadRequest);return};ctx,cancel:=context.WithTimeout(r.Context(),20*time.Second);defer cancel();count,err:=syncRightsFromJSON(ctx,body);if err!=nil{http.Error(w,"rights synchronization failed",http.StatusBadRequest);return};writeJSON(w,http.StatusAccepted,map[string]any{"status":"accepted","records":count}) }
