# Pulse Proto Contracts

`contracts/proto/ai_agent/v1` is the source of truth for the Go API and Python AI sidecar gRPC boundary.

## Toolchain

Proto generation is intentionally pinned and fail-closed:

- `protoc`: `libprotoc 3.12.4`
- `buf`: `1.71.0`
- `protoc-gen-go`: `v1.36.11`
- `protoc-gen-go-grpc`: `v1.6.2`
- Python `grpc_tools.protoc`: `libprotoc 33.5` from `apps/ai-agent/uv.lock`

Run the single generation command from the repository root:

```bash
./scripts/gen-proto.sh
```

The script must regenerate both Go and Python stubs. Missing plugins or version drift are errors, not skipped work.

## Verification

Run the contract gate from the repository root:

```bash
./scripts/check-proto-contracts.sh
```

The gate verifies:

- buf format/lint when `buf` is installed, and requires it in CI via `REQUIRE_BUF=1`;
- optional breaking check against `PROTO_BREAKING_AGAINST`;
- clean regeneration drift for Go and Python stubs;
- current descriptor equality with `contracts/proto/descriptors/ai_agent_v1.binpb`;
- Go/Python serialization fixtures through the normal test suites.

`buf.yaml` keeps the current shared streaming DTOs compatible by explicitly
excluding RPC request/response naming and uniqueness rules. Do not add new lint
exceptions without documenting the compatibility reason here.

## Compatibility Policy

Do not reuse field numbers, enum values, or service method names after removal. Reserve removed field numbers and names in the same message or enum before deleting the implementation surface.

For a deliberate compatible proto change:

1. edit `.proto` files;
2. run `./scripts/gen-proto.sh`;
3. update `contracts/proto/descriptors/ai_agent_v1.binpb` with `buf build contracts/proto -o contracts/proto/descriptors/ai_agent_v1.binpb`;
4. update serialization fixtures only when the wire contract intentionally changes;
5. run `./scripts/check-proto-contracts.sh`, Go tests, and AI agent tests.
