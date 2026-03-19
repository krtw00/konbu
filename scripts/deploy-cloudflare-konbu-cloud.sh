#!/usr/bin/env bash
set -euo pipefail

npx wrangler deploy --config deploy/cloudflare/konbu-cloud/wrangler.jsonc
