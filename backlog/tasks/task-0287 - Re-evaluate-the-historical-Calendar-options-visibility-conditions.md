---
id: TASK-0287
title: Re-evaluate the historical Calendar options visibility conditions
status: To Do
assignee: []
created_date: '2026-09-20 17:35'
labels: []
dependencies: []
references:
  - >-
    backlog/tasks/task-0286 -
    Gate-calendar-autofill-behind-VITE_ENABLE_CALENDAR_AUTOFILL.md
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/useScheduleOverlapViewModels.ts
  - frontend/src/composables/event/useEventEditing.ts
  - frontend/src/composables/event/useEventLoader.ts
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-131.md
  - docs/environments.md
priority: medium
type: spike
ordinal: 287000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Vue3 migration (7675dcf1) made the historical Calendar options visibility conditions inert: useScheduleOverlapUI returned a placeholder `showCalendarOptions` that useScheduleOverlapViewModels spread over the real computed, so the sidebar and mobile overlay view models always reported the control visible. TASK-0286 removed the placeholder, and by decision made VITE_ENABLE_CALENDAR_AUTOFILL the sole regulator, preserving the always-visible behavior that had effectively shipped since the migration. The dormant conditions encoded product intent: show calendar autofill settings only where calendar-driven filling is meaningful, meaning not while adding availability as a guest, only with calendar access (calendarPermissionGranted), and for non-group events only before the viewer's first response. With the flag enabled, an anonymous viewer or a viewer who already responded to a non-group event can now reach calendar autofill settings that never apply to their editing session. Re-evaluate each condition against the current autofill paths and decide, per condition, whether it should be reinstated, replaced, or dropped permanently.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each condition is individually decided with a rationale tied to current autofill behavior: calendarPermissionGranted, userHasResponded (including the group exception), and addingAvailabilityAsGuest
- [ ] #2 If a decision changes enabled behavior, the implementation ships with unit and e2e coverage for the affected viewer states, including an anonymous viewer, an already-responded non-group viewer, a group viewer, and a guest adding availability
- [ ] #3 If a decision changes enabled behavior, FR-131 or a new functional requirement records the Calendar options visibility rules, and docs/environments.md stays accurate about what VITE_ENABLE_CALENDAR_AUTOFILL controls
- [ ] #4 If the flag-only behavior is kept, the rationale is recorded on this task and no runtime behavior changes
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
