# AI Boundary

The AI employee is a Pulse control-plane capability. It is not a browser-owned
chat bot and not a Python-owned business service.

## Runtime Topology

```text
Browser
  -> Go HTTP/SSE AI Control Plane
  -> Python LangGraph AgentRuntimeService
  -> Go ToolGateway / ModelGateway
  -> Go domain services
```

## Browser Responsibilities

- send user body;
- send page, client, and UI context as untrusted context;
- send local date/timezone;
- render typed cards/snippets and safe progress events;
- never decide permissions;
- never call Python directly.

## Go Control Plane Responsibilities

- verify auth/session/API key;
- enforce company/project/user scope;
- normalize page context server-side;
- persist chat sessions/messages/cards/tool calls;
- expose tool catalog;
- validate tool input;
- execute read tools;
- create and patch proposals for mutating tools;
- apply writes only after approval;
- enforce idempotency;
- own provider secrets and model calls when policy requires;
- audit proposals, approvals, rejections, tool execution, model calls, and denials.

## Python LangGraph Responsibilities

- classify turns;
- maintain short-term graph state;
- plan tool usage;
- orchestrate read/write proposal flow;
- compose typed proposal cards;
- sequence stream events;
- handle interrupt/resume orchestration.

Python must not:

- read or write business tables;
- import Go domain repositories;
- decrypt provider secrets;
- call external model providers directly by default;
- execute business writes directly;
- apply workflow patches directly.

## Tool Lifecycle

```text
ListTools
  -> ValidateToolCall
  -> ExecuteReadTool for read-only tools
  -> CreateProposal for write/critical tools
  -> persist card/tool call
  -> user approval/rejection
  -> Execute write through Go domain service
  -> audit and entity links
```

Risk levels are `read`, `medium`, `high`, and `critical`. Proposal is a
lifecycle state, not a risk level.

## Approval Rules

- read tools can execute without approval when policy permits;
- `medium` tools may create or patch proposals but cannot mutate business data;
- `high` and `critical` tools require explicit approval;
- all write/critical execution requires an idempotency key;
- stale card revisions must be rejected;
- critical tools require elevated policy before enablement.

## Stream Events

User-facing stream events must be typed and redacted:

- `assistant.message.delta`;
- `assistant.message.completed`;
- `agent.mode.changed`;
- `agent.step.started`;
- `agent.step.completed`;
- `tool.call.started`;
- `tool.call.completed`;
- `draft.created`;
- `draft.updated`;
- `approval.required`;
- `approval.applied`;
- `approval.rejected`;
- `error.recoverable`.

Raw chain-of-thought is never streamed. Debug payloads are hidden by default
and redacted before persistence or display.

## Persistence

Go persists business-visible state. LangGraph checkpointer/store access is
restricted to AI runtime state only. The same browser chat session maps to a
stable LangGraph thread id.
