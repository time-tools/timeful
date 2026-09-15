---
id: TASK-0228
title: 'Clean up migration-era docs, naming, and task-record claims'
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-13 19:36'
updated_date: '2026-09-15 12:26'
labels:
  - cleanup
  - docs
dependencies: []
modified_files:
  - docs/ci.md
  - >-
    server/migrations/20260915000001_folder_events_event_unique_idx_drop_predicate.sql
  - server/postgres/folder_events_index_test.go
  - server/eventid/
  - server/routes/user.go
  - server/routes/accounts_test.go
  - server/routes/account_deletion_test.go
  - server/routes/anonymous_event_contract_test.go
  - server/routes/group_test.go
  - server/routes/signed_in_event_test.go
  - >-
    backlog/completed/task-0190.06 -
    Migrate-existing-events-and-related-organization-records-to-PostgreSQL.md
  - >-
    backlog/completed/task-0199.09 -
    Remove-MongoDB-runtime-driver-configuration-and-collections.md
  - >-
    backlog/tasks/task-0087 -
    Align-code-naming-with-canonical-glossary-terminology.md
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
- [x] #1 docs/ci.md no longer claims that a Firefox E2E job enables a PostgreSQL anonymous-event-creation flag that no workflow or stack sets.
- [x] #2 The vestigial WHERE event_id IS NOT NULL predicate on the folder_events unique index is removed with a migration, or retained with a recorded reason.
- [x] #3 eventsource and accountObjectID naming and comments no longer imply a retired store or dispatch mechanism, or a recorded reason explains the names; coordinate with TASK-0087.
- [x] #4 TASK-0190.06's malformed Final Summary is repaired or annotated through Backlog, and TASK-0199.09 receives a comment correcting the claim that the 2026-09-11 handoffs are retained when they never entered the repository.
- [x] #5 backlog/backlog.md is not modified, and changed Markdown passes the format checks.
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
1. docs/ci.md: remove the stale sentence claiming only the Firefox desktop E2E job enables PostgreSQL anonymous event creation; no workflow or stack sets that flag after TASK-0190.07.
2. Add goose migration 20260915000001_folder_events_event_unique_idx_drop_predicate.sql that recreates folder_events_event_unique_idx without `WHERE event_id IS NOT NULL`, with a Down restoring the partial form, plus a server/postgres regression test proving the non-partial definition and the down behavior.
3. Rename server/eventsource to server/eventid (keeping Canonical) and reword its package and function comments to drop retired event-store/dispatch language; update the four route imports. Rename the accounts_test.go accountObjectID helper to accountUUID and remove Mongo ObjectID vocabulary from account_deletion_test.go.
4. Repair TASK-0190.06's malformed Final Summary (embedded tool-call markup) through Backlog, add a corrective comment to TASK-0199.09 for the false "three 2026-09-11 handoffs retained" claim, and record a coordination comment on TASK-0087.
5. Verify: root format:markdown, format:markdown:check, lint:markdown; isolated Compose backend suite; graphify update.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-14 19:01
---
Coordination: TASK-0233.03 updates the PostgreSQL access-transfer qualifier wording in docs/ci.md and deliberately leaves the stale anonymous-event-creation flag sentence to this task (AC1). Coordinate if both run concurrently.
---

author: opencode
created: 2026-09-15 09:46
---
Coordination from TASK-0233.04 (final cleanup verification, 2026-09-15): docs/ci.md was left untouched by the postgres-qualifier cleanup series; TASK-0233.03 only updated the access-transfer qualifier wording. The stale anonymous-event-creation flag sentence remains yours (AC #1). TASK-0216 is Done, satisfied by TASK-0233.02.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Cleaned up the migration-era residues from the post-migration audit and corrected the two false completed-task claims.

What changed:
- docs/ci.md: removed the sentence claiming only the Firefox desktop E2E job enables PostgreSQL anonymous event creation; no workflow or stack sets that flag since TASK-0190.07, and E2E CI records that all suites create events in PostgreSQL by default.
- server/migrations/20260915000001_folder_events_event_unique_idx_drop_predicate.sql recreates folder_events_event_unique_idx without the vestigial `WHERE event_id IS NOT NULL` predicate (event_id is NOT NULL) and restores the partial form in its Down section.
- server/postgres/folder_events_index_test.go is a new regression test proving the index is not partial and that the down migration restores the predicate.
- server/eventsource/ was renamed to server/eventid/ with package and function comments that no longer imply an event-source dispatch boundary; the four route imports were updated. `accountObjectID` became `accountUUID` in server/routes/accounts_test.go and account_deletion_test.go, removing Mongo ObjectID vocabulary.
- Backlog records: TASK-0190.06's Final Summary was replaced through the Backlog CLI (temporarily restored to the active folder, then re-completed) and its embedded `</finalSummary>`/`<parameter name="status">` markup is gone; TASK-0199.09 received a correction comment about the non-existent 2026-09-11 handoffs; TASK-0087 received a coordination comment naming eventid.Canonical and accountUUID as the current names.

Evidence:
- Backend: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` -> every package ok, including timeful/server/eventid, timeful/server/postgres, and timeful/server/routes; a focused run of TestFolderEventsUniqueIndexDropsVestigialPredicate and TestMigrationReplayMatchesAppliedSchema -> PASS. The first suite run failed against a stale locally cached migrator image; rebuilding `postgres-test-migrate` applied the new migration and the suite passed.
- Go hygiene: `gofmt -l` clean on touched files; `go build ./...` and `go vet ./...` pass.
- Markdown: `npm run format:markdown` (no writes needed), `npm run format:markdown:check`, and `npm run lint:markdown` all pass.
- Graph: `graphify update .` rebuilt 5721 nodes / 10799 edges.
- backlog/backlog.md was not touched: its last modification (2026-09-15 10:01 UTC) predates this session and its pre-existing Inbox diff is unchanged.
- E2E was not required and not run: no frontend or observable route behavior changed, and the backend route suite already exercises event creation against the migrated schema.

Acceptance criteria: docs/ci.md no longer claims the removed flag (AC1); the unique-index predicate was removed by migration and covered by a regression test (AC2); eventsource and accountObjectID naming and comments no longer imply a retired store or dispatch mechanism, coordinated with TASK-0087 (AC3); TASK-0190.06's malformed Final Summary was repaired and TASK-0199.09 received the corrective comment through Backlog (AC4); backlog/backlog.md is unmodified and changed Markdown passes the format checks (AC5).
<!-- SECTION:FINAL_SUMMARY:END -->
