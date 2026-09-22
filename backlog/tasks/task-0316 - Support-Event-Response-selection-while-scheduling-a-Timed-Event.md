---
id: TASK-0316
title: Support Event Response selection while scheduling a Timed Event
status: Done
assignee:
  - opencode
created_date: '2026-09-21 22:45'
updated_date: '2026-09-22 11:42'
labels: []
dependencies: []
references:
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapSidebar.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapMobileOverlay.vue
  - frontend/src/composables/schedule_overlap/useEventScheduling.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts
  - frontend/src/views/Event.vue
  - frontend/src/composables/schedule_overlap/useEventScheduling.test.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapRendering.test.ts
  - e2e/specs/timed-event-scheduling-response-selection-firefox.spec.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-066.md
  - docs/requirements/functional/fr/FR-016.md
  - docs/requirements/functional/fr/FR-135.md
  - docs/requirements/functional/fr/FR-136.md
  - docs/requirements/README.md
  - docs/terminology/glossary.md
  - docs/requirements/functional/fr/FR-140.md
priority: medium
type: feature
ordinal: 316000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Trigger: from the Timed Event Page, the owner clicks Schedule event and then clicks an Event Response; the page currently leaves the Timed Event Scheduling Page and returns to the default timed event page, discarding the scheduling interaction. Response selection is today wired only to availability browsing, so the owner cannot focus the grid on particular respondents while choosing a Timed Event Occurrence Span. The same owner can already narrow the grid to selected Event Responses while editing availability, so scheduling should honor the same selection instead of forcing the owner to judge overlap against every response. Scheduling-time selection is a view filter only: calendar invite recipient behavior stays unchanged, and Blind Availability Mode visibility rules still apply. Selection lifetime follows user intent: cancelling scheduling is an abort, so an accidental Schedule event click preserves the selected responses, while confirming a schedule or clearing the saved schedule is a commit or destructive action that resets the selection and restores the full response set.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In Timed Event scheduling mode, selecting an Event Response does not leave scheduling mode: the Schedule, Cancel, and Clear controls remain available and the pending Timed Event Occurrence Span selection is retained.
- [x] #2 Entering scheduling mode through Schedule event preserves an existing Event Response selection made on the timed event page, and that selection filters the grid while scheduling.
- [x] #3 With one or more Event Responses selected while scheduling, the Timed Grid renders availability and calculates Schedule Overlap only for the selected responses.
- [x] #4 Selecting or deselecting responses during scheduling updates the filtered grid immediately.
- [x] #5 Clearing the last selected response restores every response's availability and overlap calculation while staying in scheduling mode.
- [x] #6 Confirming a schedule through Timeful, Google Calendar, or Outlook exits scheduling mode, clears the response selection, and returns to the timed event page showing every response's availability.
- [x] #7 Cancelling scheduling exits scheduling mode but preserves the Event Response selection; the timed event page keeps rendering the selected subset with the selected response rows still checked.
- [x] #8 Clearing the saved schedule exits scheduling mode, clears the response selection, and returns to the timed event page showing every response's availability.
- [x] #9 The behavior matches on desktop and phone viewports, through both the desktop sidebar and the phone respondents overlay.
- [x] #10 Selection during scheduling exposes no responses that the viewer could not already see under Blind Availability Mode.
- [x] #11 Existing scheduling behavior is unchanged: a schedule drag starts and ends only on Active Slots, the pending-span tooltip behavior is preserved, and the selected-slot cursor still works with responses selected.
- [x] #12 Unit tests cover the selection and scheduling-state interaction, the filtered Schedule Overlap while scheduling, and the selection lifetime across confirm, clear, and cancel exits.
- [x] #13 A browser e2e regression covers both scheduling exit paths: cancelling preserves the filtered subset, while confirming and clearing reset it.
- [x] #14 The canonical functional requirement for Event Response selection during scheduling and its lifetime across scheduling exits is recorded under docs/requirements, either as a new FR or an explicit extension of FR-066, with the README index updated.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Keep selected Event Responses (`curRespondents`) as an orthogonal filter and keep scheduling in the existing `states.SCHEDULE_EVENT`; no new state value.
2. `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts`:
   - `clickRespondent`: when the state is `SCHEDULE_EVENT`, toggle the selection without changing state (today it always assigns `SUBSET_AVAILABILITY`, which is the reported exit); keep current behavior in all other states, including returning to `defaultState` when the last response is cleared outside scheduling.
   - Add `exitScheduling("commit" | "abort")`: abort preserves the selection and lands on `SUBSET_AVAILABILITY` when a selection exists, else `defaultState`; commit lands on `defaultState`, clears the selection, and both reset the timeslot state.
   - `deselectRespondents`: ignore clicks inside a `.schedule-event-control` so scheduling controls do not count as a deselect gesture.
3. `frontend/src/composables/schedule_overlap/useEventScheduling.ts`: add an optional `onSchedulingExit` callback; call it with `abort` from `cancelScheduleEvent`, and with `commit` after a successful Timeful confirm, Google/Outlook hand-off, and `clearScheduledEvent`. Without the callback, keep setting `defaultState` so existing unit tests and callers are unaffected.
4. `frontend/src/components/schedule_overlap/ScheduleOverlap.vue`: `ui` is created after `eventSched` (drag depends on `setScheduledEventFromRowCol`), so bind the hook through a small late-bound handler that is replaced by `ui.exitScheduling` right after `ui` is created.
5. `frontend/src/views/Event.vue`: add `schedule-event-control` to the desktop and mobile Schedule event/Reschedule, Cancel, Clear, Schedule activator buttons, and destination menu items.
6. `frontend/src/components/schedule_overlap/scheduleOverlapRendering.ts`: in `getBaseTimeslotClassStyle`, treat a non-empty selection during `SCHEDULE_EVENT` as a subset view (same counts, total, and single-respondent if-needed lookup as `SUBSET_AVAILABILITY`) so both timed and days-only grids filter while scheduling.
7. Tests: unit coverage in `useScheduleOverlapUI.test.ts` (scheduling toggle, hint/state retention, exit outcomes, control exemption), `useEventScheduling.test.ts` (abort/commit hooks and fallback), `scheduleOverlapRendering.test.ts` (subset fill while scheduling); new Firefox e2e spec covering filter-while-scheduling, cancel-preserves, confirm-resets, and clear-resets.
8. Docs: add the canonical FR for scheduling-time Event Response selection and its lifetime, plus the README index row.
9. Checks: frontend lint, fmt:check, typecheck, build, unit tests; focused Firefox e2e; refresh the codebase-memory index.

Risk notes: keep non-scheduling subset rendering identical; the control exemption must not change click-outside deselection anywhere else; the exit hook must not clear selection on failed confirm/clear.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation summary (all code paths complete): `useScheduleOverlapUI.ts` makes `clickRespondent` scheduling-aware (toggles the filter without leaving `SCHEDULE_EVENT`, and clearing the last selection stays in scheduling), adds `exitScheduling("commit" | "abort")`, and exempts `.schedule-event-control` clicks from `deselectRespondents`. `useEventScheduling.ts` adds the optional `onSchedulingExit` hook and calls it with `abort` on cancel and `commit` after a successful Timeful confirm, Google/Outlook hand-off, and `clearScheduledEvent`, keeping the old `defaultState` fallback for callers and tests without the hook. `ScheduleOverlap.vue` late-binds `schedulingExitHandler` to `ui.exitScheduling` because `ui` is created after `eventSched`. `Event.vue` marks the desktop and mobile Schedule event/Reschedule, Cancel, Clear, Schedule activator, and destination menu items with `schedule-event-control`. `scheduleOverlapRendering.ts` treats a non-empty selection during `SCHEDULE_EVENT` as a subset view via `isSubsetView`. Added FR-140 plus its README index row. New spec `e2e/specs/timed-event-scheduling-response-selection-firefox.spec.ts` covers filter-while-scheduling, cancel-preserves, confirm-resets, and clear-resets; phone coverage was added to `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` as `phone scheduling keeps selected responses and their filtered grid through cancel`.

Verification evidence: unit tests `useScheduleOverlapUI.test.ts`, `useEventScheduling.test.ts`, `scheduleOverlapRendering.test.ts` pass (81 tests), and the full frontend suite passes (158 files, 1293 tests). `npm run lint` (0 errors), `npm run fmt:check`, `npm run typecheck`, and `npm run build` pass in `frontend/`. Focused Firefox desktop e2e passes: the new spec 4/4, and the new spec plus `timed-event-respondent-selection-firefox.spec.ts` and `timed-event-scheduling-tooltip-firefox.spec.ts` 9/9. `e2e` lint and typecheck pass. Root `npm run format:markdown:check` passes, and `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph (8859 nodes, 28057 edges).

Remaining work and blockers: phone browser verification for AC #9 is pending; the new firefox-touch test was added but its runs were aborted at user request, so it is unverified. The full Firefox desktop suite was likewise skipped at user request. An aborted run left the isolated test stack up (Vite on port 4174, `timeful-test-postgres-test-1`, `timeful-test-calendar-mock-1`), so kill the Vite node process and run `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down` before the next `npm run test:e2e` to avoid the port conflict. Task remains In Progress with acceptance criteria unchecked; finalization (AC checkoff, final summary, Done) is pending. No server or contract changes, so the swagger generation DoD item does not apply.

Review follow-up (2026-09-22): fixed the stray bare backtick on the desktop Schedule/Reschedule button in `Event.vue`, which the SFC compiler emitted as a `` `` `` prop that Vuetify forwarded to the DOM and browsers reject with InvalidCharacterError. Reworded the late-binding comment in `ScheduleOverlap.vue`. Added Google/Outlook hand-off commit unit tests with a stubbed `window.open`. Verified the actor/blind-mode question: the server sets `canEditSettings` only for owner visitors (`server/routes/event_routes.go`), so scheduling selection is owner-only and FR-140's Event Owner actor is accurate; Blind Availability Mode restricts non-owners only, so it does not change owner scheduling selection.

Re-verification (2026-09-22): frontend `lint`, `fmt:check`, `typecheck`, and `build` pass; the full unit suite passes (158 files, 1295 tests). Firefox desktop e2e passes 9/9 (the new spec 4/4 plus `timed-event-respondent-selection-firefox.spec.ts` and `timed-event-scheduling-tooltip-firefox.spec.ts`); the firefox-touch phone test `phone scheduling keeps selected responses and their filtered grid through cancel` passes 1/1. The e2e runs tore down the isolated stack, so the earlier stale port and leftover-container note no longer applies.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-21 22:55
---
Selection lifetime decision (middle ground): cancelling scheduling preserves the Event Response selection, while confirming a schedule or clearing the saved schedule resets it. Rationale: cancelling is an abort, so an accidental Schedule event click must not destroy a selection; confirming is a commit and clearing is destructive, so both return the page to the full response set. This also implies entry preservation: clicking Schedule event must not clear an existing selection, otherwise there is nothing to preserve on cancel.
---

created: 2026-09-21 22:55
---
Alternatives considered: resetting the selection on every scheduling exit (simpler, consistent with the current click-outside-clears rule, but destroys deliberate selections) and always transferring the selection (continuity, but it fights the click-outside-clears rule and, unless the exit also enters subset view, produces checked rows with an unfiltered grid because rendering only applies subset filtering in SUBSET_AVAILABILITY state).
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Problem

On the Timed Event Page, clicking Schedule event and then an Event Response left the Timed Event Scheduling Page and returned to the default timed event page, discarding the pending Timed Event Occurrence Span. Response selection was wired only to availability browsing, so the owner could not focus the grid on particular responses while choosing a span.

## Fix

Scheduling now treats the selected Event Responses as an orthogonal view filter. `useScheduleOverlapUI.ts` keeps `clickRespondent` in `SCHEDULE_EVENT` while toggling the selection, adds `exitScheduling("commit" | "abort")` (abort preserves the selection and lands in the subset view; commit clears it), and exempts `.schedule-event-control` clicks from click-outside deselection. `useEventScheduling.ts` reports `abort` on cancel and `commit` after a successful Timeful confirm, a Google or Outlook hand-off, and clearing the saved Timed Event Occurrence Span, with the old default-state fallback intact. `ScheduleOverlap.vue` late-binds the exit hook because `ui` is created after `eventSched`, `Event.vue` marks all desktop and mobile schedule controls, and `scheduleOverlapRendering.ts` renders a non-empty scheduling selection through the subset path. FR-140 records the behavior and the README index row was added.

## Tests

- Unit: scheduling toggle and exit outcomes plus control exemption, abort/commit hooks including Google and Outlook hand-off commits via a stubbed `window.open`, and subset fill while scheduling. Full frontend suite: 158 files, 1295 tests pass.
- Firefox desktop e2e: new `timed-event-scheduling-response-selection-firefox.spec.ts` 4/4, and 9/9 together with `timed-event-respondent-selection-firefox.spec.ts` and `timed-event-scheduling-tooltip-firefox.spec.ts`.
- Firefox touch e2e: `phone scheduling keeps selected responses and their filtered grid through cancel` passes 1/1.
- Required frontend checks lint, fmt:check, typecheck, and build pass; root `format:markdown:check` passes; the code graph was refreshed.
- Review follow-up removed a stray bare-backtick attribute on the desktop Schedule/Reschedule button that browsers reject with InvalidCharacterError, and reworded the late-binding comment.

## Risks / notes

- No server, API, Swagger, or contract changes; calendar invite recipients and Blind Availability Mode visibility are unchanged. Scheduling is gated by owner-only `canEditSettings`, so the FR-140 Event Owner actor is accurate.
- The full Chromium/Firefox E2E matrix and backend tests were not run locally; CI is the backstop.
<!-- SECTION:FINAL_SUMMARY:END -->
