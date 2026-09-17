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

type musicRightsStore interface {
	Resolve(context.Context, string, string, string) (Rights, bool, error)
	Close()
}

type cachedRights struct {
	rights    Rights
	found     bool
	expiresAt time.Time
}

type postgresMusicRightsStore struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	cache map[string]cachedRights
	ttl   time.Duration
}

func newMusicRightsStore(ctx context.Context) (musicRightsStore, error) {
	dsn := strings.TrimSpace(envOr("MUSIC_DATABASE_URL", ""))
	if dsn == "" {
		dsn = strings.TrimSpace(envOr("DATABASE_URL", ""))
	}
	if dsn == "" {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := ensureMusicRightsSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	ttlMinutes := positiveEnvInt("MUSIC_RIGHTS_CACHE_MINUTES", 5)
	return &postgresMusicRightsStore{pool: pool, cache: make(map[string]cachedRights), ttl: time.Duration(ttlMinutes) * time.Minute}, nil
}

func ensureMusicRightsSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS music_rights_entitlements (
	provider TEXT NOT NULL,
	provider_track_id TEXT NOT NULL,
	territory TEXT NOT NULL,
	licensed BOOLEAN NOT NULL DEFAULT FALSE,
	ugc_use BOOLEAN NOT NULL DEFAULT FALSE,
	streaming BOOLEAN NOT NULL DEFAULT FALSE,
	can_use_in_post BOOLEAN NOT NULL DEFAULT FALSE,
	attribution_required BOOLEAN NOT NULL DEFAULT FALSE,
	status TEXT NOT NULL DEFAULT 'pending',
	source TEXT NOT NULL DEFAULT 'manual',
	valid_from TIMESTAMPTZ,
	valid_until TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY(provider, provider_track_id, territory)
);
CREATE INDEX IF NOT EXISTS idx_music_rights_lookup ON music_rights_entitlements(provider, provider_track_id, territory);
CREATE INDEX IF NOT EXISTS idx_music_rights_expiry ON music_rights_entitlements(valid_until);
`)
	return err
}

func (s *postgresMusicRightsStore) Resolve(ctx context.Context, provider, providerTrackID, territory string) (Rights, bool, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	providerTrackID = strings.TrimSpace(providerTrackID)
	territory = strings.ToUpper(strings.TrimSpace(territory))
	key := provider + "|" + providerTrackID + "|" + territory
	now := time.Now().UTC()
	s.mu.RLock()
	cached, ok := s.cache[key]
	s.mu.RUnlock()
	if ok && now.Before(cached.expiresAt) {
		return cached.rights, cached.found, nil
	}

	var rights Rights
	err := s.pool.QueryRow(ctx, `
SELECT licensed, ugc_use, streaming, can_use_in_post, attribution_required, status
FROM music_rights_entitlements
WHERE provider=$1 AND provider_track_id=$2 AND territory IN ($3, '*')
  AND (valid_from IS NULL OR valid_from <= $4)
  AND (valid_until IS NULL OR valid_until > $4)
ORDER BY CASE WHEN territory=$3 THEN 0 ELSE 1 END
LIMIT 1`, provider, providerTrackID, territory, now).Scan(
		&rights.Licensed, &rights.UGCUse, &rights.Streaming, &rights.ProviderStatus,
		&rights.Attribution,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Rights{}, false, err
	}
	found := err == nil
	if found {
		rights.Territories = []string{territory}
	}
	s.mu.Lock()
	s.cache[key] = cachedRights{rights: rights, found: found, expiresAt: now.Add(s.ttl)}
	s.mu.Unlock()
	return rights, found, nil
}

func (s *postgresMusicRightsStore) Close() { s.pool.Close() }

var (
	musicRightsOnce  sync.Once
	musicRights      musicRightsStore
	musicRightsError error
)

func currentMusicRightsStore() (musicRightsStore, error) {
	musicRightsOnce.Do(func() {
		musicRights, musicRightsError = newMusicRightsStore(context.Background())
	})
	return musicRights, musicRightsError
}

func applyAuthoritativeRights(ctx context.Context, item Track, country string) (Track, error) {
	store, err := currentMusicRightsStore()
	if err != nil {
		return item, err
	}
	if store == nil {
		if item.Provider == "internal" {
			return item, nil
		}
		item.CanUseInPost = false
		item.Rights.Licensed = false
		item.Rights.UGCUse = false
		item.Rights.Streaming = false
		item.Rights.Territories = nil
		item.Rights.ProviderStatus = "rights-store-unavailable"
		return item, nil
	}
	rights, found, err := store.Resolve(ctx, item.Provider, item.ProviderTrackID, country)
	if err != nil {
		return item, err
	}
	if !found {
		item.CanUseInPost = false
		item.Rights.Licensed = false
		item.Rights.UGCUse = false
		item.Rights.Streaming = false
		item.Rights.Territories = nil
		item.Rights.ProviderStatus = "not-cleared"
		return item, nil
	}
	item.Rights = rights
	item.CanUseInPost = rights.Licensed && rights.UGCUse && rights.Streaming
	return item, nil
}
