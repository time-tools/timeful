---
id: TASK-0245
title: Ship structured server logs to OpenObserve over OTLP
status: To Do
assignee: []
created_date: '2026-09-16 12:35'
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
- [ ] #1 The server emits structured diagnostic records over OTLP to the configured OpenObserve organization, and the records are queryable there for a failed request driven in staging
- [ ] #2 Records carry request identity, route, outcome, latency, and error context sufficient to correlate a failed request with its diagnostic records and service health (QR-010)
- [ ] #3 No credential, secret, anonymous edit token, or token-bearing URL or header appears in the shipped records (QR-004)
- [ ] #4 With OpenObserve stopped or slow, the request success rate is unchanged, request latency stays within the bounds of QR-006 and QR-007, and local diagnostic records remain available after the exercise (QR-017)
- [ ] #5 The existing file and standard-stream diagnostics remain available, and export is asynchronous or otherwise bounded so it cannot block the request path
- [ ] #6 Unit tests cover record construction and redaction, and the task records the QR-017 operational exercise evidence
- [ ] #7 docs/environments.md documents the server-side OpenObserve variables consumed by the change
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
