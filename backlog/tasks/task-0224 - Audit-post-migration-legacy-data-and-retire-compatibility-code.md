---
id: TASK-0224
title: Audit post-migration legacy data and retire compatibility code
status: Done
assignee: []
created_date: '2026-09-13 19:35'
updated_date: '2026-09-14 17:52'
labels:
  - postgres
  - migration
  - cleanup
dependencies: []
priority: high
type: task
ordinal: 223000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
MongoDB was retired on 2026-09-11 and PostgreSQL is the only supported store. Compatibility code for migrated or legacy data still exists across the schema, repositories, and routes. The development database has been recreated from the baseline, and the remaining deployed databases are handled by recreation rather than in-place cleanup. The FR-119 event-name CHECK is still NOT VALID, and no database has been validated after cleanup.

Outcome: evidence-based retirement of the remaining legacy compatibility surface. Phases run in order and are tracked as subtasks: close the legacy-row audit (done by the recreate decision), recreate staging and production from the baseline with verified backups, validate the name constraint, then remove compatibility code and columns that the recreate decision and writer-path analysis prove unnecessary.

Context and constraints:
- Databases that predate the baseline are recreated from it rather than upgraded (docs/postgres-operations.md); TASK-0224.04 identifies each deployed database's state and performs the recreation.
- The pre-baseline development evidence is retained at ~/timeful-backups/prebaseline-dev-20260914/; no repeatable audit script is added because no live in-scope database holds legacy rows.
- Keep PostgreSQL access in server/postgres/; routes and services must not access PostgreSQL directly.
- Schema changes use new goose migrations; the baseline 20260912000000_baseline_schema.sql is immutable.
- Preserve JSONB payload round-trip, blind-availability privacy, EVCC authorization precedence, and credential encryption in Go.
- Verify with the isolated test overlay in compose.test.yaml, and with Firefox desktop e2e when event, response, signup, or calendar behavior changes.
- Public event URLs may differ from legacy identifiers; no redirect layer is in scope.

Compatibility surface: guest_id/canonical_guest_name/guest_edit_policy/guest_ownership_mode/guest_edit_token columns, nullable respondent_kind coercion, mixed-case calendar_key handling (ActualCalendarAccountMapKey/legacyKey), dual epoch-millis/RFC3339 signup-instant and group manualAvailability decodings (retained, not auditable from stored rows), and the NOT VALID postgres_events_name_length constraint.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All phase subtasks are complete and verified.
- [x] #2 Legacy compatibility behavior is either evidenced as unnecessary by the recreate decision and writer-path analysis and removed, or retained with a recorded reason grounded in those results.
- [x] #3 No event, response, signup, calendar, or account wire-shape changes are introduced.
- [x] #4 Any new schema change is a forward-only goose migration in server/migrations/.
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
Completion note (2026-09-14): all four phase subtasks are Done and verified. Phase 3's finalization re-ran the isolated backend suite after the review change (`timeful/server/postgres` 5.638s, `timeful/server/routes` 2.351s, all packages ok; log `/tmp/opencode/task-0224.03-final-backend.log`). Cross-phase checks: `git diff 40de9bb6` shows no frontend or generated API artifact changes, and the phase added only forward-only goose migrations 20260913000002, 20260914000000, and 20260915000000 under server/migrations/. Acceptance criteria #1-#4 and Definition of Done #1-#4 verified and checked.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Completed all four phases of the post-migration legacy-data audit and compatibility retirement.

Phase outcomes:
- TASK-0224.01 (audit): closed by the recreate decision; no live in-scope database holds legacy rows, with pre-baseline development evidence retained at ~/timeful-backups/prebaseline-dev-20260914/.
- TASK-0224.02 (constraint): added the forward-only validation migration 20260914000000, flipped name_constraint_test.go to the validated end state, and documented the constraint lifecycle; development, isolated test, staging, and production all end with postgres_events_name_length convalidated = true.
- TASK-0224.03 (retirement): nullable respondent_kind coercion and mixed-case calendar-key handling removed; dead guest lookup and utility code removed; the five generic-response guest columns are retained physically for the prior-release rollback window and are neither read nor written; docs updated.
- TASK-0224.04 (recreate): staging and production were detected pre-baseline, backed up with verified scratch restores, recreated from the baseline through the migration chain, and health-verified; old backups retained.

Properties preserved: JSONB payload round-trip, blind-availability privacy, EVCC authorization precedence, guest-name normalization, credential encryption, and the event, response, signup, calendar, and account wire shapes.

Verification: isolated backend suite passes on the final worktree (postgres 5.638s, routes 2.351s, all packages ok); focused migration tests 3/3; Firefox desktop e2e passes (create spec 6/6 for the constraint work, 18/18 for the retirement work); go build/go vet, format:markdown:check, and lint:markdown pass; no frontend or generated API artifact changed.

Follow-ups (not created): drop the five retained guest columns one release after the rollback window closes, and state the window end condition when the record is next touched.
<!-- SECTION:FINAL_SUMMARY:END -->
