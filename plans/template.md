# Year of Bingo → Reusable Webapp Template (Implementation Plan)

## Goal
Create a new repository (a “template repo”) derived from `yearofbingo` that preserves the production-quality scaffolding (Go backend + vanilla JS SPA, Postgres + Redis, email auth flows, containerized tests + Playwright E2E, CI) while removing Bingo-specific product code. The template should be “one command to adopt”: a user clones the template and runs a `make` target to convert it into *their* new webapp repo.

## Source Repo (required reference during implementation)
This plan is authored in (and derived from) the Year of Bingo repository. The agent implementing the template should keep a checkout of the source repo available for reference and copy patterns from it:
- Local path (this workspace): `/Users/dave/git/yearofbingo`
- Upstream: `https://github.com/example/notes-template`

## Decisions (locked in)
- Local dev is **Podman-first** (Docker still supported where it already is, e.g. GitHub Actions).
- Include **production deploy**, matching the current model: build/test → push image to **quay.io** → **SSH deploy** (plain SSH by default; Cloudflare Access optional).
- Keep **OpenAPI** scaffolding.
- Include all auth flows: **email verification**, **magic-link login**, **password login**, **password reset**.
- Sample app is **Notes**.
- License stays **Apache-2.0**.
- `make init` rewrites the **Go module path** (requires a `GO_MODULE` input).
- Deploy defaults to **plain SSH**; Cloudflare Access SSH is optional.
- Codecov upload is **optional**; CI always uploads coverage as an artifact.

## Non‑Goals (to keep scope tight)
- Do not preserve Bingo domain features (cards, friends, reactions, AI wizard, sharing, etc.).
- Do not keep any Year-of-Bingo branding, analytics, production secrets, or third-party keys.
- Do not require network access at runtime beyond standard `go mod`/container builds (and even that should be cached in CI).

## Deliverables
1. **New template repo** (e.g. `go-vanilla-webapp-template`) with:
   - Go `net/http` backend (same patterns: handlers/middleware/services).
   - Vanilla JS SPA (same routing/event-delegation patterns; no inline scripts).
   - Auth (signup/login/logout), email verification, password reset, magic-link login (as currently supported).
   - Postgres persistence + migrations; Redis for sessions/short-lived tokens if applicable.
   - Podman/Docker Compose local dev.
   - Containerized tests (`./scripts/test.sh`) and containerized Playwright E2E (`make e2e`).
   - GitHub Actions CI that runs unit tests + E2E, builds/pushes multi-arch images to Quay on version tags, creates a GitHub Release, and deploys via SSH (matching this repo’s flow).
   - Release helper (`make release`) that updates the footer version + OpenAPI version, tags, and triggers CI release/deploy.
2. **Template “adoption” command**: `make init` (or similar) that rewrites identifiers so the repo becomes a normal project.
3. **A small included sample app** to prove the template works end-to-end:
   - Home page + authenticated “app” page.
   - One CRUD resource persisted in Postgres (“Notes”).
   - E2E coverage of the auth + CRUD workflow, including email verification via Mailpit.
4. **Production doc** explaining how to provision a server and configure CI deploy (plain SSH by default, Cloudflare Access optional).

## Acceptance Criteria (Definition of Done)
- Fresh clone of the template repo passes `make test` and `make e2e`.
- Running `make init` (with provided inputs) produces a renamed project that also passes `make test` and `make e2e`.
- No remaining references to “yearofbingo”, “Year of Bingo”, `yearofbingo.com`, old container image names, or old GitHub URLs after initialization.
- Auth and email flows work locally using Mailpit (verification, reset, and magic-link).
- The sample CRUD resource persists across server restarts and is covered by unit + E2E tests.
- Security invariants preserved (CSP stays strict; no inline JS; user content is rendered safely; add an XSS regression in E2E for the sample CRUD resource).

---

## Template Parameterization (what `make init` should rewrite)
Minimum inputs (required):
- `PROJECT_SLUG` (e.g. `acmeapp`)
- `PROJECT_NAME` (e.g. `Acme App`)
- `GO_MODULE` (e.g. `github.com/acme/acmeapp`)
- `APP_BASE_URL` (e.g. `http://localhost:8080` or `https://app.example.com`)
- `EMAIL_FROM_ADDRESS` (e.g. `noreply@example.com`)

Production/CI inputs (required to deploy, but not required for local dev/tests):
- `QUAY_NAMESPACE` + `IMAGE_NAME` (e.g. `acme/acmeapp`)
- `DEPLOY_SSH_HOST`, `DEPLOY_SSH_USER`, `DEPLOY_SSH_PATH` (e.g. `/opt/<project_slug>`)
- `DEPLOY_SSH_MODE` (`ssh` | `cloudflare_access_ssh`) default `ssh`

Rewrite targets (examples; confirm by search in implementation):
- `go.mod` module path + all internal Go imports; run `go mod tidy` after rewrite.
- Container image names and tags (Containerfiles, compose files, scripts).
- App title/metadata (templates, OpenGraph defaults, favicon if any).
- README badges/links and repo URLs.
- `.env.example` defaults (domain/email).
- Any references inside JS/HTML templates (title, copy).
- GitHub Actions workflows (`.github/workflows/*.yaml`) env/vars for registry, image, deploy host/mode.
- `cloud-init.yaml` (if retained) and any “/opt/<app>” install-path strings in deploy docs/scripts.

Implementation note: prefer a dedicated script (e.g. `scripts/template-init.sh`) called by `make init`, rather than embedding logic in Make.

## `make init` Behavior
Requirements:
- Interactive by default (prompts), but supports non-interactive mode via env vars (for CI smoke checks).
- Refuses to run if the repo appears already initialized (idempotence guard).
- Performs safe, explicit replacements (avoid overly-broad `sed` that could corrupt code).
- After rewrite, runs fast verification (`make test` at minimum; optionally offers `make e2e`).
- Removes template-only files/markers (e.g. `TEMPLATE.md`, initialization docs) or flips a flag so future runs are no-ops.
- At the end, prints a **Next Steps** checklist of anything the user must do manually (notably: GitHub secrets and Quay repo setup).
- By default, offers to run a **GitHub bootstrap** step (via `gh` CLI) to minimize manual toil.

### Recommended placeholder strategy (for predictable rewrites)
- Use explicit template tokens in files that need rewriting (e.g. `__TEMPLATE_PROJECT_NAME__`, `__TEMPLATE_PROJECT_SLUG__`, `__TEMPLATE_GO_MODULE__`, `__TEMPLATE_IMAGE_NAME__`, `__TEMPLATE_DEPLOY_SSH_HOST__`).
- Avoid relying on “replace `yearofbingo` everywhere” as the only mechanism; use targeted replacements + a final `rg` sanity check.

Suggested approach:
- Maintain a `template/manifest.json` describing:
  - required inputs
  - files to rewrite
  - literal strings to replace
  - regex replacements (sparingly)
  - files to delete/rename after init

### `make init` prompts/inputs (recommended)
- `PROJECT_NAME` (UI copy)
- `PROJECT_SLUG` (used in paths, compose project name, etc.)
- `GO_MODULE` (e.g. `github.com/acme/todo-app`) — required; used for `go.mod` + imports
- `QUAY_NAMESPACE` + `IMAGE_NAME` (defaults can derive from `PROJECT_SLUG`)
- `DEPLOY_SSH_HOST` + `DEPLOY_SSH_USER` + `DEPLOY_SSH_PATH` (e.g. `/opt/<project_slug>`)
- `DEPLOY_SSH_MODE` (`ssh` | `cloudflare_access_ssh`) default `ssh`
- `APP_BASE_URL` (used to form absolute links in emails)
- `EMAIL_FROM_ADDRESS`

### `make init` end-of-run “Next Steps” output (required)
Print a short checklist that includes:
- Create the remote GitHub repo (if not created yet) and set `origin`.
- Create the Quay repo `quay.io/<QUAY_NAMESPACE>/<IMAGE_NAME>` (or confirm it exists).
- Configure GitHub Actions variables (non-secret configuration used by workflow):
  - `REGISTRY` (default `quay.io`)
  - `IMAGE_NAME` (e.g. `<quay-namespace>/<image>`)
  - `DEPLOY_SSH_HOST`, `DEPLOY_SSH_USER`, `DEPLOY_SSH_PATH`, `DEPLOY_SSH_MODE`
- Add required GitHub secrets (names should match CI workflow):
  - Quay push: `QUAY_USERNAME`, `QUAY_PASSWORD`
  - Deploy: `SSH_PRIVATE_KEY`
  - App runtime: `DB_PASSWORD`, `REDIS_PASSWORD`, email provider key(s) (e.g. `RESEND_API_KEY`), `EMAIL_FROM_ADDRESS`
  - Optional: `CODECOV_TOKEN` (if you want upload)
- Update production DNS / `APP_BASE_URL` if deploying to a real domain.
- Provision the server (see template production doc / `cloud-init.yaml` if provided).
- Run `make test` (and optionally `make e2e`) locally to confirm.
- Commit the initialized changes locally and push when ready (or accept the optional “push now?” prompt from `make github-bootstrap`).

Optional (to reduce manual toil):
- If GitHub CLI (`gh`) is installed and authenticated, run (or offer to run) `make github-bootstrap` to create/configure the GitHub repo and Actions settings.

---

## `make github-bootstrap` (required)
Goal: after `make init`, set up a working remote + GitHub Actions configuration with minimal manual steps.

Preconditions:
- `gh` is installed and `gh auth status` is authenticated.
- User provides `GITHUB_OWNER` (user/org) and `GITHUB_REPO` (repo name) unless derivable.

Behavior:
1. **Create repo** (if missing):
   - `gh repo create <owner>/<repo> --source=. --remote=origin` (visibility configurable; default **private** unless user opts public).
   - If repo already exists, ensure `origin` points to it.
   - Do **not** push any code by default.
   - Prompt at the end: “Push initialized code to GitHub now? (y/N)”. If yes, run `git push -u origin <branch>`.
2. **Set Actions variables** (non-secret):
   - `REGISTRY=quay.io`
   - `IMAGE_NAME=<quay-namespace>/<image>`
   - `DEPLOY_SSH_HOST`, `DEPLOY_SSH_USER`, `DEPLOY_SSH_PATH`, `DEPLOY_SSH_MODE`
3. **Provision secrets**:
   - Generate locally (no prompts) where safe:
     - `DB_PASSWORD`: generate strong random password.
     - `REDIS_PASSWORD`: generate strong random password.
     - `SSH_PRIVATE_KEY`: generate a new deploy keypair (e.g. ed25519) dedicated to this project.
   - Prompt user for secrets that cannot be generated:
     - Quay push: `QUAY_USERNAME`, `QUAY_PASSWORD` (or robot token).
     - Email provider key(s): e.g. `RESEND_API_KEY` (or selected provider).
     - Optional: `CODECOV_TOKEN` if they want upload enabled.
4. **Apply secrets/vars**:
   - Use `gh variable set ...` and `gh secret set ...` so secrets never appear in git history.
   - Do not echo secret values to stdout; if you print anything, print only secret *names* set.
5. **Print manual follow-ups** (still unavoidable):
   - Create/verify Quay repo exists (Quay API automation is optional; default to “manual step”).
   - Install the generated deploy public key on the server (`~deploy/.ssh/authorized_keys`) and verify `ssh` works.
   - If `DEPLOY_SSH_MODE=cloudflare_access_ssh`, confirm Cloudflare Access is configured for the hostname.
   - Run a dry-run CI: push a commit (or accept the push prompt); optionally run a `vX.Y.Z` tag release after server is ready.

Verification:
- `gh repo view` succeeds for the created repo.
- `gh secret list` / `gh variable list` shows expected keys (names only).
- A push to `main` triggers CI tests; a version tag triggers image build/push (deploy only when server is configured).

---

## Phase 0 — Discovery (in the source repo)
1. Inventory current architecture to preserve:
   - Go entrypoint (`cmd/server`), router registration, middleware chain, handler/service boundaries.
   - Session/auth design (cookies, CSRF, token issuance).
   - Email sending abstraction + Mailpit usage in E2E.
   - Postgres + Redis initialization, migrations, and seeding.
   - JS “App object” pattern, routing, and event delegation.
2. Identify Bingo-specific modules/routes to remove or isolate behind a “sample app” boundary.
3. Identify release/deploy specifics to generalize:
   - image registry name(s)
   - environment assumptions
   - any hard-coded hostnames/domains

Verification:
- Baseline: `./scripts/test.sh` and `make e2e` pass in the source repo before templating work begins.

## Phase 1 — Create the Template Repo Skeleton
1. Copy repo into a new directory/repo (or create a new branch used to generate the template).
2. Remove/replace Year-of-Bingo branding:
   - rename README, titles, assets, and default OpenGraph image copy.
3. Rename “bingo” env defaults to neutral (e.g. `app`), keeping config structure intact.
4. Keep key ops/dev tooling:
   - `Makefile`, `Containerfile*`, `compose*.yaml`, `scripts/`, `.env.example`, migrations tooling, Playwright config.
5. Remove product-specific/optional integrations that create unnecessary toil in a generic template:
   - remove Gemini/AI features, env vars, secrets, OpenAPI endpoints, and tests
   - keep the auth/email/db “spine” intact
6. Ensure the template repo still builds and runs locally (`make local`).

Verification:
- `make test` passes in the template repo *before* adding the template-init mechanism (keeps changes incremental).

## Phase 2 — Replace Bingo With a Minimal Sample App (keeps “core functionality”)
Replace Bingo-specific features with a minimal webapp that still exercises the full stack.

Recommended sample: **Notes** with:
- `GET /` home page (public) with CTA.
- Authenticated SPA route (e.g. `/#app`) showing:
  - list notes
  - create note
  - edit note
  - delete note
- Notes table columns:
  - `id`, `user_id`, `title`, `body`, `created_at`, `updated_at`

Backend tasks:
1. Implement service methods for notes CRUD (with authorization by `user_id`).
2. Add handlers and routes (`/api/notes/*`) consistent with existing routing patterns.
3. Add migrations to create the notes table.
4. Add unit tests for service + handler behavior (match repo conventions).

Frontend tasks:
1. Add SPA views for Notes.
2. Render note content safely (no `innerHTML` for user-controlled content; use DOM APIs or existing escaping utilities).
3. Add a simple nav/header pattern matching existing UI structure.

Verification:
- Add Playwright spec(s) covering:
  - sign up
  - receive verification email in Mailpit
  - verify email link works
  - create/edit/delete a note
  - password reset flow
  - magic-link login flow
  - XSS regression: create a note with payload like `"<img src=x onerror=alert(1)>"` and assert it renders as text (no `<img>` node).

## Phase 3 — Add Template Initialization (“Convert Template → Real App”)
1. Add `make init` that calls `scripts/template-init.sh`.
2. Add a single source of truth for template variables:
   - `template/manifest.json` + `template/README.md` describing the placeholders.
3. Ensure init rewrites:
   - module name (`go.mod`) + import paths
   - app name in templates/UI
   - container names, compose project names, image tags
   - `.env.example` email defaults
4. Ensure init also updates CI metadata:
   - badge URLs
   - Codecov upload behavior (default off unless token is present)
5. Ensure init verifies “no template leftovers”:
   - `rg -n "Year of Bingo|yearofbingo|yearofbingo.com" -S` must be empty (or known allowlist)
   - `go test ./...` and `node web/static/js/tests/runner.js` succeed via `make test`

Verification:
- CI smoke check job runs `make init` in a clean checkout with sample inputs, then runs `make test`.

## Phase 4 — CI/CD Generalization
1. Keep the current pipeline shape from `yearofbingo` and parameterize it:
   - secret scan (gitleaks), lint (golangci-lint), go test, js test, Playwright E2E (Docker Compose)
   - multi-arch image builds (amd64 + arm64), Trivy scan, Cosign signing, push manifest to Quay
   - create GitHub Release on version tags
   - deploy job that updates server compose with the image digest and restarts the stack
2. Replace hard-coded identifiers with template placeholders:
   - registry + image name (`quay.io/<namespace>/<image>`)
   - deploy SSH host and app install path on the server (e.g. `/opt/<app>`)
   - any domain strings used in deploy scripts/SSH config
3. Make deploy mode selectable:
   - `DEPLOY_SSH_MODE=ssh`: plain `ssh/scp`, install `known_hosts` or use `StrictHostKeyChecking accept-new`
   - `DEPLOY_SSH_MODE=cloudflare_access_ssh`: install `cloudflared` and set `ProxyCommand cloudflared access ssh --hostname %h`
   - Gate the `cloudflared` install step on `DEPLOY_SSH_MODE`
4. Document required GitHub secrets/vars (template README):
   - Quay: `QUAY_USERNAME`, `QUAY_PASSWORD`
   - Deploy: `SSH_PRIVATE_KEY` (+ either Cloudflare Access config or plain SSH host keys)
    - App secrets: `DB_PASSWORD`, `REDIS_PASSWORD`, email provider keys + from address
    - Optional: Codecov token (if enabled), Cosign keyless needs no secret but does need OIDC permissions
5. Minimize “required on day 1” secrets:
   - ensure deploy workflow does not require optional secrets (e.g. backups/AI) to be present
   - if you keep backups in the template, gate them behind separate secrets and clearly mark as optional in docs

Verification:
- GitHub Actions workflow runs successfully in the template repo (tests + E2E).

## Phase 5 — Documentation + Developer Experience
1. Update template `README.md`:
   - “Use as template” instructions
   - “Initialize” (`make init`) instructions
   - “Run local” (`make local`) and “Test” (`make test`, `make e2e`) instructions
   - “Production deploy” summary + link to a detailed doc
2. Provide an `AGENTS.md` in the template repo with progressive disclosure:
   - include only the docs required for templated workflows (auth/email/db/testing/ops)
   - remove Bingo-only sections
3. Provide `agent_docs/` equivalents where needed (testing, database, architecture, ops) and ensure paths in AGENTS are correct.
4. Add `docs/production.md` (or similar) covering:
   - provisioning a server (recommended: ship an example `cloud-init.yaml` that creates a `deploy` user, installs podman/podman-compose, opens ports, and creates `/opt/<project_slug>`)
   - how the CI deploy job works (compose-by-digest, restart procedure)
   - required/optional GitHub secrets/vars
   - how to rotate deploy keys and roll back to a previous image digest

Verification:
- `rg yearofbingo -S` returns no results after init (except in historical attribution if intentionally kept).

---

## Extension Points (how a user builds a new app on the template)
The template should document a “happy path” for adding new functionality without rework:
- **Backend**: add a service in `internal/services/`, add handlers in `internal/handlers/`, register routes in `main.go`, add migrations in `migrations/`, update `web/static/openapi.yaml`, add Go tests.
- **Frontend**: add a new SPA route + view module under `web/static/js/`, use existing event delegation (`data-action`) and safe rendering patterns, add JS unit tests where applicable.
- **E2E**: add Playwright spec(s) under `tests/e2e/` that exercise the workflow through the UI; include an XSS regression when rendering user content.

---

## Suggested File/Directory Outcomes (Template Repo)
Keep (adapt as needed):
- `cmd/server/`, `internal/` (but prune Bingo domains)
- `migrations/`
- `web/templates/`, `web/static/` (SPA + CSS)
- `scripts/`, `Makefile`
- `Containerfile`, `Containerfile.test`, `Containerfile.playwright`, `compose.yaml`, `compose.server.yaml`
- `tests/e2e/`, `playwright.config.js`
- `web/static/openapi.yaml` (if API docs are intended to remain part of the template)

Add:
- `template/manifest.json`
- `scripts/template-init.sh`
- `TEMPLATE.md` (quick checklist; optionally deleted by init)

---

## Verification Checklist (for the AI agent implementing this)
In template repo (pre-init):
- `make local` boots successfully (server, db, redis, mailpit if included).
- `make test` passes.
- `make e2e` passes.

After `make init` with sample inputs:
- `make test` passes.
- `make e2e` passes.
- `rg -n \"Year of Bingo|yearofbingo\" -S` returns no results.

---

## Open Questions (please answer before implementation)
1. None (decisions locked). If you change your mind later: confirm whether deploy should require Cloudflare Access, and whether Codecov should be on by default.
