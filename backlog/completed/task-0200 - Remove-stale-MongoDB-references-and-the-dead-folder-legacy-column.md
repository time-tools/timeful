---
id: TASK-0200
title: Remove stale MongoDB references and the dead folder legacy column
status: Done
assignee:
  - opencode
created_date: '2026-09-11 15:21'
updated_date: '2026-09-11 15:43'
labels: []
dependencies: []
references:
  - docs/postgres-operations.md
  - docs/design/architecture/adr/ADR-017.md
  - >-
    backlog/completed/task-0199 -
    Retire-MongoDB-after-moving-retained-integration-and-supporting-data-to-PostgreSQL.md
priority: low
type: chore
ordinal: 238000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After the TASK-0199 MongoDB retirement, server code still contained comments that implied a live MongoDB store, and `folder_events.legacy_event_id` was dead because folder writes only ever store a PostgreSQL `event_id`.

Outcome: the changed server comments describe the PostgreSQL-only model directly with no migration-era or prior-store references for the touched call sites, and folder memberships carry exactly one non-null PostgreSQL event reference with no legacy column.

Constraints:
- Do not edit applied goose migration files; add a new migration for the schema change.
- Leave historical records untouched: ADRs, `docs/postgres-operations.md`, deprecated requirements, candidate records, and completed backlog tasks.
- Preserve the client-visible API and wire shape; folder responses already omit legacy members.
- Comment and provenance wording changes must not alter behavior.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 No file changed by this task narrates a prior MongoDB store or migration-era state; the changed comments describe the current PostgreSQL-only model directly. The only retained "legacy" names are required identifiers, compatibility fields, or domain terms.
- [x] #2 folder_events.legacy_event_id no longer exists in the schema or in Go repository types and queries, and folder_events.event_id is NOT NULL.
- [x] #3 A new goose migration drops `legacy_event_id`, its check constraint, and its unique index, and refuses to run when any row still holds a value in that column.
- [x] #4 Folder read responses and the set-folder endpoint behave unchanged for PostgreSQL events.
- [x] #5 go build ./..., go vet ./..., and the isolated PostgreSQL-backed server test suite pass.
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
## Plan (researched 2026-09-11)

1. Wording scrub (no behavior change) in server files that still described a live MongoDB store:
   - `server/accounts/calendar.go`, `server/middleware/auth.go`, `server/models/datetime.go`, `server/models/event.go`, `server/models/id.go`, `server/postgres/{accounts,analytics,attendees,dashboard,otp,signup,types}.go`, `server/routes/{auth,user,postgres_event_routes}.go`, `server/services/auth/auth.go`.
   - Test comments: `server/postgres/otp_test.go`, `server/routes/{accounts_contract_expansion,anonymous_event_compatibility,postgres_group_response}_test.go`.
   - Fix the false `dashboard.go:29` "PostgreSQL and MongoDB own disjoint records" comment.
   - Do not touch applied migrations, ADRs, runbook, deprecated requirements, or backlog history.

2. New goose migration `server/migrations/20260911120000_drop_folder_legacy_event_id.sql`:
   - Up: raise an exception if any `folder_events.legacy_event_id IS NOT NULL`, then drop the column (its check constraint and unique index drop automatically) and set `event_id NOT NULL`.
   - Down: restore `event_id` nullability, re-add `legacy_event_id`, the XOR check constraint, and the unique index.

3. `server/postgres/folders.go`: remove `FolderMember.LegacyEventID`; simplify `AssignEventToFolder` to require `EventID`; change `DeleteFolder` to return only `error` and drop legacy-ID collection; simplify `scanFolders`.

4. `server/routes/folders.go`: update the `DeleteFolder` call to the new single return value. `canonicalFolderEventIDs` already ignores members without a PostgreSQL short ID; keep its nil guard for LEFT JOIN misses.

5. Tests:
   - `server/routes/account_deletion_test.go`: drop the legacy folder member fixture (`legacyEventID`, second insert, and the cross-store comment).
   - `server/routes/postgres_folders_test.go`: simplify the storage-reference assertion to `event_id IS NOT NULL`.
   - `server/postgres/accounts_test.go`: apply the new migration in the temp-table harness for schema fidelity.

6. Verify: `rg -i mongo server --glob '!migrations/**'` clean, `go build ./...`, `go vet ./...`, isolated `server-route-test` suite.

7. Follow-up polish: rewrite the remaining migration-era comments so they state the current PostgreSQL-only model without prior-store references, covering `server/postgres/{analytics,attendees,dashboard,folders,otp,signup,types}.go`, `server/routes/{auth,user,anonymous_event_compatibility_test}.go`, `server/services/auth/auth.go`, and the new migration's header and guard message. Only required identifiers keep "legacy" in their names.

Risks: dropping the column fails if any environment still holds legacy rows; the migration guard makes that explicit. No API/wire change because folder responses already skipped legacy members.

Rollback: revert code and restore the column with the migration Down section.

Execution note: the guard DO block is wrapped in -- +goose StatementBegin/End because goose cannot parse a dollar-quoted block with semicolons; the compose migrate image was rebuilt so /migrations included the new file. No other plan deviation.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: migration 20260911120000 drops folder_events.legacy_event_id after a RAISE EXCEPTION guard and sets event_id NOT NULL; folders repository/type and route caller simplified; account deletion and folder route tests updated; accounts/calendar/signup temp-table harnesses now apply the new migration.

Goose detail: goose splits statements on semicolons, so the guard DO block had to be wrapped in -- +goose StatementBegin/End; without it goose fails with 'unterminated dollar-quoted string'. The compose migrate image also had to be rebuilt so the new migration file was copied into /migrations.

Evidence: full isolated server-route-test suite passed (all packages ok) after the column drop; psql confirmed the schema change; a temp-table check confirmed the guard refuses when a legacy row exists (ERROR + exit 3). No browser e2e was required because no client-visible behavior changed.

Historical MongoDB wording in applied migrations, ADRs, docs/postgres-operations.md, deprecated requirements, and backlog history is intentionally unchanged.

Follow-up polish: after review, rewrote the remaining migration-era wording so comments state the current PostgreSQL-only model directly; only required identifiers (`legacy_event_id`, the migration filename, `legacyKey`, and legacy timed/guest-ownership field names) keep "legacy". The migration guard message now reads 'folder_events still holds values in legacy_event_id; refusing to drop the column'. Verification for the polish: `go build ./...`, `go vet ./...`, and the full isolated `server-route-test` suite all pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Removed the server's stale MongoDB references, dropped the dead folder `legacy_event_id` column, and reworded the remaining migration-era comments so they describe the PostgreSQL-only model directly.

### Changes
- Added `server/migrations/20260911120000_drop_folder_legacy_event_id.sql`: refuses to run while any `folder_events.legacy_event_id IS NOT NULL`, then drops the column (the XOR check constraint and legacy unique index drop with it) and sets `event_id NOT NULL`. The Down section restores the previous shape.
- `server/postgres/folders.go`: dropped `FolderMember.LegacyEventID`, simplified `AssignEventToFolder` to require a PostgreSQL `EventID`, changed `DeleteFolder` to return only an error, and simplified `scanFolders`.
- `server/routes/folders.go`: updated the `DeleteFolder` caller.
- Tests: removed the legacy folder fixture from `account_deletion_test.go`, simplified the storage assertion in `postgres_folders_test.go`, and applied the new migration in the `accounts`, `calendar`, and `signup` temp-table harnesses.
- Reworded stale MongoDB comments across server runtime and test files; the false `dashboard.go` "disjoint stores" comment is corrected. A follow-up polish then removed the remaining prior-store phrasing (`analytics.go`, `attendees.go`, `dashboard.go`, `folders.go`, `otp.go`, `signup.go`, `types.go`, `routes/auth.go`, `routes/user.go`, `services/auth/auth.go`, and the migration header/guard message) so comments state the current PostgreSQL-only model. Applied migrations, ADRs, the operations runbook, deprecated requirements, and backlog history were left untouched.

### Verification
- `go build ./...` and `go vet ./...`: pass.
- `rg -i 'mongo|retained|mongo-era' server -g '!**/migrations/**'`: no hits; applied migration comments keep their historical references intentionally. Remaining `legacy` hits in changed files are required identifiers, compatibility fields, or domain names.
- Isolated `server-route-test` suite after the migration applied: every package `ok`.
- `psql` schema check: `folder_events` has no `legacy_event_id`, `event_id` is `NOT NULL`, legacy index gone.
- Guard behavior: a row carrying a value in `legacy_event_id` makes the migration raise `folder_events still holds values in legacy_event_id; refusing to drop the column`.

### Scope/Risk
- No client-visible API or wire change: folder responses already skipped legacy members, and folder writes only ever stored `event_id`. No browser e2e was required for this backend-only change.
- The migration fails loudly if a deployed database still holds legacy folder references, which is the intended safety gate.
<!-- SECTION:FINAL_SUMMARY:END -->
