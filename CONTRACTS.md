# Notell API Contract Baseline

## Authentication
- Authentication is cookie-based.
- Backend sets the HttpOnly `auth_token` cookie.
- Frontend uses Axios `withCredentials: true`.
- The frontend must not persist JWTs in localStorage.
- Canonical current-user endpoint: `GET /api/auth/me`.

## Canonical JSON names
Use backend JSON names exactly in the frontend:
- User: `id`, `username`, `email`, `profilePicture`, `country`, `city`, `bio`, `status`, `allowFollowers`, `createdAt`, `updatedAt`.
- Post: `postId`, `userId`, `contentType`, `contentUrl`, `caption`, `createdAt`, `updatedAt`, `user`.
- Comment: `commentId`, `postId`, `userId`, `content`, `parentId`, `createdAt`, `updatedAt`, `user`.

Do not introduce compatibility aliases such as `contentURL`, `profile_picture`, `post.id`, `post.ID`, or `post.User` in new code.

## Media
`POST /api/upload` accepts multipart field `file` and returns `{ message, url }`.
Posts store the returned absolute URL in `contentUrl`.

## Posts
- `POST /api/posts`
- `GET /api/posts/feed?page=&limit=`
- `GET /api/posts/:id`
- `DELETE /api/posts/:id`
- `POST /api/posts/:id/like`
- `GET /api/posts/:id/comments`
- `POST /api/posts/:id/comments`

## Users and relationships
- `GET /api/users/:id`
- `PUT /api/users/profile`
- `POST /api/users/:id/follow`
- `DELETE /api/users/:id/unfollow`
- `DELETE /api/users/followers/:id`
- `GET /api/users/:id/followers`
- `GET /api/users/:id/following`

## Routes
Current frontend routes must correspond to actual pages. Do not expose `/explore` or `/settings` navigation until their routes/pages exist.

## Database
Every model used by an active handler must be included in migration coverage. Current active models: User, Post, Comment, Like, Relationship. Channel is included only if its feature is retained.
