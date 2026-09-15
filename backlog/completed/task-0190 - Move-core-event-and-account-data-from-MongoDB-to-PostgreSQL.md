---
id: TASK-0190
title: Move core event and account data from MongoDB to PostgreSQL
status: Done
assignee: []
created_date: '2026-09-09 21:45'
updated_date: '2026-09-11 14:54'
labels:
  - postgresql
  - migration
dependencies: []
references:
  - TASK-0189
  - server/docs/postgres-anonymous-event-compatibility.md
  - server/db/init.go
  - backlog/backlog.md
documentation:
  - BACKLOG_WORKFLOW.md
  - docs/environments.md
priority: high
type: feature
ordinal: 197000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make PostgreSQL authoritative for existing and new accounts, identities, events, responses, and event organization through separately testable stages.
Include signed-in polls, signup forms, availability groups, folders, and historical data migration.
Temporarily retain calendar connections, provider tokens and calendar preferences, OTP challenges, friend requests, and historical daily user logs in MongoDB.
Retained integration documents must not remain a second account authority.
Preserve account identity, ownership, and access rights through the legacy account mapping; public event links may change at cutover, and migrated relationships are rewritten during migration instead of using a permanent legacy-ID map.
Use one authoritative store per migrated record and avoid permanent dual writes.
Implementation and isolated rehearsal are in scope; live deployments and production data migration are separately scheduled operational actions.
Preserve historical handoffs and completed task records.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Linked subtasks deliver independently verified account, event-type, organization, migration, default-routing, and cutover-readiness stages.
- [x] #2 Existing and new core records can be served authoritatively from PostgreSQL with identity and relationship references preserved; public event URLs may change at cutover.
- [x] #3 Retained MongoDB integrations continue to work through explicit identity mappings without acting as a second source of account truth.
- [x] #4 Both anonymous-creation flags are removed after migration rehearsal and browser fixture readiness.
- [x] #5 A tested migration and backup/restore runbook records reconciliation evidence, handling of ambiguous records, and rollback boundaries.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-10 09:46
---
TASK-0190.01 narrowed AC#2: per the user's decision, pre-existing public event URLs need not keep resolving after cutover. Identity and relationship references (accounts, ownership, folder membership, retained MongoDB references) remain authoritative, and migrated relationships are rewritten during migration rather than preserved as public URLs.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
PostgreSQL is authoritative for core accounts, identities, events, responses, and event organization, with the last known event-payload correctness gap closed as TASK-0198.

## Acceptance criteria evidence
- #1: All eight linked subtasks (TASK-0190.01 through TASK-0190.08) are Done with checked acceptance criteria and recorded unit and e2e evidence spanning account, event-type, organization, migration, default-routing, and cutover-readiness stages.
- #2: Accounts, events of every kind, responses, signup forms, availability groups, and folders are served only from PostgreSQL (`server/postgres`, `server/routes/postgres_*`), identity and relationships resolve through `platform_identities.external_user_id`, and TASK-0198 removed the last data-loss gap by making the canonical event payload round-trip `duration` and the legacy schedule columns while the timed-event API keeps its documented suppression. This session's full isolated backend suite and Firefox desktop e2e pass on that code.
- #3: Retained integrations resolved through the explicit identity mapping and never became a second account authority; TASK-0199 then moved them to PostgreSQL and retired MongoDB entirely, a stronger outcome than the temporary retention this task scoped.
- #4: Both anonymous-creation flags were removed in TASK-0190.07 after migration rehearsal and browser fixture readiness, and no active references remain outside historical Backlog and graph artifacts.
- #5: TASK-0190.08 rehearsed backup and restore with reconciliation evidence (`TestBackupRestoreRehearsal`: 24 restored tables, 11 relations reconciled by row count and full-row md5), documented quarantine of ambiguous records through `migration_quarantine`, and recorded the cutover rollback boundaries; `docs/postgres-operations.md` retains the verified backup/restore and recovery boundaries after the executed migration docs were pruned.

## Verification
- Full isolated backend suite: every package ok.
- Frontend lint, fmt:check, typecheck, build, and unit (146 files, 1088 tests): pass.
- Firefox desktop e2e with `E2E_FRONTEND=bundled`: 56 passed, 1 skipped, 0 failed.
- `npm run format:markdown:check`: pass.

## Follow-ups
None in scope. Public event URLs may change at cutover as recorded in the narrowed AC#2, and the retained-MongoDB decomposition is preserved in the completed TASK-0199 records.
<!-- SECTION:FINAL_SUMMARY:END -->
