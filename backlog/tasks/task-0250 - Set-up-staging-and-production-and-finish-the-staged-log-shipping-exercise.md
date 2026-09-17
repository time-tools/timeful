---
id: TASK-0250
title: Set up staging and production and finish the staged log-shipping exercise
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-16 19:01'
updated_date: '2026-09-17 11:34'
labels: []
dependencies:
  - TASK-0245
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
type: task
ordinal: 251000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Stand up the staging and production environments on the deployment host and finish the staging log-shipping exercise deferred from TASK-0245.

TASK-0245 shipped structured server log records over OTLP and verified QR-010, QR-004, and QR-017 on the development stack, because the shared deployment host was not reachable. Its acceptance criterion accepts the development-stack exercise on that basis, and this task produces the corresponding staging evidence.

Scope: bring up the shared `timeful-observability` project and the staging and production app stacks per DEPLOYMENT.md and docs/environments.md, populate each environment's app env file, provision each environment's OpenObserve organization and environment-scoped ingest service account, and then run the staging exercise for QR-010 and QR-004 plus the QR-017 failure exercise. The development and isolated test stacks are unchanged.

Governing requirements: QR-010, QR-004, QR-017, QR-006, QR-007. Decision record: ADR-022.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The shared OpenObserve project and the staging and production app stacks run on the deployment host per DEPLOYMENT.md and docs/environments.md, with each environment's app env file populated and every interpolated variable declared
- [x] #2 Each environment's organization and environment-scoped ingest credentials are provisioned, and each server ships records to its own `timeful_server_logs` stream in its own organization (QR-019)
- [x] #3 The staging exercise succeeds: a failed staging request's `X-Request-ID` record is queryable in the staging organization with no credential, secret, or anonymous edit token, and a request failing while PostgreSQL is unavailable carries `service_readiness=unavailable` (QR-010, QR-004)
- [x] #4 With staging OpenObserve stopped or slow, the request success rate is unchanged, request latency stays within the QR-006 and QR-007 bounds, and local diagnostic output remains available (QR-017)
- [x] #5 The environment setup, provisioning, and exercise evidence are recorded in the task notes
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Pre-flight on the deployment host (`timeful-cloud-ru`): switch `~/timeful` to `set-up-openobserve`, inspect the existing edge, production stack, DNS, and PostgreSQL state, and back up the production database before any production change.
2. Create `.env.observability` with loopback port, root email, and a freshly generated strong root password; start the shared `timeful-observability` project on the external `timeful-edge` network and verify loopback-only exposure, caps, retention, and disabled telemetry.
3. Provision `timeful_staging` and `timeful_production` organizations on the shared instance and one service account per organization; record each generated organization identifier and service-account token.
4. Populate `.env.staging` and `.env.production` with the four `OPENOBSERVE_*` values and validate every documented Compose command with `docker compose config --quiet`.
5. Start the staging stack and redeploy the production stack from the branch (`up -d --build`), applying the new `folder_events` index migration to production, then confirm readiness and edge routing for both environments.
6. Exercise staging for QR-010 and QR-004: drive a failed request with token-bearing inputs, query the staging organization's `timeful_server_logs` stream by `X-Request-ID`, and confirm no credential, secret, or anonymous edit token ships.
7. Exercise staging readiness: stop staging PostgreSQL, drive a request failure, confirm `service_readiness=unavailable`, then restore.
8. Exercise QR-017 on staging: stop and pause the shared OpenObserve instance, measure request success and latency against the QR-006 and QR-007 bounds, confirm local diagnostics remain available, then restore export.
9. Confirm each environment's logs, metrics, and traces land in its own organization with its own service account, including request, PostgreSQL, and outbound HTTP spans plus PromQL queries for the readiness gauge and request histogram.
10. Record the environment setup, provisioning, and exercise evidence in the task notes, run the required Markdown checks for any documentation changes, update the task, and finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Started 2026-09-17. `set-up-openobserve` (eb8deade) is checked out on the deployment host; the running production stack predates the branch and uses the production PostgreSQL volume `timeful-production-postgres-data` with Goose migrations through `20260915000000`. The branch adds `20260915000001_folder_events_event_unique_idx_drop_predicate.sql`. The shared `timeful-observability` project is not running yet and `.env.observability` does not exist. `.env.staging` and `.env.production` lack the four `OPENOBSERVE_*` variables. DNS resolves `timeful.fun` and `staging.timeful.fun` to the host; the edge serves production (200) and staging returns 502 because no staging server is running. The branch's Caddy configuration is unchanged from the running edge.

Progress 2026-09-17: deployment host checked out to set-up-openobserve (eb8deade); production PostgreSQL backed up to /home/user1/timeful-backups/task-0250-production-pre-20260917T104609Z.dump; the host root filesystem was full and was recovered with a journal vacuum plus removal of legacy images and build cache (all data volumes retained). The shared .env.observability was created with a generated strong root password, and timeful-observability runs on 127.0.0.1:5080 over the timeful-edge network with the documented caps, retention, and disabled telemetry. Provisioned timeful_staging (identifier 3JS6le4nNGJI6oA1deXc4D5HmrG) and timeful_production (identifier 3JS6lelnPZS1dsn0bHNxp7Rr75L), each with an environment-scoped service account; own-organization ingest is 200 and cross-organization ingest and read are 401 (QR-019). Provisioning credentials are kept on the host at /home/user1/timeful-backups/task-0250-observability-credentials.env (mode 600). .env.staging and .env.production now declare all four OPENOBSERVE_* values, and every documented Compose command passes config --quiet. The staging stack runs from the branch: https://staging.timeful.fun/api/health is 200, postgres-migrate exited 0, and records ship to the staging organization. QR-010/QR-004 staging exercise: failed request 84552355-e2f0-4c2f-9e10-8283ddc4f07d (404) returned a record with route /api/events/:eventId, status 404, outcome client_error, duration 8.85 ms, error_type event-not-found, service_readiness ready, and trace/span identifiers; scans of the shipped records and of the local file diagnostics found no synthetic edit token, owner cookie value, bearer value, ingest token, service-account email, or root password. Readiness exercise: with staging PostgreSQL stopped, /api/health returned 503 (bd738717-6315-4fde-bde1-a531b30aa33f) and an event read returned 500 (55835e2e-d152-4a39-8c79-027c21d09d91), both shipping service_readiness=unavailable; restart restored 200. QR-017: with the shared OpenObserve stopped and then paused, 400 requests per condition kept their 200 statuses with curl p95 of 62-85 ms and a server-side maximum of 56.6 ms, while the local server log grew and recorded redacted export failures and export resumed after unpause. Staging traces show request server spans, PostgreSQL client spans, and an outbound HTTP client span (oauth2.googleapis.com, 400, no URL path or query), and PromQL returns the readiness gauge, request histogram series, PostgreSQL pool metrics, and Go runtime metrics. Production is still running the pre-branch code; the production redeploy and its shipping evidence remain.

Production 2026-09-17: production was redeployed from set-up-openobserve (eb8deade) with `up -d --build` after re-checking the pre-change backup at /home/user1/timeful-backups/task-0250-production-pre-20260917T104609Z.dump (pg_restore --list reads it) and the host's 12 GB free space. postgres-migrate exited 0 applying 20260915000001_folder_events_event_unique_idx_drop_predicate.sql, goose_db_version holds 20260915000001 as applied, and folder_events_event_unique_idx is recreated without a predicate as UNIQUE (platform_identity_id, event_id). timeful-production-server-1 and timeful-production-postgres-1 are healthy, https://timeful.fun/api/health is 200, and the documented production and staging Compose commands both pass `config --quiet`.

Production telemetry 2026-09-17: the production organization now holds timeful_server_logs, timeful_server_traces, and the request, PostgreSQL, and Go runtime metric streams, all readable with the production service account. Failed request 0eb9b541-fb3e-485e-81e4-d370ee4e6390 (GET /api/events/:eventId, 404, client_error, 1.81 ms, error_type event-not-found, service_readiness ready, trace 5de309de19e7e8759a5feddcbe6dbca9) is queryable in timeful_server_logs; scans of the shipped records and of the local docker and file diagnostics found no synthetic edit token, owner cookie value, bearer value, ingest token, service-account email, or root password. Trace 5de309de19e7e8759a5feddcbe6dbca9 carries the request server span and a PostgreSQL client span (db_system_name=postgresql, SELECT, child of the request span), and an outbound HTTP client span (POST, oauth2.googleapis.com, 400, url_scheme=https, no URL path or query) is queryable from a bogus-code POST /api/auth/sign-in. PromQL with the production service account returns timeful_service_readiness=1, http_server_request_duration_count by route, status, and readiness, a histogram_quantile p95, and db_client_connection_count for the PostgreSQL pool. QR-019 re-checked after the production stream existed: own-organization reads are 200 for both environments, cross-organization read and ingest are 401 in both directions, and the production organization holds only deployment_environment=production records. Production evidence is retained at /home/user1/timeful-backups/task-0250-production-evidence/ (mode 600).
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @OpenCode
created: 2026-09-16 21:04
---
TASK-0246 (metrics and traces for OpenObserve) accepted the development-stack exercise for its export acceptance criteria because the deployment host is unreachable, matching the TASK-0245 precedent. When this task stands up staging and production and finishes the deferred log-shipping exercise, extend the same exercise to metrics and traces: confirm each server's `/v1/metrics` and `/v1/traces` exports land in its own organization with the environment's service account, that request, PostgreSQL, and outbound HTTP spans are queryable, and that the readiness gauge and request histogram are queryable with PromQL.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Set up the shared timeful-observability instance plus the staging and production app stacks on the deployment host, provisioned one environment-scoped OpenObserve organization and service account per environment, and completed the staging and production log-shipping, readiness, and QR-017 exercises.

- Shared OpenObserve runs loopback-only on 127.0.0.1:5080 over the timeful-edge network with the documented caps, 14-day retention, and disabled telemetry. timeful_staging (3JS6le4nNGJI6oA1deXc4D5HmrG) and timeful_production (3JS6lelnPZS1dsn0bHNxp7Rr75L) each carry one environment-scoped service account; own-organization reads are 200 and cross-organization read and ingest are 401 in both directions (QR-019).
- Staging and production run from set-up-openobserve (eb8deade); both app env files declare the four OPENOBSERVE_* values and every documented Compose command passes config --quiet. Production was redeployed with the 20260915000001_folder_events_event_unique_idx_drop_predicate migration applied (goose version 20260915000001, unique index recreated without the predicate); both environments report healthy readiness and the edge serves 200.
- Staging exercise: failed request 84552355-e2f0-4c2f-9e10-8283ddc4f07d is queryable by X-Request-ID with no credential, secret, or anonymous edit token, and PostgreSQL-unavailable failures carried service_readiness=unavailable (QR-010, QR-004). QR-017: with OpenObserve stopped and then paused, 400 requests per condition kept their 200 statuses with curl p95 of 62-85 ms while local diagnostics stayed available.
- Production exercise and TASK-0246 extension: failed request 0eb9b541-fb3e-485e-81e4-d370ee4e6390 is queryable in timeful_production, its trace carries request and PostgreSQL spans, an outbound HTTP span to oauth2.googleapis.com is queryable, and PromQL returns the readiness gauge, request histogram series, and PostgreSQL pool metrics with the production service account.
- Evidence is recorded in the task notes and host-side artifacts under /home/user1/timeful-backups/. No runtime code or tracked documentation changed, so no unit, e2e, or Markdown format checks were required.
<!-- SECTION:FINAL_SUMMARY:END -->
