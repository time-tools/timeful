---
id: TASK-0243
title: >-
  Clarify QR-017 signal scope and add QR-019 for observability environment
  isolation
status: To Do
assignee: []
created_date: '2026-09-16 12:32'
labels: []
dependencies: []
documentation:
  - docs/requirements/AGENTS.md
  - docs/requirements/README.md
  - docs/requirements/quality/README.md
  - docs/requirements/quality/qr/TEMPLATE.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/design/architecture/adr/ADR-022.md
  - docs/design/architecture/README.md
priority: medium
type: docs
ordinal: 244000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Update the observability quality requirements so every observability signal is covered and the environment-isolation property recorded in ADR-022 is tracked.

QR-017 ("Keep serving requests when observability ingestion fails") currently reads log-centrically in its Environment, Response, and Response Measure: it says "diagnostic export", "exported diagnostics", and "diagnostic records ... locally". Revise that wording to be signal-agnostic for observability data so logs, metrics, and traces all trace to the one requirement; keep the reliability / fault tolerance classification, all six scenario subsections, and the local diagnostic output baseline.

Add QR-019 for the environment-isolation property in ADR-022's decision outcome: an environment-scoped credential may read and ingest only its own environment's organization and streams.

Link QR-019 from ADR-022's addresses.constrains metadata and its constrained-requirements body sentence, and add its row to the Quality Requirements table in docs/requirements/README.md.

Constraint: do not modify QR-004, QR-010, QR-015, QR-016, or QR-018, and do not add or change any ADR other than linking QR-019 into ADR-022. Follow docs/requirements/AGENTS.md, docs/requirements/README.md, docs/requirements/quality/README.md, qr/TEMPLATE.md, and docs/design/architecture/README.md authoring rules, including one sentence per physical source line.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 QR-017's Environment, Response, and Response Measure use signal-agnostic observability-data wording, keep all six scenario subsections, keep exactly one valid ISO/IEC 25010 classification pair (reliability / fault tolerance), and stay self-contained without citing the ADR
- [ ] #2 QR-019 exists from qr/TEMPLATE.md with exactly one valid ISO/IEC 25010 classification pair, components infrastructure, status proposed, all six scenario subsections, and no ADR citation
- [ ] #3 QR-019's scenario rejects an environment-scoped credential that attempts to read or ingest another environment's organization or streams, while the same credential succeeds against its own environment
- [ ] #4 ADR-022 lists QR-019 in addresses.constrains and in its constrained-requirements sentence, and no other ADR changes
- [ ] #5 docs/requirements/README.md lists QR-019 after QR-018 in the Quality Requirements table with a stable relative link
- [ ] #6 npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass from the repository root
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
