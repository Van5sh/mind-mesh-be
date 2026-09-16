# OAuth setup

The backend serves browser OAuth redirects; GraphQL is used after login to read
the authenticated session (`me`). Configure both providers with these callback
URLs, substituting the public API origin for `API_BASE_URL`:

- Google: `API_BASE_URL/auth/google/callback`
- GitHub: `API_BASE_URL/auth/github/callback`

Required environment variables:

```text
DATABASE_URL=postgres://...
API_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
```

Set `APP_ENV=production` (or explicitly `AUTH_COOKIE_SECURE=true`) when the API
is HTTPS. Login URLs are `/auth/google` and `/auth/github`. On success, the
backend sets the HTTP-only `session_id` cookie and redirects to `FRONTEND_URL`.

The GraphQL endpoint permits credentialed CORS requests only from the origin in
`FRONTEND_URL`. Configure the frontend client to send requests with credentials
(for example, `fetch(..., { credentials: "include" })`).
