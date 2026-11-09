<!-- Generated from registry/standards/testing/tdd-strict@1.1.0. Update the registry standard, then regenerate. -->

# TDD Strict Quality Gate Standard

This standard applies to all enabled runtime components in a generated project.

Main principle: a feature is not accepted until behavior is specified, tested or otherwise verified with honest evidence.

## May 2026 Baseline

1. Contract-first changes need contract checks.
2. Backend changes need unit and integration evidence.
3. Frontend changes need component/integration/e2e evidence according to risk.
4. Security-sensitive changes need negative tests.
5. Generated templates must not claim checks that do not exist.

## Red Green Refactor

Use the cycle for behavior changes:

1. `Red`: write or identify the failing test/check.
2. `Green`: implement the smallest correct behavior.
3. `Refactor`: improve structure without changing behavior.
4. `Evidence`: record what passed.

If a check cannot exist yet, document that honestly in the status matrix or task notes.

## Backend Minimum

For enabled backend modules:

1. service/business rules have unit tests;
2. handler -> service -> repository paths have integration tests when persistence exists;
3. migrations are exercised in test databases when migration tooling exists;
4. auth/session/security invariants include negative tests;
5. each bug fix adds a regression test when practical.

## Frontend Minimum

For enabled web modules:

1. component tests cover rendering states;
2. integration tests cover user interactions;
3. e2e smoke covers critical paths;
4. API contracts use generated clients or contract mocks;
5. accessibility states are tested for critical flows.

## Contract Minimum

For OpenAPI-enabled stacks:

1. OpenAPI root and modules exist;
2. endpoint inventory exists;
3. changed endpoints update examples and schemas;
4. generated clients are refreshed when the project uses them;
5. compatibility impact is documented for breaking changes.

## Optional Feature Testing

Only enabled features add gates:

1. worker: job idempotency, retry and DLQ tests;
2. mobile: deep link, auth handoff, offline and crash smoke tests;
3. payments: webhook signature, duplicate event, reconciliation and live-mode guard tests;
4. desktop: platform-specific tests only after a desktop standard is enabled.

Disabled feature checks must not appear in the generated project quality gate.

## Universal Regression Account

When auth and tenant flows exist, maintain a non-production seed account for regression:

1. active user;
2. membership in at least two companies;
3. identities only for enabled providers;
4. multiple active sessions when session management exists.

Credentials live in CI secrets or a secrets manager, never in the repository.

## CI Quality Gate Shape

A mature stack should run, in order:

1. `verify-structure`;
2. `verify-contracts`;
3. `verify-backend` when backend runtime exists;
4. `verify-frontend` when frontend runtime exists;
5. `verify-db` when database runtime exists;
6. optional feature gates only for enabled features;
7. build only after verify gates;
8. deploy only after build and smoke.

## Merge Gate

A feature branch is not accepted when:

1. changed behavior has no evidence;
2. contract and runtime disagree;
3. security-sensitive negative paths are untested;
4. flaky tests are ignored without a quarantine note;
5. disabled feature checks are required by accident;
6. docs and generated artifacts are out of sync.

## Advanced Testing Options

Use when risk warrants it:

1. property tests for parsers and validators;
2. fuzz tests for untrusted input;
3. mutation testing for critical business rules;
4. Testcontainers or provider-managed ephemeral services for integration tests;
5. visual regression for complex UI surfaces;
6. load smoke for high-traffic endpoints.

## Definition Of Done

A feature is done when:

1. contract and code are synchronized;
2. tests/checks prove changed behavior;
3. docs are updated at the source of truth;
4. generated artifacts are refreshed;
5. skipped checks are named with reasons;
6. `scripts/check.sh` passes for the generated project.

## Enforcement Reality

This standard is documented by default. It becomes tested/CI-blocking only when generated projects include real runtime test commands. The base generated `scripts/check.sh` currently enforces structure and disabled-folder hygiene.
