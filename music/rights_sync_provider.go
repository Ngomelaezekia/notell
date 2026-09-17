package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type providerRightsSynchronizer interface {
	SyncRights(context.Context, func(rightsSyncRecord) error) (int, error)
}

func syncConfiguredProviderRights(ctx context.Context) (int, error) {
	provider := providers.current()
	syncer, ok := provider.(providerRightsSynchronizer)
	if !ok {
		return 0, fmt.Errorf("provider %s does not expose an authoritative rights feed", provider.Name())
	}
	store, err := currentMusicRightsStore()
	if err != nil {
		return 0, err
	}
	postgres, ok := store.(*postgresMusicRightsStore)
	if !ok || postgres == nil {
		return 0, errors.New("postgres rights store is required for provider synchronization")
	}
	count, err := syncer.SyncRights(ctx, func(record rightsSyncRecord) error {
		return upsertRightsRecord(ctx, postgres, record)
	})
	if err != nil {
		return count, err
	}
	postgres.mu.Lock()
	postgres.cache = make(map[string]cachedRights)
	postgres.mu.Unlock()
	return count, nil
}

func upsertRightsRecord(ctx context.Context, store *postgresMusicRightsStore, record rightsSyncRecord) error {
	record.Provider = strings.ToLower(strings.TrimSpace(record.Provider))
	record.ProviderTrackID = strings.TrimSpace(record.ProviderTrackID)
	record.Territory = strings.ToUpper(strings.TrimSpace(record.Territory))
	if record.Provider == "" || record.ProviderTrackID == "" || record.Territory == "" {
		return errors.New("incomplete provider rights record")
	}
	if record.Status == "" {
		record.Status = "active"
	}
	if record.Source == "" {
		record.Source = "provider"
	}
	_, err := store.pool.Exec(ctx, `
INSERT INTO music_rights_entitlements(provider,provider_track_id,territory,licensed,ugc_use,streaming,can_use_in_post,attribution_required,status,source,valid_from,valid_until,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now())
ON CONFLICT(provider,provider_track_id,territory) DO UPDATE SET
licensed=EXCLUDED.licensed,ugc_use=EXCLUDED.ugc_use,streaming=EXCLUDED.streaming,can_use_in_post=EXCLUDED.can_use_in_post,
attribution_required=EXCLUDED.attribution_required,status=EXCLUDED.status,source=EXCLUDED.source,valid_from=EXCLUDED.valid_from,valid_until=EXCLUDED.valid_until,updated_at=now()`,
		record.Provider, record.ProviderTrackID, record.Territory, record.Licensed, record.UGCUse, record.Streaming, record.CanUseInPost, record.AttributionRequired, record.Status, record.Source, record.ValidFrom, record.ValidUntil)
	return err
}

func startProviderRightsSync() {
	minutes := positiveEnvInt("MUSIC_RIGHTS_SYNC_INTERVAL_MINUTES", 360)
	go func() {
		run := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			count, err := syncConfiguredProviderRights(ctx)
			if err != nil {
				logProviderSyncError(err)
				return
			}
			logProviderSyncSuccess(count)
		}
		run()
		ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}

func logProviderSyncError(err error) {
	log.Printf("music provider rights sync failed: %v", err)
}

func logProviderSyncSuccess(count int) {
	log.Printf("music provider rights sync updated %d records", count)
}
