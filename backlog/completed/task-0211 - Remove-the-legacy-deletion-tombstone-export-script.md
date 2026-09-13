---
id: TASK-0211
title: Remove the legacy deletion tombstone export script
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-12 21:15'
updated_date: '2026-09-12 21:17'
labels: []
dependencies: []
documentation:
  - server/docs/postgres-data-boundaries.md
  - docs/postgres-operations.md
  - server/migrations/20260912000000_baseline_schema.sql
priority: low
type: chore
ordinal: 212000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`server/scripts/20260912_export_legacy_deletion_tombstones/export.sh` is the operator export path for pre-cutover 24-hex account deletion tombstones that the single baseline schema (`server/migrations/20260912000000_baseline_schema.sql`) does not carry; `server/docs/postgres-data-boundaries.md` points operators to it before a pre-baseline database is recreated from the baseline. The export path is no longer required, so remove the obsolete operator procedure and its guidance so the repository no longer documents or ships it.

Constraint: before removing the export path, confirm that no pre-baseline database still needs the export, or explicitly record that retention of the legacy tombstone audit rows is waived. Until the replacement database is validated, the verified backup of the old database remains the fallback per `docs/postgres-operations.md`. Historical Backlog records and ADRs may keep their mentions; only active references must be gone.

Desired outcome: the export script and its directory are deleted, and the PostgreSQL data-boundaries documentation describes the legacy tombstone carry-forward outcome without pointing at the removed path.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The legacy pre-cutover tombstone retention decision is confirmed and recorded: no remaining pre-baseline database requires the export, or retention is explicitly waived
- [x] #2 server/scripts/20260912_export_legacy_deletion_tombstones/ is removed from the repository
- [x] #3 server/docs/postgres-data-boundaries.md no longer points operators to the removed export path, while still stating that legacy tombstones are not carried into a database recreated from the baseline
- [x] #4 No active source, test, or documentation reference to the removed script or its directory remains; historical Backlog records and ADRs may retain mentions
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
1. Record the retention decision (AC1): the user approved removing the export path, waiving legacy pre-cutover tombstone audit retention; the local development database holds no pre-baseline tombstone data, and the verified old-database backup remains the fallback per docs/postgres-operations.md.
2. Delete `server/scripts/20260912_export_legacy_deletion_tombstones/` (its only file is `export.sh`).
3. Rewrite the tombstone sentence in `server/docs/postgres-data-boundaries.md` so it still states that legacy tombstones are not carried into a database recreated from the baseline, without pointing operators at the removed export path.
4. Verify with grep that no active source, test, or documentation reference to the script or directory remains; historical Backlog records and ADRs may keep mentions.
5. Run `npm run format:markdown` for the changed Markdown (DoD #4) and `graphify update .` afterward; unit and e2e checks are exempt as this is a documentation and one-off operator script change.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Retention decision (AC1): the user approved removing the legacy tombstone export path, explicitly waiving pre-cutover tombstone audit retention. Local evidence: the running development database (`timeful-development-postgres-1`) has no `account_deletion_tombstones` table at all and only `postgres_events` plus `postgres_event_responses` (goose stuck at Aug-2026 versions), so no pre-baseline tombstone data exists locally to export. The verified backup of an old database remains the fallback per `docs/postgres-operations.md` until a replacement database is validated.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Removed the obsolete legacy deletion tombstone export path now that the single baseline schema no longer carries `account_deletion_tombstones.external_user_id` and pre-baseline databases are recreated from the baseline rather than upgraded.

### Changes

- `server/scripts/20260912_export_legacy_deletion_tombstones/export.sh` — deleted. It was the only manual export point for pre-cutover 24-hex tombstone rows, and `server/scripts/` no longer ships it.
- `server/docs/postgres-data-boundaries.md` — the tombstone sentence now ends at "legacy tombstones are not carried forward" instead of pointing operators at the removed export script.

### Decision

Legacy pre-cutover tombstone audit retention is explicitly waived, as approved by the user when they requested this removal. The local development database holds no pre-baseline tombstone data: it has no `account_deletion_tombstones` table at all and only `postgres_events` plus `postgres_event_responses` with goose at Aug-2026 versions. The verified backup of an old database remains the fallback per `docs/postgres-operations.md` until a replacement database is validated.

### Verification evidence

- AC1: retention decision recorded in the implementation notes (user approval plus the development-database inspection above).
- AC2: `git rm -r server/scripts/20260912_export_legacy_deletion_tombstones/`; `ls server/scripts/` now lists only `20240721_apple_calendar_test`.
- AC3: `git diff --cached server/docs/postgres-data-boundaries.md` shows the export clause removed while the statement that legacy tombstones are not carried into a recreated database remains.
- AC4: `grep -rn "20260912_export_legacy_deletion_tombstones" server/ docs/ frontend/ e2e/ scripts/ .github/` returns no matches; only historical Backlog records and this task mention it. The current regenerated `graphify-out/graph.json` no longer references it.

### Checks run

- `npm run format:markdown:check` (clean)
- `graphify update .` (graph rebuilt, 5609 nodes)
- Unit and e2e checks are exempt: this is a documentation plus one-off operator script change, and no tests were requested.

### Risks and follow-ups

No follow-ups. Any future need to audit pre-cutover deletions now relies on the retained old-database backup.
<!-- SECTION:FINAL_SUMMARY:END -->
