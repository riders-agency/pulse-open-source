#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT/contracts/proto"
GO_OUT="$ROOT/apps/api/internal/gen"
PY_OUT="$ROOT/apps/ai-agent/src/pulse_agent/proto"

PROTOC_VERSION="${PROTOC_VERSION:-libprotoc 3.12.4}"
PROTOC_GEN_GO_VERSION="${PROTOC_GEN_GO_VERSION:-protoc-gen-go v1.36.11}"
PROTOC_GEN_GO_GRPC_VERSION="${PROTOC_GEN_GO_GRPC_VERSION:-protoc-gen-go-grpc 1.6.2}"
PYTHON_GRPC_TOOLS_PROTOC_VERSION="${PYTHON_GRPC_TOOLS_PROTOC_VERSION:-libprotoc 33.5}"

require_version() {
  local label="$1"
  local actual="$2"
  local expected="$3"

  if [ "$actual" != "$expected" ]; then
    printf '%s version mismatch: expected "%s", got "%s"\n' "$label" "$expected" "$actual" >&2
    exit 1
  fi
}

run_python_grpc_tools() {
  if [ -n "${PYTHON_PROTO_BIN:-}" ]; then
    "$PYTHON_PROTO_BIN" -m grpc_tools.protoc "$@"
    return
  fi

  if command -v uv >/dev/null 2>&1; then
    (cd "$ROOT/apps/ai-agent" && uv run --locked python -m grpc_tools.protoc "$@")
    return
  fi

  if [ -x "$ROOT/apps/ai-agent/.venv/bin/python" ]; then
    "$ROOT/apps/ai-agent/.venv/bin/python" -m grpc_tools.protoc "$@"
    return
  fi

  python3 -m grpc_tools.protoc "$@"
}

require_version "protoc" "$(protoc --version)" "$PROTOC_VERSION"
require_version "protoc-gen-go" "$(protoc-gen-go --version)" "$PROTOC_GEN_GO_VERSION"
require_version "protoc-gen-go-grpc" "$(protoc-gen-go-grpc --version)" "$PROTOC_GEN_GO_GRPC_VERSION"
require_version "grpc_tools.protoc" "$(run_python_grpc_tools --version)" "$PYTHON_GRPC_TOOLS_PROTOC_VERSION"

mkdir -p "$GO_OUT" "$PY_OUT"

protoc -I "$PROTO_DIR" \
  --go_out="$GO_OUT" --go_opt=paths=source_relative \
  "$PROTO_DIR"/ai_agent/v1/*.proto

protoc -I "$PROTO_DIR" \
  --go-grpc_out="$GO_OUT" --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR"/ai_agent/v1/*.proto

run_python_grpc_tools -I "$PROTO_DIR" \
  --python_out="$PY_OUT" --grpc_python_out="$PY_OUT" --pyi_out="$PY_OUT" \
  "$PROTO_DIR"/ai_agent/v1/*.proto
