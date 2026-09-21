---
id: TASK-0305
title: Make the scheduling interaction hint provider-neutral (FR-133)
status: Done
assignee: []
created_date: '2026-09-21 16:21'
updated_date: '2026-09-21 16:24'
labels: []
dependencies: []
references:
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts
  - docs/requirements/functional/fr/FR-133.md
  - frontend/src/views/Event.vue
modified_files:
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts
  - docs/requirements/functional/fr/FR-133.md
priority: low
type: bug
ordinal: 305000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The scheduling-state interaction hint on the event page reads `...to schedule a Google Calendar event...`, but scheduling is not Google-specific: the Schedule menu offers Timeful, Google Calendar, and Outlook as destinations. The hint misstates what the drag does and names the wrong destination for Timeful and Outlook scheduling. The already-approved replacement wording, chosen by the user, is `schedule the event` (for example, `Click and drag on the grid below to schedule the event during those times.`), keeping the rest of the FR-133 instruction shape unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 During scheduling of a timed event, the desktop instruction reads `Click and drag on the grid below to schedule the event during those times.` and the phone viewport instruction uses `Tap and drag` with the same remainder.
- [x] #2 During scheduling of a dates-only event, the desktop instruction reads `Click and drag on the grid below to schedule the event during those days.` and the phone viewport instruction uses `Tap and drag` with the same remainder.
- [x] #3 No scheduling interaction instruction names Google Calendar or any other single scheduling destination.
- [x] #4 FR-133 acceptance criteria are updated to the new scheduling instruction strings.
- [x] #5 Unit tests cover the timed and dates-only scheduling instructions with the new strings, and no test asserts the old Google Calendar wording.
- [x] #6 Availability Editing instructions are unchanged.
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
1. Replace `schedule a Google Calendar event` with `schedule the event` in the SCHEDULE_EVENT branch of `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts`; the `Tap`/`Click` verb and `times`/`days` selection stay unchanged.
2. Update the two scheduling hint tests in `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts` to the new strings and add phone-viewport assertions (`Tap and drag`) for both timed and dates-only scheduling, since AC 1-2 cover phone and desktop.
3. Update the FR-133 acceptance criteria for timed and dates-only scheduling in `docs/requirements/functional/fr/FR-133.md` to the new strings.
4. Skip e2e per user instruction; no repo e2e spec asserts the scheduling hint (the only related spec asserts the availability hint).
5. Run frontend lint, fmt:check, typecheck, build, and test:unit; run root `npm run format:markdown` because a Markdown file changed; run `graphify update .` because code changed.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Replaced only the SCHEDULE_EVENT branch string; `verb` and `daysOrTimes` composition unchanged, so phone/desktop and timed/dates-only variants follow from the existing shared logic.

Extended the two existing scheduling hint tests with phone assertions rather than adding a separate phone test, keeping the suite's `isPhone` mutation pattern.

Added a destination-neutrality test looping over the Schedule menu destinations to guard against reintroducing provider-specific wording.

Verification: vitest `useScheduleOverlapUI.test.ts` 19/19 passed; full `npm run test:unit` 1254/1254 passed; lint, fmt:check, typecheck, and build all passed.

e2e skipped by explicit user instruction; grep confirms no remaining `Google Calendar event during` wording in source, docs, or e2e.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replaced the provider-specific scheduling hint in `useScheduleOverlapUI.ts` (`schedule a Google Calendar event` -> `schedule the event`), so the instruction now matches every scheduling destination the Schedule menu offers (Timeful, Google Calendar, Outlook). The desktop/phone verb and times/days selection are unchanged, and availability-editing hints are untouched.

Tests:
- `useScheduleOverlapUI.test.ts`: updated the timed and dates-only scheduling assertions to the new strings, added phone-viewport (`Tap and drag`) assertions for both, and added a regression test asserting the hint names no scheduling destination (`Google Calendar`, `Outlook`, `Timeful`).
- FR-133 acceptance criteria updated to the new timed and dates-only instruction strings.

Checks run from `frontend/`: `npm run lint` (0 errors; 2 pre-existing `vue/one-component-per-file` warnings in `NewSignUp.test.ts`, unrelated), `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit` (157 files, 1254 tests passed). Also ran root `npm run format:markdown` and `graphify update .` (no topology changes).

e2e was skipped at the user's explicit request; no repo e2e spec asserts the scheduling hint (the related Firefox spec only asserts the availability hint).
<!-- SECTION:FINAL_SUMMARY:END -->
