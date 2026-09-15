---
id: TASK-0213
title: Name MongoDB in ADR-018's historical context
status: Done
assignee: []
created_date: '2026-09-12 21:33'
updated_date: '2026-09-12 21:35'
labels: []
dependencies: []
modified_files:
  - docs/design/architecture/adr/ADR-018.md
priority: low
type: docs
ordinal: 213000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make the accepted ADR-018 historically specific by naming MongoDB as the retired second store instead of generic "a separate store", "both stores", and "cross-store" wording.

Context:
- ADR-018 was created in commit 7579bed2 (TASK-0205/TASK-0206) under the convention that active ADRs carry no MongoDB wording; ADR-021 later set precedent for naming "MongoDB-era" values in an accepted record.
- This task intentionally departs from the TASK-0205/0206 convention for ADR-018 only. Completed task records are not edited.
- Documentation-only change per BACKLOG_WORKFLOW.md; no requirement, index, or title changes; FR-121 already states the PostgreSQL-only behavior.

Scope:
- docs/design/architecture/adr/ADR-018.md only.
- Name MongoDB in Context, Considered Options, Decision Outcome, and Consequences.
- Bump updated_date to the revision date.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ADR-018 names MongoDB as the retired second store in its Context, Considered Options, Decision Outcome, and Consequences instead of generic second-store wording
- [x] #2 ADR-018 keeps its title, H1, status, addresses, and supersession links, and bumps updated_date to 2026-09-13
- [x] #3 No other ADR, requirement, index row, or source file changes
- [x] #4 npm run format:markdown, format:markdown:check, lint:markdown, test:markdown-format, and fmt:check pass from the repo root, and graphify update . refreshes the graph
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
## Implementation Plan

1. Mark TASK-0213 In Progress.
2. In `docs/design/architecture/adr/ADR-018.md`, bump `updated_date` to 2026-09-13.
3. Name MongoDB in Context (ADR-011 dual-store history, mutated in both PostgreSQL and MongoDB), Considered Options (keep MongoDB / dual-write both), Decision Outcome (joins between PostgreSQL and MongoDB, MongoDB holds no record kind, MongoDB is fully retired with the ADR-017 removal-order link, ADR-011 supersession wording), and Consequences (joins, backup coverage, dual-write path, purpose-built document store).
4. Leave the title, H1, status, addresses, data-ownership bullets, and index row unchanged.
5. Verify with root Markdown checks and `graphify update .`; record evidence and complete the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-13)

Documentation-only change; no runtime code, tests, configuration, or requirements touched.

- Bumped `updated_date` to 2026-09-13.
- Context: added the ADR-011 dual-store history sentence naming MongoDB's retained record kinds, and changed "authored in two stores" to "authored in both PostgreSQL and MongoDB".
- Considered Options: named MongoDB in the keep-second-store and dual-write options.
- Decision Outcome: named the PostgreSQL/MongoDB cross-store joins, replaced "No second store..." with "MongoDB holds no record kind...", added the full-retirement sentence linking the deprecated ADR-017 removal order, and reworded the ADR-011 supersession sentence.
- Consequences: named PostgreSQL and MongoDB in the join and dual-write lines and the purpose-built document-store tradeoff; kept "cover one store instead of two" because the same list names the pair.
- Title, H1, status, addresses, data-ownership bullets, H2 sections, and the docs/design/README.md index row are unchanged.

## Verification

- AC1: `rg -n -i "mongo|separate store|both stores|cross-store|second store" docs/design/architecture/adr/ADR-018.md` shows MongoDB in Context, both rejected options, Decision Outcome, supersession, and Consequences; no generic "a separate store"/"both stores"/"cross-store" phrasing remains.
- AC2: frontmatter keeps id ADR-018, its title, components, status accepted, and addresses FR-121, and updated_date is 2026-09-13; the H1 and the ADR-011/ADR-016 supersession links are intact.
- AC3: `git status --short` shows only ADR-018 plus the new TASK-0213 file from this task; backlog/backlog.md and task-0178 were already dirty with owner inbox edits before this task.
- AC4: `npm run format:markdown`, `format:markdown:check`, `lint:markdown`, `test:markdown-format` (30 tests), and `fmt:check` pass from the repo root; `graphify update .` ran and detected no code-graph topology changes (the ADR title node is unchanged).
- Unit and e2e tests are exempt: documentation-only change per BACKLOG_WORKFLOW.md and the task DoD.
- No commit was made.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Named MongoDB as the retired second store in ADR-018, replacing generic second-store wording with specific historic context.

Delivered:
- Context records the ADR-011 dual-store era: PostgreSQL authoritative for core records while MongoDB retained calendar connections, provider tokens, calendar preferences, OTP challenges, friend requests, and daily logs.
- Considered Options name keeping MongoDB as a permanent second store and dual-writing each mutation to PostgreSQL and MongoDB.
- Decision Outcome names the PostgreSQL/MongoDB cross-store joins, states MongoDB holds no record kind and no record is authored in both stores, records MongoDB as fully retired with the deprecated ADR-017 removal order as decision history, and reworks the ADR-011 supersession wording.
- Consequences name PostgreSQL and MongoDB for joins and dual writes and the purpose-built document-store tradeoff.
- updated_date is 2026-09-13; the title, H1, status, addresses, data-ownership bullets, and index row are unchanged.
- Root Markdown checks pass, graphify update . found no code-graph topology changes, and unit/e2e tests are exempt for this documentation-only change. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
