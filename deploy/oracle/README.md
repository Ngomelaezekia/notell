# Notell Oracle Always Free smoke-test deployment

This target moves compute off Render without changing the Notell service boundaries.

## Target

- Oracle Cloud Always Free Ampere A1 VM for the Go services.
- Caddy for public HTTPS and reverse proxy.
- systemd for service processes.
- Cloudflare Pages for the React/Vite frontend.
- Neon for PostgreSQL.
- Backblaze B2 for media.
- LiveKit Cloud for realtime media.
- Stripe test mode for payments.

## VM services

- core API: 127.0.0.1:8080 -> api.<DOMAIN>
- music: 127.0.0.1:8081 -> music.<DOMAIN>
- channel: 127.0.0.1:8082 -> channel.<DOMAIN>
- message: 127.0.0.1:8083 -> message.<DOMAIN>
- live: 127.0.0.1:8084 -> live.<DOMAIN>
- payment: 127.0.0.1:8085 -> payment.<DOMAIN>

Never expose ports 8080-8085 publicly. Allow only TCP 80/443 to the VM.

## Build

The repository already contains independent Go modules. Build each service natively on the ARM64 VM:

    cd /opt/notell/backend && go mod download && go build -o /opt/notell/bin/notell .
    cd /opt/notell/music && go mod download && go build -o /opt/notell/bin/music .
    cd /opt/notell/services/channel && go mod download && go build -o /opt/notell/bin/channel .
    cd /opt/notell/services/message && go mod download && go build -o /opt/notell/bin/message .
    cd /opt/notell/services/live && go mod download && go build -o /opt/notell/bin/live .
    cd /opt/notell/services/payment && go mod download && go build -o /opt/notell/bin/payment .

## Core environment

APP_ENV=production
PORT=8080
DATABASE_URL=<Neon core database>
JWT_SECRET=<shared auth signing secret>
FRONTEND_URL=https://<CLOUDFLARE_PAGES_HOSTNAME>
SERVER_URL=https://api.<DOMAIN>
GOOGLE_CLIENT_ID=<client id>
GOOGLE_CLIENT_SECRET=<client secret>
GOOGLE_REDIRECT_URL=https://api.<DOMAIN>/api/auth/google/callback
MUSIC_SERVICE_URL=https://music.<DOMAIN>
MEDIA_PUBLIC_URL=https://api.<DOMAIN>
STORAGE_DRIVER=b2
B2_ENDPOINT=<HTTPS endpoint>
B2_BUCKET=<bucket>
B2_KEY_ID=<key id>
B2_APPLICATION_KEY=<application key>
B2_REGION=<region>

The other services use their existing environment contracts from render.yaml, replacing Render hostnames with the Oracle/Caddy hostnames above and Neon database URLs.

## Frontend

Cloudflare Pages build command:

    cd frontend && npm ci && npm run build

Publish directory: frontend/dist

Set these Vite variables to the new HTTPS service hostnames:

- VITE_SERVER_URL=https://api.<DOMAIN>
- VITE_API_URL=https://api.<DOMAIN>/api
- VITE_MUSIC_SERVICE_URL=https://music.<DOMAIN>
- VITE_CHANNEL_SERVICE_URL=https://channel.<DOMAIN>
- VITE_MESSAGE_SERVICE_URL=https://message.<DOMAIN>
- VITE_LIVE_SERVICE_URL=https://live.<DOMAIN>
- VITE_PAYMENT_SERVICE_URL=https://payment.<DOMAIN>
- VITE_MUSIC_COUNTRY=TZ

## OAuth and Stripe

Google callback:
https://api.<DOMAIN>/api/auth/google/callback

Stripe test webhook:
https://payment.<DOMAIN>/v1/webhooks/stripe

## Smoke-test order

1. VM and Caddy TLS.
2. Core /api/health and /api/ready.
3. Cloudflare frontend.
4. Register/login and Google OAuth.
5. Neon persistence.
6. B2 upload and protected playback.
7. Channel.
8. Message HTTP and WebSocket.
9. LiveKit call token and live lifecycle.
10. Stripe test checkout/webhook/entitlement reconciliation.
11. Music and trends.

Do not commit real secrets to this directory.