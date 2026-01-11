# Ops & Deploy

## CI pipeline
- Runs tests, lint, and Playwright E2E.
- Builds multi-arch images and pushes to Quay.
- Creates GitHub Releases on version tags.
- Deploys via SSH using `compose.server.yaml`.

## Image naming
- Registry defaults to `quay.io`.
- Image name is configured via CI variables (`REGISTRY`, `IMAGE_NAME`).

## Deploy model
- Server uses Podman + podman-compose.
- `/opt/<project_slug>` holds the deployment.
- `compose.server.yaml` runs the app, Postgres, Redis.

See `docs/production.md` for provisioning and secrets.
