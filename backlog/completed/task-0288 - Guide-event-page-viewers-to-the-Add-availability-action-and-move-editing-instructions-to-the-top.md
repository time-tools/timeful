---
id: TASK-0288
title: >-
  Guide event-page viewers to the Add availability action and move editing
  instructions to the top
status: Done
assignee:
  - opencode
created_date: '2026-09-20 18:08'
updated_date: '2026-09-20 19:35'
labels: []
dependencies: []
references:
  - backlog/backlog.md
  - frontend/src/views/Event.vue
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
documentation:
  - docs/requirements/functional/fr/FR-041.md
  - docs/requirements/functional/fr/FR-112.md
  - docs/requirements/README.md
modified_files:
  - docs/requirements/README.md
  - docs/requirements/functional/fr/FR-132.md
  - docs/requirements/functional/fr/FR-133.md
  - e2e/helpers/availability-hint-helpers.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
  - e2e/specs/timed-event-add-availability-hint-firefox.spec.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/composables/event/types.ts
  - frontend/src/composables/event/useEventEditing.test.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/views/Event.test.ts
  - frontend/src/views/Event.vue
priority: medium
type: enhancement
ordinal: 288000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A non-technical person who receives a shared timed event link and opens it on a phone does not discover the Add availability action in the fixed bottom action bar. Before Availability Editing starts there is no guidance at all, and the existing Click/Tap and drag instruction only appears after editing begins, rendered at the bottom of the grid or inside the mobile overlay, where it is not the first thing read.

Product decisions confirmed with the user: show a text-only hint at the top of the event page while the viewer has no own Event Response, covering all viewports, with no dismiss control because it self-clears once the viewer saves a response; move the existing Availability Editing instruction to the same top-of-page slot on all viewports while keeping its existing dismiss control and remembered dismissal. Approved copy: phone banner reads "Add availability (at the bottom of the screen) to show when you're available for this event." and desktop banner reads "Add availability (in the event header) to show when you're available for this event."; editing variants read "Click/Tap and drag on the grid below to add your "available" times in green." with the equivalent "if needed", group, and scheduling variants.

The intentionally disabled filled Edit availability action that appears when responses exist but the current viewer has nothing editable stays unchanged; the hint simply names the real Add availability action.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a plain (non-group, non-sign-up, non-archived) event page, a viewer who can add availability and has no editable Event Response sees a hint at the top of the page content that names the Add availability action and its location: "at the bottom of the screen" on a phone viewport and "in the event header" on a desktop viewport, including when other responses exist and the filled Edit availability action is disabled
- [x] #2 The pre-response hint disappears once the viewer owns an editable Event Response, and it never renders for archived events, group events, or sign-up forms
- [x] #3 During Availability Editing and scheduling on the event page, the existing Click/Tap and drag instruction renders at the top of the page content on every viewport, above the event header and grid, instead of in the desktop grid strip or the mobile bottom overlay
- [x] #4 The relocated instruction keeps its dismiss control and remembered dismissal per editing state, and its copy matches the approved strings, including "on the grid below", the available, if-needed, group, and scheduling variants, and "days" wording on dates-only events
- [x] #5 The pre-response hint has no dismiss control and is unaffected by the editing instruction's dismissal state
- [x] #6 New functional requirement records document both behaviors and are added to the requirements index with the changed Markdown formatted
- [x] #7 Unit tests cover the visibility matrix and days-aware copy for both hints, and e2e coverage asserts the banner, the in-app save clearing it, and the editing-instruction placement in a mobile and a desktop project
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
1. Requirements: create FR-132 (cue Add availability on the response creation page) and FR-133 (place Availability Editing instructions at the top of the event page) in docs/requirements/functional/fr/, status proposed, frontend component; add both rows to docs/requirements/README.md; format with npm run format:markdown.

2. frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts: update hintText copy to include "on the grid below" for the EDIT_AVAILABILITY, group EDIT_AVAILABILITY, and SCHEDULE_EVENT variants; verb already switches Tap/Click by isPhone.

3. frontend/src/components/schedule_overlap/ScheduleOverlap.vue and frontend/src/composables/event/types.ts: expose hintText, hintClosed, and closeHint on ScheduleOverlapInstance so Event.vue can render the instruction at page top while the component's own grid/mobile-overlay strips are suppressed.

4. frontend/src/views/Event.vue:
   - Add showAddAvailabilityHint (can add availability, not archived/group/sign-up, not editing/scheduling/specific-times, no editable response) and viewport copy.
   - Render a top hint region in the tw:max-w-5xl content column directly after the archived banner: an always-on text-only v-alert info/tonal for the creation state, and the dismissible editing instruction with its info icon and X (reusing closeHint).
   - Pass :show-hint-text="false" to ScheduleOverlap so the old grid strip and mobile overlay hint stop rendering on the event page.
   - Keep the intentionally disabled filled Edit availability behavior untouched.

5. Tests: Event.test.ts visibility matrix and copy for both hints, close wiring, and show-hint-text false; useScheduleOverlapUI.test.ts copy per state/viewport; extend e2e/specs/event-mobile-editing-options.spec.ts (chromium-mobile) and add e2e/specs/timed-event-add-availability-hint-firefox.spec.ts (firefox-desktop) asserting placement above the grid, copy, and banner auto-clear after a saved response.

6. Checks: frontend lint, fmt:check, typecheck, build, test:unit; focused e2e with streaming output; graphify update .; root Markdown format checks.

Risks: unit tests that stub ScheduleOverlap must gain the new exposed fields; removing the mobile overlay hint changes overlay height, so mobile legend clearance tests/e2e need a check.

Review follow-up (post-review fixes before commit):

1. FR-132 requirement wording: change the trigger from 'no Event Response' to 'no editable Event Response' and document that the hint also shows when responses exist but the viewer holds nothing editable (disabled filled Edit availability plus secondary Add availability).

2. FR-133 requirement coverage: include the Dates-Only Event Scheduling Page, link Timed Grid/Dates-Only Grid, and add days-only copy acceptance criteria.

3. useScheduleOverlapUI: add a daysOnly option and use 'days' instead of 'times' for the available and if-needed variants, and 'during those days' for scheduling, on dates-only events; remove the stale 'handled by caller via override' comment; ScheduleOverlap passes daysOnly.

4. Event.vue: gate the pre-response banner on scheduleOverlapReady to avoid a pre-mount flash for returning guest owners; drop the unnecessary :key on the editing-instruction alert.

5. Tests: useScheduleOverlapUI days-only copy cases; Event.test.ts banner case for responses with nothing editable; Firefox e2e in-app guest save clears the banner without a reload.

6. README index: link Event Response in the FR-132 row.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation kept the pre-response hint out of `canEditAvailability` (that computed is an ownership gate for sign-up forms, not a general add-availability gate); a plain-event visitor without an owned response sees the banner.

Deviation from the drafted plan: the availability-group instruction previously lowercased the verb after its leading sentence (`click and drag`); the approved copy calls for `Click and drag`, so the group variant now uses the capitalized verb.

E2E helper `expectHintAboveGrid` lives in `e2e/helpers/availability-hint-helpers.ts` and is shared by the Firefox and mobile specs.

The pre-response banner condition uses `hasEditableAvailability` so a viewer who already owns an editable response sees no hint even when other responses exist; the disabled filled Edit availability state is intentionally unchanged.

Checks: frontend lint/fmt:check/typecheck/build/test:unit pass (1217 unit tests); focused e2e passes (firefox-desktop new spec, chromium-mobile new tests, plus layout/touch regression specs); `graphify update .` and root markdown format checks pass. `e2e npm run fmt:check` still flags unmodified `specs/timed-event-access-transfer-firefox.spec.ts`, a pre-existing formatting issue outside this task.

Review found and this follow-up fixes: FR-133 omitted dates-only scheduling; FR-132 stated 'no Event Response' while the code gates on no editable response; dates-only hint copy hardcoded 'times' with a stale override comment; the e2e only proved banner clearing after an API post plus reload; the disabled-Edit-availability state had no hint assertion; minor cleanup of the redundant alert :key.

Review fixes verified: useScheduleOverlapUI hint tests cover timed and dates-only available/if-needed/scheduling copy; Event.test.ts adds the responses-without-editable-response banner case; the Firefox spec adds an in-app guest save test that asserts the banner clears without a reload (3 passed); chromium-mobile editing and chromium-desktop no-responses layout specs pass (7 passed, 7 expected skips). Full frontend unit suite: 154 files / 1221 tests pass; lint, fmt:check, typecheck, and build pass; e2e lint/typecheck pass; root format:markdown:check and lint:markdown pass; graphify update . completed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The event page now guides viewers to the `Add availability` action before they ever enter Availability Editing, and the relocated editing instruction is accurate on dates-only pages.

- `frontend/src/views/Event.vue` renders a top hint region inside the event content column, directly after the archived banner: a text-only `v-alert` info/tonal hint for viewers with no editable Event Response (viewport-specific wording, no dismiss control; clears once an editable response exists), plus the dismissible grid instruction relocated from the desktop grid strip and the mobile bottom overlay on every viewport. The banner waits for the ScheduleOverlap mount so returning guest owners do not see a flash.
- `ScheduleOverlap.vue` exposes `hintText`, `hintClosed`, and `closeHint` (added to `ScheduleOverlapInstance` in `frontend/src/composables/event/types.ts`), passes `daysOnly` to `useScheduleOverlapUI`, and receives `:show-hint-text="false"` so the component renders no duplicate strips on the event page. Hint support remains intact for other embedders.
- `useScheduleOverlapUI.ts` copy now says "on the grid below" for the available, if-needed, group, and scheduling variants, with "days" instead of "times" on dates-only events; the group variant's verb is capitalized after its leading sentence.
- Requirements: FR-132 (cue Add availability on Event Response creation pages) and FR-133 (place grid interaction instructions at the top of the event page) cover no-editable-response behavior, dates-only scheduling, and days-aware copy.

## Scope notes

- Banner visibility excludes archived events, availability groups, sign-up forms, Availability Editing, scheduling, specific-times editing, and viewers who already own an editable response.
- The intentionally disabled filled `Edit availability` primary is untouched; the hint names the real `Add availability` action, including when that action is the secondary button.
- On phones the relocated instruction is in-flow at the top of the content column, so a viewer who is scrolled deep into the grid may not have it in view; a sticky presentation was deliberately out of scope.

## Verification

- Unit: 154 files / 1221 tests pass, including EventView visibility-matrix, disabled-Edit-availability banner, copy, close-wiring, and hint-suppression tests and `useScheduleOverlapUI` timed and dates-only hintText tests.
- E2E: `timed-event-add-availability-hint-firefox.spec.ts` 3 passed on firefox-desktop, including dismissal persistence across reload and the banner clearing after an in-app guest save; `event-mobile-editing-options.spec.ts` plus `event-page-no-responses-layout.spec.ts` 7 passed with 7 expected skips across chromium-mobile and chromium-desktop.
- Frontend lint, fmt:check, typecheck, build, and test:unit pass; e2e lint and typecheck pass; root `format:markdown:check` and `lint:markdown` pass; `graphify update .` completed.
- Pre-existing, unrelated: `e2e npm run fmt:check` flags `specs/timed-event-access-transfer-firefox.spec.ts`, which this change does not modify. Swagger and contract docs are unaffected.
<!-- SECTION:FINAL_SUMMARY:END -->
