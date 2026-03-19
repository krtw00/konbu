#!/usr/bin/env bash
set -euo pipefail

PROJECT_ID="${GOOGLE_CLOUD_PROJECT:-${GCP_PROJECT_ID:-}}"
CONFIG_FILE="${FIREBASE_CONFIG_FILE:-deploy/firebase/konbu-cloud.hosting.json}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "GOOGLE_CLOUD_PROJECT or GCP_PROJECT_ID is required." >&2
  exit 1
fi

npx firebase-tools deploy \
  --project "$PROJECT_ID" \
  --config "$CONFIG_FILE" \
  --only hosting
