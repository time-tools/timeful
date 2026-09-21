---
id: TASK-0304
title: Extend FR-013 to bound Display Timezone changes to projection
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 15:32'
updated_date: '2026-09-21 15:57'
labels: []
dependencies: []
references:
  - docs/requirements/functional/fr/FR-013.md
  - docs/requirements/README.md
  - >-
    backlog/tasks/task-0303 -
    Preserve-the-pending-Timed-Event-Occurrence-Span-across-Display-Timezone-changes.md
  - >-
    backlog/tasks/task-0024 -
    F-SCHEDULE-OVERLAP-006-active-slots-drop-out-of-timezone-shifted-timed-grid.md
  - e2e/specs/timed-event-viewer-tz-column-duplication-firefox.spec.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapTestUtils.ts
documentation:
  - docs/requirements/AGENTS.md
  - docs/requirements/functional/README.md
  - docs/terminology/README.md
  - docs/terminology/glossary.md
modified_files:
  - docs/requirements/functional/fr/FR-013.md
  - docs/requirements/README.md
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.displayTimezoneEventData.test.ts
  - e2e/specs/timed-event-viewer-tz-column-duplication-firefox.spec.ts
priority: medium
type: task
ordinal: 304000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The pending-span fix in task-0303 was reviewed against the constraint that Display Timezone changes must not change Event Picked Dates, the Enabled Domain, or Active Slots, cited in that task as (FR-013). FR-013 actually states only that Time Slot Instants are preserved and the Event Timezone is unchanged, so the picked-dates/domain/active-slots invariant has no requirement home and a future change could regress it unchecked. The approved decision is to extend FR-013 rather than create a new FR, because the trigger and the underlying property are the same (a Display Timezone change is projection-only), and to back the amendment with unit and Firefox e2e regression coverage. FR-013 keeps status accepted; FR-136 and task-0303 stay unchanged, and the amendment makes task-0303's citation accurate.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 FR-013 states that changing the Display Timezone preserves Time Slot Instants and does not change the Event Timezone, Event Picked Dates, the Enabled Domain, or Active Slots
- [x] #2 FR-013's requirement index row reflects the amended title with canonical glossary links and stays on one physical source line
- [x] #3 FR-013's acceptance criteria each occupy one physical source line and the root Markdown format and lint checks pass
- [x] #4 A unit regression test shows that changing the Display Timezone on a Timed Event page leaves the event's Event Picked Dates and Active Slots unchanged and sends no event write
- [x] #5 A Firefox e2e regression test shows that switching the Display Timezone sends no event write and leaves the event's picked dates, Active Slots, and Event Timezone unchanged
- [x] #6 All required frontend checks pass (lint, fmt:check, typecheck, build, test:unit)
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Decision context (recorded 2026-09-21, user-approved): extend FR-013 rather than create a new FR, because the trigger and the underlying property are the same (a Display Timezone change is projection-only); keep status accepted; cover with both a unit test and a Firefox e2e test. FR-136 and task-0303 are unchanged, and this amendment makes task-0303's "(FR-013)" citation accurate.

1. Authoring rules first: read docs/requirements/AGENTS.md, docs/requirements/functional/README.md, docs/requirements/README.md, and docs/terminology/README.md. Link the first occurrence of each controlled term per paragraph, list item, and table cell; bold repeats; one physical source line per sentence and per table row.

2. docs/requirements/functional/fr/FR-013.md
   - Front matter title -> "Keep the Display Timezone from changing event data"; H1 title-cased to match.
   - Statement -> "Changing the [Display Timezone](...#display-timezone) shall preserve every [Time Slot's](...#time-slot) [Instant](...#instant) and shall not change the [Event Timezone](...#event-timezone), the [Event Picked Dates](...#event-picked-dates), the [Enabled Domain](...#enabled-domain), or [Active Slots](...#active-slots)." on one physical line.
   - Add acceptance criterion -> "Changing the [Display Timezone](...) leaves the [Event Picked Dates](...), the [Enabled Domain](...), and [Active Slots](...) unchanged."
   - Join the existing acceptance criterion currently split across two physical lines ("A Time Slot renders no more / than once ...") onto one line.
   - Keep id, type, components: frontend, status: accepted.

3. docs/requirements/README.md: replace the FR-013 index row title cell with the amended title using first-occurrence glossary links (Display Timezone, Event Picked Dates, Enabled Domain, Active Slots); keep the row on one physical line; run the Markdown formatter for table padding.

4. Unit regression coverage: new frontend/src/components/schedule_overlap/ScheduleOverlap.displayTimezoneEventData.test.ts using the harness in scheduleOverlapTestUtils.ts (mountScheduleOverlap, installScheduleOverlapTestGlobals, getTimedGridPresentation, zdt).
   - Mount a canonical Timed Event with Active Slots and a display timezone offset.
   - Capture the event prop's Event Picked Dates and Active Slots instants and the grid's projected slot-instant set.
   - Switch vm.curTimezone to a different offset; await nextTick.
   - Assert the event prop's Event Picked Dates and Active Slots are unchanged and no PUT request was issued; assert the projected slot-instant set is unchanged.
   - Fixture caveat: TASK-0024 (open) can drop early-edge slots from the grid for positive-offset events, so keep the stored-data assertions primary and do not assert a complete rendered cell set.

5. Firefox e2e coverage: extend the first test in e2e/specs/timed-event-viewer-tz-column-duplication-firefox.spec.ts ("event page renders unique projected civil date labels at UTC+6 and UTC+7"), which already uses changeDisplayTimezone and seeded Active Slots.
   - Register a PUT counter for /api/events/ before the switches.
   - fetchEventByShortId before and after the UTC+6 and UTC+7 switches.
   - Assert timedRecurrence.selectedDays, sortIsoInstants(activeSlots), and eventTimezone are unchanged and the PUT count is zero.
   - Update the spec's FR-013 header comment to cover the amended acceptance criterion.

6. Checks:
   - Root: npm run format:markdown; npm run format:markdown:check; npm run lint:markdown.
   - Frontend: npm run lint; npm run fmt:check; npm run typecheck; npm run build; npm run test:unit.
   - E2E from e2e/: npm run test:e2e -- --project=firefox-desktop (stream full output; run the edited spec first).
   - graphify update .
   - No Swagger annotations change; root scripts/ and prettier/ are untouched.

7. Record evidence on the task (command results, the amended requirement and index row), mark Done, and note that task-0303's citation is now accurate.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-21: Starting implementation per the recorded plan. Re-read docs/requirements/AGENTS.md, docs/requirements/functional/README.md, docs/requirements/README.md, and docs/terminology/README.md before editing FR-013. FR-013 currently bounds Display Timezone changes to Time Slot Instants and the Event Timezone only; the amendment adds Event Picked Dates, the Enabled Domain, and Active Slots, and joins the split acceptance-criterion line.

2026-09-21 implementation:
- FR-013 amended: title is now 'Keep the Display Timezone from changing event data'; the statement preserves Time Slot Instants and forbids changes to the Event Timezone, Event Picked Dates, the Enabled Domain, and Active Slots; a new acceptance criterion covers the picked-dates/domain/active-slots invariant; the previously split 'renders no more than once' criterion is now one physical line. status stays accepted.
- Requirements README FR-013 row updated to the amended title with the Display Timezone glossary link on one physical line.
- New unit regression ScheduleOverlap.displayTimezoneEventData.test.ts: mounts a canonical UTC workday event at Etc/GMT-8, switches the display timezone to Etc/GMT-7, and asserts the event prop's Event Picked Dates and Active Slots instants are unchanged and no PUT to /api/events/ was issued.
- Firefox e2e: the first test in timed-event-viewer-tz-column-duplication-firefox.spec.ts now records /api/events/ PUTs, fetches the event before and after the UTC+6 and UTC+7 switches, and asserts eventTimezone, timedRecurrence.selectedDays, and sorted activeSlots are unchanged with zero PUTs; the FR-013 header comment describes the amended criterion.
Evidence: focused unit test passed; full unit 157 files / 1253 tests passed; e2e typecheck and lint clean; focused Firefox e2e 3 passed; full firefox-desktop e2e 71 passed with 1 pre-existing skip; frontend lint 0 errors (2 pre-existing warnings); fmt:check, typecheck, build passed; root format:markdown, format:markdown:check, lint:markdown clean; graphify update completed.
DoD notes: #5 swagger and #7 scripts/prettier are not applicable because no Swagger annotations or root JS changed; #8 is covered by the amended FR-013 and its index row.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Amended FR-013 to bound Display Timezone changes to projection: the requirement now preserves Time Slot Instants and forbids changes to the Event Timezone, Event Picked Dates, the Enabled Domain, and Active Slots, with a new acceptance criterion and a single-line restatement of the no-duplicate-render criterion; the requirements index row reflects the amended title. Backed the invariant with a component unit test that switches the Display Timezone and asserts event dates, active slots, and the absence of event writes, and extended the Firefox column-duplication spec to compare fetched event data and the PUT count across UTC+6/UTC+7 switches. All required frontend checks and the full Firefox desktop e2e suite pass.
<!-- SECTION:FINAL_SUMMARY:END -->
