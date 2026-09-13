---
id: TASK-0224
title: Audit post-migration legacy data and retire compatibility code
status: To Do
assignee: []
created_date: '2026-09-13 19:35'
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
MongoDB was retired on 2026-09-11 and PostgreSQL is the only supported store. A 2026-09-13 audit found that compatibility code for migrated/legacy data still exists across the schema, repositories, and routes, while no one has ever audited the deployed databases for the legacy row shapes that keep that compatibility necessary. The FR-119 event-name CHECK is still NOT VALID, and no database was ever validated after cleanup.

Outcome: evidence-based retirement of the remaining legacy compatibility surface. Phases run in order and are tracked as subtasks: audit legacy rows, clean violations and validate the name constraint, then remove compatibility code and columns that the audit proves unnecessary.

Context and constraints:
- Databases that predate the baseline are recreated from it rather than upgraded (docs/postgres-operations.md); the audit identifies which deployed databases are in that state.
- Keep PostgreSQL access in server/postgres/; routes and services must not access PostgreSQL directly.
- Add one-off audit tooling under dated server/scripts/YYYYMMDD_description/ directories; run it manually.
- Schema changes use new goose migrations; the baseline 20260912000000_baseline_schema.sql is immutable.
- Preserve JSONB payload round-trip, blind-availability privacy, EVCC authorization precedence, and credential encryption in Go.
- Verify with the isolated test overlay in compose.test.yaml, and with Firefox desktop e2e when event, response, signup, or calendar behavior changes.
- Public event URLs may differ from legacy identifiers; no redirect layer is in scope.

Known compatibility surface the audit covers: guest_id/canonical_guest_name/guest_edit_policy/guest_ownership_mode/guest_edit_token columns, nullable respondent_kind coercion, mixed-case calendar_key handling (ActualCalendarAccountMapKey/legacyKey), dual epoch-millis/RFC3339 signup-instant decoding, and the NOT VALID postgres_events_name_length constraint.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All phase subtasks are complete and verified.
- [ ] #2 Legacy compatibility behavior is either evidenced as unnecessary by audit results and removed, or retained with a recorded reason grounded in those results.
- [ ] #3 No event, response, signup, calendar, or account wire-shape changes are introduced.
- [ ] #4 Any new schema change is a forward-only goose migration in server/migrations/.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
