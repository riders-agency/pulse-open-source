#!/usr/bin/env bash
# Смоук MAX Bot API. Только переменные окружения — токены в git не попадают.
#
# export MAX_BOT_TOKEN='...'
# export MAX_CHAT_ID='-74739869656479'
# chmod +x scripts/smoke-max-send.sh && ./scripts/smoke-max-send.sh

set -euo pipefail

BASE="${MAX_API_BASE:-https://platform-api.max.ru}"
CHAT="${MAX_CHAT_ID:-}"
TOKEN="${MAX_BOT_TOKEN:-}"
MSG="${MAX_MESSAGE:-Riders Pulse: smoke через MAX Bot API}"

if [[ -z "$TOKEN" ]]; then
  echo "Нужно: export MAX_BOT_TOKEN" >&2
  exit 1
fi

if [[ -z "$CHAT" ]]; then
  echo "Нужно: export MAX_CHAT_ID" >&2
  exit 1
fi

# Если прислали ссылку — берём последний сегмент пути.
if [[ "$CHAT" == *"/"* ]]; then
  CHAT="${CHAT##*/}"
  CHAT="${CHAT%%\?*}"
fi

AUTH_MODE="${MAX_AUTHORIZATION_SCHEME:-raw}" # raw | bearer

TRIM_TOKEN="$TOKEN"
shopt -s nocasematch
if [[ "$TOKEN" == bearer\ * ]]; then
  TRIM_TOKEN="${TOKEN#*[Bb][Ee][Aa][Rr][[:space:]]}"
fi
shopt -u nocasematch

AUTH="$TRIM_TOKEN"
if [[ "${AUTH_MODE}" == "bearer" ]] || [[ "${AUTH_MODE}" == "Bearer" ]]; then
  AUTH="Bearer ${TRIM_TOKEN}"
fi

if command -v jq >/dev/null 2>&1; then
  PAYLOAD=$(jq -nc --arg t "$MSG" '{text:$t}')
else
  PAYLOAD=$(python3 -c 'import json,os,sys; print(json.dumps({"text":sys.argv[1]}))' "$MSG")
fi

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

code=$(curl -sS -o "$TMP" -w "%{http_code}" \
  -X POST "${BASE%/}/messages?chat_id=${CHAT}" \
  -H "Content-Type: application/json" \
  -H "Authorization: ${AUTH}" \
  --data-binary "$PAYLOAD")

echo "HTTP ${code}"
cat "$TMP"

if [[ "${code}" -lt 200 ]] || [[ "${code}" -ge 300 ]]; then
  exit 1
fi

echo "--- OK ---"
