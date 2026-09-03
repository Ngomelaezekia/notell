# API Contracts

## Authentication

- Authentication uses an HttpOnly `auth_token` cookie.
- Frontend requests use Axios `withCredentials: true`.
- JWTs are not stored in browser localStorage.
- Google OAuth state is stored in a short-lived HttpOnly cookie and must match the callback state.

## Users

Canonical user keys are `id`, `username`, `email`, `profilePicture`, `country`, `city`, `bio`, `status`, `allowFollowers`, `createdAt`, and `updatedAt`.

Profile update input is bounded server-side. Username/email uniqueness is enforced by the database and surfaced as a conflict response.

## Posts

Canonical post keys are `postId`, `userId`, `contentType`, `contentUrl`, `caption`, `createdAt`, `updatedAt`, and `user`.

`contentUrl` must be an absolute managed media URL under `/uploads/`. In production this normally points to the configured durable media origin (such as the R2 public domain); legacy media may still use the API server origin during migration. Post creation verifies the media URL origin/path, validates the media type, verifies ownership through the upload record, and prevents reuse by another post.

Feed ordering is deterministic: `created_at DESC, id DESC`, with one-record lookahead pagination.

## Uploads

`POST /api/upload` accepts multipart field `file` and returns `{message,url}`. Supported types are JPEG, PNG, WebP, MP4, and MOV. The server ceiling is 100 MiB.

Upload records retain ownership and claim metadata so unowned/reused media cannot be attached to arbitrary posts. A post claims exactly one upload as part of the same database transaction that creates the post. New production uploads are stored in durable object storage; the local file is treated as temporary instance-local storage. Deleting a post removes its associated upload record and schedules/removes the corresponding stored object through the storage cleanup path.

## Search

Post search covers captions and author usernames. LIKE wildcards are escaped and result ordering is deterministic with an ID tie-breaker.

User search covers username, bio, city, and country with deterministic username + ID ordering.

## Comments

Replies may target top-level comments only. Parent comments must belong to the same post.

## Relationships

Follower/following lists are paginated. Follow operations respect the target user's `allowFollowers` setting.

## Notifications

Notifications support list/read/read-all operations. Ordering is deterministic with `created_at DESC, id DESC`.

## Active models

The database currently migrates User, Post, Comment, Like, Relationship, Channel, Notification, and Upload. Channel remains schema-only until channel routes/UI are implemented.
