#!/usr/bin/env bash
# API (Air / hot reload) + Next dev.
# По умолчанию :9820 и :9821 — чтобы не пересекаться с systemd (9810/9811).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

export HTTP_ADDR="${HTTP_ADDR:-:9820}"
DEV_WEB_PORT="${DEV_WEB_PORT:-9821}"
API_PORT="${HTTP_ADDR##*:}"
export NEXT_PUBLIC_RP_API_URL="${NEXT_PUBLIC_RP_API_URL:-http://127.0.0.1:${API_PORT}}"

cleanup() {
  local p
  for p in "${PIDS[@]:-}"; do
    kill "$p" 2>/dev/null || true
  done
}
trap cleanup INT TERM

PIDS=()
"$ROOT/scripts/dev-api.sh" &
PIDS+=($!)

(
  cd "$ROOT/apps/web"
  export NEXT_PUBLIC_RP_API_URL
  exec pnpm exec next dev -p "$DEV_WEB_PORT"
) &
PIDS+=($!)

wait
