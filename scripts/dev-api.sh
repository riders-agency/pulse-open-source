#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/apps/api"

DEV_HTTP_ADDR="${HTTP_ADDR:-:9820}"
DEV_MIGRATIONS_PATH="${MIGRATIONS_PATH:-$ROOT/db/migrations}"

# После `go install github.com/air-verse/air@v1.65.1` бинарь в $(go env GOPATH)/bin
export PATH="$(go env GOPATH 2>/dev/null)/bin:${PATH}"

# Опционально: те же переменные, что у systemd на сервере.
runtime_env="$ROOT/.runtime/env/riders-pulse.env"
if [[ -r "$runtime_env" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$runtime_env"
  set +a
fi

# Дев-дефолты поверх /etc (порты и путь миграций от корня репозитория).
# Скрипт `scripts/dev.sh` обычно задаёт HTTP_ADDR=:9820, чтобы не бить systemd на 9810.
export HTTP_ADDR="$DEV_HTTP_ADDR"
export MIGRATIONS_PATH="$DEV_MIGRATIONS_PATH"

if command -v air >/dev/null 2>&1; then
  exec air
fi
exec go run github.com/air-verse/air@v1.65.1
