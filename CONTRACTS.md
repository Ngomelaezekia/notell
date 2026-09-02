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
- `GET /api/posts/search?q=&page=&limit=` — authenticated; searches post captions; returns `data.posts` plus `data.pagination` (`page`, `limit`, `total`, `hasMore`).
- `DELETE /api/posts/:id`
- `POST /api/posts/:id/like`
- `GET /api/posts/:id/comments`
- `POST /api/posts/:id/comments`

## Users and relationships
- `GET /api/users/:id`
- `PUT /api/users/profile`
- `GET /api/users/search?q=&page=&limit=` — authenticated; searches username, bio, city, and country; returns `data` plus pagination metadata (`page`, `limit`, `total`, `hasMore`).
- `POST /api/users/:id/follow`
- `DELETE /api/users/:id/unfollow`
- `GET /api/users/:id/relationship`
- `DELETE /api/users/followers/:id`
- `GET /api/users/:id/followers`
- `GET /api/users/:id/following`

## Search frontend contract
- Frontend search page: `/search?q=&type=`.
- Supported tabs: `all`, `people`, `posts`.
- Search results use canonical User/Post JSON names.
- People and Posts support incremental pagination through their respective `hasMore` metadata.

## Routes
Current frontend routes must correspond to actual pages. Do not expose `/explore` or `/settings` navigation until their routes/pages exist.

## Database
Every model used by an active handler must be included in migration coverage. Current active models: User, Post, Comment, Like, Relationship. Channel is included only if its feature is retained.
