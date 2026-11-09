<!-- Generated from registry/standards/frontend/next@1.1.0. Update the registry standard, then regenerate. -->

# Next.js Frontend Engineering Standard

This standard applies to `apps/web` when the active stack uses Next.js.

UI density, shadcn usage, color policy and compact layout rules live in `ui/shadcn-compact`. Keep this standard focused on framework architecture, runtime boundaries, contracts and testability.

## May 2026 Baseline

1. Next.js baseline: `16.x`.
2. React baseline: `19.2.x`.
3. Node.js production baseline: `24.x LTS`.
4. TypeScript baseline: `6.0.x`.
5. ESLint baseline: `10.x`.
6. Server Components are the default for route data-loading views.

## Stack Contract

1. `NEXT_MAJOR = 16`
2. `REACT_MAJOR = 19`
3. `NODE_MAJOR = 24`
4. `TYPESCRIPT_MAJOR = 6`
5. `SERVER_COMPONENT_DEFAULT_PERCENT = 100`

## App Router Boundaries

1. Route files stay thin and delegate to feature components or server functions.
2. Server Components are the default for data-loading views.
3. Client Components require an interaction reason.
4. Server actions stay narrow and typed.
5. Direct access to `process.env` is allowed only in configuration modules.
6. Route handlers do not become a second backend unless explicitly documented.
7. Cache behavior is named at the data boundary.

## Next 16 Runtime Notes

1. Turbopack is the default build/dev path unless the project documents why not.
2. The removed framework-provided lint script is not the standard command; use the ESLint CLI.
3. Middleware/proxy conventions must follow the active Next.js version.
4. React Compiler may be enabled only with an explicit compatibility note.
5. Image, cache and prefetch behavior should be treated as runtime contract, not styling.

## Contract-First Frontend

1. API calls go through a shared typed client generated or maintained from OpenAPI.
2. Feature code must not invent hidden response shapes.
3. Error states use the shared problem/error model.
4. Request IDs from API responses must remain visible to observability and support flows.
5. Generated clients are regenerated when OpenAPI changes.
6. Frontend validation mirrors API constraints but does not replace backend validation.

## State And Data

1. Server state belongs in server components, actions or a shared client cache.
2. Client state has one owner.
3. Do not duplicate domain state across local component state and remote cache.
4. Loading, empty, error and permission-denied states are part of the feature.
5. Optimistic updates require rollback behavior.

## Reuse-First Rules

1. `REUSE_REQUIRED_PERCENT = 100` for repeated UI patterns.
2. `DUPLICATE_FEATURE_COMPONENTS_ALLOWED = 0`.
3. `DIRECT_FETCH_OUTSIDE_SHARED_API_CLIENT = 0`.
4. `BUNDLER_ABSTRACTION_COUNT = 1`.
5. `UNEXPLAINED_USE_CLIENT = 0`.

## File And Component Limits

1. `MAX_TS_FILE_LINES = 260`
2. `MAX_COMPONENT_LINES = 70`
3. `MAX_FUNCTION_LINES = 45`
4. `MAX_HOOK_LINES = 55`
5. `MAX_PAGE_FILE_LINES = 120`
6. `MAX_SERVER_ACTION_LINES = 60`
7. `MAX_COMPONENT_PROPS = 10`
8. `MAX_FUNCTION_PARAMS = 4`
9. `MAX_NESTING_DEPTH = 3`
10. `MAX_CYCLOMATIC_COMPLEXITY = 10`
11. `MAX_COMMENT_RATIO_PERCENT = 5`
12. `MAX_UNTESTED_CHANGED_FILES = 0`

## Performance Budgets

1. `WEB_VITALS_LCP_MS_P75_MAX = 2500`
2. `WEB_VITALS_INP_MS_P75_MAX = 200`
3. `WEB_VITALS_CLS_P75_MAX = 0.1`
4. `JS_PAYLOAD_KB_GZIP_INITIAL_MAX = 170`
5. `CRITICAL_CSS_KB_MAX = 35`
6. `ABOVE_THE_FOLD_IMAGE_LAZYLOAD = 0`
7. `NON_CRITICAL_IMAGE_LAZYLOAD = 100%`
8. `USE_CLIENT_WITHOUT_REASON = 0`
9. Server-rendered routes should not ship avoidable client JavaScript.
10. Route-level performance evidence is required for UX-critical pages when runtime checks exist.

## Accessibility

1. WCAG 2.2 AA is the target.
2. Keyboard operation is required for interactive flows.
3. Focus states must be visible and not obscured.
4. Form errors must be programmatically associated with fields.
5. Authentication must avoid cognitive puzzles as the only path.

## Test Gates

1. Unit/component pass rate: `100%`.
2. Integration pass rate: `100%`.
3. E2E smoke pass rate: `100%`.
4. Contract compatibility check pass rate: `100%`.
5. Performance budget check pass rate: `100%`.

## Automated Checks

1. `scripts/check.sh` must call the active frontend checks when the project has runnable frontend code.
2. Until runtime checks exist, enforcement must be marked as documented only.

## Enforcement Reality

This standard is documented by default. It becomes linted/tested/CI-blocking only when the generated project includes real Next.js, ESLint, typecheck, component, e2e and performance checks.
