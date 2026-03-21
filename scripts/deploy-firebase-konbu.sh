#!/usr/bin/env bash
set -euo pipefail

PROJECT_ID="${GOOGLE_CLOUD_PROJECT:-${GCP_PROJECT_ID:-}}"
CONFIG_FILE="${FIREBASE_CONFIG_FILE:-deploy/firebase/konbu-cloud.hosting.json}"
HOSTING_SITE="${FIREBASE_HOSTING_SITE:-}"
SERVICE_NAME="${FIREBASE_CLOUD_RUN_SERVICE:-${CLOUD_RUN_SERVICE:-konbu}}"
REGION="${GOOGLE_CLOUD_REGION:-${GCP_REGION:-asia-northeast1}}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "GOOGLE_CLOUD_PROJECT or GCP_PROJECT_ID is required." >&2
  exit 1
fi

if [[ -n "$HOSTING_SITE" ]]; then
  CONFIG_FILE="$(mktemp)"
  cat > "$CONFIG_FILE" <<EOF
{
  "hosting": {
    "site": "$HOSTING_SITE",
    "public": "proxy-public",
    "ignore": [
      "firebase.json",
      "**/.*",
      "**/node_modules/**"
    ],
    "rewrites": [
      {
        "source": "**",
        "run": {
          "serviceId": "$SERVICE_NAME",
          "region": "$REGION"
        }
      }
    ]
  }
}
EOF
fi

npx firebase-tools deploy \
  --project "$PROJECT_ID" \
  --config "$CONFIG_FILE" \
  --only hosting
