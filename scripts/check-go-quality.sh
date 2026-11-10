#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="$ROOT_DIR/apps/api"

STATICCHECK_PACKAGE="${GO_QUALITY_STATICCHECK_PACKAGE:-honnef.co/go/tools/cmd/staticcheck@v0.7.0}"
GOVULNCHECK_PACKAGE="${GO_QUALITY_GOVULNCHECK_PACKAGE:-golang.org/x/vuln/cmd/govulncheck@v1.4.0}"
FUZZTIME="${GO_QUALITY_FUZZTIME:-5s}"
STATICCHECK_CHECKS="${GO_QUALITY_STATICCHECK_CHECKS:-all,-ST1000,-ST1003}"

RACE_PACKAGES="${GO_QUALITY_RACE_PACKAGES:-./internal/executor/... ./internal/platform/outbox/... ./internal/runtime/... ./internal/modules/auth/... ./internal/modules/sitechat/... ./internal/modules/clients/runtime/... ./internal/modules/httpclient/runtime/... ./internal/modules/providers/runtime/... ./internal/platform/httptransport/...}"
FUZZ_PACKAGES="${GO_QUALITY_FUZZ_PACKAGES:-./internal/executor ./internal/modules/clients/repository ./internal/modules/ingest/transport/http ./internal/modules/person/fio ./internal/modules/person/phone ./internal/modules/workflows/definition ./internal/modules/workflows/definitionvalidate ./internal/modules/workflows/paymentworkflow}"

cd "$API_DIR"

GO_TOOLCHAIN="${GO_QUALITY_GOTOOLCHAIN:-$(awk '$1 == "toolchain" { print $2 }' go.mod)}"
if [ -z "$GO_TOOLCHAIN" ]; then
  GO_TOOLCHAIN="$(go env GOVERSION)"
fi
export GOTOOLCHAIN="$GO_TOOLCHAIN"

echo "==> gofmt"
files="$(gofmt -l .)"
if [ -n "$files" ]; then
  printf 'Go files need gofmt:\n%s\n' "$files" >&2
  exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test ./...

echo "==> go test -race"
# shellcheck disable=SC2086
go test -race $RACE_PACKAGES

echo "==> go fuzz smoke (${FUZZTIME})"
for package in $FUZZ_PACKAGES; do
  go test -run=^$ -fuzz=Fuzz -fuzztime="$FUZZTIME" "$package"
done

echo "==> staticcheck"
GOTOOLCHAIN="$GO_TOOLCHAIN" go run "$STATICCHECK_PACKAGE" -checks="$STATICCHECK_CHECKS" ./...

echo "==> govulncheck"
GOTOOLCHAIN="$GO_TOOLCHAIN" go run "$GOVULNCHECK_PACKAGE" ./...

echo "==> go build commands"
go build ./cmd/...
