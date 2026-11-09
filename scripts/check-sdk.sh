#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
python3 sdk/check_contract.py
pnpm --dir sdk/typescript install --frozen-lockfile
pnpm --dir sdk/typescript test
PYTHONPATH=sdk/python/src python3 -m unittest discover -s sdk/python/tests -v
(cd sdk/generated/go && GOWORK=off go test ./... && GOWORK=off go vet ./...)
