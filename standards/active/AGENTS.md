<!-- Generated from registry/standards/agent/base@1.1.0. Update the registry standard, then regenerate. -->

# Agent Base Standard

This standard defines the operating contract for humans and coding agents in a generated project.

The goal is not to slow agents down. The goal is to make the active context small, explicit and hard to misread.

## May 2026 Baseline

AI-native projects should assume:

1. multiple coding agents may touch the same repository;
2. generated files and registry source files have different authority;
3. context windows are finite, so the project must expose a short read path;
4. quality gates are evidence, not decoration;
5. public starter repositories must be safe to clone, fork and publish.

## Source Of Truth Order

When instructions conflict, use this order:

1. user task and repository owner decisions;
2. generated `AGENTS.md`;
3. active standards in `standards/active`;
4. `.vibe/profile.yaml`;
5. `.vibe/enabled.yaml`;
6. `.vibe/registry.lock`;
7. registry standards when updating the template itself;
8. local conventions in code.

Generated `standards/active/*` files are read-only project context. The source of truth is `registry/standards/<id>/standard.md`.

## Required Generated Files

Every generated project must contain:

1. `AGENTS.md`
2. `standards/active/AGENTS.md`
3. `.vibe/profile.yaml`
4. `.vibe/enabled.yaml`
5. `.vibe/registry.lock`
6. `scripts/check.sh`

`AGENTS.md` is generated from the active stack profile. It must describe only the selected stack and enabled optional features.

## Agent Read Order

For a normal task, read:

1. `AGENTS.md`
2. the relevant files in `standards/active/`
3. `.vibe/profile.yaml`
4. `.vibe/enabled.yaml` when adding or removing project parts
5. `.vibe/registry.lock` when changing standards or stack composition

For a registry maintenance task, read:

1. `docs/architecture.md`
2. `docs/contributing-standards.md`
3. the target `standard.yaml`
4. the target `standard.md`
5. the target `CHANGELOG.md`
6. affected stack profiles and examples

## Forbidden

1. Do not add unused folders.
2. Do not add disabled technologies without updating `.vibe/profile.yaml` and `.vibe/enabled.yaml`.
3. Do not bypass quality gates.
4. Do not commit real secrets.
5. Do not duplicate business logic across API/web/mobile/worker.
6. Do not change standard versions without updating `.vibe/registry.lock` through generation.
7. Do not edit generated `standards/active/*` as if they were the registry source.
8. Do not introduce new package managers, runtimes or services without documenting the decision.
9. Do not claim a rule is enforced unless a real check exists.

## Task Flow

Every non-trivial change follows:

```text
scope -> contract -> failing check/test -> implementation -> green checks -> docs -> evidence
```

If a project cannot run a check yet, the limitation must be visible in the status matrix and in the task notes.

## Evidence Standard

A finished agent task should leave enough evidence for a reviewer to trust it:

1. files changed;
2. behavior changed;
3. checks run;
4. checks not run and why;
5. migrations or lockfile updates;
6. remaining risk.

Do not summarize "all good" without naming the checks.

## Context Hygiene

Agents should keep the project clean:

1. prefer editing existing active components over creating parallel folders;
2. keep temporary scripts out of committed templates;
3. remove debug logs before completion unless they are observability events;
4. update docs at the registry source, then regenerate examples;
5. keep generated examples reproducible.

## Multi-Agent Safety

When more than one agent may work in the repo:

1. avoid broad refactors outside the task scope;
2. do not revert unknown changes;
3. keep ownership boundaries explicit;
4. regenerate examples only after registry changes settle;
5. run `vibe verify` before handing off.

## Enforcement Reality

The generated-file presence checks are linted and CI-blocking through `check_standards`.

Behavioral agent discipline is documented and review-enforced unless a project adds stronger local automation.
