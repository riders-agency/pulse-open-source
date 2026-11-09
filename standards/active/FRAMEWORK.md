<!-- Generated from registry/standards/framework/lifecycle@1.1.0. Update the registry standard, then regenerate. -->

# Framework Lifecycle Standard

This standard defines the delivery lifecycle for generated projects that use `vibecoding-base`.

The core idea is simple: coding agents are fast, but the project must give them rails. The rails are active standards, stack profile metadata, contracts, checks, generated docs and human checkpoints.

## May 2026 Baseline

Modern starter repositories should be registry-driven, not folder-copy-driven:

1. reusable standards live once in `registry/standards`;
2. stack profiles compose standards and templates;
3. generated projects contain only active context;
4. examples are reproducible artifacts;
5. readiness is shown as a matrix, not as vague "production ready" marketing.

## Principles

### 1. Active Context Before Code

Every non-trivial change starts by reading the generated project context:

1. `AGENTS.md`
2. relevant files in `standards/active/`
3. `.vibe/profile.yaml`
4. `.vibe/registry.lock` when changing stack composition or standard versions
5. existing contracts and migrations

If the agent does not know the active standards, it should not write code.

### 2. Scope Before Speed

Fast implementation is useful only inside a clear boundary. New product domains, architecture changes, data model changes and deployment changes need an explicit task note, ADR, or standard/profile update.

Default generated project scope is intentionally narrow:

1. selected backend;
2. selected frontend;
3. selected database;
4. selected contracts;
5. selected standards;
6. explicitly enabled optional features.

Everything else is opt-in through `vibe enable`.

### 3. Contract Before Runtime

The contract is the source of truth:

1. API behavior starts in `contracts/openapi`.
2. Data behavior starts in `db/migrations`.
3. UI behavior follows API contracts and selected UI standards.
4. SDKs follow generated or maintained contracts.

Runtime code should not invent hidden APIs.

### 4. Red Before Green

A task is not implementation-ready until the expected failure is clear:

1. failing unit test;
2. failing integration test;
3. failing e2e scenario;
4. failing contract coverage;
5. failing static gate;
6. documented reason why another gate is used.

### 5. Vertical Slice Over Horizontal Piles

Prefer one thin complete path over many unfinished layers.

Good:

```text
OpenAPI -> migration -> service -> handler -> test -> docs
```

Risky:

```text
many schemas -> many empty handlers -> no tests -> no behavior
```

### 6. Observable By Default

Every real feature should carry:

1. request ID propagation;
2. structured logs;
3. traces for request flow;
4. metrics for latency/errors/security-sensitive flows;
5. redaction for secrets and tokens.

### 7. Human Checkpoints For Dangerous Work

Humans must review before:

1. deployment;
2. schema/data migration with real data impact;
3. auth/session/security changes;
4. enabling integrations with write access;
5. destructive git operations;
6. secret rotation.

### 8. Public Template Hygiene

Generated projects and registry templates must stay reusable:

1. no real secrets;
2. no private hostnames or server paths;
3. no product-specific leftovers;
4. no committed `.env`;
5. no dependency vendor folders;
6. safe placeholder domains only.

### 9. Readiness Is Specific

Do not describe a stack as fully ready when its template, runtime, checks and docs have different maturity.

Use this matrix:

1. `template: ready | partial | missing`
2. `runtime: ready | partial | missing`
3. `checks: ready | partial | missing`
4. `docs: ready | partial | missing`

### 10. Standards Move Through SemVer

Standard updates must say whether they are:

1. patch: text clarity or non-breaking checks;
2. minor: new recommendations or optional checks;
3. major: stricter required behavior or incompatible structure.

## Lifecycle

Every meaningful task moves through this lifecycle:

```text
Idea
  -> Scope
  -> Contract
  -> Test plan
  -> Red
  -> Implementation
  -> Green
  -> Observability
  -> Docs
  -> Verify
  -> Review
  -> Merge
```

## Task Anatomy

A good task has:

1. title;
2. scope;
3. non-goals;
4. files/modules likely to change;
5. contract impact;
6. data impact;
7. tests/gates;
8. security impact;
9. docs impact;
10. acceptance criteria.

Example:

```md
# Task: Auth login vertical slice

Scope:
- apps/api
- contracts/openapi
- db/migrations
- standards/active

Non-goals:
- OAuth providers
- UI polish
- mobile

Acceptance:
- login endpoint follows OpenAPI contract
- token stored only as hash
- request_id is returned and logged
- integration test covers success and invalid password
- scripts/check.sh passes
```

## Generated Docs Rule

Do not edit generated `standards/active/*` as the source of truth. Update the matching `registry/standards/<id>/standard.md`, then regenerate the project.

`vibe verify` should fail when a generated active standard drifts from the registry.

## Registry Change Anatomy

A standards change is complete only when:

1. `standard.md` reflects the rule;
2. `standard.yaml` version and enforcement are truthful;
3. `CHANGELOG.md` explains the change;
4. affected stack profiles still resolve;
5. examples are regenerated;
6. lockfiles pin the resolved versions;
7. `vibe verify` passes.

## Definition Of Ready

A task is ready when:

1. the scope is narrow;
2. the owner understands the active standards;
3. API/data impact is known;
4. security impact is known;
5. expected validation is listed;
6. non-goals are explicit.

## Definition Of Done

A task is done when:

1. implementation matches the contract;
2. tests or gates prove behavior;
3. docs are updated at the source of truth;
4. generated artifacts are refreshed;
5. hygiene checks still pass;
6. unresolved risks are documented;
7. human checkpoints were completed where required.

## Public Release Gate

Before publishing the repository:

1. no real secrets or private endpoints;
2. no old copyable framework folders;
3. no generated-example drift;
4. no duplicate standards as separate sources of truth;
5. status/readiness matrix is current;
6. examples regenerate cleanly;
7. verification commands are documented.

## Standard Workflow For A New Endpoint

1. Add endpoint to `contracts/openapi/modules/<module>.yaml`.
2. Add or update schemas in `contracts/openapi/components/schemas`.
3. Add row to `contracts/openapi/endpoints.inventory.tsv`.
4. Add examples and code samples.
5. Add backend handler/service tests.
6. Implement backend service and handler.
7. Add observability.
8. Run `scripts/check.sh`.
9. Update registry docs if the pattern creates a reusable rule.

## Standard Workflow For A New Table

1. Read `standards/active/DATABASE.md`.
2. Add append-only up/down migrations.
3. Use UUID entity IDs and canonical FK names.
4. Add versioning tables/triggers for business entities when required.
5. Run the active project checks.
6. Update docs if the data model changes.

## Standard Workflow For UI

1. Read `standards/active/FRONTEND.md`.
2. Read `standards/active/UI.md` when enabled.
3. Reuse shared components and API clients.
4. Do not duplicate backend validation.
5. Use selected design tokens/profile.
6. Add component/integration/e2e coverage for changed flows where runtime exists.
7. Run `scripts/check.sh`.

## Enforcement Reality

This lifecycle standard is documented by default. Registry drift and generated example drift are enforced by `vibe verify`; human checkpoint discipline is review-enforced unless a project adds workflow automation.
