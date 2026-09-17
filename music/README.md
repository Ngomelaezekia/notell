# Notell Music Service

Independent backend service for music discovery, catalog management, previews, rights checks, playback descriptors, post-track selection, and trend analytics.

## Trends storage

The trend API supports a shared PostgreSQL event store for multi-instance deployments. Set `MUSIC_TRENDS_STORE=postgres` and provide `MUSIC_DATABASE_URL` (or `DATABASE_URL`). The service creates the required event table and indexes on startup. Leave the store as `memory` for local development.

PostgreSQL is the durable shared source for first-party music events; high-volume rollups and retention can be added later without changing the trend API contract.

## Current API

- `GET /health`
- `GET /api/music/health`
- `GET /api/music/search?q=<query>`
- `GET /api/music/trending`
- `POST /api/music/events` with `X-Music-Event-Key`
- `GET /api/music/tracks/<id>`
- `GET /api/music/tracks/segment?trackId=<id>&startSec=<seconds>&endSec=<seconds>`

## Licensed-provider direction

The production catalog is designed for a licensed B2B music provider rather than scraping consumer music services. The primary provider target is MassiveMusic / 7digital. Provider integration must be enabled only after a commercial agreement explicitly covers Notell's intended use: UGC synchronization, streaming/playback to viewers, territory rights, attribution, usage reporting, takedowns, and publisher clearances.

## Rights model

Every production track should carry enough information for the service to decide whether it can be exposed to a user in a territory and whether it can be attached to a post. The service rejects tracks when required rights are not confirmed.

## Environment

- `PORT` — Render-provided HTTP port.
- `FRONTEND_URL` — allowed browser origin for CORS.
- `MUSIC_PROVIDER` — provider selector; defaults to `internal` until a licensed provider is configured.
- `MUSIC_MAX_SEGMENT_SECONDS` — maximum selectable segment length; defaults to `60`.
- `MUSIC_EVENT_INGEST_KEY` — required secret for event ingestion.
- `MUSIC_TRENDS_STORE` — `memory` or `postgres`; defaults to `memory`.
- `MUSIC_DATABASE_URL` — PostgreSQL connection string when using the PostgreSQL trend store. Falls back to `DATABASE_URL`.

## Render

The service is deployed as a separate Render Web Service named `music`, independent of the main Notell backend and frontend.
