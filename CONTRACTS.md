# API Contracts

## Authentication

- Core Notell authentication uses an HttpOnly `auth_token` cookie.
- Frontend requests use Axios `withCredentials: true`.
- JWTs are not stored in browser localStorage.
- Google OAuth state is stored in a short-lived HttpOnly cookie and must match the callback state.
- Independent Channel, Message, Live, and Payment services validate the same HS256 Notell JWT contract (`iss=notell-api`).

## Users

Canonical user keys are `id`, `username`, `email`, `profilePicture`, `country`, `city`, `bio`, `status`, `allowFollowers`, `createdAt`, and `updatedAt`.

Profile update input is bounded server-side. Username/email uniqueness is enforced by the database and surfaced as a conflict response.

## Posts

Canonical post keys are `postId`, `userId`, `contentType`, `contentUrl`, `caption`, `createdAt`, `updatedAt`, and `user`.

Managed posts also expose the linked `upload` relation when available, including `upload.mediaMetadata` for processing/playback metadata. The API keeps the existing `contentUrl` field for frontend compatibility.

`contentUrl` must be an absolute managed media URL under `/uploads/`. The application validates the managed media origin/path, validates the media type, verifies ownership through the upload record, and prevents reuse by another post.

Feed ordering is deterministic: `created_at DESC, id DESC`, with one-record lookahead pagination. The categorized feed endpoint may apply its documented engagement/freshness ranking before the deterministic ID tie-breaker.

## Uploads

`POST /api/upload` accepts multipart field `file` and returns `{message,url}`. Supported types are JPEG, PNG, WebP, MP4, and MOV. The server ceiling is 100 MiB.

Upload records retain ownership and claim metadata so unowned/reused media cannot be attached to arbitrary posts. A post claims exactly one upload as part of the same database transaction that creates the post. New production uploads are stored in durable object storage; the local file is treated as temporary instance-local storage. Deleting a post removes its associated upload record and schedules/removes the corresponding stored object through the storage cleanup path.

## Channel service

The independent Channel service owns channel metadata, memberships, team roles, channel plans, and channel entitlements. It does not own user credentials, posts, media objects, live transport, or payment processing.

- Public discovery: `GET /v1/channels`, `GET /v1/channels/:id`.
- Owner management: `POST/PATCH/DELETE /v1/channels...`.
- Membership: follow, subscribe, member and team-management endpoints.
- Channel plans include monthly price, storage allowance, live-minute allowance, manager limits, privacy, and monetization capabilities.
- New channels begin with a `PENDING` entitlement; billing must activate the entitlement before paid channel capabilities are treated as enabled.

## Payment service

The independent Payment service owns provider-neutral payment records, Stripe integration, webhooks, ledger/audit records, usage metering, and subscription entitlement persistence.

- Subscription resources are identified by `(resourceType, resourceId)` so the same billing system can support channels and future paid services.
- `GET /v1/entitlements/:resourceType/:resourceId` returns the authenticated user's current entitlement state.
- Subscription state is `incomplete`, `active`, `trialing`, `past_due`, `paused`, or `canceled` as provider synchronization is added.
- Payment and subscription mutations remain outside the Channel and Live databases.

## Live service

The independent Live service owns stream lifecycle state and live-stream credentials. It does not own channels, payments, or object storage.

- `POST /v1/channels/:channelId/streams` creates a draft stream and returns a one-time ingest credential.
- `POST /v1/streams/:id/start` and `/end` control lifecycle and require stream ownership.
- `GET /v1/channels/:channelId/live` discovers the current live stream.
- `POST /v1/streams/:id/playback-token` issues a ten-minute signed playback token.
- Live service persistence is isolated in its own PostgreSQL database.
- Channel/payment entitlement checks are an explicit integration boundary and must be enforced before production stream start is enabled.

## Messages

The independent Message service owns messaging/realtime state and may use the configured realtime provider. It remains independent from Channel, Live, Payment, Music, and the core backend.

## Media metadata

`MediaMetadata` is created with each successful upload and currently records the uploaded file size and processing status. Duration, dimensions, codec, and thumbnail fields remain empty until a real media-processing pipeline populates them; the API does not fabricate those values.

## Search

Post search covers captions and author usernames. LIKE wildcards are escaped and result ordering is deterministic with an ID tie-breaker.

User search covers username, bio, city, and country with deterministic username + ID ordering.

## Comments

Replies may target top-level comments only. Parent comments must belong to the same post.

## Relationships

Follower/following lists are paginated. Follow operations respect the target user's `allowFollowers` setting and return authoritative relationship counts after mutations.

## Notifications

Notifications support list/read/read-all operations. Ordering is deterministic with `created_at DESC, id DESC`.

## Active core models

The core database migrates User, Post, Comment, Like, Relationship, Channel, Notification, Upload, MediaMetadata, MediaJob, PostView, and PostMusic. Channel business functionality now lives in the independent Channel service; the legacy core Channel model remains for compatibility/migration safety.
