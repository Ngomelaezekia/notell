# Notell Message Service

Independent Go service for realtime messaging and group calling.

## Features

- Direct and group conversations
- Persistent PostgreSQL messages
- Authenticated WebSocket transport
- Instant message delivery to connected members
- Typing indicators
- Read receipts
- Client message IDs for optimistic UI/deduplication foundations
- Group call rooms through LiveKit
- Short-lived LiveKit participant tokens
- Health/readiness endpoints

## Runtime

- Go 1.27.1
- PostgreSQL
- Gin
- Gorilla WebSocket
- LiveKit Server SDK
- No Docker/Kubernetes

## Environment

```text
PORT=10000
APP_ENV=production
DATABASE_URL=<independent PostgreSQL URL>
JWT_SECRET=<same auth signing secret as Notell>
FRONTEND_URL=https://notell-web.onrender.com
LIVEKIT_URL=wss://<project>.livekit.cloud
LIVEKIT_API_KEY=<server key>
LIVEKIT_API_SECRET=<server secret>
```

LiveKit is an external realtime media provider. The Message service owns conversation/call authorization and issues short-lived room tokens; it does not own the media transport.

## API

```text
GET  /health
GET  /ready
POST /v1/conversations
GET  /v1/conversations
GET  /v1/conversations/:id/messages
POST /v1/conversations/:id/read
GET  /v1/ws/:conversationId
POST /v1/conversations/:id/calls
POST /v1/calls/:id/token
```

WebSocket messages:

```json
{"type":"message","body":"hello","clientId":"local-id"}
{"type":"typing"}
{"type":"stop_typing"}
{"type":"read","messageId":"..."}
```

The service is independently deployable and must not import Notell backend packages or share its database schema with the core application.
