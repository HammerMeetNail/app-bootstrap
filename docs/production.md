# Production Deployment Guide

This template ships with a Podman-first production workflow that builds, pushes, and deploys a container image over SSH.

## Prerequisites
- A Linux server reachable via SSH.
- Podman + podman-compose installed.
- A Quay repo for the application image.

## Provision a server
You can provision manually or with the included cloud-init.

Manual checklist:
- Create a `deploy` user with SSH access.
- Install Podman and podman-compose.
- Create `/opt/__TEMPLATE_PROJECT_SLUG__` and set ownership to `deploy`.
- Open ports `80` (and `443` if terminating TLS elsewhere).

If you use `cloud-init.yaml`, replace placeholders (including the SSH public key) and apply it on your server. The file creates the deploy user, installs Podman, and prepares `/opt/__TEMPLATE_PROJECT_SLUG__`.

## Required GitHub Actions variables
Set these as Actions **variables** (not secrets):
- `REGISTRY` (default `quay.io`)
- `IMAGE_NAME` (example: `acme/notes-app`)
- `DEPLOY_SSH_HOST`
- `DEPLOY_SSH_USER`
- `DEPLOY_SSH_PATH` (example: `/opt/__TEMPLATE_PROJECT_SLUG__`)
- `DEPLOY_SSH_MODE` (`ssh` or `cloudflare_access_ssh`)

## Required GitHub Actions secrets
- `QUAY_USERNAME`
- `QUAY_PASSWORD`
- `SSH_PRIVATE_KEY`
- `DB_PASSWORD`
- `REDIS_PASSWORD`
- `EMAIL_FROM_ADDRESS`
- `EMAIL_FROM_NAME`
- Email provider key(s) (e.g., `RESEND_API_KEY`)

Optional:
- `CODECOV_TOKEN` (if you want Codecov uploads)

## How deploy works
1. CI builds and pushes a new image tag to Quay.
2. Deploy job SSHes to the server.
3. The job updates the compose file to the new image digest and restarts the stack.

## Cloudflare Access (optional)
Set `DEPLOY_SSH_MODE=cloudflare_access_ssh` and ensure `cloudflared` is installed on the CI runner. The deploy job will use `cloudflared access ssh` as a ProxyCommand.

## Rollbacks
- Find the last known-good image digest in Quay.
- Update the server compose to that digest and restart the stack.
