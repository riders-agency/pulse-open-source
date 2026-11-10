#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT/contracts/proto"
DESCRIPTOR_BASELINE="$PROTO_DIR/descriptors/ai_agent_v1.binpb"
BUF_VERSION="${BUF_VERSION:-1.71.0}"

require_version() {
  local label="$1"
  local actual="$2"
  local expected="$3"

  if [ "$actual" != "$expected" ]; then
    printf '%s version mismatch: expected "%s", got "%s"\n' "$label" "$expected" "$actual" >&2
    exit 1
  fi
}

if command -v buf >/dev/null 2>&1; then
  has_buf=1
  require_version "buf" "$(buf --version)" "$BUF_VERSION"
  (cd "$PROTO_DIR" && buf format --diff --exit-code)
  buf lint "$PROTO_DIR"
  if [ -n "${PROTO_BREAKING_AGAINST:-}" ]; then
    buf breaking "$PROTO_DIR" --against "$PROTO_BREAKING_AGAINST"
  fi
elif [ "${REQUIRE_BUF:-}" = "1" ]; then
  echo "buf is required for CI proto formatting, lint, and breaking checks" >&2
  exit 1
else
  has_buf=0
  echo "buf not found; skipped proto format/lint/breaking checks outside CI" >&2
fi

generated_snapshot="$(mktemp)"
generated_snapshot_after="$(mktemp)"
tmp_descriptor="$(mktemp)"
cleanup() {
  rm -f "$tmp_descriptor" "$generated_snapshot" "$generated_snapshot_after"
}
trap cleanup EXIT

snapshot_generated() {
  find "$ROOT/apps/api/internal/gen" "$ROOT/apps/ai-agent/src/pulse_agent/proto" -type f -print0 \
    | sort -z \
    | xargs -0 sha256sum
}

snapshot_generated > "$generated_snapshot"

"$ROOT/scripts/gen-proto.sh"
if [ "${ALLOW_DIRTY_PROTO:-}" = "1" ]; then
  snapshot_generated > "$generated_snapshot_after"
  diff -u "$generated_snapshot" "$generated_snapshot_after"
else
  git -C "$ROOT" diff --exit-code -- apps/api/internal/gen apps/ai-agent/src/pulse_agent/proto
fi

if [ "$has_buf" = "1" ]; then
  buf build "$PROTO_DIR" -o "$tmp_descriptor"
else
  echo "buf not found; skipped descriptor baseline drift check outside CI" >&2
  exit 0
fi

cmp -s "$tmp_descriptor" "$DESCRIPTOR_BASELINE" || {
  echo "proto descriptor baseline drift: run scripts/check-proto-contracts.sh after updating contracts/proto/descriptors/ai_agent_v1.binpb intentionally" >&2
  exit 1
}
