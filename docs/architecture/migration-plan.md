# Migration Plan

This document records the original incremental hardening order. Current module
boundaries are implemented and checked without exceptions; see
[the accepted decision](controlled-module-boundaries.md). It is not an open-issue list.

This is the incremental hardening order. Each phase must keep existing product
behavior working.

## Phase 1: Architecture Freeze

- Update target architecture documents.
- Add module, runtime, AI, worker, and contract boundary docs.
- Add checks for forbidden paths, secret shapes, AI sidecar DB access, migration
  pairing, and size budgets.

Exit criteria:

- docs describe all planes;
- `scripts/check.sh` passes;
- no product code path changed.

## Phase 2: Composition Root

- Introduce `platform/app`.
- Add shared `Deps` and `Module` contract.
- Keep `router.go` behavior unchanged while modules migrate.
- Add tests for module registration helpers.

Exit criteria:

- module contract compiles;
- old routes remain mounted;
- no route deletions.

## Phase 3: Runtime Registry

- Extend runtime registry to support triggers.
- Keep existing action registry behavior.
- Add duplicate/empty lookup tests for actions and triggers.
- Document legacy executor fallback.

Exit criteria:

- `go test ./internal/runtime/...` passes;
- no new action type is added to `executeActionLegacy`.

## Phase 4: Outbox + Worker

- Add `outbox_events` migration.
- Add `platform/outbox` repository, registry, and worker loop.
- Add `apps/api/cmd/worker`.
- Keep registry empty until domain modules register handlers.

Exit criteria:

- migration has up/down files;
- worker can start with no handlers and no pending rows;
- claim uses `FOR UPDATE SKIP LOCKED`.

## Phase 5: Workflow Definition Contract

- Add `riders-pulse-workflow-definition-v1` normalizer.
- Support legacy definitions through compatibility normalization.
- Add validator tests for v1 and legacy payloads.
- Publish path later switches to this one validator.

Exit criteria:

- v1 and legacy definitions normalize to the same internal shape;
- invalid settings fail deterministic validation.

## Phase 6: Module Migration

Move modules one at a time:

1. AI;
2. workflows;
3. runs;
4. action dependencies;
5. clients/worktracker.

Exit criteria per module:

- HTTP routes are still identical;
- runtime handlers register through module contract;
- outbox handlers register through module contract where needed;
- repository stops owning orchestration.

## Phase 7: AI Control Plane

- Keep Browser -> Go only.
- Keep Python orchestration-only.
- Make Go ToolGateway mandatory for tools.
- Keep approvals and writes in Go.
- Add service-level 30-scenario runner as acceptance gate.

Exit criteria:

- all 30 scenario catalog checks pass;
- live runner passes when project/auth env is available;
- write tools require approval and idempotency.

## Phase 8: UI Feature Split

- Move workflow editor model/API/UI logic into `features/workflow-editor`.
- Move AI employee stream/cards/model/API logic into `features/ai-employee`.
- Keep route pages as thin adapters.
- Do not perform the global visual style/token/CSS architecture refactor in this
  phase.

Exit criteria:

- typecheck/lint pass;
- Playwright smoke passes for touched flows;
- no page-level file grows further.

## Phase 9: Contract Gates

- Add route inventory drift checks.
- Add frontend API path inventory checks.
- Add proto generation check.
- Add migration up/down check when disposable DB is available.

Exit criteria:

- `scripts/check.sh` is a real merge gate;
- drift failures are actionable and deterministic.

## Phase 10: Final Architecture Audit

- Scan key Go, Python, and web files for line-count hotspots.
- Ratchet oversized files through `.vibe/size-budgets.tsv`.
- Check for non-modular data paths: Python business SQL, AI direct writes outside
  Go ToolGateway, legacy executor action switch growth, and domain follow-up work
  that bypasses outbox.
- Fix concrete findings before the final deploy, not just document them.

Exit criteria:

- `scripts/check.sh` passes with size budgets, redaction keys, workflow-editor
  route/module boundary guard, route inventory drift, frontend API path drift,
  and migration up/down checks;
- workflow editor route remains a thin page adapter;
- workflow editor feature files stay under ratcheted budgets;
- remaining large legacy files are listed in budgets or split before additional
  behavior is added;
- production deploy includes API, Web, and Worker service verification.

Style architecture follow-up:

- global CSS, token extraction, repeated visual constants, and inline visual
  cleanup are handled in the next style-focused iteration.

## Rollback Rules

- Every phase should be revertable independently.
- Legacy route/executor paths remain until replacement tests are green.
- No data migration should remove legacy columns/tables before a full release
  cycle with compatibility reads.
