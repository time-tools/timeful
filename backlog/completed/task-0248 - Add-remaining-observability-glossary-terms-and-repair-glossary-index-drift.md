---
id: TASK-0248
title: Add remaining observability glossary terms and repair glossary index drift
status: Done
assignee:
  - OpenCode
created_date: '2026-09-16 15:11'
updated_date: '2026-09-16 15:26'
labels: []
dependencies: []
references:
  - docs/terminology/glossary.md
  - docs/terminology/README.md
  - docs/requirements/quality/qr/QR-010.md
  - docs/requirements/quality/qr/QR-015.md
  - docs/requirements/quality/qr/QR-016.md
  - docs/requirements/quality/qr/QR-018.md
  - docs/requirements/quality/qr/QR-019.md
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/migration/backlog-fr-inventory-candidates/CAND-216.md
priority: medium
type: docs
ordinal: 249000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0243 added Observability Data and Diagnostic Output, but six further observability concepts recur across ADR-022, QR-010, QR-015, QR-016, QR-017, QR-018, QR-019, and TASK-0247 without controlled forms: Signal, Structured Log Records, Service Health Status, Retention Window, Stream, and Environment-Scoped Observability Credential. Add them to the Observability section of docs/terminology/glossary.md with concise definitions, authoritative-context links, and TOC entries; restore the glossary TOC entries for Archived Event, Account Profile, and the Integrations section so the TOC matches the body; fix the stale glossary anchor in CAND-216; and update TASK-0247 to align the corpus against the expanded glossary.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/terminology/glossary.md defines Signal, Structured Log Records, Service Health Status, Retention Window, Stream, and Environment-Scoped Observability Credential with concise definitions, authoritative-context links, and TOC entries in the Observability section, and TOC order matches body order
- [x] #2 The glossary TOC lists Archived Event, Account Profile, Integrations, and Calendar Connection and its section order matches the body
- [x] #3 The new definitions use controlled terms with first-occurrence links and bold repeats per docs/terminology/README.md, and no requirement or ADR text changes in this task
- [x] #4 CAND-216's glossary link resolves to the current heading anchor
- [x] #5 TASK-0247's Backlog description and acceptance criteria cover the expanded Observability term set and its dependency on this task is recorded
- [x] #6 npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass from the repository root
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
1. Create this task and record TASK-0247's dependency on it.
2. Add the six Observability entries and their TOC lines to docs/terminology/glossary.md.
3. Restore the missing TOC entries and match TOC order to the body.
4. Fix the stale glossary anchor in docs/requirements/migration/backlog-fr-inventory-candidates/CAND-216.md.
5. Update TASK-0247's description and acceptance criteria for the expanded term set and reset it to To Do.
6. Run npm run format:markdown, then format:markdown:check, lint:markdown, test:markdown-format, and fmt:check from the repository root.
7. Record the final summary and mark this task Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Audit that led to this task: QR-010's "structured log records", ADR-022's "structured diagnostic records", QR-018/019's streams, QR-015/018's retention windows, QR-016/019's environment-scoped credentials, and ADR-022's signals had no controlled forms, which blocked TASK-0247 from aligning them without broadening or leaving references generic. Glossary TOC drift and the CAND-216 stale anchor were found during the same audit. Notation decisions: Retention Window is defined observability-scoped and the generic uses in ADR-013/016/017 stay untouched; Stream is defined as an OpenObserve storage container; Signal is a class of Observability Data.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the six remaining Observability glossary terms and repaired glossary index drift.

- Glossary: new Signal, Structured Log Records, Service Health Status, Retention Window, Stream, and Environment-Scoped Observability Credential entries in the Observability section, each with a concise definition, authoritative-context links, and a TOC line; definitions link Observability Data, Diagnostic Output, and each other per the terminology rules, and the Structured Log Records, Service Health Status, and Retention Window wording was tightened after review.
- TOC: restored the missing Archived Event, Account Profile, and Integrations/Calendar Connection entries; a script check confirms 97 body headings and 97 TOC entries in identical order.
- CAND-216: repointed the stale Show all hours link and label to the current "Collapse disabled times" Option anchor; a repository-wide audit of glossary.md anchors finds no dangling anchors.
- TASK-0247: description and acceptance criteria broadened to the expanded term set, and its dependency on TASK-0248 recorded.
- Verification: npm run format:markdown made no changes; format:markdown:check, lint:markdown, test:markdown-format (2 files, 30 tests), and fmt:check all pass from the repository root. No requirement or ADR text changed in this task.
<!-- SECTION:FINAL_SUMMARY:END -->
