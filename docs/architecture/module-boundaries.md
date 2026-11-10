# Module Boundaries

Current implementation: [owner commands and versioned read contracts](controlled-module-boundaries.md).
The exception register is empty. Cross-module reads use non-updatable, owner-declared
projections; credentials use internal owner query ports. Application commands own
transactions and outbox persistence.

Modules are business capabilities with explicit registration hooks. HTTP
endpoints must be assembled from module-owned registration functions, not from a
central router file.

## Module Contract

```go
type Module interface {
    Name() string
    RegisterHTTP(r chi.Router, deps Deps)
    RegisterAPIV1(scope APIV1Scope, deps Deps)
    RegisterRuntime(reg *registry.Registry, deps Deps)
    RegisterOutboxHandlers(reg *outbox.Registry, deps Deps)
}
```

`Deps` is owned by `platform/app` and contains shared infrastructure plus
approved runtime-facing infrastructure only:

- DB pool;
- auth service;
- runtime registry;
- outbox registry;
- runtime notifier interface for outbox/run enqueue consumers;
- secret resolver/master-key access;
- shared HTTP/model/provider clients where approved.

Modules must not accept `*executor.Engine` directly. If a module still needs a
compatibility path during migration, it must be isolated behind a package-local
adapter and documented as the next extraction target; `platform/app.Deps` must
not expose executor action-runner interfaces.

## Package Shape

```text
modules/<name>/
  module.go
  transport/http/
  service/
  repository/
  runtime/
  outbox/
  contracts/
  tests/
```

## Responsibility Rules

### transport/http

- parse request;
- call service/application boundary;
- map errors to problem responses;
- never own business orchestration.

### service

- authorization-aware application operations;
- validation;
- publish/apply/proposal decisions;
- domain side effects;
- transaction boundaries when needed.

Mutating application services should use the shared `platform/uow` transaction
boundary. Repositories may expose `...Tx` persistence operations, but should not
open or commit their own transactions.

### repository

- data persistence only;
- no HTTP provider clients;
- no cross-module orchestration;
- no policy decisions beyond data constraints.
- no direct outbox/provider side effects unless explicitly listed as a dated
  `repository_side_effect_import` exception.

`apps/api/cmd/check-module-contracts` fails new repository imports of
`net/http`, `platform/httptransport`, and `platform/outbox` unless the import is
listed in `docs/architecture/module-ownership-exceptions.tsv` with owner,
reason, and expiry. Existing exceptions are migration inventory, not precedent
for new repository behavior.

The same gate rejects new `.Begin(...)` transaction ownership inside
`modules/*/repository` and `modules/*/transport` unless a dated
`transaction_owner` exception exists. Transaction ownership belongs to
application services, runtime services, or platform infrastructure.

### runtime

- `ActionHandler` and `TriggerHandler` adapters;
- config validation delegated to handler/schema metadata;
- no additions to legacy runtime switch.
- no dependency on `apps/api/internal/executor` from module packages.

Domain-owned AI tools follow the same rule: the domain module publishes tool
descriptors, intent metadata, risk and idempotency policy; AI gateway assembles
the catalog and rejects duplicate tool/intent keys.

Scoped HTTP routes and AI tools must also be represented in
`docs/architecture/tenant-isolation-catalog.tsv`. The
`apps/api/cmd/check-tenant-isolation` gate derives scoped runtime routes from
`apps/api/internal/platform/httpserver/doc_inventory.tsv` and tool keys from the
AI tool registry, then fails when any scoped operation lacks a negative
tenant/project isolation fixture row or when a row points at a stale
operationId. Use `go run ./cmd/check-tenant-isolation --write-baseline` only
when intentionally refreshing the catalog after adding/removing scoped
operations, then review the generated `scope`, `negative_case`, and
`expected_guard` columns before committing.

### outbox

- idempotent domain event handlers;
- no direct HTTP routing.

## Initial Migration Targets

1. AI: control plane, tools, approvals, proposals, providers.
2. Workflows: publish, definition validation, patches, runtime handlers.
3. Runs: run read/explain/replay/cancel operations.
4. Action dependencies: provider secrets, model/provider catalog, integration diagnostics.
5. Clients/worktracker: client/payment/task tool adapters and triggers.

Worktracker is the current example for the target command shape: HTTP and AI
tool entrypoints use `worktracker/service.Commands`, task repositories no longer
own command transactions, and service-owned task events use a narrow
`TaskEventPublisher` port for atomic `platform/outbox` enqueueing.

## Current HTTP Boundary

`platform/httpserver/router.go` is restricted to middleware, response defaults,
and route-error policy. It must not instantiate repositories or list business
endpoints.

`platform/httpserver/api_v1.go` owns the single `/api/v1` dispatcher and its
public, protected, work, and project route scopes. It knows middleware and path
prefixes only; it must not import business modules or call endpoint handlers.

`platform/app.BuildModules()` owns the installed module list and high-level
module ordering. It must not call module endpoint mount functions such as
`*.MountProjectHTTP(...)`, `*.MountProtectedHTTP(...)`, or transport route
mounts directly. A central list of installed modules is allowed; a central list
of business endpoints is not.

Each module owns its HTTP route registration from `modules/<name>/module.go`.
Root routes outside `/api/v1` use `RegisterHTTP(r, deps)`. API routes use
`RegisterAPIV1(scope, deps)` and switch on `public`, `protected`, `work`, or
`project` scope. Current module-owned HTTP seams include:

```text
modules/ai/module.go
modules/action_dependencies/module.go
modules/apikeys/module.go
modules/auth/module.go
modules/catalog/module.go
modules/clients/module.go
modules/companies/module.go
modules/desktop/module.go
modules/ingest/module.go
modules/integrationdocs/module.go
modules/projects/module.go
modules/runs/module.go
modules/sitechat/module.go
modules/telegramrelay/module.go
modules/templates/module.go
modules/webhooks/module.go
modules/workflows/module.go
modules/worktracker/module.go
```

Legacy module route files still own their endpoint handlers:

```text
modules/auth/transport/http/routes.go
modules/apikeys/transport/http/routes.go
modules/catalog/transport/http/routes.go
modules/companies/transport/http/routes.go
modules/clients/transport/http/routes.go
modules/projects/transport/http/routes.go
modules/sitechat/transport/http/routes.go
modules/action_dependencies/transport/http/routes.go
modules/ai/transport/http/routes.go
modules/templates/transport/http/routes.go
modules/workflows/transport/http/routes.go
modules/webhooks/transport/http/routes.go
modules/runs/transport/http/routes.go
modules/desktop/transport/http/routes.go
```

`platform/httpserver/routes.go` only mounts platform health and invokes the
module list supplied by the composition root. `Modules.RegisterHTTP` then mounts
the API v1 dispatcher once and lets modules fill their own scopes. Route
ownership belongs to the modules; platform/app is not a router registry.

## Frontend API Route Ownership

The web app must not grow a handwritten global route registry such as
`apps/web/src/lib/api/routes.ts`. Runtime route inventory and generated client
artifacts come from OpenAPI. Handwritten path constants are allowed only as
module-owned contracts, for example `auth-routes.ts` for browser auth transport
bootstrap paths. Feature API adapters should import their owning module's
contract or use generated OpenAPI operations, not append paths to a central
frontend routes file.
