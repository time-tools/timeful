---
id: TASK-0203
title: >-
  Remove remaining MongoDB-era and store-specific wording from live docs and
  active requirements
status: Done
assignee:
  - opencode
created_date: '2026-09-11 20:12'
updated_date: '2026-09-11 20:14'
labels: []
dependencies: []
modified_files:
  - docs/requirements/README.md
  - docs/requirements/functional/fr/FR-081.md
  - docs/requirements/functional/fr/FR-082.md
  - docs/requirements/functional/fr/FR-083.md
  - docs/requirements/functional/fr/FR-121.md
  - docs/requirements/functional/fr/FR-123.md
  - docs/requirements/quality/qr/QR-003.md
  - docs/requirements/quality/qr/QR-011.md
  - docs/requirements/quality/qr/QR-013.md
priority: low
type: chore
ordinal: 242000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Finish the PostgreSQL-only cleanup of active product requirements after the TASK-0199 retirement.

Scope:
- Active/proposed requirements only: remove redundant or migration-era store wording. FR-081/FR-082/FR-083 titles and H1s drop the "PostgreSQL" qualifier; FR-121 drops retained/legacy/two-store framing while keeping PostgreSQL as the single authoritative store; FR-123, QR-003, QR-011, and QR-013 drop store or cutover wording; docs/requirements/README.md index rows match any changed titles.

Note: the runbook and prettier-fixture portions of this cleanup were already delivered by TASK-0202 (commit dc2781a3), so they are out of scope here.

Constraints:
- Deprecated FR-122 and QR-012, CAND-212/CAND-213 provenance, ADRs, applied goose migration comments, and backlog/handoff history stay unchanged.
- Requirement IDs and filenames are permanent; only titles and statements change.
- Follow docs/requirements/AGENTS.md and the terminology linking rules.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 No active or proposed FR/QR carries migration-era or redundant store wording: FR-081/FR-082/FR-083 titles and H1s drop the PostgreSQL qualifier, FR-121 keeps only the PostgreSQL single-store rule without retained/legacy/two-store wording, FR-123/QR-003/QR-011/QR-013 are storage-neutral, and docs/requirements/README.md index rows match the changed titles.
- [x] #2 Deprecated FR-122/QR-012, CAND-212/CAND-213, ADRs, applied migration comments, and backlog/handoff history are unchanged.
- [x] #3 Root npm run fmt:check, npm run format:markdown:check, npm run lint:markdown, and npm run test:markdown-format pass, and graphify update . refreshes the graph.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-11)

- FR-081/FR-082/FR-083: dropped the redundant PostgreSQL qualifier from front matter titles and H1s; bodies unchanged.
- FR-121: merged the core/retained record-kind lists into one list and restated the durable rule as PostgreSQL sole authoritative store; removed legacy-document, legacy-account, and two-store wording.
- FR-123: "Deletion mechanics remain implementation details."
- QR-003 Environment: "The event is stored."; QR-011 Environment: "A transfer is pending..."; QR-013 Environment: "Calendar data is persisted with provider credentials encrypted under a configured key...".
- docs/requirements/README.md: FR-081/FR-082/FR-083 index rows updated; `npm run format:markdown` realigned the table.
- Deprecated FR-122/QR-012, CAND-212/CAND-213, ADRs, applied migration comments, and backlog/handoff history unchanged.

## Scope note

The runbook and prettier-fixture parts of the original cleanup plan were delivered concurrently by TASK-0202 (commit dc2781a3); this task was narrowed to the requirement wording before implementation.

## Verification

- AC1: `rg -i mongo docs/requirements docs/postgres-operations.md prettier/` returns only FR-122, QR-012, CAND-212/CAND-213, and the FR-122 index row; no active FR/QR names a store except FR-121's single-store rule; the only remaining "retained-data cutover" hits are ADR-016/ADR-017 history.
- AC2: `git status` shows only the nine requirement files; no deprecated requirement, candidate, ADR, migration, or backlog/handoff file is modified (backlog/backlog.md was already dirty with the user's inbox edits before this task).
- AC3: `npm run fmt:check`, `npm run format:markdown:check`, `npm run lint:markdown`, and `npm run test:markdown-format` (30 tests) pass from the repo root; `graphify update .` rebuilt the graph (5495 nodes, 9713 edges, 406 communities).
- Unit and E2E tests are exempt: documentation-only change per BACKLOG_WORKFLOW.md and the task DoD.
- No commit was made.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Removed migration-era and redundant store wording from active, proposed product requirements; the runbook and test-fixture portions were already delivered by TASK-0202 (commit dc2781a3), so this task covered only the requirement wording. Documentation-only change: no runtime code, configuration, or tests were modified.

## Changes

- FR-081/FR-082/FR-083: dropped the redundant "PostgreSQL" qualifier from front matter titles and H1s; bodies unchanged.
- FR-121: merged the core/retained record-kind lists into one list, restated the durable rule as "PostgreSQL is the sole authoritative store for every record kind, and the system shall not read or write any record kind from a second store," and removed legacy-document, legacy-account, and two-store wording.
- FR-123: "Deletion mechanics remain implementation details."
- QR-003 Environment: "The event is stored."; QR-011 Environment: "A transfer is pending..."; QR-013 Environment: "Calendar data is persisted with provider credentials encrypted under a configured key...".
- docs/requirements/README.md: FR-081/FR-082/FR-083 index rows match the new titles; `npm run format:markdown` realigned the table.
- Deprecated FR-122/QR-012, CAND-212/CAND-213, ADRs, applied migration comments, and backlog/handoff history were left unchanged.

## Verification per acceptance criterion

- #1: `rg -i mongo docs/requirements docs/postgres-operations.md prettier/` returns only FR-122, QR-012, CAND-212/CAND-213, and the FR-122 index row; no active FR/QR names a store except FR-121's single-store rule; the only remaining "retained-data cutover" hits are ADR-016/ADR-017 history.
- #2: `git status` shows only the nine requirement files changed; no deprecated requirement, candidate, ADR, migration, or backlog/handoff file was modified (backlog/backlog.md was already dirty with the user's inbox edits before this task).
- #3: `npm run fmt:check`, `npm run format:markdown:check`, `npm run lint:markdown`, and `npm run test:markdown-format` (30 tests) pass from the repo root; `graphify update .` rebuilt the graph (5495 nodes, 9713 edges, 406 communities).

## Notes

- Unit and E2E tests are exempt for this documentation-only change per BACKLOG_WORKFLOW.md and the task DoD.
- No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
