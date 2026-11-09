# Worker Topology Contract

Production topology is `split-worker`.

`cmd/api` owns HTTP request handling only:

- authenticate and authorize;
- validate requests;
- write domain mutations;
- enqueue workflow runs through `runtime/runqueue`;
- enqueue follow-up work through the transactional outbox;
- notify workers opportunistically.

It must not process workflow runs in production. Production startup validation
rejects `PULSE_RUNTIME_TOPOLOGY=embedded` for the API role.

`cmd/worker` owns background execution:

- runtime kernel claim/drain loop for `workflow_runs`;
- trigger sync/enqueue through the runtime registry;
- Telegram relay polling;
- transactional outbox claim/dispatch loop.

The worker uses the same module registrations as the API but starts no HTTP
server. It sets `SkipMigrations=true` during application construction so worker
instances do not race the release/API migration path.

Embedded runtime remains a local/development profile only. It is useful for
`scripts/dev.sh`, isolated tests, and one-process debugging, but it is not a
production topology.

Verification:

- production config validation rejects API/worker roles unless
  `PULSE_RUNTIME_TOPOLOGY=split-worker`;
- `go run ./cmd/check-worker-topology` builds multiple API application instances
  without starting background execution, confirms a pending run stays pending,
  then runs multiple worker runtime instances and verifies exactly one claim
  completes the run.
