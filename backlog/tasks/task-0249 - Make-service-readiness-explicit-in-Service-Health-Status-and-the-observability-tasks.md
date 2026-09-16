---
id: TASK-0249
title: >-
  Make service readiness explicit in Service Health Status and the observability
  tasks
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-16 15:46'
updated_date: '2026-09-16 15:46'
labels: []
dependencies: []
references:
  - docs/terminology/glossary.md
  - docs/requirements/quality/qr/QR-010.md
  - docs/environments.md
  - TASK-0245
  - TASK-0246
modified_files:
  - docs/terminology/glossary.md
priority: medium
type: docs
ordinal: 250000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Service Health Status glossary entry (TASK-0248, used by QR-010 and ADR-022) describes only a generic reported health state. The server's actual reports are process liveness (/api/health/live) and PostgreSQL-dependent readiness (/api/health), documented in docs/environments.md, so the term never tells a reader that readiness is included.

Make the definition concrete instead of adding a separate Service Readiness Status term: name liveness and readiness in docs/terminology/glossary.md and list docs/environments.md as a second authoritative context. Keep QR-010's normative text unchanged because Service Health Status already covers the required correlation. Make readiness explicit in the acceptance criteria of TASK-0245 (structured records) and TASK-0246 (metrics and traces) so the instrumentation ships a concrete readiness signal rather than a generic health phrase.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/terminology/glossary.md defines Service Health Status as covering the service's process liveness and PostgreSQL-dependent readiness, lists QR-010 and docs/environments.md as authoritative contexts, keeps the TOC unchanged, and adds no new glossary term
- [x] #2 QR-010 Artifact, Response, and Response Measure are unchanged, and no requirement's normative meaning changes
- [x] #3 TASK-0245 and TASK-0246 acceptance criteria name readiness as part of the health signal to correlate or export
- [x] #4 npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass from the repository root
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
1. Rewrite the Service Health Status definition in docs/terminology/glossary.md to name the service's process liveness and PostgreSQL-dependent readiness to serve requests, keeping the existing Structured Log Records cross-link.
2. Add [Environment Files](../environments.md) to the entry's authoritative context alongside QR-010; leave the heading, TOC, and QR-010 text unchanged.
3. Append readiness wording to TASK-0245 acceptance criterion #2 and TASK-0246 acceptance criterion #3 through Backlog MCP.
4. Verify QR-010 is unchanged and run the root Markdown checks: format:markdown, format:markdown:check, lint:markdown, test:markdown-format, and fmt:check.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Made service readiness explicit through the existing Service Health Status term and the observability follow-ups.

- docs/terminology/glossary.md: Service Health Status now covers the service's process liveness and PostgreSQL-dependent readiness to serve requests and lists QR-010 plus Environment Files as authoritative contexts; heading and TOC unchanged, and no new glossary term was added.
- QR-010 is unchanged: git diff for the file is empty, so readiness coverage comes through the Service Health Status definition and no requirement's normative meaning changed.
- TASK-0245 acceptance criterion #2 now requires the shipped records to include the service's readiness state when a failure coincides with an unavailable dependency; TASK-0246 acceptance criterion #3 now requires metrics to expose PostgreSQL-dependent readiness that explains dependency-related request failures.
- Verification: npm run format:markdown made no changes; format:markdown:check, lint:markdown, test:markdown-format (2 files, 30 tests), and fmt:check all pass from the repository root. Graphify resync was skipped because this repository keeps Graphify cache updates in separate chore(graphify) commits.
<!-- SECTION:FINAL_SUMMARY:END -->
