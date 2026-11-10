# Riders Pulse AI Agent

Standalone Python LangGraph runtime for the project AI employee.

The browser still talks only to the Go API. Go owns authz, business data, tool execution, provider secrets, model calls, audit and usage metering. This service owns agent orchestration, short-term graph state, proposal composition, human-in-the-loop interrupts and stream events.

## Local

```bash
export LANGGRAPH_STRICT_MSGPACK=true
export AI_AGENT_GRPC_ADDR=127.0.0.1:9097
export AI_AGENT_SERVICE_TOKEN="$(openssl rand -hex 32)"
export PULSE_TOOL_GRPC_ADDR=127.0.0.1:9098
export PULSE_MODEL_GRPC_ADDR=127.0.0.1:9099
export AI_AGENT_CHECKPOINT_DATABASE_URL=postgres://pulse_ai@127.0.0.1:5432/pulse_ai?sslmode=disable
export AI_AGENT_CHECKPOINT_SCHEMA=ai_agent_checkpoint
export AI_AGENT_RUN_MIGRATIONS=true
uv sync
uv run pytest
uv run pulse-ai-agent
```

Set the same generated `AI_AGENT_SERVICE_TOKEN` in the Go API environment before
starting it. Create the dedicated checkpoint database and restricted role first;
the database URL above is an example, not a provisioned account. Never commit the
generated value or use it in browser configuration. The local example deliberately
binds gRPC to loopback. Deployment on separate hosts requires authenticated TLS.

Python must not connect to Riders Pulse business tables. LangGraph checkpoint and store tables use
`AI_AGENT_CHECKPOINT_DATABASE_URL` with a restricted role and `AI_AGENT_CHECKPOINT_SCHEMA`; in production the
checkpoint role must not receive grants on application tables.

Checkpoint state includes `riders-pulse-dialogue-state-v1`, a versioned
conversation envelope for active topic, suspended topics, resolved entities,
referents, unresolved slots, pending action, last tool observations and page
frame. Go exposes the envelope through `GetThreadState` and still mirrors legacy
top-level fields for older UI consumers.

`riders-pulse-topic-policy-v1` runs after deterministic/model intent
classification and before proposal routing. It applies the fixed priority order
for unsafe/cancel/approval/correction/switch/return/follow-up turns, resolves
deictic references from DialogueStateV1, keeps pending write proposals alive
during read-only detours, and asks targeted clarification when several recent
referents match.

## Rollout

Go owns the LangGraph v3 rollout policy and blocks execution before calling this
service when gates are not satisfied. Required production flags:

```bash
export AI_EMPLOYEE_RUNTIME=langgraph
export PULSE_MODEL_GATEWAY_READY=true
export PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP=true
export PULSE_AI_CHECKPOINTER_READY=true
export PULSE_AI_RESUME_READY=true
```

Canary stages are controlled by `PULSE_AI_V3_STAGE`:

- `shadow`: classify/plan comparison window only; Go rejects execution.
- `read_only`: read turns are allowed, proposals and approvals are gated.
- `proposal_only`: proposal cards are allowed, approvals/writes are gated.
- `approved_writes`: full v3 path through ToolGateway approval lifecycle.

Scope can be narrowed with `PULSE_AI_V3_COMPANY_ALLOWLIST` and
`PULSE_AI_V3_PROJECT_ALLOWLIST` comma-separated IDs. Automatic rollback is
triggered by `PULSE_AI_V3_ROLLBACK=true` or by thresholds
`PULSE_AI_V3_MIN_EVAL_SCORE`, `PULSE_AI_V3_MAX_ERROR_RATE`, and
`PULSE_AI_V3_MAX_P95_LATENCY_MS` compared with their `CURRENT_*` values. During
rollback Go downgrades LangGraph to read-only and persists the rollout policy in
the turn payload with schema, prompt, and release versions.
