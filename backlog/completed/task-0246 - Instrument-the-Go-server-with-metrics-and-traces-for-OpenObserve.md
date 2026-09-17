---
id: TASK-0246
title: Instrument the Go server with metrics and traces for OpenObserve
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-16 12:35'
updated_date: '2026-09-16 21:23'
labels: []
dependencies:
  - TASK-0244
  - TASK-0245
documentation:
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/quality/qr/QR-004.md
  - docs/requirements/quality/qr/QR-015.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/requirements/quality/qr/QR-018.md
  - docs/environments.md
priority: medium
type: feature
ordinal: 247000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Extend the OpenObserve adoption from logs to metrics and traces per ADR-022, exporting OpenTelemetry metrics and traces directly from the Go server to OpenObserve, since the platform is adopted for all three signals and logs ship first.

This task depends on TASK-0245, which establishes the OpenTelemetry exporter bootstrap, the environment wiring, and the non-blocking export behavior that metrics and traces reuse. The same data-exclusion rules apply: no credential, secret, anonymous edit token, or token-bearing URL or header may be exported.

Governing requirements: QR-004, QR-015, QR-017, QR-018. Decision record: ADR-022.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The server exports metrics and traces over OTLP to the configured OpenObserve organization, and they are queryable from the development-stack exercise that used the same code and OTLP contract; TASK-0250 produces the staging setup and evidence
- [x] #2 Exported telemetry never blocks or fails request handling, including when OpenObserve is unavailable, and the local diagnostic output remains available (QR-017)
- [x] #3 Traces cover representative request paths, and metrics expose health signals usable together with logs for diagnosis, including the service's PostgreSQL-dependent readiness that explains dependency-related request failures
- [x] #4 No credential, secret, anonymous edit token, or token-bearing URL or header appears in the exported telemetry (QR-004)
- [x] #5 Unit tests cover instrumentation setup and attribute construction, and docs/environments.md lists any new variables
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 Changed Markdown files are formatted with npm run format:markdown
- [x] #4 Browser E2E is not required for this server-only change: the user approved the exemption, and the isolated backend suite plus the development-stack exercise cover the affected request path
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Dependencies: add `otlptracehttp` v1.46.0, `otlpmetrichttp` v1.46.0, `sdk/metric` v1.46.0, and `contrib/instrumentation/runtime` v0.71.0 (pairs with otel v1.46.0); `go mod tidy` and compile.
2. Config (`server/observability/config.go`): generalize the signal URL builder, add `MetricsURL()` and `TracesURL()` (`/api/<org>/v1/metrics|traces`), add `ServerTraceStream = "timeful_server_traces"`. No new environment variables; the four existing `OPENOBSERVE_*` values plus `APP_ENV` cover all three signals.
3. Recorder/provider: extend `Recorder` to logs + traces + metrics over one shared resource (`service.name`, `deployment.environment`) and Basic auth. Bounded trace batch processor (2048 queue / 512 batch / 5s), periodic metric reader (60s interval) with a semconv HTTP duration-bucket view. Instruments: `http.server.request.duration` (seconds, bounded attributes), `timeful.service.readiness` gauge (1/0), `db.client.connection.count{state=used|idle}` and `db.client.connection.max` via an injected pool-stats source; runtime metrics started in `Start`. Explicit providers (no global tracer/meter registration except `TraceContext` propagation and the existing error handler), nil recorder is a no-op, and shutdown flushes metrics, traces, and logs concurrently within one shared budget; raise `DefaultShutdownTimeout` to 10s to fit the three flushes inside Docker's 30s stop grace.
4. Request middleware: start one always-on server span per request, set final attributes and `Error` status for 5xx after `c.Next()`, record the duration histogram, and emit the log record with the span context so log records carry native `trace_id`/`span_id`. Add testable `Attributes`/`MetricAttributes`/status methods on `RequestCompletion`.
5. PostgreSQL (`server/postgres`): implement a small `pgx.QueryTracer` (client span with `db.system.name`, `db.namespace`, `server.address`, operation name, SQLSTATE as `db.response.status_code`; never SQL text or error messages) and register it in `postgres.Init`; add a connection-stats accessor for the pool gauges.
6. Outbound HTTP: add a QR-004-safe `observability.Transport` RoundTripper that records only method, scheme, host/port, status, and a sanitized error class, injects `traceparent`, and never records `url.full`, query strings, headers, or `err.Error()`; install it in `main.go` by wrapping `http.DefaultTransport`, which covers every `http.DefaultClient`/`http.Get` call site.
7. Tests: extend the httptest collector to all three signals; in-memory trace/metric assertions for instrumentation setup, span and metric attribute construction, buckets, readiness and pool gauges, log-to-trace correlation, disabled no-op; pgx tracer attribute/error tests; transport leak tests (query tokens, Authorization headers, URL paths).
8. Docs: extend `docs/environments.md` with the metrics and traces contract (endpoints, trace stream, observed metric stream naming, span coverage and attributes, correlation, always-on sampling, export-failure behavior, explicit no-new-variables statement); add `DEPLOYMENT.md` operator query notes; run the root Markdown pipeline.
9. Verify: `gofmt`, `go vet`, `go build`, host unit tests, the isolated Compose suite, then the development-stack exercise (success/404/500 requests, trace and PromQL queries, PostgreSQL-stopped readiness flip, secret scan of shipped telemetry, QR-017 stop/slow repeat for the new pipelines), and record the evidence.
10. Confirm the dependency scope recorded and presentation decisions: follow the user-approved scope where outbound dependency tracing covers request, PostgreSQL, and outbound HTTP only; Cloud Tasks gRPC instrumentation is dropped because that path is dormant (no production callers and unused `Event.TaskIds`), and Discord is skipped because the package has no production caller. A follow-up spike task captures the in-server reminder scheduling evaluation.

Approval record: the user approved development-stack evidence for AC #1/#3 with an AC amendment (staging unreachable, TASK-0250 precedent), a browser-E2E exemption for this server-only change, and a followed-up evaluation of replacing Google Cloud Tasks.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation

- `server/observability`: `config.go` gained `MetricsURL`/`TracesURL` through a shared signal URL builder plus `ServerTraceStream`; `provider.go` now owns log, trace, and metric providers over one shared resource with a bounded batch span processor, a periodic metric reader, a nil-safe no-op recorder, and a bounded concurrent shutdown of all three pipelines (`DefaultShutdownTimeout` raised to 10s); `metrics.go` defines the duration histogram with semconv buckets, the readiness gauge, pool gauges, and runtime instrumentation; `transport.go` is the QR-004-safe outbound HTTP RoundTripper; `record.go` exposes shared log/span/metric attribute construction; `middleware.go` starts one always-on server span per request, puts it into `c.Request` so handlers and repositories inherit it, records the histogram, and emits the log record with the span context so records carry `trace_id`/`span_id`.
- `server/postgres`: `tracing.go` adds a `pgx.QueryTracer` (operation, namespace, host, SQLSTATE; never SQL text or error messages) and `ConnectionStats`; `pool.go` receives the tracer provider from the recorder.
- `server/main.go`: installs the outbound transport, wires the readiness observer and pool stats, and passes the tracer provider to `postgres.Init`.
- Scope recorded: Cloud Tasks gRPC and Discord instrumentation were dropped because both paths are dormant (no production callers); TASK-0251 captures the Cloud Tasks evaluation with preliminary pros and cons.
- Fixes found during live verification: the pgx tracer originally resolved the global `otel.Tracer` while providers stayed explicit, so database spans were no-ops; `postgres.Init(trace.TracerProvider)` now receives the recorder provider. Request context propagation was also missing, so database spans started root traces; the middleware now stores the span context in `c.Request`.

## Checks

- `gofmt -l`, `go vet ./...`, `go build ./...`, and `go test -race -count=1 ./observability/ ./postgres/` pass.
- Isolated Compose suite `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` passes for every package.
- Root Markdown: `npm run format:markdown`, `format:markdown:check`, `lint:markdown`, `test:markdown-format` (30 passed), and `fmt:check` pass.
- `graphify update .` run after the code changes.

## Development-stack exercise (same code and OTLP contract; staging is unreachable)

- Rebuilt `timeful-development-server` against organization `3JQ1a8basz7mdePUqGq6P7jk0ot`; OpenObserve access logs show `POST /v1/logs`, `/v1/traces` (OTel OTLP Exporter Go/1.46.0), and `/v1/metrics` all returning 200.
- Log-to-trace correlation: request `4ddaa639-8d3e-460f-ab95-77476aa858dc` (404) produced a log record whose `trace_id` equals the request span trace, and the PostgreSQL `SELECT` span shares that trace id (post-fix nesting).
- Traces are queryable in `timeful_server_traces` with `operation_name`, `http_route`, `http_response_status_code`, `service_readiness`, `error_type`, and `span_kind` covering request server spans, PostgreSQL client spans, and outbound HTTP client spans.
- Metrics are queryable with PromQL: `http_server_request_duration_count` labeled by route, method, status, and readiness; `timeful_service_readiness` (1 ready); `db_client_connection_count` (idle 1, used 0); `db_client_connection_max` (10); and Go runtime streams such as `go_goroutine_count` and `go_memory_*`. Metric streams use dots-to-underscores naming, and the histogram also produces `_bucket`, `_count`, `_sum`, `_min`, and `_max` streams.
- PostgreSQL-dependent readiness: with `postgres` stopped, `/api/health` returned 503 (`cd702e68-e087-4420-b392-95563cea6ca0`) and an event read returned 500 (`6b6d21ec-d9ae-4430-b826-41f6f6822b9a`) with `service_readiness=unavailable` and `error_type=failed-to-load-event` in the trace; `timeful_service_readiness` moved to 0 and back to 1 after the restart.
- Outbound HTTP spans: a `POST` to `host.docker.internal` carried `error.type=transport_error` from the OTP path's Listmonk call (a root trace because that caller passes no request context), with no URL path, query, or headers recorded.
- QR-004: scans of 28 trace records, the dedicated 404-with-synthetic-secrets trace, and the metric label output found none of the synthetic edit token, owner cookie value, bearer token, ingest token, service-account email, or root password; `url_full`, `url_path`, `url_query`, and raw request headers are absent. Unit tests assert the same for the transport, span attributes, and metric attributes.
- QR-017 with OpenObserve stopped: 100/100 health requests returned 200 (max 3.0 ms, p95 2.1 ms) and 50/50 event reads returned 404 (max 3.7 ms, p95 1.8 ms). With OpenObserve paused: 100/100 health 200 (max 6.7 ms, p95 3.3 ms) and 50/50 event reads 404 (max 16.3 ms, p95 5.3 ms). The local log grew to 5,636 lines with 8 redacted export-failure lines, and logs, traces, and metrics exports resumed after unpausing. Evidence lives in `/tmp/opencode/t0246-*`.

## Review and scope notes

- AC #1 was amended with approval to accept the development-stack exercise because staging is unreachable, following the TASK-0245 precedent; TASK-0250 carries the staging setup and deferred staging evidence, and a comment there records the metrics and traces expectation.
- The browser E2E exemption was approved for this server-only change; the isolated backend suite and development-stack exercise cover the affected request path.
- Outbound HTTP spans appear as root traces unless a caller passes a request context; threading context through service signatures was intentionally left out of scope, and docs/environments.md records the behavior.

## Review follow-up (2026-09-17)

A follow-up review found mechanical and consistency issues in the staged work; all were addressed before commit:

- `go mod tidy` was rerun after imports changed; `go mod tidy -diff` is empty, and `server/go.mod`/`go.sum` now mark `runtime`, `otlpmetrichttp`, `otlptracehttp`, `sdk/metric`, `otel/metric`, and `otel/trace` as direct dependencies. The tidy check is now recorded alongside the other checks.
- `docs/environments.md`: the first `Stream` use in the metrics paragraph now links the glossary, lowercase `signal` uses that refer to the controlled term are canonicalized per `docs/terminology/README.md`, and the outbound client-span failure rule is documented.
- `error.type` was dropped from `MetricAttributes()` and remains on spans and log records because handler-authored error codes do not guarantee a bounded metric label set; a middleware regression test asserts its absence from the histogram.
- `db.client.connection.count` and `db.client.connection.max` switched from `Int64ObservableGauge` to `Int64ObservableUpDownCounter` to match the OpenTelemetry semantic conventions; tests assert the non-monotonic sum export, and the stream names stay unchanged.
- Outbound HTTP client spans now mark `Error` for 4xx and 5xx responses, matching otelhttp; the transport test covers both classes.
- `Recorder.readiness` is now an `atomic.Pointer`, mirroring `poolStats`, removing the latent `SetReadiness`-vs-request race.
- `provider_test.go` asserts the metrics export sends no `stream-name` header, and the pool metric assertions read non-monotonic sums.

Re-verified after the changes: `gofmt`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./observability/ ./postgres/`, and the root Markdown pipeline (`format:markdown`, `format:markdown:check`, `lint:markdown`, `test:markdown-format`, `fmt:check`) all pass.

The development-stack exercise was not re-run: the changes are covered by unit tests, and the recorded PromQL stream names and query evidence remain valid because non-monotonic sums keep the metric stream names.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Shipped OpenTelemetry metrics and traces from the Go server to OpenObserve over OTLP, extending the log pipeline built in TASK-0245 to every signal in ADR-022.

`server/observability` now owns log, trace, and metric providers over one shared resource (`service.name`, `deployment.environment`) with the same environment-scoped Basic credentials and no new environment variables. A bounded batch span processor carries always-on request spans; a periodic metric reader exports the `http.server.request.duration` histogram with semconv buckets, the `timeful.service.readiness` gauge, PostgreSQL pool gauges, and Go runtime metrics. The request middleware starts one server span per request, stores its context in `c.Request` so handlers and the repository layer inherit it, records the histogram, and emits the log record with the span context, so log records carry native `trace_id`/`span_id`. `server/postgres` adds a `pgx.QueryTracer` that records operation, namespace, host, and SQLSTATE only, and `server/main.go` installs a QR-004-safe outbound HTTP transport that records method, scheme, host, port, status, and a coarse error class, never URLs, query strings, headers, bodies, or transport error text.

Verified on the development stack with the same code and OTLP contract: logs, traces, and metrics all reached the `timeful_development` organization; traces are queryable in `timeful_server_traces` with request, PostgreSQL, and outbound HTTP spans; PromQL returns the request histogram, readiness gauge, pool gauges, and runtime metrics; and log records correlate with traces by native trace id. Stopping PostgreSQL flipped `timeful_service_readiness` to 0 and produced `service_readiness=unavailable` on dependency-failure traces. QR-004 scans of traces and metric labels found none of the synthetic secrets and no URL, query, or header fields. With OpenObserve stopped or paused, 100 health requests and 50 event reads kept their status codes and stayed far inside the QR-006/QR-007 latency bounds while local file and standard-stream diagnostics remained available and the exporters recovered after unpausing.

Checks: `gofmt`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./observability/ ./postgres/`, the isolated Compose suite (all packages pass, including route tests), the root Markdown format/lint checks, and `graphify update .`.

Staging is unreachable from this environment, so AC #1 was amended with approval to accept the development-stack exercise following the TASK-0245 precedent; TASK-0250 carries the staging setup and deferred evidence and now records the metrics and traces expectation in a comment. Browser E2E is exempt by user approval for this server-only change. Outbound HTTP spans are root traces unless a caller passes a request context; threading context through service signatures was intentionally out of scope and is documented in docs/environments.md. TASK-0251 was created to evaluate replacing the dormant Google Cloud Tasks reminder scheduler with a server-owned implementation.
<!-- SECTION:FINAL_SUMMARY:END -->
