# Template Initialization

Run the initializer to convert this template into your own app:

```bash
make init
```

Required inputs:
- `PROJECT_NAME`
- `PROJECT_SLUG`
- `GO_MODULE`
- `APP_BASE_URL`
- `EMAIL_FROM_ADDRESS`

Optional inputs:
- `EMAIL_FROM_NAME` (defaults to `PROJECT_NAME`)
- `QUAY_NAMESPACE` (defaults to `PROJECT_SLUG`)
- `IMAGE_NAME` (defaults to `PROJECT_SLUG`)
- `DEPLOY_SSH_HOST`
- `DEPLOY_SSH_USER`
- `DEPLOY_SSH_PATH` (defaults to `/opt/<project_slug>`)
- `DEPLOY_SSH_MODE` (`ssh` or `cloudflare_access_ssh`)
- `DEPLOY_SSH_PUBLIC_KEY` (optional; used in `cloud-init.yaml`)

Non-interactive example:
```bash
PROJECT_NAME="Acme Notes" \
PROJECT_SLUG="acme-notes" \
GO_MODULE="github.com/acme/acme-notes" \
APP_BASE_URL="https://notes.acme.com" \
EMAIL_FROM_ADDRESS="noreply@acme.com" \
make init
```

The initializer:
- Rewrites template placeholders across the repo.
- Updates the Go module path and runs `go mod tidy`.
- Prints a next-steps checklist for CI/deploy setup.
