package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type rightsSyncRun struct {
	ID            string
	Provider      string
	Mode          string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Status        string
	RecordCount   int
	Error         string
	Source        string
}

func ensureMusicRightsSyncRunsSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS music_rights_sync_runs (
 id TEXT PRIMARY KEY,
 provider TEXT NOT NULL,
 mode TEXT NOT NULL,
 started_at TIMESTAMPTZ NOT NULL,
 completed_at TIMESTAMPTZ,
 status TEXT NOT NULL,
 record_count INTEGER NOT NULL DEFAULT 0,
 error TEXT,
 source TEXT
);
CREATE INDEX IF NOT EXISTS idx_music_rights_sync_runs_provider_started
 ON music_rights_sync_runs(provider, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_music_rights_sync_runs_status_started
 ON music_rights_sync_runs(status, started_at DESC);`)
	return err
}

func beginRightsSyncRun(ctx context.Context, pool *pgxpool.Pool, provider, mode, source string) (string, error) {
	id := newRightsSnapshotID()
	provider = strings.ToLower(strings.TrimSpace(provider))
	mode = strings.ToLower(strings.TrimSpace(mode))
	if provider == "" || mode == "" {
		return "", errors.New("rights sync run requires provider and mode")
	}
	_, err := pool.Exec(ctx, `
INSERT INTO music_rights_sync_runs(id,provider,mode,started_at,status,source)
VALUES($1,$2,$3,$4,'running',$5)`, id, provider, mode, time.Now().UTC(), strings.TrimSpace(source))
	if err != nil {
		return "", err
	}
	return id, nil
}

func finishRightsSyncRun(ctx context.Context, pool *pgxpool.Pool, id string, status string, count int, syncErr error) {
	if pool == nil || strings.TrimSpace(id) == "" {
		return
	}
	errText := ""
	if syncErr != nil {
		errText = syncErr.Error()
		if len(errText) > 1000 {
			errText = errText[:1000]
		}
	}
	_, _ = pool.Exec(ctx, `
UPDATE music_rights_sync_runs
SET completed_at=$2,status=$3,record_count=$4,error=NULLIF($5,''),source=COALESCE(source,'provider-sync')
WHERE id=$1`, id, time.Now().UTC(), strings.ToLower(strings.TrimSpace(status)), count, errText)
}
