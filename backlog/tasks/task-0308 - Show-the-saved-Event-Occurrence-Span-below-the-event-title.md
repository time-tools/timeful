---
id: TASK-0308
title: Show the saved Event Occurrence Span below the event title
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 17:50'
updated_date: '2026-09-21 18:23'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - frontend/src/types/index.ts
  - frontend/src/utils/dateFormatting.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/composables/event/types.ts
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-012.md
  - docs/requirements/functional/fr/FR-086.md
  - docs/requirements/functional/fr/FR-136.md
  - docs/requirements/AGENTS.md
  - docs/terminology/glossary.md
modified_files:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - frontend/src/utils/dateFormatting.ts
  - frontend/src/utils/dateFormatting.test.ts
  - frontend/src/utils/index.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/composables/event/types.ts
  - frontend/src/composables/event/useEventEditing.test.ts
  - e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts
  - docs/requirements/functional/fr/FR-138.md
  - docs/requirements/README.md
priority: medium
type: enhancement
ordinal: 309000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When an event has a saved Event Occurrence Span, the confirmed date and time are only discoverable inside the grid (the highlighted block and its hover tooltip). The header that identifies the event never states when it is, so a visitor landing on the page cannot tell the confirmed date/time at a glance, and the Reschedule control is action-oriented rather than informational. Show the saved span directly below the event title in the left header column so the confirmed occurrence reads as part of the event's identity, while the right-hand column stays reserved for actions. The span must agree with the grid: same Display Timezone, same 12h/24h time format, and no separate copy of timezone or occurrence state.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 When the event has a saved Event Occurrence Span, the event header shows its start and end below the event title for every Event Visitor who can view the page, including signed-out visitors.
- [x] #2 When no Event Occurrence Span is saved, no occurrence span text is shown below the event title.
- [x] #3 The displayed span follows the Display Timezone and the 12h/24h time format preference and updates without a page reload when either changes.
- [x] #4 A Timed Event Occurrence Span shows its date and time range, and when start and end fall on different Civil Dates both dates are identifiable.
- [x] #5 A Dates-Only Event Occurrence Span shows a single date and no time range.
- [x] #6 After scheduling, rescheduling, or clearing the Event Occurrence Span, the header span reflects the saved state without a page reload.
- [x] #7 While a reschedule is pending and unsaved, the header keeps showing the saved span rather than the unsaved selection.
- [x] #8 Unit tests cover span formatting (12h and 24h, non-UTC Display Timezones, ranges spanning different Civil Dates, and Dates-Only Events) and header visibility.
- [x] #9 An e2e test schedules an event, asserts the header span matches the saved occurrence, and asserts the span follows a Display Timezone change.
- [x] #10 The requirements corpus records the header rendering behavior using canonical glossary terms.
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

Placement decision (approved by the user): render the saved Event Occurrence Span directly below the event title in the left header column. No code changes started yet; this plan is pending user approval before implementation.

### Research findings (current tree)

- Saved span data: `event.scheduledEvent` is the internal `CalendarEvent` at `frontend/src/types/index.ts:103-106`, with `startDate`/`endDate` as `Temporal.ZonedDateTime` instants tagged UTC, decoded from epoch millis by `fromRawCalendarEvent` (`frontend/src/types/transport.ts:633-640`). `event.daysOnly` separates Timed Events from Dates-Only Events; `event.scheduledEvent` presence already backs `hasSavedTimefulSchedule` (`frontend/src/views/Event.vue:1442-1444`).
- Display Timezone already reaches the parent: `ScheduleOverlap` exposes `curTimezone` (`ScheduleOverlap.vue:1203`) and `ScheduleOverlapInstance.curTimezone` is typed (`frontend/src/composables/event/types.ts:14`). Event.vue already reads exposed child state in computeds (`Event.vue:1487-1499`).
- The 12h/24h `timeType` lives in `useCalendarGrid.ts:81-86` and is not exposed yet.
- Formatters: the grid tooltip helpers `getTooltipTimeFormat` / `formatScheduledSpanTooltipContent` live in `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts:904-997`; `getStartEndDateString` (`frontend/src/utils/dateFormatting.ts:16-38`) is not Display Timezone aware. The shared conversion helper is `getDateInTimezone` (`frontend/src/utils/timezoneDateRules.ts:63-78`).
- No transport or server changes are needed; the span is already decoded.

### Steps

1. Tests first, per the frontend working defaults:
   - `frontend/src/utils/dateFormatting.test.ts` (exists): formatter cases for 12h and 24h, a non-UTC Display Timezone, a range spanning different Civil Dates, and a Dates-Only span.
   - `frontend/src/views/Event.test.ts`: extend `ScheduleOverlapStub` (`Event.test.ts:303+`) with `curTimezone` and `timeType`; add header tests for visible span, hidden when unsaved, hidden for dates-only time range, and reactive updates when the stub values change.
2. Define the time format once: move the 12h/24h `Intl.DateTimeFormatOptions` out of `getTooltipTimeFormat` into `frontend/src/utils/dateFormatting.ts` and import it from `scheduleOverlapRendering.ts`, so the grid tooltip and the header cannot diverge.
3. Add a Display Timezone aware span formatter to `frontend/src/utils/dateFormatting.ts`: inputs are the span start/end, `Timezone`, `TimeType`, and `daysOnly`; convert with `getDateInTimezone` (no call-site reimplementation); timed output is date-first with the time range (for example `Tue, Sep 15, 2026 · 8:00 AM – 8:15 AM`) and adds the end date when the Civil Dates differ; dates-only output is the single occurrence date with no time range.
4. Expose `timeType` from `ScheduleOverlap.vue` `defineExpose` and add it to `ScheduleOverlapInstance` (`frontend/src/composables/event/types.ts`).
5. In `Event.vue`, add computeds: Display Timezone from `scheduleOverlap.value?.curTimezone ?? props.initialTimezone`, and the span text from `event.scheduledEvent` plus that timezone. Render the text below the title/chips row inside the existing left column (`Event.vue:172-207`) with a stable hook such as `id="event-header-scheduled-span"`. Do not introduce a second copy of timezone or occurrence state.
6. Behavior while scheduling: keep showing the saved span, never the pending unsaved selection; after Timeful confirm/clear, `refreshEvent` reloads the saved span and the computed follows. Render nothing until a Display Timezone is known, to avoid wrong-zone text before `ScheduleOverlap` mounts (`props.initialTimezone` covers `?tz=` links).
7. Requirements docs: read `docs/requirements/AGENTS.md` and `docs/requirements/README.md` first, then add an FR for header rendering with the README index row, using canonical glossary terms (Event Occurrence Span, Timed Event Occurrence Span, Dates-Only Event Occurrence Span, Display Timezone, Event Visitor). FR-136 already requires a saved span to follow the Display Timezone.
8. Browser coverage: extend `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (or a focused sibling spec) to schedule an event, assert the header span matches the saved occurrence, then use the `changeDisplayTimezone` helper and assert the header follows. Cover the dates-only case in the days-only spec if practical.
9. Checks: frontend `lint`, `fmt:check`, `typecheck`, `build`, `test:unit`; one focused Firefox e2e run; Markdown formatting; `graphify update .`.

### Open decisions / risks

- Exact typography (separator, date-first vs time-first, timezone suffix) is a design detail; the plan is date-first with an en dash and no timezone suffix because the Display Timezone control is adjacent. Confirm before implementation if you want a different style.
- A Dates-Only span rendered in a Display Timezone other than the Event Timezone can land on a different Civil Date; follow the grid projection via `getDateInTimezone` for consistency with FR-136.
- Multi-day and midnight-crossing timed spans must show both Civil Dates.
- Pre-mount fallback: when the child is not mounted and no `initialTimezone` prop exists, the span is omitted until `curTimezone` is available rather than rendered in a guessed timezone.

## Two-date rendering (approved 2026-09-21)

The user confirmed a Timed Event Occurrence Span can cross midnight and approved the two-date rendering: same date `Sat, Jul 4, 2026 · 11:00 PM – 11:30 PM`; two dates `Sat, Jul 4, 2026 · 11:30 PM – Sun, Jul 5, 2026 · 12:30 AM`. Dates-Only spans stay a single date. Formatter unit tests must cover the two-date case, mirroring the grid tooltip case `scheduleOverlapRendering.test.ts:189` (`reports both dates when a scheduled span crosses midnight`).

## Post-review refresh (2026-09-21)

- The pre-mount `props.initialTimezone` fallback was removed: the header span now waits for the exposed `curTimezone`/`timeType`, so it never renders a guessed display format (the plan's step 6 rule, now enforced for the format as well).
- `specificDatesDateFormatOptions` is shared by the header formatter and the grid tooltip, extending the step 2 "define it once" rule to the full date format.
- Header unit tests moved to a dedicated `Event header occurrence span` describe; a stub with a differing pending range guards the saved-vs-pending boundary; a fixed-offset Display Timezone case was added.
- The e2e reschedule step now drags a different pending range and asserts the header still shows the saved span.
- FR-138 records the scheduling-mode behavior: the saved span stays and a pending unsaved selection does not change it.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation followed the recorded plan. TDD: formatter unit tests failed first (`getEventOccurrenceSpanString is not a function`) and header unit tests failed before the template existed (missing `#event-header-scheduled-span`). Added `getTimeFormatOptions` + `getEventOccurrenceSpanString` to `dateFormatting.ts`, reused the shared clock format and full date options in `scheduleOverlapRendering.ts`, exposed `timeType` from `ScheduleOverlap`, typed it on `ScheduleOverlapInstance`, and rendered the span below the title from `event.scheduledEvent` plus the exposed `curTimezone`/`timeType`. Two-date and dates-only rendering mirror the grid tooltip through `getDateInTimezone`. No server or transport changes.

Post-review refresh (2026-09-21):
- The header span no longer falls back to `props.initialTimezone` with a hardcoded 24-hour format; it renders only after `ScheduleOverlap` exposes `curTimezone`/`timeType`, so no frame can show the wrong display format.
- `specificDatesDateFormatOptions` is shared between the header formatter and the grid tooltip, so the full date format cannot drift either.
- Header tests moved to a dedicated `Event header occurrence span` describe; the pending-selection unit test uses a stub carrying a differing pending range; a fixed-offset Display Timezone formatter case was added.
- The e2e reschedule step now drags a different pending range and asserts the header still shows the saved span. FR-138 records the scheduling-mode behavior.

Checks: frontend lint (only the 2 pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, build, test:unit (157 files / 1271 tests, +10 new). E2E firefox-desktop `specs/timed-event-scheduling-tooltip-firefox.spec.ts` 3/3. E2E lint and typecheck clean; e2e fmt:check flags only the pre-existing unformatted access-transfer spec. Root fmt:check, format:markdown:check, and lint:markdown clean; `graphify update .` ran.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The saved Event Occurrence Span now renders in the event header directly below the event title, so any Event Visitor sees the confirmed date/time at a glance instead of only inside the grid block and its hover tooltip. The rendered text follows the Display Timezone and the 12h/24h display time format, and it agrees with the grid because both share one clock-format definition, one full-date definition, and the centralized timezone conversion.

**Placement and behavior**
- `frontend/src/views/Event.vue`: `#event-header-scheduled-span` below the title/chips row in the left header column, driven by a computed over `event.scheduledEvent` and the exposed `curTimezone`/`timeType`. The span renders only once `ScheduleOverlap` exposes the display state, so it never shows a guessed timezone or format. Hidden when no span is saved; the saved span stays visible while a reschedule is pending and updates after save or clear without a page reload.
- `frontend/src/utils/dateFormatting.ts`: new `getEventOccurrenceSpanString` (date-first, `Tue, Sep 15, 2026 · 8:00 AM – 8:15 AM`; both Civil Dates when the span crosses midnight; dates-only renders a single date with no time range), shared `getTimeFormatOptions`, and shared `specificDatesDateFormatOptions`.
- `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`: the grid tooltip imports the shared clock-format and full-date helpers instead of local copies.
- `frontend/src/components/schedule_overlap/ScheduleOverlap.vue` + `frontend/src/composables/event/types.ts`: expose and type `timeType` next to the already-exposed `curTimezone`; no duplicate timezone or occurrence state.

**Tests**
- `frontend/src/utils/dateFormatting.test.ts` (5 cases): 12h/24h, non-UTC IANA Display Timezone, fixed-offset Display Timezone, midnight-crossing two-date span, dates-only single date.
- `frontend/src/views/Event.test.ts` (5 cases in a dedicated `Event header occurrence span` describe): span shown for a signed-out viewer, hidden when unsaved, reactive to Display Timezone and format changes, saved span kept while a different pending selection is unsaved, and dates-only.
- `frontend/src/composables/event/useEventEditing.test.ts`: mock instance updated for the new `timeType` field.
- `e2e/specs/timed-event-scheduling-tooltip-firefox.spec.ts` (new test): seeded saved span renders in the header, follows a Display Timezone change (`Asia/Dhaka`), stays while a different pending reschedule range is unsaved, disappears after Clear, and reappears after re-scheduling through the grid without a reload.

**Requirements**
- New `docs/requirements/functional/fr/FR-138.md` (proposed) with its `docs/requirements/README.md` index row, using canonical glossary terms; it records the scheduling-mode behavior where the saved span survives a pending selection.

**Checks**
- Frontend: lint (2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, build, test:unit (157 files / 1271 tests).
- E2E: firefox-desktop focused spec 3/3; e2e lint and typecheck; e2e fmt:check flags only the pre-existing unformatted access-transfer spec.
- Root `fmt:check`, `format:markdown:check`, `lint:markdown`, and `graphify update .` ran clean. Swagger, contract documents, and server code are unaffected.

**Risk / follow-up**
- None material. The span intentionally appears only after `ScheduleOverlap` mounts, which happens automatically during page bootstrap.
<!-- SECTION:FINAL_SUMMARY:END -->
