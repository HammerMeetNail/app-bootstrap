# Notes Webapp Template

A production-ready Go + vanilla JS template with email-based auth, Postgres + Redis, and a Notes sample app. This repo is designed to be cloned and initialized into your own app with a single command.

## What’s Included
- Go `net/http` backend with services/handlers/middleware layout.
- Vanilla JS SPA with hash routing (no inline scripts; CSP-friendly).
- Auth flows: email verification, magic-link login, password login, password reset.
- Postgres migrations + Redis-backed sessions.
- Podman-first local dev with Compose.
- Containerized unit tests and Playwright E2E.
- CI/CD pipeline for tests, multi-arch builds, Quay push, and SSH deploy.

## Quick Start (Local)
```bash
make local
```

Then visit `http://localhost:8080`.

## Tests
```bash
make test
make e2e
```

## Template Initialization
Run the template initializer to rewrite identifiers (project name, module path, image name, etc.).

```bash
make init
```

See `template/README.md` for details and required inputs.

Optional GitHub bootstrap:
```bash
make github-bootstrap
```

## Replacing the Notes App (AI Agent Checklist)
The Notes UI and API are intentionally small so you can swap in your own webapp quickly. Follow this exact checklist.

Backend:
- Edit the schema in `migrations/000001_initial_schema.up.sql` and `.down.sql`; keep users/sessions/tokens unless you plan to rewrite auth.
- Create/adjust models in `internal/models/`, services in `internal/services/`, and handlers in `internal/handlers/`.
- Register your new routes in `cmd/server/main.go` and wire dependencies in the service constructors.
- Update `web/static/openapi.yaml` to reflect your endpoints and payloads.

Frontend:
- Replace the SPA logic in `web/static/js/app.js`; use `web/static/js/api.js` as the HTTP client or swap it out.
- Update `web/templates/index.html` for your base layout and `web/static/css/styles.css` for styling.
- Keep `index.html` as the SPA entry point; update `web/templates/404.html` and `web/templates/500.html` if your routes or copy change.

Tests:
- Update unit tests under `internal/` to reflect your service/handler behavior.
- Rewrite Playwright specs in `tests/e2e/` to cover your user flows end-to-end.

## Production Deploy
See `docs/production.md` for provisioning, CI secrets/vars, and deploy workflow details.

## Repository Layout
- `cmd/server` - app entrypoint
- `internal/` - config, middleware, handlers, services
- `web/` - SPA + templates + OpenAPI
- `migrations/` - Postgres schema
- `tests/e2e/` - Playwright specs

## License
Apache-2.0
