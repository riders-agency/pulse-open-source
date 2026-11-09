# OpenAPI Contracts

`openapi.root.yaml` is the source of truth for HTTP contracts. Handwritten
module files live in `modules/`; `modules/runtime-inventory.generated.yaml` is
generated from `endpoints.inventory.tsv` to keep every runtime route represented
even before a curated schema is written.

Run the gate after route or contract changes:

```bash
./scripts/check-openapi-contracts.sh
```

The gate runs the runtime supplement generator, Redocly lint/bundle with the
pinned CLI package, compares the deterministic bundle against
`dist/openapi.yaml`, compares the API embedded docs mirror against that bundle,
verifies method/path/operationId coverage against the runtime inventory, and
regenerates the web contract artifacts in
`apps/web/src/lib/api/generated/openapi-types.ts` and `openapi-client.ts`.

`x-runtime-status: preview` is the only accepted way to keep a documented
OpenAPI operation that is intentionally not mounted in the runtime inventory.
