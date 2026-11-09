# Workflow step output: `expose` and chain data

Structured action results are written to `action_runs.output_json` (per attempt: `action_attempts.output_json`). Integrations that emit a safe, flat payload for downstream steps and for the run UI populate:

- **`expose`** — `object` (flat key/value). Only non-sensitive values the product explicitly allows (ids, counts, host names of integration endpoints, etc.). Must not contain tokens, secrets, or full message bodies.
- **`criteria_results`** — `array` of `{ type, passed, detail?, message? }` from the runner’s success-criteria evaluator.
- **`criteria_passed`** — `boolean` when `criteria_results` is present.

## Per action type (runner contract)

| `action_type_key`     | Typical `expose` keys                          |
|----------------------|-------------------------------------------------|
| `telegram.send_message` | `chat_id`, `message_id`                    |
| `max.send_message`     | `recipient_type`, `recipient_id`, `chat_id`, `user_id`, `message_id` |
| `email.send`           | `recipient_count`, `smtp_host`, `smtp_port` |

Future custom steps (e.g. price tracking) should agree on stable `expose` keys in this document before shipping.

Templates for a future `run_context` may reference prior steps as `{{ steps.<action_key>.<expose_key> }}` once the orchestrator merges `expose` from completed actions.

## Single-action replay (input)

Created runs may include `input_json._pulse`:

- `retry_action_key` — `action_key` to execute alone (other nodes skipped).
- `base_workflow_run_id` — originating run when using `POST /api/v1/runs/{workflow_run_id}/retry-action`.
