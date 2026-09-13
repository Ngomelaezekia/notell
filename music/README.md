# Notell Music Service

Independent backend service for music discovery, catalog management, previews, rights checks, playback descriptors, post-track selection, and future trend analytics.

## Current API

- `GET /health`
- `GET /api/music/health`
- `GET /api/music/search?q=<query>`
- `GET /api/music/trending`
- `GET /api/music/tracks/<id>`
- `GET /api/music/tracks/segment?trackId=<id>&startSec=<seconds>&endSec=<seconds>`

## Licensed-provider direction

The production catalog is designed for a licensed B2B music provider rather than scraping Spotify, YouTube, Apple Music, or other consumer services.

The primary provider target is **MassiveMusic / 7digital**. Its current platform documentation explicitly supports social-media services, catalog metadata, audio delivery, playback, reporting, territory availability, and rights workflows. Songtradr is being treated as a commercial/licensing relationship to evaluate, but the current Songtradr Marketplace is transitioning to MassiveMusic and Bandcamp, so Notell should not couple its application code to the legacy Songtradr Marketplace.

Provider integration must be enabled only after a commercial agreement explicitly covers Notell's intended use: UGC synchronization, streaming/playback to viewers, territory rights, attribution, usage reporting, takedowns, and any required publisher clearances.

## Rights model

Every production track should carry enough information for the service to decide whether it can be exposed to a user in a territory and whether it can be attached to a post:

- recording/label availability
- publishing availability
- UGC/synchronization permission
- streaming permission
- territory list
- start/end availability dates
- attribution requirements
- provider status/takedown state

The service must reject a track when the required rights are not confirmed. A preview URL is not treated as a permanent post asset unless the provider agreement permits that use.

## Segment model

Post music selection uses a bounded segment:

`0 <= startSec < endSec <= durationSec`

The default maximum segment length is 60 seconds and can be changed with `MUSIC_MAX_SEGMENT_SECONDS`. This limit is a product default, not a statement about provider licensing; the licensed provider's actual rules remain authoritative.

## Trends roadmap

The service is designed to calculate Notell-owned rankings from first-party events, including:

- global trending
- country/territory trending
- city/local trending
- fastest rising
- most used in posts
- most played/listened
- most saved
- 24-hour and 7-day growth

Provider usage reporting and Notell recommendation/trending analytics remain separate concerns.

## Environment

- `PORT` — Render-provided HTTP port.
- `FRONTEND_URL` — allowed browser origin for CORS.
- `MUSIC_PROVIDER` — provider selector; defaults to `internal` until a licensed provider is configured.
- `MUSIC_MAX_SEGMENT_SECONDS` — maximum selectable segment length; defaults to `60`.

## Render

The service is deployed as a separate Render Web Service named `music`, independent of the main Notell backend and frontend.
