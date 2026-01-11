# Notes Webapp - Agent Guide

This project is a Go + vanilla JS webapp with email-based auth and a Notes feature set. It was created from the Notes Webapp Template, then customized for this repo.

## Start here
- Architecture overview: `agent_docs/architecture.md`
- Testing & tooling: `agent_docs/testing.md`
- Ops & deploy: `agent_docs/ops.md`

## Upstream reference
- Template source: `https://github.com/HammerMeetNail/app-bootstrap` (use for upstream patterns or missing context).

## If you change backend behavior
- Models live in `internal/models/`, services in `internal/services/`, handlers in `internal/handlers/`.
- Route registration is in `cmd/server/main.go`.
- Schema changes go in `migrations/000001_initial_schema.up.sql` and `.down.sql`.
- Keep auth flows consistent unless you intentionally rewrite them.

## If you change the frontend
- SPA logic is in `web/static/js/app.js`, API client in `web/static/js/api.js`.
- Templates and base layout are in `web/templates/`.
- Keep CSP-friendly patterns (no inline scripts).

## If you touch auth or email flows
- Auth endpoints are in `internal/handlers/auth.go`.
- Email behavior is in `internal/services/email.go`.
- Verify email verification, magic link login, and password reset flows still work.

## Tests and verification
- Unit tests: `make test` (or `go test ./...` and `node web/static/js/tests/runner.js`).
- E2E: `make e2e` (Playwright).

## Safety checks
- Preserve `web/static/openapi.yaml` when APIs change.
- Update `tests/e2e/` to reflect any UI flow changes.
