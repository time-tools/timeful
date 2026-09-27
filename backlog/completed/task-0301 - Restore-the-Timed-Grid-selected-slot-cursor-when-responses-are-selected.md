---
id: TASK-0301
title: Restore the Timed Grid selected-slot cursor when responses are selected
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 13:52'
updated_date: '2026-09-21 14:10'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.test.ts
  - frontend/src/components/schedule_overlap/useTimedGridInteractions.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
  - e2e/specs/timed-event-respondent-selection-firefox.spec.ts
  - docs/requirements/functional/fr/FR-066.md
documentation:
  - docs/requirements/README.md
  - docs/terminology/README.md
priority: medium
type: bug
ordinal: 301000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reported: on the phone viewport, after selecting one or more Event Responses in the responses list, tapping a Timed Grid cell does not show the selected-slot highlight (Grid Pointer), even though the mobile tooltip anchors to the tapped slot.

Reproduced on the isolated test stack (chromium-mobile, iPhone 13 viewport, temporary scratch spec, removed afterwards): seed a canonical Timed Event with two guest responses, tap a slot with no selection (the tapped cell gets class `schedule-overlap-time-grid__selected-timeslot` and a `::after` cursor), then select Guest One through the respondents checkbox and tap a second slot. The tooltip moves to the second slot, but that cell never receives the class and its `::after` computed style stays `none`.

Root cause: in `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`, `getTimeGridTimeslotClassStyle` gates the cursor on `state === states.HEATMAP || state === states.BEST_TIMES || editing || state === states.SET_SPECIFIC_TIMES`, which excludes `states.SUBSET_AVAILABILITY` (the state `useScheduleOverlapUI.clickRespondent` sets when responses are selected). The days-only grid (`getDayGridTimeslotClassStyle`) still checks `respondents.length > 0 || state === states.EDIT_AVAILABILITY`, so the two grids are inconsistent. This is a migration regression: the pre-refactor Vue 2 code gated the timed-grid cursor on `respondents.length > 0 || editing || state === states.SET_SPECIFIC_TIMES` (see commit `c8f8d704^`).

Constraints: keep the existing zero-response HEATMAP/BEST cursor behavior and its unit coverage, keep the cursor off collapsed-hours strips (FR-095) and out of disabled/out-of-range cells, and keep FR-042 (highlight stays within the cell). Determine whether the intended selected-responses pointer behavior belongs in FR-066 or the Grid Pointer glossary entry and amend requirements only if needed, following `docs/requirements/README.md`.

A separate follow-up task covers the Vue expand-transition `parentNode` TypeError observed in a real browser session; it did not appear in the automated reproduction of this highlight defect and is out of scope here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 With one or more Event Responses selected in the responses list, a Timed Grid cell under the pointer on desktop and under a phone tap shows the selected-slot cursor (`schedule-overlap-time-grid__selected-timeslot`), matching the dates-only grid
- [x] #2 The cursor appears after pointer hover (desktop) and after tap (phone), and clears when the response selection is cleared, in the same way the dates-only grid does
- [x] #3 No regression: the cursor still appears for the existing HEATMAP/BEST_TIMES/editing/specific-times states, including the zero-response case, and never appears on collapsed-hours strips or disabled/out-of-range cells
- [x] #4 Regression coverage: rendering unit tests assert the timed-grid cursor class for SUBSET_AVAILABILITY, and browser coverage asserts the cursor after selecting a response and tapping/hovering a slot (touch and mouse), with the test failing before the fix
- [x] #5 Existing checks pass: frontend lint, fmt:check, typecheck, build, and test:unit, plus the focused e2e specs touched by the change; changed Markdown is formatted
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

1. Reproduce at the unit layer first: add a `scheduleOverlapRendering.test.ts` case for `states.SUBSET_AVAILABILITY` with a selected response, `respondents.length > 0`, and `curTimeslot` on the cell; confirm it fails before the fix.
2. Fix `getTimeGridTimeslotClassStyle` in `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`: restore the pre-refactor/dates-only gate by adding `respondents.length > 0` to the cursor condition alongside the existing zero-response `HEATMAP`/`BEST_TIMES`, `editing`, and `SET_SPECIFIC_TIMES` cases, keeping the `!isDisabled` and `curTimeslot` match checks.
3. Add browser regression coverage that fails before the fix:
   - mouse hover on `e2e/specs/timed-event-respondent-selection-firefox.spec.ts` (firefox-desktop),
   - touch tap on `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` (firefox-touch),
   each seeding a canonical timed event with a guest response, selecting the response, then asserting the tapped/hovered cell carries `schedule-overlap-time-grid__selected-timeslot` and a non-`none` `::after` cursor.
4. Run the focused e2e specs against the unfixed code to capture the pre-fix failure, then rerun after the fix.
5. Requirements determination: the Grid Pointer glossary entry already defines the highlight during Availability Editing, which includes the selected-responses state, and FR-095 covers collapsed strips; amend FR-066 or the glossary only if the determination surfaces a real gap.
6. Run the required checks (`lint`, `fmt:check`, `typecheck`, `build`, `test:unit`), then `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Research: `clickRespondent` sets `state = SUBSET_AVAILABILITY` (useScheduleOverlapUI.ts:207) while `respondents` still holds every response, so the dates-only grid's `respondents.length > 0` gate shows the cursor but the timed-grid gate excludes it. Pre-refactor Vue 2 `ScheduleOverlap.vue` at `c8f8d704^` gated the timed cell on `respondents.value.length > 0 || editing.value || state.value === states.SET_SPECIFIC_TIMES`, confirming the migration drop is the regression.

Pre-fix evidence: unit case failed (`tw:relative` absent from `tw:border-t tw:border-r tw:border-l tw:border-r tw:border-b`), and both new e2e tests failed with `toHaveClass(/schedule-overlap-time-grid__selected-timeslot/)` receiving `timeslot tw:h-full tw:w-full tw:border-t tw:border-r tw:border-l tw:border-r`.
After adding `respondents.length > 0` to the timed-grid cursor gate: the focused unit file passes (34/34) and both focused e2e tests pass (firefox-desktop hover and firefox-touch tap).

Requirements determination: no amendment needed. The Grid Pointer glossary entry defines the highlight during Availability Editing, which includes the selected-responses state (`clickRespondent` is one of FR-066's authoritative links), and FR-095 keeps it off collapsed strips. FR-066 governs which availability is shown and how overlap is calculated, not the pointer, so the regression was code deviating from existing requirements rather than a requirement gap.
Browser clearing evidence: the firefox-touch test taps outside the grid after selecting a respondent, asserts the checkbox returns to `aria-pressed="false"`, and asserts the cursor class is removed, confirming the selection-clear path clears the pointer on mobile where mouseleave does not reset `curTimeslot`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Restored the Timed Grid Grid Pointer while Event Responses are selected.

**Root cause and fix**
`getTimeGridTimeslotClassStyle` dropped the pre-refactor `respondents.length > 0` clause from the selected-slot cursor gate, so `states.SUBSET_AVAILABILITY` (set by `clickRespondent`) never rendered `schedule-overlap-time-grid__selected-timeslot` even though `curTimeslot` tracked the pointed cell. The renderer now matches the dates-only grid and the pre-refactor Vue 2 gate: `respondents.length > 0 || HEATMAP || BEST_TIMES || editing || SET_SPECIFIC_TIMES`, still guarded by the `curTimeslot` match and `!isDisabled`, so collapsed-hours strips, disabled, and out-of-range cells never receive the cursor.

**Tests**
- `frontend/src/components/schedule_overlap/scheduleOverlapRendering.test.ts`: new SUBSET_AVAILABILITY cursor case; failed before the fix (`tw:relative` missing) and passes after.
- `e2e/specs/timed-event-respondent-selection-firefox.spec.ts`: desktop cursor after selecting a response and hovering a slot (firefox-desktop).
- `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts`: phone cursor after selecting a response and tapping a slot, plus tapping outside clears the selection and the cursor (firefox-touch).
- Both browser tests failed before the fix with the cell missing the cursor class; both pass after, and the full focused spec files pass (12 tests).

**Checks**
Frontend lint (only two pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, test:unit (154 files / 1239 tests); e2e lint, fmt:check, typecheck; focused firefox-desktop + firefox-touch specs; `graphify update .` (no topology changes).

**Requirements determination**
No requirement amendment needed. The Grid Pointer glossary entry already defines the highlight as appearing during Availability Editing, which includes the selected-responses state, and FR-095 already keeps it off collapsed strips; FR-066 governs availability filtering and overlap calculation rather than the pointer. No Markdown, swagger, contract, `scripts/`, or `prettier/` files changed, so those conditional checks are not applicable (`backlog/**` is excluded from the Markdown pipeline).
<!-- SECTION:FINAL_SUMMARY:END -->
