#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v gcloud >/dev/null 2>&1; then
  echo "gcloud is required" >&2
  exit 1
fi

PROJECT_ID="${GOOGLE_CLOUD_PROJECT:-${GCP_PROJECT_ID:-}}"
REGION="${GOOGLE_CLOUD_REGION:-${GCP_REGION:-asia-northeast1}}"
ARTIFACT_REPO="${ARTIFACT_REGISTRY_REPOSITORY:-apps}"
SERVICE_NAME="${CLOUD_RUN_SERVICE:-konbu}"
ENV_FILE="${RUNTIME_ENV_FILE:-$ROOT_DIR/deploy/google/konbu.runtime.env}"
DATABASE_URL_SECRET_NAME="${DATABASE_URL_SECRET_NAME:-konbu-database-url}"
SMTP_PASSWORD_SECRET_NAME="${SMTP_PASSWORD_SECRET_NAME:-}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "GOOGLE_CLOUD_PROJECT or GCP_PROJECT_ID is required." >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Runtime env file not found: $ENV_FILE" >&2
  echo "Copy deploy/google/konbu.runtime.env.example to deploy/google/konbu.runtime.env and fill in values." >&2
  exit 1
fi

IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${ARTIFACT_REPO}/${SERVICE_NAME}:latest"
DEPLOY_ENV_FILE="$(mktemp)"

cd "$ROOT_DIR"

python3 - "$ENV_FILE" "$DEPLOY_ENV_FILE" "$DATABASE_URL_SECRET_NAME" "$SMTP_PASSWORD_SECRET_NAME" <<'PY'
import json
import sys

src, dst, database_secret_name, smtp_password_secret_name = sys.argv[1:]
env = {}

with open(src, encoding="utf-8") as f:
    for row_num, raw_line in enumerate(f, start=1):
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            print(
                f"warning: skipping invalid runtime env line {row_num} (length={len(line)})",
                file=sys.stderr,
            )
            continue
        key, value = line.split("=", 1)
        if key == "DATABASE_URL" and database_secret_name:
            continue
        if key == "SMTP_PASSWORD" and smtp_password_secret_name:
            continue
        env[key] = value

print(f"runtime env keys: {sorted(env.keys())}", file=sys.stderr)
with open(dst, "w", encoding="utf-8") as f:
    json.dump(env, f)
PY

gcloud builds submit \
  --project "$PROJECT_ID" \
  --config deploy/google/cloudbuild.konbu.yaml \
  --substitutions=_IMAGE="$IMAGE_URI"

DEPLOY_CMD=(
  gcloud run deploy "$SERVICE_NAME"
  --project "$PROJECT_ID"
  --region "$REGION"
  --image "$IMAGE_URI"
  --platform managed
  --allow-unauthenticated
  --port 8080
  --env-vars-file "$DEPLOY_ENV_FILE"
)

SECRETS_ARGS=""
if [[ -n "$DATABASE_URL_SECRET_NAME" ]]; then
  SECRETS_ARGS+="DATABASE_URL=${DATABASE_URL_SECRET_NAME}:latest,"
fi
if [[ -n "$SMTP_PASSWORD_SECRET_NAME" ]]; then
  SECRETS_ARGS+="SMTP_PASSWORD=${SMTP_PASSWORD_SECRET_NAME}:latest,"
fi
SECRETS_ARGS="${SECRETS_ARGS%,}"
if [[ -n "$SECRETS_ARGS" ]]; then
  DEPLOY_CMD+=(--set-secrets "$SECRETS_ARGS")
fi

"${DEPLOY_CMD[@]}"
