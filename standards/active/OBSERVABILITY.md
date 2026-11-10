<!-- Generated from registry/standards/observability/three-signals@1.1.0. Update the registry standard, then regenerate. -->

# Three Signals Observability Standard

This standard applies to enabled runtime components only. In the base stacks that means `apps/api` and `apps/web`. Mobile, desktop, worker and payments telemetry applies only after those features are enabled.

## May 2026 Baseline

1. OpenTelemetry is the preferred instrumentation model.
2. Semantic conventions should follow the selected OpenTelemetry semconv version.
3. Three core signals are traces, logs and metrics.
4. Profiling is optional and must be explicitly enabled.
5. PII and secrets are redacted before leaving the process.

## Three-Signal Model

1. `traces`: causal request flow.
2. `logs`: structured events.
3. `metrics`: aggregates for SLOs, capacity and alerts.

Logs without traces or metrics are incomplete for production behavior.

## Request Correlation

Required identifiers:

1. `request_id` in `X-Request-Id`;
2. `trace_id` in span context;
3. `span_id` for the current span.

Rules:

1. honor incoming `X-Request-Id` when safe;
2. generate one when absent;
3. return it in responses;
4. include it in logs;
5. include it in span attributes as `vibe.request_id` or the project-approved namespace.

## Structured Logging

Server logs use JSON lines.

Required server log fields:

1. `timestamp`
2. `level`
3. `message`
4. `service`
5. `env`
6. `module`
7. `operation`
8. `request_id`
9. `trace_id`
10. `span_id`
11. `http_method`
12. `http_route`
13. `http_status`
14. `latency_ms`
15. `result`

Conditional fields:

1. `user_id`
2. `company_id`
3. `session_id`
4. `error_code`
5. `validate`

Forbidden in logs:

1. plaintext passwords;
2. access or refresh tokens;
3. OAuth codes;
4. raw third-party signed payloads;
5. environment secrets;
6. full payment payloads.

## Tracing

Server span name:

```text
HTTP <METHOD> <ROUTE_TEMPLATE>
```

Example:

```text
HTTP POST /api/v1/auth/login
```

Required span attributes:

1. HTTP method;
2. route template;
3. response status;
4. client address when available;
5. user agent when available;
6. authenticated user ID when available;
7. tenant/company ID when available;
8. request ID;
9. module;
10. operation.

Use stable route templates, not raw URLs with IDs.

## Status Mapping

1. `2xx` -> span status `Ok`.
2. Correctable `4xx` -> span status `Ok` with validation event.
3. `5xx` -> span status `Error`.
4. Canceled requests should record cancellation reason when safe.

## Metrics

Minimum API metrics:

1. `http_server_requests_total{method,route,status,result}`
2. `http_server_request_duration_ms{method,route}`
3. `auth_attempts_total{provider,result}`
4. `auth_session_revocations_total{mode}`
5. `active_sessions_gauge`

Baseline SLO targets:

1. read endpoint `P95 < 300ms`;
2. auth/session mutation endpoint `P95 < 500ms`;
3. error budget defined before production.

## Sampling And Retention

Starter defaults:

1. `5xx` traces: `100%`;
2. validation traces: sampled;
3. success traces: sampled;
4. security-sensitive events: retained according to compliance needs;
5. high-cardinality debug data: short retention.

Do not sample away audit-critical payment or security events without an explicit reason.

## Web Telemetry

When `apps/web` is enabled:

1. collect Web Vitals: `LCP`, `INP`, `CLS`;
2. propagate `request_id` in API calls;
3. capture client runtime errors;
4. avoid logging PII from forms;
5. connect route changes to user-visible performance metrics.

## Optional Feature Telemetry

When enabled:

1. worker jobs expose job ID, attempt, queue, duration and result;
2. mobile exposes crash-free sessions, startup, deep links and network errors;
3. payments exposes webhook event ID, payment intent ID, idempotency key hash and reconciliation status;
4. desktop follows its own optional standard.

## Definition Of Done

A feature is observability-ready when:

1. new endpoints have trace coverage;
2. success and error paths produce structured logs;
3. request count and latency metrics exist;
4. sensitive data is redacted;
5. dashboards/alerts are updated where runtime exists.

## Enforcement Reality

This standard is documented by default. It becomes linted/tested/CI-blocking only when the generated project includes real telemetry checks, snapshot tests or runtime smoke tests.
