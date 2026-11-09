<!-- Generated from registry/standards/backend/go@1.1.0. Update the registry standard, then regenerate. -->

# Go Backend Engineering Standard

This standard applies to `apps/api` when the active stack uses Go.

## May 2026 Baseline

1. Go baseline: `1.26.x`.
2. Use Go modules and pin the toolchain in `go.mod`.
3. Use `context.Context` through request, service and repository boundaries.
4. Use structured logging through `log/slog` or the stack-approved adapter.
5. Run `go test ./...`; add race/fuzz/vulnerability checks when runtime maturity allows it.

## Canonical Structure

```text
apps/api/
├─ cmd/api/main.go
├─ internal/platform/httpserver/router.go
├─ internal/platform/httpserver/middleware/*.go
├─ internal/platform/problem/writer.go
├─ internal/platform/db/*.go
├─ internal/modules/auth/module.go
├─ internal/modules/auth/transport/http/routes.go
├─ internal/modules/auth/transport/http/*_handler.go
├─ internal/modules/auth/service/*.go
├─ internal/modules/auth/repository/*.go
├─ internal/modules/company/module.go
├─ internal/modules/company/transport/http/routes.go
├─ internal/modules/company/transport/http/*_handler.go
├─ internal/modules/company/service/*.go
└─ internal/modules/company/repository/*.go
```

## Numeric Limits

1. `MAX_GO_FILE_LINES = 250`
2. `MAX_FUNCTION_LINES = 40`
3. `MAX_HANDLER_LINES = 35`
4. `MAX_ROUTES_FILE_LINES = 100`
5. `MAX_FUNCTION_PARAMS = 4`
6. `MAX_FUNCTION_RETURNS = 3`
7. `MAX_NESTING_DEPTH = 3`
8. `MAX_SWITCH_CASES = 7`
9. `MAX_CYCLOMATIC_COMPLEXITY = 10`
10. `MAX_HANDLER_DB_CALLS = 0`
11. `MAX_HANDLER_REPOSITORY_CALLS = 0`
12. `MAX_HANDLER_SERVICE_CALLS = 1`
13. `MAX_HANDLER_RESPONSE_WRITES = 1`
14. `MAX_MODULE_ROUTES_PER_FILE = 12`
15. `MAX_MIDDLEWARE_PER_ROUTE = 6`
16. `MAX_COMMENT_RATIO_PERCENT = 5`
17. `MAX_PANIC_IN_REQUEST_PATH = 0`
18. `MAX_GLOBAL_MUTABLE_STATE = 0`
19. `MAX_TODO_IN_MAIN_BRANCH = 0`
20. `MAX_UNTESTED_CHANGED_FILES = 0`

These limits are documented unless the generated project includes a real limits checker.

## Router Decomposition

1. `GLOBAL_ROUTER_FILES = 1` (`internal/platform/httpserver/router.go`)
2. `MODULE_ROUTER_FILES_MIN = 1` per module
3. `ENDPOINTS_PER_HANDLER_FILE = 1`
4. `HANDLER_FILES_PER_MODULE_MIN = 1`
5. `ROUTES_WITHOUT_MODULE_OWNER = 0`

## Handler Contract

1. Parse input: `1`
2. Validate input: `1`
3. Service call: `1`
4. Error mapping: `1`
5. Response write: `1`

Запрещено в handler:
1. SQL calls: `0`
2. Repository orchestration: `0`
3. Domain policy branches: `0`
4. Transaction control: `0`

Handlers map protocol to application calls. Business policy lives in services.

## Service Contract

1. HTTP imports in service package: `0`
2. DB driver imports in service package: `0`
3. Repository interfaces per service file: `<= 4`
4. Unit tests required per public service method: `>= 1`
5. Hidden goroutine launches without lifecycle ownership: `0`
6. Business errors must be typed or wrapped with stable sentinels.

Use `errors.Is`, `errors.As` and `errors.Join` where they make error handling clearer.

## Repository Contract

1. `context.Context` in public methods: `100%`
2. Non-UUID domain IDs: `0`
3. Direct HTTP dependency in repository: `0`
4. Integration tests per repository method: `>= 1`
5. SQL is parameterized: `100%`
6. Transactions are explicit and owned by a service/application boundary.

Prefer generated or typed query layers when they exist in the project. Do not add a new ORM by default.

## Concurrency And Runtime Safety

1. Request goroutines must be cancellable.
2. Background goroutines require a lifecycle owner and shutdown path.
3. Mutable package globals are forbidden.
4. Shared maps and caches require synchronization and tests.
5. Use `context.WithCancelCause` for long-running flows where cancellation reason matters.
6. Race-sensitive code needs `go test -race` evidence before merge.

## Security And Supply Chain

1. `govulncheck` is the preferred vulnerability gate when available.
2. Keep `GOTOOLCHAIN` or toolchain pinning explicit for reproducible builds.
3. Do not log tokens, passwords, session secrets or raw credentials.
4. External calls must set timeouts and handle retries deliberately.
5. File/path/network inputs need validation before use.

## Observability

1. Every request has `request_id`.
2. Logs are structured.
3. Errors include safe codes, not raw internals.
4. Traces propagate through service and repository calls.
5. Metrics use stable route templates, not raw URLs.

## Team Readability

1. Constructor per component: `1`
2. Hidden singleton mutable state: `0`
3. Public symbol count per file: `<= 12`
4. Distinct responsibilities per file: `1`

## Test Gates

1. Unit suite pass rate: `100%`
2. Integration suite pass rate: `100%`
3. Contract checks pass rate: `100%`
4. Migration up/down checks: `100%`
5. Critical auth/session/security tests pass: `100%`
6. Race test for concurrency-sensitive code: `100%`
7. Fuzz tests for parsers/normalizers where malformed input is a risk.

## Merge Gate

Feature work is not accepted when:

1. documented limits are knowingly violated without an ADR;
2. changed behavior has no test evidence;
3. handlers bypass services/repositories;
4. contract and runtime disagree;
5. secrets or unsafe logs are introduced.

## Enforcement Reality

This standard is documented by default. It becomes linted/tested/CI-blocking only when the generated project includes real Go checks such as formatting, vet, staticcheck, tests, race checks, limits checks or `govulncheck`.

## Related Standards

1. `standards/active/TESTING.md`
2. `standards/active/API.md`
3. `standards/active/DATABASE.md`
4. `standards/active/SECURITY.md`
5. `standards/active/OBSERVABILITY.md`
