# Runtime Boundaries

The workflow runtime must be a kernel with registered extensions, not a
provider/domain switch.

## Kernel Ownership

The application-facing runtime boundary is `runtime/kernel.Kernel`.
`executor.Engine` adapts workflow definitions to the kernel runtime contract,
but action and trigger execution must enter through `runtime/registry.Registry`.
Application/module wiring must depend on kernel, registry, notifier, and
explicit infrastructure clients instead of accepting `*executor.Engine` or
executor action-runner interfaces directly.

The runtime kernel owns:

- claiming pending `workflow_runs`;
- loading immutable workflow version definitions;
- building execution plans from nodes and edges;
- attempt scheduling;
- retry and waiting state;
- persistence through stores;
- run completion/failure;
- output reduction.

The kernel must not import domain/provider modules. It can depend only on
runtime contracts, runtime stores, platform observability, and platform
configuration.

## Action Boundary

Actions are registered through:

```go
type ActionHandler interface {
    TypeKey() string
    ValidateConfig(ctx context.Context, cfg json.RawMessage) error
    Execute(ctx context.Context, req ActionRequest) ActionResult
}
```

Rules:

- New action types must be registered in `runtime/registry.Registry`.
- Unknown action types fail as `missing_handler`.
- Handler packages validate their own config.
- Handler packages receive dependencies through module registration.
- Secret access goes through approved resolver interfaces.
- Provider/domain handlers must not call executor compatibility.

## Trigger Boundary

Triggers are registered through:

```go
type TriggerHandler interface {
    TypeKey() string
    Sync(ctx context.Context) error
    ClaimDue(ctx context.Context, limit int) ([]TriggerEvent, error)
    MarkEnqueued(ctx context.Context, event TriggerEvent, workflowRunID string) error
}
```

Trigger types:

- `webhook.trigger`;
- `schedule.trigger`;
- `telegram.relay.trigger`;
- `clients.payment_due.trigger`;
- `module.event.trigger`;
- `manual.trigger`;
- `test.trigger`.

Trigger events must carry idempotency keys and must not enqueue duplicate
workflow runs on retry.

## Worker Boundary

Domain mutations that require follow-up work write an outbox row in the same DB
transaction. Workflow run producers write `workflow_runs` through the runqueue.
`apps/api/cmd/worker` claims both pending workflow runs and outbox rows with
`FOR UPDATE SKIP LOCKED` and dispatches to registered runtime/outbox handlers.

The worker owns:

- workflow run claim leases, heartbeats, fencing, bounded retries and
  dead-letter/replay handling;
- outbox leases;
- processed/failed state;
- Telegram relay polling;
- structured logs and metrics;
- idempotent handler dispatch.

The API process may run an embedded runtime only in dev. Production deploys a
separate worker process and rejects embedded API topology at startup.

## Executor Adapter

`executor.Engine` remains only as the workflow-definition adapter and registry
composition helper. Growth guardrails:

- engine line count is ratcheted to facade size and must not grow without an
  explicit architecture debt note;
- router line count must not grow without module migration reason;
- registered action/trigger counts must be visible in tests or diagnostics.
- modules must not take `*executor.Engine` in `platform/app.Deps`;
- provider and client/payment actions must stay module-owned;
- bot state is a registered action handler with a narrow bot-session store, not
  a compatibility fallback.
