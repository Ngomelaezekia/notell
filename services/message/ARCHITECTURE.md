# Message Service Backend Hardening

The Message service is an independent realtime service. Backend hardening targets low-latency delivery, authenticated WebSocket sessions, message idempotency, bounded payloads, read receipts, typing presence, and group-call authorization.

## Realtime contract

- WebSocket sessions are authenticated with the Notell JWT.
- Conversation membership is checked before the socket is accepted.
- Message payloads are bounded before persistence/broadcast.
- Client IDs are used for idempotent optimistic delivery.
- Typing events are ephemeral and are never persisted.
- Read receipts are persisted and broadcast to members.
- WebRTC media is handled by LiveKit; this service owns room authorization and short-lived tokens.

## Scale boundary

The in-process WebSocket hub is suitable for the first deployment. Before horizontal scaling, replace the process-local fan-out with a shared realtime broker/pub-sub layer without changing the public message API.

## Reliability

Persistence remains authoritative. WebSocket delivery is a realtime notification path; clients must reconcile from the message history endpoint after reconnects.

## Security

Never expose the LiveKit API secret or database credentials to clients. Only the Message service issues participant tokens after checking conversation membership.
