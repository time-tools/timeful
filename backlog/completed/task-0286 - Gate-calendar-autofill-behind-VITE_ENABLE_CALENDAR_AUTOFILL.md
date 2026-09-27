---
id: TASK-0286
title: Gate calendar autofill behind VITE_ENABLE_CALENDAR_AUTOFILL
status: Done
assignee: []
created_date: '2026-09-20 16:41'
updated_date: '2026-09-20 17:30'
labels: []
dependencies: []
modified_files:
  - .env.development.example
  - .env.production.example
  - .env.staging.example
  - .env.test.example
  - compose.yaml
  - docs/environments.md
  - docs/requirements/README.md
  - docs/requirements/functional/fr/FR-131.md
  - frontend/Dockerfile
  - frontend/env.d.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.calendarOptionsGating.test.ts
  - frontend/src/composables/event/useEventEditing.ts
  - frontend/src/composables/event/useEventEditing.calendarAutofill.test.ts
  - frontend/src/composables/event/useEventLoader.ts
  - frontend/src/composables/event/useEventLoader.test.ts
  - frontend/src/composables/schedule_overlap/useAvailabilityData.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/utils/calendarAutofillAvailability.ts
  - frontend/src/utils/calendarAutofillAvailability.test.ts
  - frontend/src/utils/featureAvailability.ts
  - frontend/src/utils/index.ts
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
priority: medium
type: feature
ordinal: 286000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add a frontend build-time flag `VITE_ENABLE_CALENDAR_AUTOFILL` (defaults enabled when unset or blank) that disables filling availability from a calendar on event pages.

When disabled:
- The "How would you like to add your availability?" modal never appears; Add/Edit availability starts manual Availability Editing.
- No event-page path fills availability from calendar data automatically (add-availability autofill, post-loader refill, reanimate on refresh, group invitation flow, Apple link redirect).
- The "Calendar options" control (buffer time and working hours dialog) is not rendered during Availability Editing on desktop or mobile.

When enabled or unset, all current behavior is unchanged. Calendar accounts/overlays, Settings calendar access, and the landing page stay as they are.

Record the disabled behavior as a new functional requirement and document the flag in docs/environments.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 VITE_ENABLE_CALENDAR_AUTOFILL is declared in frontend env typing, the frontend Docker build, Compose frontend-artifacts build args, and the example env files, and defaults to enabled when unset or blank
- [x] #2 With the flag false, Add availability and Edit availability never open the source-choice dialog and start manual Availability Editing
- [x] #3 With the flag false, no event-page path fills availability from calendar data automatically, including the post-loader refill, reanimate-on-refresh, group invitation flow, and Apple link redirect
- [x] #4 With the flag false, the Calendar options control (buffer time and working hours) is not rendered during Availability Editing on desktop or mobile
- [x] #5 With the flag enabled or unset, existing event-page behavior is unchanged, including calendar options, calendar overlays, and connected calendar accounts
- [x] #6 FR-131 records the disabled behavior and the requirements index links it
- [x] #7 docs/environments.md documents the flag semantics
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
1. Add VITE_ENABLE_CALENDAR_AUTOFILL to featureAvailability.ts, a calendarAutofillAvailability.ts wrapper, utils barrel, env.d.ts, Dockerfile, compose.yaml, and the four tracked .env.*.example files; add the variable to local .env.development and .env.test.
2. Gate event-page autofill: skip the choice dialog and autofill in useEventEditing.addAvailability, guard useEventEditing.setAvailabilityAutomatically, skip the post-loader refill in useEventLoader, guard useAvailabilityData.reanimateAvailability, skip the linkApple dialog and group invitation auto-open in Event.vue, and force ScheduleOverlap.showCalendarOptions false.
3. Add unit coverage: flag parsing, useEventEditing flag-off behavior, ScheduleOverlap showCalendarOptions gating, loader refill gating, and Event.vue open-site gating.
4. Author FR-131, add its requirements README row, and document the flag in docs/environments.md.
5. Run frontend lint, fmt:check, typecheck, build, test:unit, Markdown formatting, and graphify update.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Progress on 2026-09-20 (implementation complete; task remains In Progress):

Done:
- Flag plumbing: featureAvailability.ts, calendarAutofillAvailability.ts, utils barrel, frontend/env.d.ts, frontend/Dockerfile, compose.yaml, four tracked .env.*.example files, plus local ignored .env.development and .env.test.
- Behavior gating: choice dialog skipped for Add/Edit availability, all event-page auto-fill paths skipped, Calendar options control hidden when disabled; enabled behavior preserved.
- Tests added for flag parsing, calendar-options visibility, manual start/no dialog/inert autofill, loader refill on/off, Event.vue linkApple and group-invitation gating.
- FR-131 authored with its requirements README row; docs/environments.md documents the flag.

Checks run and passing: frontend lint (0 errors; 2 pre-existing warnings in NewSignUp.test.ts), fmt:check, typecheck, build, test:unit (154 files / 1209 tests); root format:markdown and format:markdown:check; graphify update .

Remaining for the next session:
- E2E regression run was started for e2e/specs/event-mobile-editing-options.spec.ts --project=chromium-mobile and aborted at user request; no E2E result was recorded. The isolated test stack and test-mode Vite process were torn down (docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down -v). Decide whether to run E2E before finalizing; DoD #3 is the only outstanding hygiene item. DoD #7 is not applicable because no scripts/ or prettier/ files changed.
- Finalize the task (final summary, mark Done); do not move it to backlog/completed/.
- Worktree note: implementation changes are staged in the git index; backlog/backlog.md has an unrelated pre-existing unstaged change.

Implementation decisions: flag helper follows the signInAvailability/landingAvailability pattern (featureAvailability.ts parser + calendarAutofillAvailability.ts build-time constant + utils barrel re-export); Calendar options visibility was extracted to calendarOptionsVisibility.ts (`isCalendarOptionsVisible`) for unit testability; MarkAvailabilityDialog stays mounted but is never opened when disabled; the landing-page animation and Settings calendar access are intentionally ungated; local ignored .env.development and .env.test now carry VITE_ENABLE_CALENDAR_AUTOFILL=true, and compose.yaml requires the variable via `:?` interpolation.

Bug fix later on 2026-09-20 after a user report that VITE_ENABLE_CALENDAR_AUTOFILL=false in .env.development still showed the Calendar options button on the timed event response editing page.

Root cause: useScheduleOverlapUI returned a placeholder `showCalendarOptions = computed(() => true)`, and useScheduleOverlapViewModels spreads `input.ui` after `input.derived`, so the placeholder overrode the caller's gated computed in both the sidebar and mobile overlay view models; this also defeated the pre-existing calendarPermissionGranted / addingAvailabilityAsGuest / userHasResponded gating.

Fix: removed the placeholder from useScheduleOverlapUI (definition and return object), making the ScheduleOverlap `isCalendarOptionsVisible` computed the single source for `showCalendarOptions`.

Regression coverage: new ScheduleOverlap.calendarOptionsGating.test.ts mounts ScheduleOverlap with the flag mocked false and asserts the sidebar and mobile overlay view models report `showCalendarOptions` false; the test was verified to fail before the fix and pass after.

Evidence: live dev-server check on 4173 with the flag false shows 0 `.calendar-options-button` elements and 0 "Calendar options" texts after Add availability; frontend lint (0 errors), fmt:check, typecheck, build, and test:unit (155 files / 1211 tests) pass; graphify update . was run.

Focused E2E on 2026-09-20 (specs event-page-no-responses-layout, event-mobile-editing-options, timed-event-group-firefox across chromium-desktop, chromium-mobile, firefox-desktop): 6 passed, 5 skipped, 2 failed. Both failures were `.calendar-options-button` missing: chromium-desktop "timed add availability controls stay close to the Legend" (anonymous viewer, calendarPermissionGranted false) and chromium-mobile "mobile editing with responses keeps Show best times..." (browser owns a guest response, userHasResponded true, non-group). Both were independent of the flag because .env.test sets VITE_ENABLE_CALENDAR_AUTOFILL=true. Root cause: removing the useScheduleOverlapUI placeholder restored not only the flag gate but also the dormant calendarPermissionGranted / userHasResponded / addingAvailabilityAsGuest conditions that the placeholder had overridden since the Vue3 migration; the gate and placeholder are both from 7675dcf1, and the failing e2e assertions were added 2026-08-29 onward.

Decision (user-approved): VITE_ENABLE_CALENDAR_AUTOFILL directly regulates the Calendar options button. Enabled or unset keeps the always-visible editing behavior observed on main; only the flag hides the control. Follow-up implementation: ScheduleOverlap.vue now computes `showCalendarOptions = computed(() => calendarAutofillEnabled)`; calendarOptionsVisibility.ts and its test were deleted; ScheduleOverlap.calendarOptionsGating.test.ts now asserts hidden when the flag is off and visible when the flag is on without calendar permission (default props). No e2e spec changes were needed.

Final verification on 2026-09-20: focused e2e rerun of event-page-no-responses-layout and event-mobile-editing-options on chromium-desktop and chromium-mobile: 5 passed, 5 skipped, 0 failed (both former failures pass). Frontend: lint 0 errors (2 pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, build, test:unit 154 files / 1205 tests; root npm run fmt:check passes; graphify update . run. DoD #3 and DoD #7 are now satisfied.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Gate calendar autofill behind VITE_ENABLE_CALENDAR_AUTOFILL.

What changed:
- Added a frontend build-time flag `VITE_ENABLE_CALENDAR_AUTOFILL` (defaults enabled when unset or blank). Plumbing: `featureAvailability.ts` parser + `calendarAutofillAvailability.ts` build-time constant re-exported from the utils barrel, `frontend/env.d.ts`, the frontend Docker build, `compose.yaml` build args, and all four tracked `.env.*.example` files; local ignored `.env.development`/`.env.test` carry it locally and `.env.test` keeps it true for browser E2E.
- When disabled, Add/Edit availability skips the availability-source choice dialog and starts manual Availability Editing, and no event-page path fills availability from calendar data: add-availability autofill, post-loader refill, reanimate-on-refresh, group invitation auto-open, and the Apple link redirect. The Calendar options control is not rendered during Availability Editing.
- Calendar options visibility is regulated solely by the flag. The previous fix had removed the `useScheduleOverlapUI` placeholder and, as a side effect, restored dormant `calendarPermissionGranted` / `userHasResponded` / `addingAvailabilityAsGuest` conditions, which focused E2E caught hiding the button for anonymous and already-responded viewers. Per the approved decision, `ScheduleOverlap.vue` now computes `showCalendarOptions` from the flag alone, so enabled/unset keeps the always-visible behavior observed on main and only the flag hides the control. The `calendarOptionsVisibility` helper and its test were deleted, and `ScheduleOverlap.calendarOptionsGating.test.ts` now covers hidden-when-off and visible-when-on-without-calendar-permission.
- Docs: FR-131 records the disabled behavior, the requirements README indexes it, and `docs/environments.md` documents the flag semantics.

Tests and checks:
- Unit coverage for flag parsing, `useEventEditing` flag-off behavior (no dialog, manual start, inert autofill), `useEventLoader` refill gating, `Event.vue` linkApple/group-invitation gating, and `ScheduleOverlap` calendar-options gating.
- Frontend `lint` (0 errors; 2 pre-existing NewSignUp.test.ts warnings), `fmt:check`, `typecheck`, `build`, `test:unit` (154 files / 1205 tests), root `npm run fmt:check`, and `graphify update .` pass.
- Focused e2e (user-requested): `event-page-no-responses-layout`, `event-mobile-editing-options`, and `timed-event-group-firefox`; final rerun 5 passed / 5 skipped / 0 failed, with the choice-dialog and group-invitation journeys green.

Enabled behavior, calendar accounts/overlays, Settings calendar access, and the landing page remain unchanged.
<!-- SECTION:FINAL_SUMMARY:END -->
