#!/usr/bin/env bash
set -euo pipefail

if [[ -f ".template-initialized" ]]; then
  echo "Template already initialized (.template-initialized present)."
  exit 1
fi

if ! rg -q "__TEMPLATE_PROJECT_NAME__" .; then
  echo "Template appears initialized (no template tokens found)."
  exit 1
fi

is_interactive=false
if [[ -t 0 ]]; then
  is_interactive=true
fi

prompt_value() {
  local var_name="$1"
  local prompt="$2"
  local default_value="$3"
  local value="${!var_name:-}"

  if [[ -z "$value" ]]; then
    if [[ "$is_interactive" == true ]]; then
      if [[ -n "$default_value" ]]; then
        read -r -p "${prompt} [${default_value}]: " value
        value="${value:-$default_value}"
      else
        read -r -p "${prompt}: " value
      fi
    fi
  fi

  if [[ -z "$value" ]]; then
    echo "Error: ${var_name} is required." >&2
    exit 1
  fi

  export "$var_name"="$value"
}

prompt_value PROJECT_NAME "Project name" "Notes App"
prompt_value PROJECT_SLUG "Project slug (used in paths and compose)" "notes-app"
prompt_value GO_MODULE "Go module path (e.g. github.com/acme/app)" "github.com/acme/${PROJECT_SLUG}"
prompt_value APP_BASE_URL "App base URL" "http://localhost:8080"
prompt_value EMAIL_FROM_ADDRESS "Email from address" "noreply@example.com"
EMAIL_FROM_NAME="${EMAIL_FROM_NAME:-$PROJECT_NAME}"
export EMAIL_FROM_NAME

QUAY_NAMESPACE="${QUAY_NAMESPACE:-$PROJECT_SLUG}"
IMAGE_NAME="${IMAGE_NAME:-$PROJECT_SLUG}"
DEPLOY_SSH_HOST="${DEPLOY_SSH_HOST:-}"
DEPLOY_SSH_USER="${DEPLOY_SSH_USER:-}"
DEPLOY_SSH_PATH="${DEPLOY_SSH_PATH:-/opt/${PROJECT_SLUG}}"
DEPLOY_SSH_MODE="${DEPLOY_SSH_MODE:-ssh}"
DEPLOY_SSH_PUBLIC_KEY="${DEPLOY_SSH_PUBLIC_KEY:-}"

if [[ "$is_interactive" == true && -z "$DEPLOY_SSH_PUBLIC_KEY" ]]; then
  read -r -p "Deploy SSH public key for cloud-init (optional): " DEPLOY_SSH_PUBLIC_KEY
fi

export QUAY_NAMESPACE IMAGE_NAME DEPLOY_SSH_HOST DEPLOY_SSH_USER DEPLOY_SSH_PATH DEPLOY_SSH_MODE DEPLOY_SSH_PUBLIC_KEY

replace_token() {
  local token="$1"
  local value="$2"

  rg -l "$token" \
    -g '!yearofbingo/**' \
    -g '!.git/**' \
    -g '!scripts/template-init.sh' | while read -r file; do
    python - "$file" "$token" "$value" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
token = sys.argv[2]
value = sys.argv[3]

data = path.read_text(encoding='utf-8')
path.write_text(data.replace(token, value), encoding='utf-8')
PY
  done
}

replace_token "__TEMPLATE_PROJECT_NAME__" "$PROJECT_NAME"
replace_token "__TEMPLATE_PROJECT_SLUG__" "$PROJECT_SLUG"
replace_token "__TEMPLATE_GO_MODULE__" "$GO_MODULE"
replace_token "__TEMPLATE_APP_BASE_URL__" "$APP_BASE_URL"
replace_token "__TEMPLATE_EMAIL_FROM_ADDRESS__" "$EMAIL_FROM_ADDRESS"
replace_token "__TEMPLATE_EMAIL_FROM_NAME__" "$EMAIL_FROM_NAME"
replace_token "__TEMPLATE_QUAY_NAMESPACE__" "$QUAY_NAMESPACE"
replace_token "__TEMPLATE_IMAGE_NAME__" "$IMAGE_NAME"
replace_token "__TEMPLATE_DEPLOY_SSH_HOST__" "$DEPLOY_SSH_HOST"
replace_token "__TEMPLATE_DEPLOY_SSH_USER__" "$DEPLOY_SSH_USER"
replace_token "__TEMPLATE_DEPLOY_SSH_PATH__" "$DEPLOY_SSH_PATH"
replace_token "__TEMPLATE_DEPLOY_SSH_MODE__" "$DEPLOY_SSH_MODE"
replace_token "__TEMPLATE_DEPLOY_SSH_PUBLIC_KEY__" "$DEPLOY_SSH_PUBLIC_KEY"

# Update Go module path
old_module="github.com/example/notes-template"
if rg -q "$old_module" .; then
  rg -l "$old_module" \
    -g '!yearofbingo/**' \
    -g '!.git/**' \
    -g '!scripts/template-init.sh' | while read -r file; do
    python - "$file" "$old_module" "$GO_MODULE" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
old = sys.argv[2]
new = sys.argv[3]

data = path.read_text(encoding='utf-8')
path.write_text(data.replace(old, new), encoding='utf-8')
PY
  done
fi

# Mark initialized
cat <<EOF_MARKER > .template-initialized
PROJECT_NAME=${PROJECT_NAME}
PROJECT_SLUG=${PROJECT_SLUG}
GO_MODULE=${GO_MODULE}
APP_BASE_URL=${APP_BASE_URL}
EMAIL_FROM_ADDRESS=${EMAIL_FROM_ADDRESS}
EOF_MARKER

# Go module tidy (optional but recommended)
if command -v go >/dev/null 2>&1; then
  go mod tidy
fi

echo ""
echo "Template initialized successfully."
echo ""
echo "Next steps:"
cat <<EOF
- Create the remote GitHub repo (if not created) and set origin.
- Create the Quay repo: quay.io/${QUAY_NAMESPACE}/${IMAGE_NAME}
- Set GitHub Actions variables:
  - REGISTRY=quay.io
  - IMAGE_NAME=${QUAY_NAMESPACE}/${IMAGE_NAME}
  - DEPLOY_SSH_HOST=${DEPLOY_SSH_HOST}
  - DEPLOY_SSH_USER=${DEPLOY_SSH_USER}
  - DEPLOY_SSH_PATH=${DEPLOY_SSH_PATH}
  - DEPLOY_SSH_MODE=${DEPLOY_SSH_MODE}
- Add GitHub Actions secrets:
  - QUAY_USERNAME, QUAY_PASSWORD
  - SSH_PRIVATE_KEY
  - DB_PASSWORD, REDIS_PASSWORD
  - EMAIL_FROM_ADDRESS, EMAIL_FROM_NAME, email provider key(s)
  - Optional: CODECOV_TOKEN
- Update APP_BASE_URL and DNS if deploying to production.
- Provision the server (see docs/production.md).
- Update cloud-init.yaml with the deploy public key if you plan to use it.
- Run make test (and optionally make e2e).
- Commit initialized changes locally and push when ready.
EOF

if [[ "$is_interactive" == true ]]; then
  read -r -p "Run 'make test' now? (y/N): " run_tests
  if [[ "$run_tests" =~ ^[Yy]$ ]]; then
    make test
  fi

  read -r -p "Run 'make e2e' now? (y/N): " run_e2e
  if [[ "$run_e2e" =~ ^[Yy]$ ]]; then
    make e2e
  fi

  if command -v gh >/dev/null 2>&1; then
    read -r -p "Run 'make github-bootstrap' now? (y/N): " run_bootstrap
    if [[ "$run_bootstrap" =~ ^[Yy]$ ]]; then
      make github-bootstrap
    fi
  fi
fi
