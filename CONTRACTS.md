# Notell API Contract Baseline

## Authentication
- Authentication is cookie-based.
- Backend sets the HttpOnly `auth_token` cookie.
- Frontend uses Axios `withCredentials: true`.
- The frontend must not persist JWTs in localStorage.
- Canonical current-user endpoint: `GET /api/auth/me`.
- Google OAuth uses a short-lived HttpOnly `oauth_state` cookie and validates the returned state before creating a session.

## Canonical JSON names
Use backend JSON names exactly in the frontend:
- User: `id`, `username`, `email`, `profilePicture`, `country`, `city`, `bio`, `status`, `allowFollowers`, `createdAt`, `updatedAt`.
- Post: `postId`, `userId`, `contentType`, `contentUrl`, `caption`, `createdAt`, `updatedAt`, `user`.
- Comment: `commentId`, `postId`, `userId`, `content`, `parentId`, `createdAt`, `updatedAt`, `user`.

Do not introduce compatibility aliases such as `contentURL`, `profile_picture`, `post.id`, `post.ID`, or `post.User` in new code.

## Media
`POST /api/upload` accepts multipart field `file` and returns `{ message, url }`.
- Supported uploads are JPEG, PNG, WebP, MP4, and MOV.
- The backend validates the detected file content instead of trusting the browser MIME type.
- The backend enforces a 100 MiB upload ceiling; the frontend currently applies a stricter 50 MB UX limit.
- Posts store an absolute URL under the configured server `/uploads/` path in `contentUrl`.
- `POST /api/posts` rejects arbitrary external media URLs.

## Posts
- `POST /api/posts`
- `GET /api/posts/feed?page=&limit=` — authenticated; returns `data` plus `pagination` (`page`, `limit`, `hasMore`). Feed pagination uses a one-record lookahead so `hasMore` does not require a separate count query. Ordering is deterministic by `created_at DESC, id DESC`.
- `GET /api/posts/:id`
- `GET /api/posts/search?q=&page=&limit=` — authenticated; searches post captions and author usernames; returns `data.posts` plus `data.pagination` (`page`, `limit`, `total`, `hasMore`). Search ordering includes a deterministic `id` tie-breaker.
- `DELETE /api/posts/:id`
- `POST /api/posts/:id/like`
- `GET /api/posts/:id/comments`
- `POST /api/posts/:id/comments`
- Comment replies may target top-level comments only; replies to replies are rejected consistently by the API and UI.

## Users and relationships
- `GET /api/users/:id`
- `PUT /api/users/profile`
- `GET /api/users/search?q=&page=&limit=` — authenticated; searches username, bio, city, and country; returns `data` plus pagination metadata (`page`, `limit`, `total`, `hasMore`).
- `POST /api/users/:id/follow`
- `DELETE /api/users/:id/unfollow`
- `GET /api/users/:id/relationship`
- `DELETE /api/users/followers/:id`
- `GET /api/users/:id/followers?page=&limit=` — authenticated; returns `data` plus pagination metadata.
- `GET /api/users/:id/following?page=&limit=` — authenticated; returns `data` plus pagination metadata.

## Notifications
- `GET /api/notifications?page=&limit=` — authenticated.
- `POST /api/notifications/:id/read`
- `POST /api/notifications/read-all`
- Notification types currently include `like`, `comment`, `reply`, and `follow`.

## Search frontend contract
- Frontend search page: `/search?q=&type=`.
- Supported tabs: `all`, `people`, `posts`.
- Search results use canonical User/Post JSON names.
- People and Posts support incremental pagination through their respective `hasMore` metadata.

## Routes
Current frontend routes must correspond to actual pages. Do not expose `/explore` or `/settings` navigation until their routes/pages exist.

## Database
Every model used by an active handler must be included in migration coverage. Current active models: User, Post, Comment, Like, Relationship, Notification. Channel remains schema-only until its handlers/routes/UI are implemented or the model is intentionally removed.
