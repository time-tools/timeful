---
id: TASK-0245
title: Ship structured server logs to OpenObserve over OTLP
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-16 12:35'
updated_date: '2026-09-16 19:02'
labels: []
dependencies:
  - TASK-0244
documentation:
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/quality/qr/QR-004.md
  - docs/requirements/quality/qr/QR-010.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/requirements/quality/qr/QR-006.md
  - docs/requirements/quality/qr/QR-007.md
  - docs/environments.md
  - DEPLOYMENT.md
priority: medium
type: enhancement
ordinal: 246000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Rework Go server diagnostics so an operator can correlate a failed staging or production request with structured records and service health (QR-010), shipping logs as OpenTelemetry OTLP records directly to OpenObserve with the OpenTelemetry Go SDK and no sidecar log shipper.

Shipped records must exclude credentials, secrets, anonymous edit tokens, and token-bearing URLs and headers (QR-004). Export must never block or fail the request path, and the existing local diagnostic output must remain available while OpenObserve is unavailable (QR-017).

The OpenObserve endpoint, organization, and environment-scoped ingestion credentials come from the environment contract established by TASK-0244, which this task depends on.

Governing requirements: QR-010, QR-004, QR-017. Decision record: ADR-022.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The server emits structured diagnostic records over OTLP to the configured OpenObserve organization, and the records are queryable for a failed request driven in a deployed environment; the development stack exercised the same code and OTLP contract, and TASK-0250 produces the staging evidence
- [x] #2 Records carry request identity, route, outcome, latency, and error context sufficient to correlate a failed request with its diagnostic records and service health (QR-010), including the service's readiness state when the failure coincides with an unavailable dependency
- [x] #3 No credential, secret, anonymous edit token, or token-bearing URL or header appears in the shipped records (QR-004)
- [x] #4 With OpenObserve stopped or slow, the request success rate is unchanged, request latency stays within the bounds of QR-006 and QR-007, and local diagnostic records remain available after the exercise (QR-017)
- [x] #5 The existing file and standard-stream diagnostics remain available, and export is asynchronous or otherwise bounded so it cannot block the request path
- [x] #6 Unit tests cover record construction and redaction, and the task records the QR-017 operational exercise evidence
- [x] #7 docs/environments.md documents the server-side OpenObserve variables consumed by the change
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
1. Add the OpenTelemetry Go logs stack to `server/go.mod`: `go.opentelemetry.io/otel/log`, `.../sdk/log`, `.../exporters/otlp/otlplog/otlploghttp`, `.../sdk`, `.../sdk/resource` pinned to the latest stable line (otel v1.46.0 / log v0.22.0), raising the module Go directive as the dependencies require; run `go mod tidy` and compile.
2. Add a new `server/observability` package:
   - `config.go`: read `OPENOBSERVE_ENDPOINT`, `OPENOBSERVE_ORGANIZATION_ID`, `OPENOBSERVE_INGEST_USERNAME`, `OPENOBSERVE_INGEST_PASSWORD` and `APP_ENV`; compute the OTLP/HTTP logs URL `<endpoint>/api/<org-id>/v1/logs`; enabled only when all four values are non-blank; never fail startup.
   - `redact.go`: `Redact` for credentials, secrets, token-bearing URLs and headers, URL userinfo, and sensitive query parameters (`token`, `access_token`, `id_token`, `refresh_token`, `edit_token`, `owner_token`, `code`, `password`, `secret`, `api_key`, etc.).
   - `record.go`: canonical `RequestCompletion` internal struct and its OTel `log.Record` construction with request.id, http.request.method, http.route, http.response.status_code, http.response.outcome, http.response.duration_ms, error.type, error.message, service.readiness.
   - `provider.go`: `Recorder` that builds a `sdklog.LoggerProvider` over an OTLP HTTP exporter with Basic auth and `stream-name: timeful_server_logs`, and a batch processor (bounded queue, async, drop-on-full) so export never touches the request path; nil-safe no-op when disabled; `Shutdown` flushes with a bounded context; local export errors go through `Redact` to `logger.StdErr`.
   - `middleware.go`: Gin middleware that generates the request correlation UUID, returns it as `X-Request-ID`, captures the bounded JSON `error` field only for responses >= 400, and emits the completion record after `c.Next()`, including the injected readiness state.
   - `readiness.go`: background PostgreSQL readiness monitor around an injected probe with interval/timeout and an atomic last state.
3. Wire `server/main.go`: initialize the recorder from the environment after the logger; install the request middleware before `gin.Recovery()`; start the readiness monitor against `postgres.Ping`; add graceful shutdown so the export queue is flushed and the monitor stops on SIGINT/SIGTERM.
4. Add package unit tests: redaction table tests, config parsing/URL/auth-header tests, record construction tests, middleware emission tests against an in-memory exporter (200/404/500, token-bearing request headers/URLs, response header), error-field capture tests, and readiness monitor tests.
5. Document the server-side contract in `docs/environments.md`: variables consumed, OTLP logs URL, Basic auth, stream name, disabled-export behavior for incomplete values, asynchronous bounded export, and shipped record fields; add a short operator query note in `DEPLOYMENT.md`.
6. Verify: run `go test` for the new package on the host, then the full suite through the isolated Compose stack; run root Markdown format/lint checks.
7. Operational exercise on the development Compose stack (the shared instance is not reachable here): start the development OpenObserve instance, provision the `timeful_development` organization and service account, drive a failed request, confirm the records are queryable with no secrets, then exercise QR-017 with OpenObserve stopped/slow and confirm unchanged success/latency and retained local logs; record the evidence in the task notes.

Executed: dependencies pinned to otel v1.46.0 / log v0.22.0, package implemented as planned, middleware installed before Recovery, graceful shutdown added, docs updated, unit plus isolated-route suites green, and the development-stack QR-010/QR-004/QR-017 exercises recorded. Deviation: the injected-exporter constructor is unexported (`newRecorder`) because only in-package tests use it.

Remaining: run `cd e2e && npm run test:e2e -- --project=firefox-desktop` after `npx playwright install firefox` (or record why browser E2E is not required), review the diff, write the Final Summary, and move the task to Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation

- Added `server/observability`: environment contract parsing (`OPENOBSERVE_ENDPOINT`, `OPENOBSERVE_ORGANIZATION_ID`, `OPENOBSERVE_INGEST_USERNAME`, `OPENOBSERVE_INGEST_PASSWORD`, `APP_ENV`), an OTLP/HTTP logs exporter with Basic auth and the `stream-name: timeful_server_logs` header, a bounded batch processor (2048-record queue, 5s export interval, 5s export timeout, drop-on-full), Gin request middleware, redaction, and a background PostgreSQL readiness monitor.
- `server/main.go` installs `observability.RequestMiddleware` before `gin.Recovery()` (so recovered panics still produce a record), starts the readiness monitor against `postgres.Ping`, and drains the export queue on SIGINT/SIGTERM through a graceful `http.Server` shutdown.
- `server/observability` is inert when the contract is incomplete: `Start` returns an error, `main` logs a local warning, and the nil `Recorder` makes every method a no-op, so the isolated test stack and local diagnostics are unaffected.
- Dependencies: `go.opentelemetry.io/otel` v1.46.0, `otel/log` v0.22.0, `otel/sdk` v1.46.0, `otel/sdk/log` v0.22.0, `otel/exporters/otlp/otlplog/otlploghttp` v0.22.0. The module Go directive rose to 1.25.0 as those modules require; the Dockerfile already builds with Go 1.26.4.
- Docs: `docs/environments.md` gained the "Server-Side Structured Diagnostics" subsection (URL, Basic auth, stream, record fields, X-Request-ID, bounded export, incomplete-contract behavior, QR-004), and `DEPLOYMENT.md` gained the operator query note.

## Checks run

- `gofmt -l`, `go vet ./...`, `go test ./observability/ ./appenv/ .` pass.
- Isolated Compose suite passes: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` (`go test ./... -count=1`) — all packages `ok`, including the new `observability` package.
- Root Markdown: `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format` (30 passed), `npm run fmt:check` all pass.
- `graphify update .` run after the code changes.

## QR-010 / QR-004 exercise (development stack, same code and OTLP contract)

- Rebuilt and restarted `timeful-development-server` against the `timeful_development` organization (`3JQ1a8basz7mdePUqGq6P7jk0ot`) and service account recorded by TASK-0244.
- Failed request `GET /api/events/NOTAREAL?editToken=anonymous-edit-token-secret` with an `Authorization: Bearer` header and a `timeful_owner_*` cookie returned 404 `{"error":"event-not-found"}` with `X-Request-ID: b1115265-8647-420d-9911-de336f8a357d`.
- OpenObserve SQL over the `timeful_server_logs` stream returned the matching record: `request_id` equal to the response header, `http_route=/api/events/:eventId`, `http_response_status_code=404`, `http_response_outcome=client_error`, `http_response_duration_ms=6.35`, `error_type=event-not-found`, `service_readiness=ready`.
- Readiness on dependent failure: with `postgres` stopped, the monitor sampled `unavailable` and the `/api/health` 503 record (`X-Request-ID: 84b43e3d-6f96-42bb-8b41-7ed260c7172e`) shipped `service_readiness=unavailable` and `error_type=http_5xx`.
- Recovered panic: `POST /api/auth/otp/send` returned 500 through `gin.Recovery` and shipped a record with `http_response_status_code=500` and `error_type=http_5xx`.
- QR-004: all 307 records in the window were fetched and searched for the synthetic edit token, owner cookie value, Bearer header value, the ingest token, the service-account email, and the root password; none appear. Shipped fields are only body, environment, error type, method, duration, outcome, status, route, request id, service name, readiness, severity, and instrumentation library (no raw URLs, headers, cookies, or bodies).
- Staging is not reachable from this environment; the development instance with the same code and OTLP contract substitutes for it, matching the TASK-0244 precedent.

## QR-017 exercise (OpenObserve stopped, then paused)

- Stopped: 200/200 `GET /api/health/live` returned 200 (max 6.0 ms, p95 2.3 ms); 100/100 failed event reads returned the expected 404 (max 11.9 ms, p95 1.6 ms).
- Paused (slow, no response): 200/200 health requests returned 200 (max 3.7 ms, p95 0.8 ms); 100/100 event reads returned 404 (max 1.3 ms, p95 1.0 ms).
- Representative event workload with OpenObserve stopped: 500/500 event views returned 200 (max 7.8 ms, p95 4.3 ms) and 100/100 availability saves returned 200 (max 5.1 ms, p95 4.6 ms), all far inside the two-second 95th-percentile bounds of QR-006 and QR-007.
- Local diagnostics survived: `/app/logs/server.log` grew to 4,929 lines during the exercises and still contains GIN access lines plus redacted `observability export failed: ...` lines; stdout output stayed live. New records shipped again after OpenObserve restarted.
- Evidence logs: `/tmp/opencode/qr017-obs-down-health.txt`, `/tmp/opencode/qr017-obs-down-events.txt`, `/tmp/opencode/qr017-obs-slow-health.txt`, `/tmp/opencode/qr017-obs-slow-events.txt`, `/tmp/opencode/qr006-views-obs-down.txt`, `/tmp/opencode/qr006-saves-obs-down.txt`, `/tmp/opencode/qr004-records.json`.

## Remaining work for the next session

- DoD #3 browser E2E has not been run. Playwright's Firefox build is not installed on this machine (`~/.cache/ms-playwright` has only ffmpeg); the attempted `npx playwright install firefox` was interrupted. The isolated test stack and frontend dependencies are present, and the full backend suite plus the live development-stack exercise already covered the integrated server path.
- Final review of the diff, Final Summary, and status transition to Done remain.

## Observed pre-existing issues (out of scope, not fixed)

- The existing GIN text access log records the raw query string, so a synthetic token-bearing query appears in the local file output; shipped records never contain it, and the task scopes QR-004 here to shipped records.
- `POST /api/auth/otp/send` panics with a nil-pointer dereference in `services/listmonk` when the configured Listmonk endpoint is unreachable; `gin.Recovery` returns 500 and the middleware records it. Pre-existing and unrelated.

Implementation complete and verified except the browser E2E run; see Implementation Notes for the full evidence and remaining work. Another session should finish DoD #3 (or record why it is not required), write the Final Summary, and move the task to Done.

## Follow-up session: review fixes and finalization

- A review of the staged change found and fixed: UTF-8-safe truncation in `redact.go` (a split rune made `proto.Marshal` fail the whole export batch), the `TestReadinessMonitorTracksProbeResult` data race, asymmetric password trimming, `LogsURL` string concatenation (now parsed `JoinPath` with query and fragment rejection), and the shutdown budget versus Docker's default stop grace (`compose.yaml` now sets `stop_grace_period: 30s`).
- QR-004 local gap closed: the GIN access log now passes its request path and handler error messages through `observability.Redact`, so token-bearing query parameters no longer reach the file or standard-stream diagnostic output; `docs/environments.md` records the behavior.
- AC #1 amended with approval to accept the development-stack exercise; TASK-0250 tracks the staging and production setup and the deferred staging evidence.
- The browser E2E requirement was replaced by an explicitly user-approved exemption; the isolated backend suite and the development-stack exercise cover the affected request path.
- Checks after the fixes: `gofmt -l`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./observability/`, the isolated Compose suite (`go test ./... -count=1`, all packages `ok`), and the root Markdown format and lint checks all pass. `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml config --quiet` validates the `stop_grace_period` addition.
- Deferred minor items, recorded but not fixed: the first readiness probe runs synchronously in `Start`, `otel.SetErrorHandler` is a process-global side effect, and queue drops during an outage are not surfaced in local diagnostics.
- This section supersedes the "Remaining work for the next session" section above and the GIN raw-query bullet under "Observed pre-existing issues".
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @OpenCode
created: 2026-09-16 19:02
---
## Review and decisions (2026-09-16)

Findings and decisions recorded before finalization:

- AC #1 accepted the development-stack exercise because staging is unreachable; TASK-0250 tracks the staging and production setup and the deferred staging evidence.
- The browser E2E requirement was replaced by an explicitly user-approved exemption; the isolated backend suite and the development-stack exercise cover the affected request path.
- QR-004 gap in the local GIN access log (raw query string with edit tokens): confirmed, fixed in `server/main.go` by passing the access-log path and handler error messages through `observability.Redact`; scope expanded to TASK-0245 with approval.
- Code fixes applied: UTF-8-safe truncation in `redact.go` (a split rune failed `proto.Marshal` and dropped the whole export batch), the `TestReadinessMonitorTracksProbeResult` data race, asymmetric password trimming, `LogsURL` construction (parsed `JoinPath`, query and fragment rejected), and `stop_grace_period: 30s` in `compose.yaml`.
- Deferred minor items, recorded but not fixed: the first readiness probe runs synchronously in `Start`, `otel.SetErrorHandler` is a process-global side effect, and `X-Request-ID` is not exposed through CORS.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Shipped structured server log records to OpenObserve over OTLP directly from the OpenTelemetry Go SDK, with no sidecar shipper. Added `server/observability` (environment contract and OTLP/HTTP logs URL, Basic auth with `stream-name: timeful_server_logs`, bounded asynchronous batch export with drop-on-full, Gin request middleware emitting `X-Request-ID` and bounded error context, defense-in-depth redaction, and a background PostgreSQL readiness monitor) and wired it into `server/main.go` before `gin.Recovery()` with graceful shutdown that flushes the export queue.

Verified QR-010, QR-004, and QR-017 on the development stack: a failed request's correlation identifier returned its record with route, status, outcome, latency, error type, and readiness; a PostgreSQL outage shipped `service_readiness=unavailable`; 307 shipped records contained none of the synthetic edit token, owner cookie, Bearer header, ingest token, service-account email, or root password; and with OpenObserve stopped or slow, health, event-view, and availability-save requests stayed within the QR-006 and QR-007 latency bounds while local file and standard-stream diagnostics remained available. Staging was unreachable, so AC #1 was amended with approval to accept the development-stack exercise, and TASK-0250 carries the staging and production setup plus the deferred staging evidence.

Review fixes hardened the change: UTF-8-safe error truncation (a split rune previously failed protobuf marshalling and dropped a whole export batch), the readiness test data race, symmetric password trimming, URL construction via parsed `JoinPath` with query and fragment rejection, `stop_grace_period: 30s` so the final flush outlives Docker's default stop grace, and redaction of the local GIN access-log path so QR-004 holds for file and standard-stream output too.

Checks: `gofmt`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./observability/`, the isolated Compose suite (`go test ./... -count=1`, all packages `ok`, including the new regression tests), Compose config validation, and the root Markdown format and lint checks all pass. Browser E2E is not required by an explicitly user-approved exemption recorded in the Definition of Done.
<!-- SECTION:FINAL_SUMMARY:END -->
