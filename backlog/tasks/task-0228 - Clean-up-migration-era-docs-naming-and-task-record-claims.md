---
id: TASK-0228
title: 'Clean up migration-era docs, naming, and task-record claims'
status: To Do
assignee: []
created_date: '2026-09-13 19:36'
labels:
  - cleanup
  - docs
dependencies: []
modified_files:
  - docs/ci.md
  - server/eventsource/
  - server/routes/accounts_test.go
priority: low
type: chore
ordinal: 230000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A post-migration quality audit found small stale residues and two incorrect completed-task claims. Outcome: live documentation and task history no longer assert things the repository contradicts.

Scope:
- docs/ci.md still says the Firefox desktop job enables PostgreSQL anonymous event creation, but that flag was removed from every workflow and its removal was completed by TASK-0190.07.
- The folder_events_event_unique_idx predicate WHERE event_id IS NOT NULL is vestigial because event_id is NOT NULL (server/migrations/20260912000000_baseline_schema.sql).
- The server/eventsource package now only validates Crockford short IDs, and the accounts_test.go accountObjectID helper parses UUIDs; both names carry retired-store/dispatch connotations.
- TASK-0190.06's Final Summary is malformed (embedded tool-call markup at the end of the file).
- TASK-0199.09 AC4 claims backlog/handoffs/ retains the 2026-09-11 entries, but no 2026-09-08-or-later handoff except 2026-09-04 ever existed in the repository or git history.

Out of scope: backlog/backlog.md checkboxes, and TASK-0216, which remains the owner of the anonymous-event-compatibility route-test rename.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/ci.md no longer claims that a Firefox E2E job enables a PostgreSQL anonymous-event-creation flag that no workflow or stack sets.
- [ ] #2 The vestigial WHERE event_id IS NOT NULL predicate on the folder_events unique index is removed with a migration, or retained with a recorded reason.
- [ ] #3 eventsource and accountObjectID naming and comments no longer imply a retired store or dispatch mechanism, or a recorded reason explains the names; coordinate with TASK-0087.
- [ ] #4 TASK-0190.06's malformed Final Summary is repaired or annotated through Backlog, and TASK-0199.09 receives a comment correcting the claim that the 2026-09-11 handoffs are retained when they never entered the repository.
- [ ] #5 backlog/backlog.md is not modified, and changed Markdown passes the format checks.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
