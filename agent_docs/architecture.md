# Architecture Overview

## Backend
- Entry point: `cmd/server/main.go`
- Config: `internal/config`
- Database: `internal/database` (Postgres + Redis, migrations on boot)
- Services: `internal/services` (auth, user, email, notes)
- Handlers: `internal/handlers` (auth, notes, health, pages)
- Middleware: `internal/middleware` (auth, CSRF, security headers, cache control, compression)

### Notes API
- `GET /api/notes` list notes for the authenticated user.
- `POST /api/notes` create a note.
- `GET /api/notes/{id}` fetch a note.
- `PUT /api/notes/{id}` update a note.
- `DELETE /api/notes/{id}` delete a note.

### Auth API
- Register/login/logout, email verification, magic-link login, and password reset.
- Sessions stored in Redis with Postgres fallback.

## Frontend
- SPA lives in `web/static/js/app.js` + `web/static/js/api.js`.
- Hash-based routing (`#home`, `#login`, `#register`, `#app`, etc.).
- Event delegation via `data-action` and `data-action` on forms.
- Always render user content with `textContent` to avoid XSS.

## Templates
- HTML shell: `web/templates/index.html`
- OpenAPI spec: `web/static/openapi.yaml`
