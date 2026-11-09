# Effect Ledger

External side effects use `effect_ledger` before the provider call.

Covered action types:

- `telegram.*`
- `max.*`
- `email.send`
- `http.request`
- `uon.create_lead`
- `clients.*`

Flow:

1. Kernel builds a stable `effect_id` from `workflow_run_id`, `action_key`,
   and `attempt_no`.
2. Before calling the runtime action, kernel starts a ledger record with
   `status = started`, input fingerprint, local idempotency key, and provider
   idempotency key (`pulse:<effect_id>`).
3. If the same effect is seen again with `status = succeeded`, kernel reuses
   the stored output/provider response and does not call the provider again.
4. If the same effect is seen again with `status = started`, kernel marks it
   `unknown` and returns `unknown_outcome`; the action fails closed and requires
   manual reconciliation instead of automatic duplicate delivery.
5. Provider failures mark the effect `failed`. Retry attempts use a different
   `attempt_no` and therefore a different `effect_id`.

This does not replace provider-native idempotency. Provider adapters that offer
idempotency headers or keys should pass `provider_idempotency_key` to the
provider. The ledger still protects adapters without native provider support by
preventing automatic replay after an unknown outcome.

CI evidence:

- `apps/api/cmd/check-runtime-claim-leases` verifies `started` duplicate to
  `unknown` and `succeeded` duplicate replay on the migrated PostgreSQL schema.
