---
id: TASK-0250
title: Set up staging and production and finish the staged log-shipping exercise
status: To Do
assignee: []
created_date: '2026-09-16 19:01'
updated_date: '2026-09-16 21:04'
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
- [ ] #1 The shared OpenObserve project and the staging and production app stacks run on the deployment host per DEPLOYMENT.md and docs/environments.md, with each environment's app env file populated and every interpolated variable declared
- [ ] #2 Each environment's organization and environment-scoped ingest credentials are provisioned, and each server ships records to its own `timeful_server_logs` stream in its own organization (QR-019)
- [ ] #3 The staging exercise succeeds: a failed staging request's `X-Request-ID` record is queryable in the staging organization with no credential, secret, or anonymous edit token, and a request failing while PostgreSQL is unavailable carries `service_readiness=unavailable` (QR-010, QR-004)
- [ ] #4 With staging OpenObserve stopped or slow, the request success rate is unchanged, request latency stays within the QR-006 and QR-007 bounds, and local diagnostic output remains available (QR-017)
- [ ] #5 The environment setup, provisioning, and exercise evidence are recorded in the task notes
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @OpenCode
created: 2026-09-16 21:04
---
TASK-0246 (metrics and traces for OpenObserve) accepted the development-stack exercise for its export acceptance criteria because the deployment host is unreachable, matching the TASK-0245 precedent. When this task stands up staging and production and finishes the deferred log-shipping exercise, extend the same exercise to metrics and traces: confirm each server's `/v1/metrics` and `/v1/traces` exports land in its own organization with the environment's service account, that request, PostgreSQL, and outbound HTTP spans are queryable, and that the readiness gauge and request histogram are queryable with PromQL.
---
<!-- COMMENTS:END -->
