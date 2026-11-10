# System Architecture

Riders Pulse is a self-hosted operational platform. The system is organized by
runtime responsibility, not by UI feature names.

## Process Topology

```text
nginx
  -> apps/web Next server
  -> apps/api cmd/api
  -> Postgres
  -> apps/api cmd/worker
  -> apps/ai-agent Python LangGraph sidecar
```

Expected production services:

- `pulse-web.service`;
- `pulse-api.service`;
- `pulse-worker.service`;
- `pulse-ai-agent.service`;
- Postgres 17+.

## Request Paths

### Browser API

```text
Browser
  -> same-origin /api/v1
  -> Go API auth/session/project scope
  -> module service/repository/runtime
```

The browser never decides authorization and never calls Python directly.

### Workflow Runtime

```text
API/manual/webhook/domain trigger
  -> run enqueue service
  -> workflow_runs pending
  -> cmd/worker runtime kernel
  -> registered ActionHandler
  -> action_runs/action_attempts/run_events
```

### Follow-Up Work

```text
domain mutation transaction
  -> versioned outbox_events row
  -> cmd/worker claim with FOR UPDATE SKIP LOCKED
  -> registered outbox handler
  -> processed/failed state
```

### AI Employee

```text
Browser
  -> Go AI HTTP/SSE
  -> Python LangGraph orchestration
  -> Go ToolGateway/ModelGateway
  -> Go domain services
```

Go remains source of truth. Python is orchestration-only.

## Data Ownership

- Companies/projects/users/auth: Go API modules.
- Clients/tasks/payments/workflows/runs: Go API modules and runtime stores.
- Secrets/provider credentials: Go API secrets/action dependencies layer.
- AI chat sessions/messages/cards/tool calls/audit/usage: Go AI module.
- LangGraph checkpoints: restricted AI-agent schema only.

## Non-Negotiable Constraints

- No `apps/ai`; the Python sidecar is `apps/ai-agent`.
- No `apps/worker`; worker command lives under `apps/api/cmd/worker`.
- No Redis/NATS/Kafka in the first production worker boundary.
- No executor action compatibility fallback; unknown actions fail as
  `missing_handler`.
- No direct business table access from `apps/ai-agent`.
- No secrets in logs, SSE, traces, audit, prompt snapshots, or test artifacts.
