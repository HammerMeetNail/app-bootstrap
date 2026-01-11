#!/usr/bin/env bash
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "Error: gh CLI is required." >&2
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "Error: gh auth status failed. Run 'gh auth login' first." >&2
  exit 1
fi

current_owner="${GITHUB_OWNER:-}"
if [[ -z "$current_owner" ]]; then
  current_owner=$(gh api user -q .login 2>/dev/null || true)
fi

repo_default="${GITHUB_REPO:-$(basename "$(pwd)")}" 

read -r -p "GitHub owner [${current_owner}]: " owner_input
owner="${owner_input:-$current_owner}"
if [[ -z "$owner" ]]; then
  echo "Error: GITHUB_OWNER is required." >&2
  exit 1
fi

read -r -p "GitHub repo name [${repo_default}]: " repo_input
repo="${repo_input:-$repo_default}"
if [[ -z "$repo" ]]; then
  echo "Error: GITHUB_REPO is required." >&2
  exit 1
fi

visibility="${GITHUB_VISIBILITY:-private}"

if gh repo view "${owner}/${repo}" >/dev/null 2>&1; then
  echo "Repo exists: ${owner}/${repo}"
else
  echo "Creating repo ${owner}/${repo} (${visibility})..."
  gh repo create "${owner}/${repo}" --source=. --remote=origin --${visibility} --confirm
fi

if git remote get-url origin >/dev/null 2>&1; then
  git remote set-url origin "git@github.com:${owner}/${repo}.git"
else
  git remote add origin "git@github.com:${owner}/${repo}.git"
fi

registry="${REGISTRY:-quay.io}"
quay_namespace="${QUAY_NAMESPACE:-${PROJECT_SLUG:-}}"
image_name="${IMAGE_NAME:-${PROJECT_SLUG:-}}"
if [[ -z "$quay_namespace" || -z "$image_name" ]]; then
  read -r -p "Quay namespace: " quay_namespace
  read -r -p "Image name: " image_name
fi

deploy_host="${DEPLOY_SSH_HOST:-}"
deploy_user="${DEPLOY_SSH_USER:-deploy}"
deploy_path="${DEPLOY_SSH_PATH:-/opt/${PROJECT_SLUG:-app}}"
deploy_mode="${DEPLOY_SSH_MODE:-ssh}"

read -r -p "Deploy SSH host: " deploy_host
read -r -p "Deploy SSH user [${deploy_user}]: " deploy_user_input
deploy_user="${deploy_user_input:-$deploy_user}"
read -r -p "Deploy SSH path [${deploy_path}]: " deploy_path_input
deploy_path="${deploy_path_input:-$deploy_path}"
read -r -p "Deploy SSH mode (ssh/cloudflare_access_ssh) [${deploy_mode}]: " deploy_mode_input
deploy_mode="${deploy_mode_input:-$deploy_mode}"

if [[ -z "$deploy_host" ]]; then
  echo "Error: DEPLOY_SSH_HOST is required." >&2
  exit 1
fi

# Set Actions variables
gh variable set REGISTRY --body "$registry" -R "${owner}/${repo}"
gh variable set IMAGE_NAME --body "${quay_namespace}/${image_name}" -R "${owner}/${repo}"
gh variable set DEPLOY_SSH_HOST --body "$deploy_host" -R "${owner}/${repo}"
gh variable set DEPLOY_SSH_USER --body "$deploy_user" -R "${owner}/${repo}"
gh variable set DEPLOY_SSH_PATH --body "$deploy_path" -R "${owner}/${repo}"
gh variable set DEPLOY_SSH_MODE --body "$deploy_mode" -R "${owner}/${repo}"

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 24
  else
    python - <<'PY'
import secrets
print(secrets.token_urlsafe(24))
PY
  fi
}

DB_PASSWORD="${DB_PASSWORD:-$(generate_secret)}"
REDIS_PASSWORD="${REDIS_PASSWORD:-$(generate_secret)}"

tmpdir="$(mktemp -d)"
key_path="${tmpdir}/deploy_key"
ssh-keygen -t ed25519 -N "" -f "$key_path" >/dev/null

read -r -p "Quay username: " quay_user
read -s -r -p "Quay password/token: " quay_pass
printf "\n"

read -r -p "Email from address: " email_from
default_from_name="${PROJECT_NAME:-}"
read -r -p "Email from name [${default_from_name}]: " email_from_name_input
email_from_name="${email_from_name_input:-$default_from_name}"
if [[ -z "$email_from" || -z "$email_from_name" ]]; then
  echo "Error: Email from address and name are required." >&2
  exit 1
fi
read -r -p "Email provider key (e.g. RESEND_API_KEY): " email_provider_key
if [[ -z "$email_provider_key" ]]; then
  echo "Error: Email provider key is required." >&2
  exit 1
fi

read -r -p "Enable Codecov? (y/N): " codecov_choice
codecov_token=""
if [[ "$codecov_choice" =~ ^[Yy]$ ]]; then
  read -r -p "Codecov token: " codecov_token
fi

# Set secrets
gh secret set QUAY_USERNAME --body "$quay_user" -R "${owner}/${repo}"
gh secret set QUAY_PASSWORD --body "$quay_pass" -R "${owner}/${repo}"
gh secret set DB_PASSWORD --body "$DB_PASSWORD" -R "${owner}/${repo}"
gh secret set REDIS_PASSWORD --body "$REDIS_PASSWORD" -R "${owner}/${repo}"
gh secret set SSH_PRIVATE_KEY --body "$(cat "$key_path")" -R "${owner}/${repo}"
gh secret set EMAIL_FROM_ADDRESS --body "$email_from" -R "${owner}/${repo}"
gh secret set EMAIL_FROM_NAME --body "$email_from_name" -R "${owner}/${repo}"
gh secret set RESEND_API_KEY --body "$email_provider_key" -R "${owner}/${repo}"

if [[ -n "$codecov_token" ]]; then
  gh secret set CODECOV_TOKEN --body "$codecov_token" -R "${owner}/${repo}"
fi

cat <<EOF

Bootstrap complete.

Public deploy key (add to server ~/.ssh/authorized_keys):
$(cat "${key_path}.pub")

Next steps:
- Create/verify the Quay repo: quay.io/${quay_namespace}/${image_name}
- Install the deploy public key on the server and confirm SSH access.
- Push a commit to trigger CI.
EOF
