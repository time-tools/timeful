---
id: TASK-0302
title: Show the pending event duration in the scheduling tooltip
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 14:13'
updated_date: '2026-09-21 14:28'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/useTimedGridInteractions.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/composables/schedule_overlap/useEventScheduling.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
  - e2e/specs/timed-event-owner-authority-firefox.spec.ts
  - docs/requirements/functional/fr/FR-024.md
  - docs/requirements/functional/fr/FR-043.md
  - docs/requirements/functional/fr/FR-086.md
  - docs/requirements/functional/fr/FR-135.md
documentation:
  - docs/requirements/README.md
  - docs/requirements/AGENTS.md
  - docs/terminology/glossary.md
  - docs/terminology/README.md
modified_files:
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.schedulingTooltip.test.ts
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
  - docs/requirements/functional/fr/FR-135.md
  - docs/requirements/README.md
priority: medium
type: bug
ordinal: 302000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the Timed Event Scheduling Page, dragging out the pending Timed Event Occurrence Span makes the tooltip report only the Time Slot currently under the pointer, for example `8:45 PM to 9:00 PM` for a 15-minute Slot Duration, even though the blue pending span and the event's actual duration are larger. The Event Owner cannot confirm the start, end, or duration of the event they are about to save. The scheduling tooltip should report the pending Timed Event Occurrence Span (the event duration), not the hovered Time Slot range.

Reproduction: as the Event Owner, open a timed event and enter scheduling mode, then drag a span longer than one Time Slot. The tooltip tracks the pointer and shows the last Time Slot's range instead of the full pending span shown by the blue block.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On the Timed Event Scheduling Page, once a schedule drag defines a pending Timed Event Occurrence Span, the tooltip reports the pending span's start and end clock times (the event duration) instead of the hovered Time Slot's range
- [x] #2 The tooltip updates live while the drag endpoints move, and continues to report the pending span after the drag is released until the span is changed, cleared, saved, or cancelled
- [x] #3 Tooltip times follow the Display Time Format and Display Timezone (FR-024, FR-013), the tooltip date reflects the pending span's start date, and a span crossing midnight reports the correct start and end times and dates
- [x] #4 Availability Editing tooltips are unchanged and still report the hovered Time Slot's range
- [x] #5 Tooltip placement and visibility behavior is unchanged: it still appears at the interaction location, is absent when no Time Slot is selected, and keeps the existing mobile containment behavior (FR-043, FR-020, FR-096)
- [x] #6 Regression coverage fails before the fix and passes after: unit coverage for scheduling tooltip content with a span longer than one Time Slot, and browser coverage that drags a scheduling span and asserts the tooltip reports the full pending span
- [x] #7 A new canonical functional requirement (next free ID after FR-134, added per docs/requirements/AGENTS.md and functional/README.md, with its requirements README index row) states that on the Timed Event Scheduling Page the tooltip reports the pending Timed Event Occurrence Span's span, not the hovered Time Slot's range
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

1. Rendering helper (`frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`): add `formatScheduledSpanTooltipContent` that resolves the pending span's first and last Time Slots, computes the span end as the last slot plus the Slot Duration, and formats `start to end · date` under the existing Display Time Format and Display Timezone rules, appending the end date when the span crosses a display date. Leave `formatTooltipContent` unchanged so Availability Editing tooltips are untouched.
2. View wiring (`ScheduleOverlap.vue` `getTooltipContent`): when `state === SCHEDULE_EVENT` and a pending span exists (in-progress drag from `dragStart`/`dragCur`, otherwise `curScheduledEvent`), return the span content; otherwise keep the existing hovered-Time-Slot content.
3. Unit tests first: add `ScheduleOverlap.schedulingTooltip.test.ts` mounting the real component to assert the pending span during a scheduling drag and after release, and that a non-scheduling hover still reports the hovered Time Slot; add pure formatter cases to `scheduleOverlapRendering.test.ts` (span longer than one slot, cross-midnight, 12h/24h). Confirm the component test fails before the fix.
4. Browser coverage: add `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (firefox-desktop, context timezone UTC) that seeds a canonical timed event as the owner, enters scheduling, drags a three-slot span, and asserts the tooltip reports the full span during and after the drag. Capture the pre-fix failure before wiring the view.
5. Checks: frontend `lint`, `fmt:check`, `typecheck`, `build`, `test:unit`; focused firefox-desktop e2e; root Markdown check if Markdown changes; `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Pre-fix evidence: `ScheduleOverlap.schedulingTooltip.test.ts` failed 3 tests before the view wiring (hover returned `10:00 AM to 11:00 AM` instead of the pending span), and the focused firefox-desktop spec failed with received `02:00 to 03:00 · Thu, May 28, 2026`. Both pass after the fix; logs in /tmp/opencode/task-0302-prefix.log and task-0302-postfix.log.

Design decisions: the scheduling branch only triggers in `SCHEDULE_EVENT` state when a drag or `curScheduledEvent` exists, so hover before a pending span and all Availability Editing tooltips keep the hovered-slot content. The span end is the last slot plus the Slot Duration, matching `useEventScheduling.getSelectedScheduleRange`; the span formatter uses the same display-date resolver (`getDateFromRowCol` with `getDisplayDateFromRowCol` fallback) as the hover tooltip so weekly/group display dates stay consistent.

Checks: frontend lint/fmt:check/typecheck/build/test:unit pass (155 files, 1247 tests); e2e lint/typecheck pass; focused firefox-desktop and firefox-touch specs pass; root `format:markdown:check` and `graphify update .` run. E2E `fmt:check` flags `e2e/specs/timed-event-access-transfer-firefox.spec.ts`, which is unmodified in git and untouched by this task.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-21 14:16
---
AC #7 satisfied ahead of execution: added `docs/requirements/functional/fr/FR-135.md` ("Show the pending Timed Event Occurrence Span in the scheduling tooltip", status proposed) and its index row in `docs/requirements/README.md`. The remaining work is the tooltip fix plus unit and e2e regression coverage.
---

created: 2026-09-21 14:16
---
FR-135 records that on the Timed Event Scheduling Page the tooltip for a pending Timed Event Occurrence Span reports that span's start and end times and date instead of the hovered Time Slot's range, that the tooltip tracks the pending span during and after the schedule drag, and that Availability Editing tooltips are unchanged.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

On the Timed Event Scheduling Page, the timed-grid tooltip now reports the pending Timed Event Occurrence Span (the event duration) instead of the Time Slot under the pointer, so the Event Owner can confirm the event's start, end, and duration before saving.

**Fix**
- `frontend/src/components/schedule_overlap/ScheduleOverlap.vue` `getTooltipContent`: in `SCHEDULE_EVENT` state, prefer the pending span (in-progress drag from `dragStart`/`dragCur`, otherwise the released `curScheduledEvent`) and fall back to the existing hovered-slot tooltip when no pending span exists or its boundary slots cannot be resolved.
- `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`: new `formatScheduledSpanTooltipContent` resolves the span's first and last slots through the same display-date resolver as the hover tooltip, computes the end as the last slot plus the Slot Duration, follows the Display Time Format and Display Timezone, and appends the end date when the span crosses a display date. `formatTooltipContent` output is unchanged and now shares time/date format helpers.

**Tests**
- Pre-fix failure captured: `ScheduleOverlap.schedulingTooltip.test.ts` failed 3 tests with the hovered slot (`10:00 AM to 11:00 AM`) instead of the span, and the focused firefox-desktop spec failed with received `02:00 to 03:00 \u00b7 Thu, May 28, 2026`.
- New `ScheduleOverlap.schedulingTooltip.test.ts` (5 tests): pending span on hover, live drag updates, span after release, cancel returns to the hovered slot, Availability Editing unchanged.
- `scheduleOverlapRendering.test.ts`: span longer than one slot, cross-midnight span in 24h specific-date and 12h weekly-date forms, missing boundary slot returns no content.
- New `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts`: owner enters scheduling, drags three hourly slots, and the tooltip reports `00:00 to 03:00 \u00b7 Thu, May 28, 2026` during and after the drag.
- Regression checks: `timed-event-specific-times-edit-firefox.spec.ts` (6 passed) and `schedule-overlap-mobile-touch-firefox.spec.ts` (10 passed).

**Requirements**
- New `docs/requirements/functional/fr/FR-135.md` records the behavior; its row was added to `docs/requirements/README.md`.

**Checks**
- Frontend: lint (only the two pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, test:unit (155 files / 1247 tests).
- E2E: lint, typecheck, focused specs above.
- Root `format:markdown:check` and `graphify update .` run.
- E2E `fmt:check` reports one pre-existing unformatted file (`e2e/specs/timed-event-access-transfer-firefox.spec.ts`) that this task did not touch.
- Swagger, `scripts/`/`prettier/`, and contract-document checks are not applicable (no server, tooling, or contract changes).
<!-- SECTION:FINAL_SUMMARY:END -->
