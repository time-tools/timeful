---
id: TASK-0307
title: Show the Grid Pointer only on desktop hover while scheduling
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 16:52'
updated_date: '2026-09-21 17:35'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/components/schedule_overlap/useTimedGridPresentation.ts
  - frontend/src/components/schedule_overlap/useTimedGridInteractions.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/composables/schedule_overlap/useDragPaint.ts
  - frontend/src/composables/schedule_overlap/useAvailabilityData.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.test.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
  - docs/requirements/functional/fr/FR-137.md
  - docs/requirements/functional/fr/FR-135.md
  - docs/requirements/functional/fr/FR-095.md
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
  - docs/terminology/README.md
  - docs/terminology/glossary.md
priority: medium
type: bug
ordinal: 307000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the Timed Event Scheduling Page, the Grid Pointer (the black striped cell highlight) is currently rendered from click, tap, drag, and release interactions rather than only from non-pressing pointer interaction. On a phone viewport, tapping a Timed Grid cell shows the Grid Pointer together with the pending Timed Event Occurrence Span, and a subsequent drag from another cell moves the span while the Grid Pointer stays stuck on the first tapped cell. The extra highlight obscures which range is actually being scheduled.

The recorded product decision (FR-137, with the matching glossary Grid Pointer entry) is that during scheduling the Grid Pointer renders only from non-pressing pointer interaction on a desktop viewport. Click, tap, drag, or release must not render or retain it, and a phone viewport must never render it during scheduling. A release hides it even though the pointer is physically over the release cell; the next unpressed pointer movement shows it again at the cell under the pointer, including within the released cell. The pending Timed Event Occurrence Span remains the range marker during and after the interaction. Availability Editing keeps its existing Grid Pointer behavior.

Reproduction: on a phone viewport, open a timed event with Event Responses as the Event Owner, enter scheduling, tap a grid cell (the black striped Grid Pointer appears with the pending span), then drag from a different cell. The span moves to the drag range but the Grid Pointer stays at the first tapped cell.

FR-137 and the glossary Grid Pointer entry are already recorded in the working tree and must be committed with this fix; this task is the implementation and verification work order for them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On the Timed Event Scheduling Page, a desktop viewport renders the Grid Pointer only from non-pressing pointer interaction on a Timed Grid cell: hover shows it at the hovered cell; clicking, dragging, or releasing hides it; after release the next unpressed pointer movement shows it again at the cell under the pointer, including within the released cell.
- [x] #2 On a phone viewport, tapping, dragging, or releasing during scheduling never renders the Grid Pointer, including a tap on one cell followed by a drag from another cell.
- [x] #3 The pending Timed Event Occurrence Span still renders during and after the interaction, and its tooltip content, placement, and timezone projection are unchanged (FR-135, FR-136).
- [x] #4 Availability Editing keeps its existing Grid Pointer behavior, including the selected-responses state and the collapsed-hours exclusion (FR-066, FR-095).
- [x] #5 Regression coverage fails before the fix and passes after: unit coverage for the scheduling Grid Pointer state, and browser coverage on firefox-desktop (hover shows; click, drag, and release hide; the next movement re-shows it, including inside the released cell) and firefox-touch (tap and a subsequent drag never show).
- [x] #6 FR-137 and the glossary Grid Pointer entry match the implemented behavior and are committed with the fix, including the FR-137 row in docs/requirements/README.md.
- [x] #7 Frontend checks pass: lint, fmt:check, typecheck, build, and test:unit; changed Markdown is formatted.
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
## Implementation plan

Scope confirmed: only the Timed Grid during `states.SCHEDULE_EVENT`. Availability Editing (including the selected-responses `SUBSET_AVAILABILITY` cursor), the days-only grid, collapsed-hours exclusion, and the FR-135 tooltip stay unchanged.

1. Unit tests first (must fail before the fix):
   - `scheduleOverlapRendering.test.ts`: add a scheduling `Grid Pointer` gate to `TimeGridTimeslotArgs`; cases for `SCHEDULE_EVENT` with respondents and `schedulingGridPointerVisible` false (no cursor) and true (cursor).
   - `useTimedGridInteractions.test.ts`: extend the harness with `isScheduling` and `curTimeslot`; cases for press suppressing the pointer, end keeping it suppressed, the next unpressed move re-arming and updating the cell, and phone never arming.
2. Renderer: `getTimeGridTimeslotClassStyle` replaces the scheduling branch of the cursor gate with `schedulingGridPointerVisible`; editing/heatmap/best-times/specific-times/respondent conditions remain the non-scheduling gate.
3. Interactions: `useTimedGridInteractions` gains `isScheduling` and `curTimeslot` options, a `schedulingPointerSuppressed` ref set on `startTimedGridDrag`/`endTimedGridDrag` during non-days-only scheduling, and a `rearmSchedulingGridPointer` step from `moveTimedGridDrag` for unpressed desktop moves (`buttons === 0`) that updates `curTimeslot` when the resolved selectable cell differs and clears suppression. Expose a `schedulingGridPointerVisible` computed (`isScheduling && !isPhone && !daysOnly && !dragging && !suppressed`). Leaving scheduling resets suppression.
4. Wiring: `ScheduleOverlap.vue` passes the new interactions options and forwards `schedulingGridPointerVisible` to `useTimedGridPresentation`, which passes it to `buildTimeGridTimeslotClassStyles`.
5. Browser regression coverage:
   - `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (firefox-desktop): seed a response so the pre-fix gate is active; hover shows the cursor, drag and release leave none, and the next move inside the released cell shows it again.
   - `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` (firefox-touch): with a scheduled event and a response, tapping and dragging during scheduling never adds the cursor class while the pending span still updates.
6. Checks: frontend lint, fmt:check, typecheck, build, test:unit; focused firefox-desktop and firefox-touch specs; changed Markdown formatted; `graphify update .`.

## Review follow-up (2026-09-21)

1. Entering `SCHEDULE_EVENT` arms suppression on the transition, so a pointer parked over the grid cannot render the Grid Pointer before the first unpressed move; add a unit test for the transition.
2. `schedulingGridPointerVisible` also requires `interactable`; add a unit case.
3. Add the plain-click case from AC #5: a unit case that invokes the cell click handler while scheduling, and a firefox-desktop step that clicks a cell, asserts no cursor, then moves and asserts it re-arms.
4. Qualify the glossary and FR-137 to the Timed Grid and scope the click/tap/drag/release sentence to scheduling; the days-only grid keeps its existing behavior and is recorded in a follow-up task.
5. Correct the Final Summary labels that call the pre-existing test files `new`.
6. Re-run frontend checks and both focused Firefox specs, then re-finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause: the scheduling cursor was driven by the same `curTimeslot` state that click, tap, and drag set, while `updateCurTimeslot` no-ops in `SCHEDULE_EVENT`, so the highlight stuck at the press cell while the pending span moved. The fix separates scheduling pointer visibility from the interaction that set the cursor.

Pre-fix evidence captured by staging the tests before the behavior: `shows the scheduling cursor only for the active hover pointer` failed with the class still present; `suppresses the scheduling Grid Pointer from press until the next unpressed move` failed with `true !== false`. Browser evidence captured by temporarily reverting only the gate and suppression/rearm logic (production files restored byte-identically, `md5sum -c` clean): both new specs failed with `Expected: 0, Received: 1` cursor.

Post-fix checks: frontend lint, fmt:check, typecheck, build, `test:unit` (157 files / 1258 tests), focused firefox-desktop spec (2/2), full firefox-touch spec (12/12), root Markdown format check, and `graphify update .`.

E2E authoring note: Playwright has no touch-drag API for Firefox, so the phone regression keeps a native `touchscreen.tap` for the tap case and uses `page.mouse` for the drag case under the phone viewport; the test scrolls cells with `behavior: "instant"` because the app sets `scroll-behavior: smooth`, which otherwise made captured boxes go stale mid-animation.

Pre-existing condition, not introduced here: e2e `fmt:check` reports `specs/timed-event-access-transfer-firefox.spec.ts` as unformatted; the file is untouched by this task.

Review follow-up implemented. Behavior: entering SCHEDULE_EVENT arms suppression (`watch` sets `schedulingPointerSuppressed = isScheduling`) so a pointer parked over the grid cannot render the Grid Pointer before the first unpressed move; `schedulingGridPointerVisible` also requires `interactable`. New unit cases: parked pointer entering scheduling, non-interactable grid, click handler suppression.

Pre-fix evidence for the follow-up behavior tests: with only the entry-arming and interactable production changes temporarily reverted (file backed up and restored byte-identically via `md5sum`), both new tests failed with `expected true to be false`. Browser coverage for the plain click was added to `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (hover shows, click hides, next move re-shows).

Docs qualified to the Timed Grid in `FR-137.md` (phone line and ACs) and `glossary.md` (scheduling rule and click/tap/drag/release sentence scoped to scheduling on the Timed Grid). The open days-only scheduling Grid Pointer decision is recorded as TASK-0307.01.

Follow-up checks: frontend lint, fmt:check, typecheck, build, test:unit (157 files / 1261 tests), `format:markdown` (no changes), `format:markdown:check`, root `fmt:check`; e2e lint, typecheck, and fmt:check (only the pre-existing unformatted access-transfer spec); focused Firefox run 14/14; `graphify update .`. Final Summary corrected: the two test files are modified pre-existing files, not new files.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-21 16:52
---
Recorded ahead of execution: new `docs/requirements/functional/fr/FR-137.md` (status proposed) with its row in `docs/requirements/README.md`, plus the updated Grid Pointer entry in `docs/terminology/glossary.md` (authoritative context now FR-095 and FR-137). They are uncommitted in the working tree and AC #6 requires committing them with the fix. Product decision: on desktop, hover shows the Grid Pointer while click, drag, and release hide it; phone never shows it during scheduling. This supersedes the stale inbox note in `backlog/backlog.md` about the timeslot pointer being visible while dragging to schedule.
---

created: 2026-09-21 16:57
---
Release semantics refined after review: a release hides the Grid Pointer even though the pointer is physically over the release cell, and the next unpressed pointer movement shows it again at the cell under the pointer, including within the released cell. This avoids a black box reappearing at the release point while keeping the first subsequent hover responsive; the existing per-cell `mouseover` alone cannot re-show it within the same cell, so the implementation needs an explicit re-arm path. FR-137 and the glossary entry were updated to match.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

During Timed Event Occurrence Span scheduling, the Grid Pointer no longer renders from click, tap, drag, or release interactions on the Timed Grid. On desktop it renders only while an unpressed pointer moves over a Timed Grid cell; on a phone viewport it never renders during scheduling. Availability Editing, the days-only grid, the collapsed-hours exclusion, and the FR-135 scheduling tooltip are unchanged.

**Fix**
- `frontend/src/components/schedule_overlap/useTimedGridInteractions.ts`: `isScheduling`/`curTimeslot` options; `schedulingPointerSuppressed` armed on `startTimedGridDrag`/`endTimedGridDrag` and on entering `SCHEDULE_EVENT`, so a parked pointer must move before the pointer renders; `schedulingGridPointerVisible` is `isScheduling && !daysOnly && interactable && !isPhone && !dragging && !suppressed`; and `rearmSchedulingGridPointer` from unpressed desktop moves (`buttons === 0`) repositions `curTimeslot` to the cell under the pointer and re-shows the pointer, including inside the released cell.
- `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`: the scheduling branch of the cursor gate uses `schedulingGridPointerVisible`; every other state keeps the existing gate.
- `frontend/src/components/schedule_overlap/ScheduleOverlap.vue` and `useTimedGridPresentation.ts`: wiring; `useTimedGridPresentation.test.ts` updated for the new option.

**Tests**
- `useTimedGridInteractions.test.ts` (modified, 3 cases added): a parked pointer entering scheduling stays hidden until the first unpressed move; a non-interactable grid never shows; a click handler does not re-arm. Existing cases cover press/release suppression, re-arm, pressed moves, and phone.
- `scheduleOverlapRendering.test.ts` (modified): scheduling cursor hidden/shown by `schedulingGridPointerVisible`.
- `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (modified): hover shows; a plain click hides and the next move re-shows; drag/release hide and the next move inside the released cell re-shows.
- `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` (modified): tap, then drag from another cell, never shows the cursor while the pending span and tooltip update.
- Pre-fix evidence: the two new behavior unit tests failed with `expected true to be false` when only the entry-arming and interactable changes were temporarily reverted (production file restored by `md5sum`); the original e2e regressions failed pre-fix with `Received: 1` cursor.
- Post-fix: frontend `test:unit` (157 files / 1261 tests); one focused run of both specs passed 14/14 (firefox-desktop 2/2, firefox-touch 12/12).

**Requirements**
- `docs/requirements/functional/fr/FR-137.md` and the glossary Grid Pointer entry are qualified to the Timed Grid, including the phone line and the click/tap/drag/release sentence; the days-only scheduling Grid Pointer decision is tracked by TASK-0307.01.

**Checks**
- Frontend lint (2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, test:unit; `npm run format:markdown` and `format:markdown:check`; root `npm run fmt:check`; e2e lint and typecheck; `graphify update .`.
- E2E `fmt:check` still flags only the pre-existing unformatted `timed-event-access-transfer-firefox.spec.ts`, untouched by this task.
- Swagger annotations, contract documents, and root `scripts/`/`prettier/` checks are not applicable.
<!-- SECTION:FINAL_SUMMARY:END -->
