#!/usr/bin/env bash
# Проверка входящего вебхука Pulse (тот же путь, что Bitrix/внешние системы).
# Пример:
#   export PULSE_BASE_URL=https://pulse.riders.agency
#   export PULSE_HOOK_KEY=…        # из UI проекта
#   export PULSE_ENDPOINT_KEY=bitrix-order-to-max
#   export PULSE_PROJECT_SECRET=… # открытый текст секрета (не hash)
#   ./scripts/smoke-webhook-ingest.sh
#
# Локально на API без nginx:
#   PULSE_BASE_URL=http://127.0.0.1:9810 ./scripts/smoke-webhook-ingest.sh

set -euo pipefail
BASE="${PULSE_BASE_URL:-https://pulse.riders.agency}"
HK="${PULSE_HOOK_KEY:?set PULSE_HOOK_KEY}"
EK="${PULSE_ENDPOINT_KEY:?set PULSE_ENDPOINT_KEY}"
SEC="${PULSE_PROJECT_SECRET:?set PULSE_PROJECT_SECRET}"

BODY='{"source":"smoke-script","order_id":"smoke-'$(date +%s)'","price":1,"currency":"RUB"}'

echo "POST $BASE/hooks/$HK/$EK"
exec curl -sv -X POST "$BASE/hooks/$HK/$EK" \
  -H "Content-Type: application/json" \
  -H "X-Project-Secret: $SEC" \
  -H "Idempotency-Key: smoke-ingest-$RANDOM" \
  --data-binary "$BODY"
