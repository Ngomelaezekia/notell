# API Contracts

## Authentication

- Authentication uses an HttpOnly `auth_token` cookie.
- Frontend requests use Axios `withCredentials: true`.
- JWTs are not stored in browser localStorage.
- Google OAuth state is stored in a short-lived HttpOnly cookie and must match the callback state.

## Upload lifecycle

- `POST /api/upload` is authenticated and records each uploaded object against the current user.
- `POST /api/posts` accepts only media uploaded by the current user and only while that upload is unclaimed.
- Post creation claims the upload in the same database transaction as post creation.
- Deleting a post removes its associated upload record and attempts to remove the stored media file.
- Supported uploads remain JPEG, PNG, WebP, MP4, and MOV; the server ceiling is 100 MiB.
