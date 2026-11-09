#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REDOCLY_PACKAGE="${REDOCLY_PACKAGE:-@redocly/cli@2.34.0}"
OPENAPI_TS_PACKAGE="${OPENAPI_TS_PACKAGE:-openapi-typescript@7.13.0}"
SOURCE="$ROOT/contracts/openapi/openapi.root.yaml"
DIST="$ROOT/contracts/openapi/dist/openapi.yaml"
EMBEDDED="$ROOT/apps/api/internal/platform/httpserver/openapi_pulse_stub.yaml"
GENERATED_MODULE="$ROOT/contracts/openapi/modules/runtime-inventory.generated.yaml"
WEB_TYPES="$ROOT/apps/web/src/lib/api/generated/openapi-types.ts"
WEB_CLIENT="$ROOT/apps/web/src/lib/api/generated/openapi-client.ts"

tmp_bundle="$(mktemp --suffix=.yaml)"
tmp_root_before="$(mktemp)"
tmp_module_before="$(mktemp)"
tmp_web_types_before="$(mktemp)"
tmp_web_client_before="$(mktemp)"
cleanup() {
  rm -f "$tmp_bundle" "$tmp_root_before" "$tmp_module_before" "$tmp_web_types_before" "$tmp_web_client_before"
}
trap cleanup EXIT

cp "$SOURCE" "$tmp_root_before"
if [ -f "$GENERATED_MODULE" ]; then
  cp "$GENERATED_MODULE" "$tmp_module_before"
fi
if [ -f "$WEB_TYPES" ]; then
  cp "$WEB_TYPES" "$tmp_web_types_before"
fi
if [ -f "$WEB_CLIENT" ]; then
  cp "$WEB_CLIENT" "$tmp_web_client_before"
fi

(cd "$ROOT/apps/api" && go run ./cmd/sync-openapi-inventory --inventory ../../contracts/openapi/endpoints.inventory.tsv --root ../../contracts/openapi/openapi.root.yaml --module ../../contracts/openapi/modules/runtime-inventory.generated.yaml)

if [ "${ALLOW_DIRTY_OPENAPI:-0}" = "1" ]; then
  cmp -s "$tmp_root_before" "$SOURCE" || {
    echo "OpenAPI root drift after sync: commit regenerated contracts/openapi/openapi.root.yaml" >&2
    exit 1
  }
  cmp -s "$tmp_module_before" "$GENERATED_MODULE" || {
    echo "OpenAPI runtime supplement drift after sync: commit regenerated contracts/openapi/modules/runtime-inventory.generated.yaml" >&2
    exit 1
  }
else
  if [ -n "$(git -C "$ROOT" status --porcelain -- contracts/openapi/openapi.root.yaml contracts/openapi/modules/runtime-inventory.generated.yaml)" ]; then
    echo "OpenAPI runtime supplement drift: run apps/api/cmd/sync-openapi-inventory and commit the result" >&2
    exit 1
  fi
fi

pnpm --package="$REDOCLY_PACKAGE" dlx redocly lint "$SOURCE"
pnpm --package="$REDOCLY_PACKAGE" dlx redocly bundle "$SOURCE" --output "$tmp_bundle" >/dev/null

cmp -s "$tmp_bundle" "$DIST" || {
  echo "OpenAPI bundle drift: run pnpm --package=$REDOCLY_PACKAGE dlx redocly bundle contracts/openapi/openapi.root.yaml --output contracts/openapi/dist/openapi.yaml" >&2
  exit 1
}

cmp -s "$DIST" "$EMBEDDED" || {
  echo "Embedded OpenAPI drift: copy contracts/openapi/dist/openapi.yaml to apps/api/internal/platform/httpserver/openapi_pulse_stub.yaml after intentional contract changes" >&2
  exit 1
}

(cd "$ROOT/apps/api" && go run ./cmd/check-openapi-inventory --inventory ../../contracts/openapi/endpoints.inventory.tsv --openapi ../../contracts/openapi/dist/openapi.yaml)
pnpm --package="$OPENAPI_TS_PACKAGE" dlx openapi-typescript "$DIST" --output "$WEB_TYPES" --root-types
(cd "$ROOT/apps/api" && go run ./cmd/gen-openapi-web-client --openapi ../../contracts/openapi/dist/openapi.yaml --output ../web/src/lib/api/generated/openapi-client.ts)

if [ "${ALLOW_DIRTY_OPENAPI:-0}" = "1" ]; then
  cmp -s "$tmp_web_types_before" "$WEB_TYPES" || {
    echo "Generated OpenAPI TypeScript types drift: regenerate apps/web/src/lib/api/generated/openapi-types.ts" >&2
    exit 1
  }
  cmp -s "$tmp_web_client_before" "$WEB_CLIENT" || {
    echo "Generated OpenAPI web client drift: regenerate apps/web/src/lib/api/generated/openapi-client.ts" >&2
    exit 1
  }
else
  if [ -n "$(git -C "$ROOT" status --porcelain -- apps/web/src/lib/api/generated/openapi-types.ts apps/web/src/lib/api/generated/openapi-client.ts)" ]; then
    echo "Generated OpenAPI TypeScript artifacts drift: regenerate and commit apps/web/src/lib/api/generated" >&2
    exit 1
  fi
fi
