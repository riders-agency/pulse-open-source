<!-- Generated from registry/standards/api/openapi-modular@1.1.0. Update the registry standard, then regenerate. -->

# OpenAPI Modular Contract Standard

This standard applies to every HTTP API endpoint in an active stack.

The contract is the first runtime surface. Backend code, frontend clients, tests and documentation follow the OpenAPI contract, not the other way around.

## May 2026 Baseline

1. OpenAPI Specification: `3.2.0`.
2. HTTP semantics: RFC 9110.
3. Problem Details: RFC 9457.
4. JSON Schema alignment follows the OpenAPI 3.2 model.
5. Endpoint inventory is mandatory for agent-readable coverage.

## Fixed Structure

1. `OPENAPI_VERSION = 3.2.0`
2. `ROOT_FILE = contracts/openapi/openapi.root.yaml`
3. `MODULE_FILES = one file per module`
4. `INVENTORY_FILE = contracts/openapi/endpoints.inventory.tsv`
5. `DIST_DIR = contracts/openapi/dist`

Each module file owns a bounded API area such as `auth`, `company`, `docs` or `billing`.

## Endpoint Coverage Rules

The allowed count is `0` for:

1. endpoint in code without an OpenAPI operation;
2. OpenAPI operation without an inventory row;
3. inventory row without an OpenAPI operation;
4. missing `operationId`;
5. missing `summary`;
6. missing success response;
7. missing client-error response;
8. missing `tags`;
9. duplicate `operationId`;
10. undocumented auth requirement.

## Operation Design

Every operation must define:

1. stable `operationId`;
2. module tag;
3. short summary;
4. request body schema when applicable;
5. success response schema;
6. RFC 9457-compatible error response schema;
7. request examples for write operations;
8. response examples;
9. curl sample through `x-codeSamples` when tooling supports it;
10. test hint in the endpoint catalog.

## HTTP Method Policy

Allowed methods:

1. `GET` for reads with no business side effects.
2. `POST` for commands, creates and non-idempotent mutations.
3. `PUT` for full idempotent replacement.
4. `PATCH` for partial updates with documented idempotency semantics.
5. `DELETE` for revoke/remove operations.
6. `HEAD` for GET metadata mirrors.
7. `OPTIONS` for CORS and capability negotiation.

Disallowed for business APIs:

1. `TRACE`
2. `CONNECT`

Body rules:

1. `GET` request body count: `0`.
2. `HEAD` request body count: `0`.
3. `OPTIONS` request body count: `0`.
4. `POST|PUT|PATCH` request body examples minimum: `1`.

Same path with different methods is allowed only when each operation has distinct semantics, summaries, examples, responses and tests.

## Response Contract

Do not wrap every success response in a universal `data/meta/errors` envelope by default. Business payload shape is owned by the endpoint schema.

Error responses must use a shallow, stable shape compatible with Problem Details:

1. `status`
2. `title`
3. `detail`
4. `action`
5. `validate`
6. `request_id`

Field-level validation may include structured extensions, but must remain explicitly documented in the schema.

## Mandatory Response Headers

Every API response includes:

1. `X-Request-Id`
2. `X-App-Result`
3. `Content-Language`

`X-App-Result` values:

1. `ok` for successful responses;
2. `fix_required` for client-correctable `4xx`;
3. `error` for `5xx` and infrastructure failures.

Conditional headers:

1. `X-App-Action` when a caller can fix the request.
2. `X-App-Validate` when specific fields need attention.

## Idempotency And Concurrency

Write endpoints must document their idempotency behavior:

1. create endpoints either accept an idempotency key or document why retries are safe;
2. payment, billing, webhook and worker-triggering endpoints require idempotency keys;
3. update endpoints that can overwrite user data should use ETags, version fields or explicit conflict responses;
4. retryable failures must be distinguishable from final failures.

## Pagination, Filtering And Sorting

List endpoints must define:

1. pagination style: cursor preferred for mutable lists, offset allowed for stable admin lists;
2. maximum page size;
3. default sort;
4. allowed filter fields;
5. stable response schema for `items` and continuation metadata.

Do not invent per-endpoint pagination shapes without documenting the reason.

## Versioning And Deprecation

Breaking changes in `v1` without a migration plan are not allowed.

Deprecation requires:

1. `deprecated: true` in the OpenAPI operation or schema property;
2. removal date in `YYYY-MM-DD` format;
3. migration note;
4. minimum production deprecation window of `90` days;
5. support and observability plan for old and new consumers.

Backward-compatible additions must be optional or have safe default behavior.

## Docs Endpoints

Generated projects should expose docs endpoints when runtime exists:

1. `GET /api/v1/docs/openapi.json`
2. `GET /api/v1/docs/openapi.yaml`
3. `GET /api/v1/docs/endpoints`
4. `GET /api/v1/docs/testing/postman`

The endpoint catalog returns method, path, summary, tag, operation ID, examples and test hints.

## Security Requirements

Every operation must define:

1. public/private access model;
2. auth scheme when required;
3. tenant or ownership boundary when applicable;
4. rate-limit or abuse note for sensitive business flows;
5. SSRF-safe URL handling when accepting external URLs;
6. safe third-party API consumption contract.

This aligns the API standard with OWASP API Security Top 10 2023 risks.

## Feature Sync Rule

Changing a feature means updating:

1. module OpenAPI file;
2. schemas;
3. examples;
4. endpoint inventory;
5. tests or test hints;
6. docs endpoint catalog when runtime exists;
7. generated clients when the project uses them.

## Enforcement Reality

`check_openapi` verifies the required OpenAPI project structure.

Deep semantic coverage, examples, idempotency and deprecation rules are documented until a stack adds a real OpenAPI linter such as Spectral, Redocly or custom coverage scripts.
