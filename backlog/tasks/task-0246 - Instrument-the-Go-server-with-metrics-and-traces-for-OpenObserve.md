---
id: TASK-0246
title: Instrument the Go server with metrics and traces for OpenObserve
status: To Do
assignee: []
created_date: '2026-09-16 12:35'
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
- [ ] #1 The server exports metrics and traces over OTLP to the configured OpenObserve organization
- [ ] #2 Exported telemetry never blocks or fails request handling, including when OpenObserve is unavailable, and the local diagnostic output remains available (QR-017)
- [ ] #3 Traces cover representative request paths, and metrics expose health signals usable together with logs for diagnosis
- [ ] #4 No credential, secret, anonymous edit token, or token-bearing URL or header appears in the exported telemetry (QR-004)
- [ ] #5 Unit tests cover instrumentation setup and attribute construction, and docs/environments.md lists any new variables
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
