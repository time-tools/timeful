---
id: TASK-0295
title: Move the Scheduled event Legend entry directly below Unavailable
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 22:02'
updated_date: '2026-09-20 22:04'
labels: []
dependencies: []
modified_files:
  - frontend/src/components/schedule_overlap/ColorLegend.vue
  - frontend/src/components/schedule_overlap/ColorLegend.test.ts
priority: medium
type: enhancement
ordinal: 295000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In the schedule-overlap Legend (ColorLegend), the "Scheduled event" indicator currently renders after both structural "Disabled" entries, which buries the event-level indicator below unrelated structural states. The user wants "Scheduled event" grouped with the availability palette, directly below "Unavailable, change in Add/Edit availability", so the Legend reads from availability states into the scheduled-event indicator before the disabled structural entries.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In the schedule-overlap ColorLegend, the "Scheduled event" entry renders immediately below the "Unavailable, change in Add/Edit availability" entry whenever that entry is shown.
- [x] #2 When the "Unavailable, change in Add/Edit availability" entry is hidden, "Scheduled event" still renders before the "Disabled, inside the event dates in the event timezone" entry, with the hidden conditional entries collapsed in place.
- [x] #3 ColorLegend unit tests assert the new label order for the no-active-slots, active-slots, and response-palette cases.
- [x] #4 No indicator styling, indicator classes, or visibility conditions change; only the order of Legend entries changes.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implementation plan (researched against the current system):

1. `frontend/src/components/schedule_overlap/ColorLegend.vue`: move the unconditional Scheduled event entry to immediately after the conditional Unavailable entry, leaving every class, indicator, and visibility condition unchanged.
2. `frontend/src/components/schedule_overlap/ColorLegend.test.ts`: update label-order expectations for the no-active-slots, active-slots, and response-palette cases so Scheduled event is asserted directly below Unavailable (or after the availability palette when Unavailable is hidden).
3. Check for other consumers/tests that assert Legend entry order (ScheduleOverlapSidebar.test.ts, e2e specs); update only if they assert this order.
4. Run the required frontend checks: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`; run `graphify update .`; record evidence and finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: moved the Scheduled event entry in ColorLegend.vue to sit directly after the conditional Unavailable entry; no indicator classes, fills, labels, or visibility conditions changed.

Verification: focused ColorLegend.test.ts 9/9; full test:unit 154 files / 1233 tests; lint 0 errors (2 pre-existing warnings in untouched NewSignUp.test.ts); fmt:check, typecheck, and build pass; git diff limited to ColorLegend.vue and ColorLegend.test.ts; `graphify update .` run (no topology changes).

E2E decision: no e2e spec asserts Legend entry order, and an order-only DOM move cannot change the Legend bounding box that legend-touching layout specs measure, so DoD #3 is left unchecked per the TASK-0268 precedent; the frontend required check set excludes e2e.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved the Scheduled event entry directly below the Unavailable entry in the schedule-overlap Legend, per request.

Why: the event-level indicator previously sat after both structural "Disabled" entries, separated from the availability palette and easy to miss among structural states.

Changes:
- frontend/src/components/schedule_overlap/ColorLegend.vue: relocated the unconditional Scheduled event block to immediately after the conditional Unavailable block. Every class, indicator, and visibility condition is unchanged, and the Disabled inside/outside/collapsed entries follow unchanged.
- frontend/src/components/schedule_overlap/ColorLegend.test.ts: replaced the length-only response-palette assertions with full label-order assertions and updated the no-active-slots and active-slots expectations to assert Available/If needed, Unavailable, Scheduled event, Disabled inside, Disabled outside.

Evidence: focused ColorLegend.test.ts 9/9 passed; full test:unit 154 files / 1233 tests passed; npm run lint 0 errors (2 pre-existing NewSignUp.test.ts warnings in an untouched file); fmt:check, typecheck, and build passed; git diff is limited to the two component files; `graphify update .` completed with no topology changes.

E2E not run: no e2e spec asserts Legend entry order, and an order-only DOM move cannot affect the Legend bounding-box measurements used by the legend-touching layout specs; the required frontend check set (frontend/AGENTS.md) does not include e2e. DoD #3 left unchecked for that reason.
<!-- SECTION:FINAL_SUMMARY:END -->
