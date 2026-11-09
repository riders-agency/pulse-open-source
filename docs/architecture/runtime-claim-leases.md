# Runtime Claim Leases

Workflow execution claims are fenced by the `workflow_runs` claim columns:

- `run_claim_owner` identifies the worker process instance.
- `run_lease_expires_at` is compared with database `now()` for bounded
  clock-skew behavior.
- `run_heartbeat_at` records the last lease extension.
- `run_fencing_token` increments on every claim/reclaim.
- `run_claim_count` records first claims and reclaims.

Claim behavior:

1. The kernel claims `pending` runs and expired `running` runs through
   `FOR UPDATE SKIP LOCKED`.
2. Claiming sets status `running`, owner, lease expiry, heartbeat timestamp,
   increments `run_fencing_token`, and increments `run_claim_count`.
3. Long action execution starts a heartbeat loop. If the heartbeat loses the
   owner/token fence, the action context is cancelled and no late result is
   written by that worker.
4. Workflow terminal and waiting transitions are guarded by
   `workflow_run_id + run_claim_owner + run_fencing_token + status=running`.
5. Reclaimed runs emit `workflow.reclaimed`.

Run state semantics:

- workflow and action state audit is documented in
  `docs/architecture/run-state-machine.md`;
- user cancellation is transactional and idempotent through `runs.Cancel`;
- expired waiting deadlines are processed by the kernel as
  `terminal_reason=waiting_timeout`.

CI evidence:

- `apps/api/cmd/check-runtime-claim-leases` creates a fixture workflow on the
  migrated PostgreSQL database.
- It verifies heartbeat during an action longer than the initial lease.
- It verifies reclaim of an expired `running` run.
- It verifies that the stale owner/fencing token cannot update the reclaimed
  run.
- It verifies pending/running/completed cancellation semantics and single
  `workflow.cancelled` event emission.
- It verifies `waiting_deadline_at` timeout transition to failed workflow and
  action rows with `workflow.waiting_timed_out`.
