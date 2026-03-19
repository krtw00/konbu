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

cd "$ROOT_DIR"

gcloud builds submit \
  --project "$PROJECT_ID" \
  --config deploy/google/cloudbuild.konbu.yaml \
  --substitutions=_IMAGE="$IMAGE_URI"

gcloud run deploy "$SERVICE_NAME" \
  --project "$PROJECT_ID" \
  --region "$REGION" \
  --image "$IMAGE_URI" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --env-vars-file "$ENV_FILE"
