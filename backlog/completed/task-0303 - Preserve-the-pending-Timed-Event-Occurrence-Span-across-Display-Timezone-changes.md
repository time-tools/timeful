---
id: TASK-0303
title: >-
  Preserve the pending Timed Event Occurrence Span across Display Timezone
  changes
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 14:56'
updated_date: '2026-09-21 15:57'
labels: []
dependencies: []
references:
  - frontend/src/composables/schedule_overlap/useEventScheduling.ts
  - frontend/src/composables/schedule_overlap/useDragPaint.ts
  - frontend/src/components/schedule_overlap/useScheduleOverlapController.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
modified_files:
  - frontend/src/composables/schedule_overlap/types.ts
  - frontend/src/composables/schedule_overlap/useEventScheduling.ts
  - frontend/src/composables/schedule_overlap/useEventScheduling.test.ts
  - frontend/src/composables/schedule_overlap/useDragPaint.ts
  - frontend/src/composables/schedule_overlap/useDragPaint.test.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapTimeGrid.vue
  - frontend/src/components/schedule_overlap/useScheduleOverlapController.ts
  - >-
    frontend/src/components/schedule_overlap/useScheduleOverlapController.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.schedulingTooltip.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.scheduledSpanDisplayTimezone.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlapGridDragBinding.test.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapTestUtils.ts
  - >-
    frontend/src/components/schedule_overlap/scheduleOverlapViewModelContracts.ts
  - frontend/src/components/schedule_overlap/useScheduleOverlapViewModels.ts
  - frontend/src/components/schedule_overlap/useTimedGridPresentation.ts
  - frontend/src/utils/scheduleOverlap.regressions.test.ts
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
  - docs/requirements/functional/fr/FR-136.md
  - docs/requirements/README.md
priority: high
type: bug
ordinal: 303000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the Timed Event Scheduling Page, changing the Display Timezone re-projects the Timed Grid but leaves a pending (unsaved) Timed Event Occurrence Span on its old grid coordinates, so the span stays on the same clock rows while the Time Slots keep their Instants and shift. The pending span tooltip and the range that confirmScheduleEvent later saves also silently change meaning.

Reproduction:
- Event 15CFEPHH (canonical timed event, Europe/Moscow, 09:00-17:00 local, 15-minute slots, Sep 15-17) has scheduledEvent: null, so the span on the page is the pending selection set by a scheduling drag. GET /api/events/15CFEPHH returns the public payload with no saved span.
- Enter scheduling, drag a span over active slots, change the sidebar Display Timezone from GMT+8 to GMT+7.
- Reproduced in a component harness: the instant-based saved span moves from row 72 to row 68 (18:00 to 17:00 display), while the pending curScheduledEvent stays at row 72; getSelectedScheduleRange would then save the new instants.

Constraints:
- Keep one canonical internal shape for the pending span: back it with the selected start/end Instants and derive the rendered row/col/numRows, mirroring savedScheduledEvent.
- Do not change the saved-span path; it already follows the Display Timezone.
- Display Timezone changes must not change Event Picked Dates, the Enabled Domain, or Active Slots (FR-013).

Outcome:
- A pending span keeps its selected Instants when the Display Timezone changes, renders over the re-projected slots, reports the preserved range in its tooltip, and saves the preserved range on confirmation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 When the Display Timezone changes while a pending Timed Event Occurrence Span exists, the pending span's selected start and end Instants are preserved
- [x] #2 The pending span renders over the re-projected Time Slots that carry those Instants, including with collapse disabled times enabled by default
- [x] #3 The pending span tooltip reports the preserved start and end after the Display Timezone change
- [x] #4 Confirming the scheduling saves the preserved start and end Instants
- [x] #5 A saved Timed Event Occurrence Span continues to follow Display Timezone changes
- [x] #6 Unit regression tests cover the pending and saved span behavior
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
Research (2026-09-21):
- curScheduledEvent is a writable {row, col, numRows} ref in useEventScheduling. useDragPaint writes it at drag end; useScheduleOverlapController writes it from the scheduled_event URL param and clears it when leaving SCHEDULE_EVENT; confirmScheduleEvent derives the saved range from its coordinates through getDateFromRowCol.
- savedScheduledEvent already projects a saved start/end Instant range into the current grid, so it follows Display Timezone changes.
- startDrag and moveDrag only accept cells where getDateFromRowCol returns a slot, so a pending selection can always be converted to Instants at drag end.

Plan:
1. Add ScheduledEventRange { startDate; endDate } (Instants) to schedule_overlap types and make it the canonical pending selection in useEventScheduling.
2. Store curScheduledRange in useEventScheduling; derive curScheduledEvent from it with the same projection helper as savedScheduledEvent; expose setScheduledEventFromRowCol(ScheduledEvent | null) to convert coordinates to Instants; make getSelectedScheduleRange return the stored Instants for a pending selection; clear the range on cancel, confirm, clear, and state exit.
3. Replace the curScheduledEvent ref option in useDragPaint with setScheduledEventFromRowCol and call it at schedule drag end.
4. Replace the curScheduledEvent ref option in useScheduleOverlapController with setScheduledEventFromRowCol for the URL consume and the state-exit clear; wire it in ScheduleOverlap.vue.
5. Add FR-136 plus its README row for the span following the Display Timezone.
6. Add component and composable regression tests for pending and saved spans; update existing tests that write curScheduledEvent directly.
7. Run frontend checks, then graphify update.

Follow-up fixes plan (post-review, 2026-09-21):
1. Selection consistency: expose hasPendingScheduledEvent from useEventScheduling and derive selectedScheduledEvent = hasPending ? curScheduledEvent : savedScheduledEvent. Use it for allowScheduleEvent and scheduledEventStyle; pass the pending flag through useTimedGridPresentation, ScheduleOverlapTimeGridViewModel, useScheduleOverlapViewModels, and ScheduleOverlapTimeGrid.vue so a pending range never falls back to rendering or confirming the saved span. An off-grid pending range keeps its Instants, renders nothing, and disables the Schedule buttons until it projects again.

2. Deep link: in useScheduleOverlapController.onMounted set opts.state.value from getInitialState before calling setScheduledEventFromRowCol so coordinate-to-Instant conversion uses scheduling-state grid semantics; cover with a controller ordering assertion and a component test that mounts with ?scheduled_event=.

3. Drag typing: type UseDragPaintOptions.setScheduledEventFromRowCol as (ScheduledEvent | null) => void and pass getScheduledEventFromDragRange(ds, dc) unconditionally so a null drag range clears the pending selection instead of leaving a stale one.

4. Tests: make TimedGridPresentationForTest.renderedRows.rowTop required; add component coverage that a saved span is not substituted while a pending range exists, that a collapsed disabled-time run splits the pending span into fragments, and that a non-projecting pending range neither renders nor enables Schedule; keep the pending and saved Display Timezone tests.

5. E2E: extend timed-event-scheduling-tooltip-firefox.spec.ts to switch the Display Timezone after the drag and confirm that the preserved Instants reach the /schedule PUT body.

6. Checks: frontend lint, fmt:check, typecheck, build, test:unit; root format:markdown, format:markdown:check, lint:markdown; Firefox e2e for the edited specs; graphify update.

Plan adjustment during implementation (2026-09-21): the non-projecting pending-range regression is covered at the composable level (hasPendingScheduledEvent, selectedScheduledEvent, allowScheduleEvent, scheduledEventStyle, and confirmScheduleEvent still saving the pending Instants) instead of the component level. The Timed Grid derives its displayed date columns from the projected slots, so an off-grid pending state is not reachable through the real component grid; the component tests instead cover the reprojection, the collapsed-run fragments, and the scheduled_event URL mount.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-21: Confirmed the repro on event 15CFEPHH: GET /api/events/15CFEPHH returns scheduledEvent: null, so the reported span is the pending selection set by a scheduling drag. In a component harness the saved span moved row 72 -> 68 (18:00 -> 17:00 display) on GMT+8 -> GMT+7 while the coordinate-backed pending span stayed at row 72.

Fix: the pending selection is now backed by its selected start/end Instants in useEventScheduling.curScheduledRange, curScheduledEvent derives its row/col/numRows through the same projection helper as savedScheduledEvent, and setScheduledEventFromRowCol converts drag-end/URL coordinates into Instants. getSelectedScheduleRange returns the stored Instants, so confirmation after a Display Timezone change saves the originally selected range; cancel, confirm, clear, and leaving SCHEDULE_EVENT clear the range.

Writers moved from the curScheduledEvent ref to setScheduledEventFromRowCol in useDragPaint and useScheduleOverlapController; ScheduleOverlap.vue wires the setter. The saved-span path and the scheduled_event URL coordinate shape are unchanged.

Regression evidence: new ScheduleOverlap.scheduledSpanDisplayTimezone.test.ts covers the pending span (row 72 -> 68, rendered style aligned to the reprojected row, tooltip "5:00 PM to 7:00 PM - Tue, Jun 2, 2026", confirm saves 2026-06-02T10:00Z/12:00Z) and the saved-span guard. With the old coordinate pinning temporarily restored, the pending test fails with row 72 vs 68 while the saved test passes; after the fix both pass.

Checks: unit 156 files / 1249 tests passed; lint 0 errors (2 pre-existing NewSignUp.test.ts warnings); fmt:check clean; typecheck passed; build passed; focused Firefox e2e specs/timed-event-scheduling-tooltip-firefox.spec.ts passed (1 passed, 51.1s); npm run format:markdown and format:markdown:check clean; lint:markdown clean; graphify update completed.

The report's span was pending, not saved (the public payload has scheduledEvent: null). The saved-span path was already instant-based and still follows the Display Timezone; it is now guarded by a regression test.

The scheduled_event URL query parameter keeps its {row, col, numRows} shape; it is converted to Instants at consume time.

2026-09-21 review follow-up: re-opened to record and implement the review fixes. Findings addressed: (1) rendered span vs confirm target divergence when a pending range stops projecting while a saved span exists, (2) page/week navigation behavior for pending ranges, (3) deep-link conversion ordering and lossy coordinate mapping, (4) setScheduledEventFromRowCol nullability and dead guard in useDragPaint, (5) test gaps (optional rowTop, collapse rendering, off-grid policy), (6) e2e coverage for the preserved-Instant save. The FR-013 citation issue is handled by TASK-0304; FR-136 stays unchanged.

2026-09-21 follow-up implementation:
- Selection consistency: useEventScheduling now exposes hasPendingScheduledEvent and selectedScheduledEvent (pending range exists ? curScheduledEvent : savedScheduledEvent). allowScheduleEvent and scheduledEventStyle use it, and useTimedGridPresentation.scheduledEventStyles plus ScheduleOverlapTimeGrid.vue use it through the new ScheduleOverlapTimeGridViewModel.hasPendingScheduledEvent. A pending range that no longer projects renders nothing, does not fall back to the saved span, and disables the Schedule buttons; the preserved Instants stay in state so returning the view restores them.
- Deep link: useScheduleOverlapController.onMounted sets state from getInitialState before setScheduledEventFromRowCol; the controller test records the state at setter-call time and asserts SCHEDULE_EVENT, and ScheduleOverlap.scheduledSpanDisplayTimezone.test.ts mounts with ?scheduled_event= and asserts the pending span projects and the query param is consumed.
- Drag typing: UseDragPaintOptions.setScheduledEventFromRowCol is (ScheduledEvent | null) => void and endDrag passes getScheduledEventFromDragRange(ds, dc) unconditionally, so a null range clears the pending selection.
- Tests: TimedGridPresentationForTest.renderedRows.rowTop is now required; new off-grid composable regression (saved span projects, pending does not -> selectedScheduledEvent null, allowScheduleEvent false, empty schedule style, confirm still saves the preserved pending Instants); new collapsed-disabled-run component regression (base rows 38..54 across a collapsed 10:00-12:45 run renders two fragments at base rows 38 and 52); rowTop/height assertions hardened.
- E2E: timed-event-scheduling-tooltip-firefox.spec.ts switches the Display Timezone to Asia/Dhaka after the drag, asserts the reprojected tooltip 06:00-09:00 for the preserved 00:00-03:00Z Instants, and asserts the /schedule PUT body carries the preserved Instants.
Evidence: focused vitest 89 passed; full unit 157 files / 1253 tests passed; frontend lint 0 errors (2 pre-existing NewSignUp.test.ts warnings); fmt:check, typecheck, build passed; root format:markdown, format:markdown:check, lint:markdown clean; focused Firefox e2e 3 passed; full firefox-desktop e2e 71 passed with 1 pre-existing skip; graphify update completed.
DoD notes: #5 swagger and #7 scripts/prettier are not applicable because no Swagger annotations or root JS changed; #8 is covered by FR-136/README from the original round and the FR-013 amendment in TASK-0304.

Plan adjustment: the off-grid policy regression is composable-level because the component grid derives its columns from the projected slots; recorded in the plan.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Made a pending Timed Event Occurrence Span follow the Display Timezone by preserving its selected Instants instead of pinning it to grid coordinates.

Why: GET /api/events/15CFEPHH has no saved span, so the reported span was the pending selection set by a scheduling drag. It was stored as {row, col, numRows}, so a Display Timezone change re-projected the grid but left the span on its old clock rows; its tooltip and the range confirmScheduleEvent would save silently changed meaning. The saved span was already instant-based and correct.

Changes:
- useEventScheduling now stores the pending selection as start/end Instants (curScheduledRange), derives curScheduledEvent with the same projection helper as savedScheduledEvent, converts drag-end and scheduled_event URL coordinates through setScheduledEventFromRowCol, and returns the stored Instants from getSelectedScheduleRange.
- useDragPaint and useScheduleOverlapController write through setScheduledEventFromRowCol; ScheduleOverlap.vue wires the setter.
- Added FR-136 and its requirements index row.
- Added component regression tests for the pending and saved span; updated tests that wrote curScheduledEvent directly.

Evidence: with the old coordinate pinning restored, the new pending-span test fails at row 72 vs 68 while the saved-span guard passes; after the fix both pass. Unit 156 files / 1249 tests passed; lint 0 errors (2 pre-existing warnings); fmt:check, typecheck, and build passed; focused Firefox e2e specs/timed-event-scheduling-tooltip-firefox.spec.ts passed (51.1s); npm run format:markdown, format:markdown:check, and lint:markdown clean; graphify update completed.

Follow-up review fixes (2026-09-21): made the pending-versus-saved selection consistent by exposing hasPendingScheduledEvent and selectedScheduledEvent, so a pending range never renders or confirms as the saved span; fixed the scheduled_event deep-link conversion order; unified setScheduledEventFromRowCol nullability; hardened the presentation test types; added off-grid, collapsed-run, URL-mount, and controller-order regressions; extended the Firefox scheduling-tooltip spec to switch the Display Timezone and assert the preserved Instants reach the save request. The full Firefox desktop e2e suite passed. FR-013 now documents the projection-only invariant in TASK-0304.
<!-- SECTION:FINAL_SUMMARY:END -->
