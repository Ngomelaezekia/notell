# Notell Channel Service

Independent Channel domain service for Notell. This service owns channel metadata, membership, team roles, channel plans, and channel entitlements. It does not import the Notell monolith or own user credentials, posts, media storage, live streaming, or payments.

## Phases implemented

- C1: service foundation, configuration, health/readiness, PostgreSQL persistence, JWT authentication contract, CORS, Render runtime contract.
- C2: channel creation, lookup/discovery, public/private channel type, ownership, slug generation, lifecycle status, soft disable.
- C3: follower/subscriber/team membership endpoints, owner/team roles, plan catalog and channel entitlement foundation.

## Local development

Requirements: Go 1.27+ and PostgreSQL.

Required environment variables:

- `DATABASE_URL`
- `JWT_SECRET` (must match the Notell API JWT signing secret)
- `PORT` (default `8080`)
- `FRONTEND_URL` (optional)
- `APP_ENV` (optional)

Run:

```text
go mod tidy
go run .
```

Health: `GET /health`
Readiness: `GET /ready`
API root: `/v1`

## Boundary rules

- No Docker or Kubernetes.
- No imports from `backend/`, `music/`, Message, or Live services.
- Identity is represented by the authenticated Notell JWT `userId` claim.
- Payments are represented only by channel entitlement state; payment processing belongs to Billing.
- Media is represented by media IDs/references; object storage belongs to the media system.
- Live streaming is an external service contract.
