# Notell Music Service

Independent backend service for music discovery, catalog management, previews, and post-track selection.

## Current API

- `GET /health`
- `GET /api/music/health`
- `GET /api/music/search?q=<query>`
- `GET /api/music/trending`
- `GET /api/music/tracks/<id>`

## Architecture

This service is intentionally isolated from the main Notell API. The main app will consume this service over HTTP when the music picker is integrated.

The catalog currently contains one internal demo track so deployment and integration can be verified without depending on an external music provider.

The provider layer will be added next. It will keep discovery metadata, licensed previews, and post-usage rights separate so Notell does not scrape or re-host music from unauthorized sources.

## Environment

- `PORT` — Render-provided HTTP port.
- `FRONTEND_URL` — allowed browser origin for CORS.

## Render

The service is deployed from the repository `music/` root as a separate Render Web Service named `music`.
