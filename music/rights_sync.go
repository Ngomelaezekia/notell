package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type rightsSyncRecord struct {
	Provider             string     `json:"provider"`
	ProviderTrackID      string     `json:"providerTrackId"`
	Territory            string     `json:"territory"`
	Licensed             bool       `json:"licensed"`
	UGCUse               bool       `json:"ugcUse"`
	Streaming            bool       `json:"streaming"`
	CanUseInPost         bool       `json:"canUseInPost"`
	AttributionRequired  bool       `json:"attributionRequired"`
	Status               string     `json:"status"`
	Source               string     `json:"source"`
	ValidFrom            *time.Time `json:"validFrom,omitempty"`
	ValidUntil           *time.Time `json:"validUntil,omitempty"`
}

type rightsSyncPayload struct {
	Records []rightsSyncRecord `json:"records"`
}

func syncRightsFromJSON(ctx context.Context, body []byte) (int, error) {
	store, err := currentMusicRightsStore()
	if err != nil {
		return 0, err
	}
	postgres, ok := store.(*postgresMusicRightsStore)
	if !ok || postgres == nil {
		return 0, errors.New("postgres rights store is required for synchronization")
	}
	var payload rightsSyncPayload
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return 0, err
	}
	if len(payload.Records) == 0 || len(payload.Records) > 10000 {
		return 0, errors.New("rights sync must contain between 1 and 10000 records")
	}
	tx, err := postgres.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	for _, record := range payload.Records {
		record.Provider = strings.ToLower(strings.TrimSpace(record.Provider))
		record.ProviderTrackID = strings.TrimSpace(record.ProviderTrackID)
		record.Territory = strings.ToUpper(strings.TrimSpace(record.Territory))
		record.Status = strings.TrimSpace(record.Status)
		record.Source = strings.TrimSpace(record.Source)
		if record.Provider == "" || record.ProviderTrackID == "" || record.Territory == "" {
			return 0, errors.New("provider, providerTrackId and territory are required")
		}
		if record.Status == "" { record.Status = "active" }
		if record.Source == "" { record.Source = "sync" }
		if _, err := tx.Exec(ctx, `
INSERT INTO music_rights_entitlements(provider,provider_track_id,territory,licensed,ugc_use,streaming,can_use_in_post,attribution_required,status,source,valid_from,valid_until,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now())
ON CONFLICT(provider,provider_track_id,territory) DO UPDATE SET
licensed=EXCLUDED.licensed,ugc_use=EXCLUDED.ugc_use,streaming=EXCLUDED.streaming,can_use_in_post=EXCLUDED.can_use_in_post,
attribution_required=EXCLUDED.attribution_required,status=EXCLUDED.status,source=EXCLUDED.source,valid_from=EXCLUDED.valid_from,
valid_until=EXCLUDED.valid_until,updated_at=now()`, record.Provider, record.ProviderTrackID, record.Territory, record.Licensed, record.UGCUse, record.Streaming, record.CanUseInPost, record.AttributionRequired, record.Status, record.Source, record.ValidFrom, record.ValidUntil); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil { return 0, err }
	postgres.mu.Lock()
	postgres.cache = make(map[string]cachedRights)
	postgres.mu.Unlock()
	return len(payload.Records), nil
}

func syncRightsHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	secret := strings.TrimSpace(os.Getenv("MUSIC_RIGHTS_SYNC_KEY"))
	if secret == "" || r.Header.Get("X-Music-Rights-Key") != secret { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	if r.ContentLength > 5<<20 { http.Error(w, "payload too large", http.StatusRequestEntityTooLarge); return }
	body := make([]byte, 0, 64<<10)
	limited := http.MaxBytesReader(w, r.Body, 5<<20)
	buf := make([]byte, 32<<10)
	for { n, err := limited.Read(buf); if n > 0 { body = append(body, buf[:n]...) }; if err != nil { if err.Error() == "EOF" { break }; http.Error(w, "invalid payload", http.StatusBadRequest); return } }
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second); defer cancel()
	count, err := syncRightsFromJSON(ctx, body)
	if err != nil { http.Error(w, "rights synchronization failed", http.StatusBadRequest); return }
	writeJSON(w, http.StatusAccepted, map[string]any{"status":"accepted","records":count})
}

var _ = pgx.ErrNoRows
