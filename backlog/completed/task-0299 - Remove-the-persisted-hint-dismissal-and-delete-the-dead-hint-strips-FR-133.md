---
id: TASK-0299
title: Remove the persisted hint dismissal and delete the dead hint strips (FR-133)
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 23:22'
updated_date: '2026-09-20 23:37'
labels:
  - frontend
  - event-page
dependencies:
  - TASK-0296
references:
  - docs/requirements/functional/fr/FR-133.md
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/views/Event.vue
  - frontend/src/composables/event/types.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapTimeGrid.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapDaysOnlyGrid.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapMobileOverlay.vue
  - frontend/src/components/schedule_overlap/useScheduleOverlapViewModels.ts
  - >-
    frontend/src/components/schedule_overlap/scheduleOverlapViewModelContracts.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapTestUtils.ts
  - frontend/src/components/landing/LandingPageCalendar.vue
  - e2e/specs/timed-event-add-availability-hint-firefox.spec.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-133.md
  - docs/requirements/README.md
  - frontend/AGENTS.md
  - e2e/AGENTS.md
  - docs/requirements/AGENTS.md
  - docs/AGENTS.md
modified_files:
  - docs/requirements/functional/fr/FR-133.md
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts
  - frontend/src/composables/event/types.ts
  - frontend/src/composables/event/useEventEditing.test.ts
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - frontend/src/components/schedule_overlap/ScheduleOverlap.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapTimeGrid.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapDaysOnlyGrid.vue
  - frontend/src/components/schedule_overlap/ScheduleOverlapMobileOverlay.vue
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlapMobileOverlay.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlapGridDragBinding.test.ts
  - >-
    frontend/src/components/schedule_overlap/ScheduleOverlap.childViewModels.test.ts
  - frontend/src/components/schedule_overlap/useScheduleOverlapViewModels.ts
  - >-
    frontend/src/components/schedule_overlap/scheduleOverlapViewModelContracts.ts
  - frontend/src/components/schedule_overlap/scheduleOverlapTestUtils.ts
  - frontend/src/components/landing/LandingPageCalendar.vue
  - e2e/specs/timed-event-add-availability-hint-firefox.spec.ts
  - e2e/specs/event-mobile-editing-options.spec.ts
priority: medium
type: bug
ordinal: 299000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
FR-133 (Place grid interaction instructions at the top of the event page) is implemented in commit b3b9ed80 (TASK-0288). That commit is on origin/fix-frontend, and it is not on main; landing it on main is out of scope for this task.

The implementation inherited the legacy dismissal behavior of the old bottom-of-grid / mobile-overlay hint: frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts reads and writes localStorage["closedHintText" + state], and Event.vue hides the top instruction when ScheduleOverlap reports it closed. A dismissal stored in an earlier session (including the pre-TASK-0288 strip, which used the same key) permanently suppresses the relocated instruction. A user entering Availability Editing on a browser that has closedHintTextedit_availability stored sees no instruction at all, while the FR-132 Add availability banner still renders because it has its own independent condition.

Product decisions confirmed with the user:

- The grid instruction is not dismissible and has no dismissal state. It renders for as long as the applicable editing or scheduling state is active and disappears when that state ends (save or cancel). It reappears the next time that state starts.
- Browser-stored dismissal state from an earlier version must never suppress the instruction.
- On a phone viewport, entering Availability Editing must bring the instruction into view: the action that starts editing lives in the fixed bottom action bar while the instruction is at the top of the page content, so a phone user otherwise never sees it.
- The pre-response Add availability banner (FR-132) is unchanged: non-dismissible and self-clearing once the viewer owns an editable Event Response.

The embedded hint strips are dead code and are removed in this task:

- ScheduleOverlapTimeGrid.vue, ScheduleOverlapDaysOnlyGrid.vue, and ScheduleOverlapMobileOverlay.vue render sticky hint strips with an X, but the only two ScheduleOverlap embedders (Event.vue and LandingPageCalendar.vue) pass show-hint-text="false", so they can never render.
- The whole machinery goes: the showHintText prop, hintTextShown, hintState, hintStateLocalStorageKey, hintClosed, closeHint, the strip templates, the view-model fields and actions, and their test utilities/tests. ScheduleOverlapInstance keeps only hintText as the single source of the instruction copy.

Instruction copy stays exactly as FR-133's acceptance criteria already state; no copy changes.

Working-tree constraints:

- TASK-0296 (Done) is already committed as 77273260 ("fix(frontend): make the mobile Add availability action solid green like desktop"); its Event.vue, Event.test.ts, and e2e/specs/event-mobile-editing-options.spec.ts changes are in that commit and their working tree is clean. No baseline commit is needed; before starting, verify git status shows only the staged TASK-0299 record and the dirty backlog/backlog.md.
- Leave the dirty backlog/backlog.md untouched and out of every commit.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 During Availability Editing and scheduling on the event page, the instruction renders at the top of the page content, above the event header and the grid, on every viewport, with the copy FR-133 lists for the available, if-needed, group, and timed/dates-only scheduling variants and Tap on phone / Click on desktop
- [x] #2 The instruction renders without any dismiss control, and Event.vue no longer consults dismissal state from ScheduleOverlap
- [x] #3 No code reads or writes closedHintText* localStorage keys, and a pre-seeded closedHintTextedit_availability key does not prevent the instruction from rendering
- [x] #4 The instruction disappears when Availability Editing or scheduling ends by save or cancel, and reappears the next time that state starts, including after a page reload
- [x] #5 On a phone viewport, entering Availability Editing scrolls the instruction into view
- [x] #6 The dead hint machinery is fully removed: the showHintText prop, hintTextShown, hintState, hintStateLocalStorageKey, hintClosed, closeHint, the three strip templates, the view-model fields/actions, and their utility and component-test references; ScheduleOverlapInstance keeps only hintText
- [x] #7 FR-133 records the new behavior (not dismissible, active-state-only, immune to stored browser state, phone in view) with updated acceptance criteria, is formatted, and keeps its existing README index row and proposed status
- [x] #8 Unit tests cover strip visibility during editing, absence outside it, no close control, rendering despite a stored closedHintTextedit_availability key, phone scroll into view, and the existing per-state/viewport copy
- [x] #9 E2E proves on firefox-desktop that there is no Close button, the strip is visible during editing, disappears after cancel/save, and renders again on the next editing session even with a pre-seeded legacy key; and on chromium-mobile that the strip is in the viewport after entering editing
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
## 0. Baseline and separation

1. Read `BACKLOG_WORKFLOW.md`, `frontend/AGENTS.md`, `e2e/AGENTS.md`, `docs/requirements/AGENTS.md`, and `docs/AGENTS.md` before editing.
2. Confirm the baseline: `git log --oneline -1 77273260` shows the committed TASK-0296 work, and `git status` shows only the staged TASK-0299 record plus the dirty `backlog/backlog.md`. No TASK-0296 baseline commit is needed; do not create one.
3. Mark TASK-0299 In Progress and assign it before implementation.

## 1. Requirement first: docs/requirements/functional/fr/FR-133.md

1. Replace the sentence "The instruction shall remain dismissible, and a dismissal shall persist for the applicable editing state in the browser." with "The instruction shall not be dismissible, and it shall render only while the applicable editing or scheduling state is active."
2. Add a sentence that browser-stored dismissal state from an earlier application version shall not suppress the instruction.
3. Add a sentence that on a phone viewport, entering Availability Editing shall bring the instruction into view.
4. Extend the acceptance criteria with: no dismiss control; absent when neither Availability Editing nor scheduling is active; renders even when the legacy `closedHintTextedit_availability` key is present; phone viewport brings the instruction into view when Availability Editing starts.
5. Follow `docs/AGENTS.md` (one sentence per physical line, canonical glossary terms linked on first use per paragraph). Run root `npm run format:markdown`.

## 2. Remove localStorage persistence and the event-page dismissal

1. `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts`
   - Keep `hintText` (line 178) as the single copy source.
   - Delete `hintState` (155), `hintStateLocalStorageKey` (173-176), `hintClosed` (197-200), `hintTextShown` (202-204), `closeHint` (206-209), and their entries in the returned object (420, 431-434, 451).
   - Remove the `showHintText` option from `UseScheduleOverlapUIOptions` (line 28).
2. `frontend/src/views/Event.vue`
   - Template (144-163): remove `closable` and `@click:close`; keep `type="info"`, `variant="tonal"`, the classes, and `data-testid="availability-editing-hint"`.
   - Script (1389-1417): `scheduleOverlapHintTextShown` becomes `scheduleOverlapHintText.value !== ""`; delete `scheduleOverlapHintClosed` and `closeScheduleOverlapHint`.
   - Remove `:show-hint-text="false"` from `ScheduleOverlap` (line 840).
3. `frontend/src/composables/event/types.ts` (28-30): keep `hintText: string`; delete `hintClosed` and `closeHint`.
4. `frontend/src/composables/event/useEventEditing.test.ts` (53-54, 143-144): drop the removed stub fields.

## 3. Phone in-view scroll

1. In `Event.vue`, add a template ref on the editing-instruction alert.
2. Watch `isEditing` (line 1262); when it transitions false to true and `isPhone.value` is true, `await nextTick()` and call `scrollIntoView({ block: "start", behavior: "smooth" })` on the alert element (use the component instance's `$el`). Trigger only on the transition, not on hint copy changes (available vs if-needed), so the page does not re-scroll mid-editing.
3. Note the mount path at Event.vue:2497 (`fromSignIn`/`editingMode` starts editing automatically); the watcher firing then is desirable.
4. Precedent for scroll-on-state-change: `useScheduleOverlapController.ts:252-260`. Prefer element `scrollIntoView` over `window.scrollTo` for precision; `scheduleOverlapTestUtils.ts` stubs `scrollTo` (around line 117) but not `scrollIntoView`, so stub/guard it in happy-dom tests.

## 4. Delete the dead hint strips and machinery

1. `frontend/src/components/schedule_overlap/ScheduleOverlap.vue`: remove the `showHintText` prop and default (153, 175), the `showHintText` option (490), the destructured `hintState`, `hintStateLocalStorageKey`, `hintTextShown`, `hintClosed`, `closeHint` (691-715), the `closeHint` listeners (`mobileOverlayListeners` 1039, `daysOnlyGridActions` 1061, `timedGridActions` 1073), and `hintClosed`/`closeHint` from `defineExpose` (1229-1231; keep `hintText`).
2. `ScheduleOverlapTimeGrid.vue` (365-383) and `ScheduleOverlapDaysOnlyGrid.vue` (73-91): delete the sticky hint strip blocks and any now-unused `MdiInformationOutline`/`MdiClose` imports.
3. `ScheduleOverlapMobileOverlay.vue` (14-32): delete the hint block and the `closeHint` emit (141), plus any now-unused `MdiInformationOutline`/`MdiClose` imports.
4. `useScheduleOverlapViewModels.ts` (156-157, 416-417, 471-472, 518-519) and `scheduleOverlapViewModelContracts.ts` (53, 64, 158-159, 212-213, 258-259): remove `hintTextShown`, `hintText`, and `closeHint` from options, builders, and contracts.
5. `frontend/src/components/landing/LandingPageCalendar.vue` (14): remove `:show-hint-text="false"`.
6. `scheduleOverlapTestUtils.ts` (310-311): remove `hintTextShown`/`hintText` from the overlay builder.
7. Update component tests that construct the removed fields: `ScheduleOverlapMobileOverlay.test.ts` (33-34), `ScheduleOverlapGridDragBinding.test.ts` (53, 124-125, 219, 248-249), `ScheduleOverlap.childViewModels.test.ts` (573-582), `useScheduleOverlapUI.test.ts` (44).
8. Grep for `hintClosed|closeHint|showHintText|hintTextShown|hintState` and confirm only `hintText` remains, exposed by ScheduleOverlap and consumed by Event.vue.

## 5. Tests

1. `frontend/src/views/Event.test.ts`
   - Replace "dismisses the editing instruction through the component close handler" (991-1004) and "keeps the add availability hint independent from the editing instruction dismissal" (1006-1021).
   - Add: the strip renders whenever `hintText` is non-empty and has no button/close control; it is absent when `hintText` is empty.
   - Add the regression: set `localStorage["closedHintTextedit_availability"] = "true"` before mount and assert the strip still renders; clear it after.
   - Add: with `isPhone` true, entering editing calls `scrollIntoView` on the strip (stub it).
   - Keep the top placement and `data-testid` assertions (965-989) and drop the `props("showHintText")` assertion (987).
2. `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.test.ts`: drop the `showHintText` option; keep all `hintText` copy cases; remove any dismissal expectations.
3. `e2e/specs/timed-event-add-availability-hint-firefox.spec.ts`
   - Replace the "dismissal persists for the editing state" step (83-92) with: the strip has no Close button; it stays visible while editing; after cancelling editing it is gone; entering editing again shows it again.
   - Add a regression test that runs `page.addInitScript(() => localStorage.setItem("closedHintTextedit_availability", "true"))` before `openEventPage`, enters editing, and asserts the strip is visible.
4. `e2e/specs/event-mobile-editing-options.spec.ts` (261-300): after entering editing, assert the strip is in the viewport with `await expect(hint).toBeInViewport()`; keep the overlay-absence assertion.

## 6. Checks

1. `cd frontend && npm run lint && npm run fmt:check && npm run typecheck && npm run build && npm run test:unit`
2. From `e2e`, run focused specs with full streaming output (never pipe through `head`/`tail`; use `2>&1 | tee /tmp/opencode/<name>.log` when a log is needed): `npm run test:e2e -- --project=firefox-desktop specs/timed-event-add-availability-hint-firefox.spec.ts` and `npm run test:e2e -- --project=chromium-mobile specs/event-mobile-editing-options.spec.ts`.
3. Run `graphify update .` from the repo root.
4. Run root `npm run format:markdown` and the corresponding check, plus `lint:markdown` if configured.
5. Swagger is untouched; no contract documents change.

## 7. Finalize

1. Verify each acceptance criterion with evidence and record it in implementation notes.
2. Write the final summary, mark TASK-0299 Done, and leave the record in `backlog/tasks/`.
3. Commit the change including the TASK-0299 record in one commit; do not include `backlog/backlog.md`. The TASK-0296 baseline is already committed (77273260), so no separate baseline commit is needed.
4. Do not push, open a PR, or touch origin branches; landing during a merge/deploy is out of scope.

## Risks and watch items

- `Event.test.ts` stubs ScheduleOverlap heavily; removing exposed fields can break unrelated assertions. Grep first and update deliberately.
- Removing the mobile-overlay hint changes overlay height; TASK-0288 flagged mobile legend clearance/layout specs as sensitive, so run the chromium-mobile layout specs if time allows.
- `scrollIntoView` is not implemented in happy-dom; stub it in unit tests.
- `isEditing` starts automatically on mount for `fromSignIn`/`editingMode`; the phone scroll will run then too, which is intended, but watch for test flakiness from async `nextTick`.
- A `v-alert` template ref holds the component instance; use its `$el` element for `scrollIntoView`, and guard the method's presence for happy-dom.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Baseline correction: TASK-0296 (Done) was already committed as 77273260, so no baseline commit was needed, and b3b9ed80 is on origin/fix-frontend (not local-only). The task description and plan were corrected before implementation.

Implementation:
- docs/requirements/functional/fr/FR-133.md now records that the instruction is not dismissible, renders only while the applicable editing or scheduling state is active, ignores browser-stored dismissal state from an earlier version, and is brought into view when Availability Editing starts on a phone viewport; four matching acceptance criteria were added. The README index row and proposed status are unchanged.
- frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts keeps hintText as the only copy source and removes hintState, hintStateLocalStorageKey, hintClosed, hintTextShown, closeHint, and the showHintText option, so no closedHintText* localStorage key is read or written.
- frontend/src/views/Event.vue drops closable and @click:close from the top alert, removes the ScheduleOverlap show-hint-text="false" prop, makes scheduleOverlapHintTextShown a non-empty hintText check, and adds a watcher on the isEditing false-to-true transition that scrolls the alert into view on a phone viewport. The watcher supports both an HTMLElement template ref and a component instance $el, and guards scrollIntoView for non-browser test environments.
- The fixed app header is h-14 on phones and h-16 from sm up (App.vue:28), so the alert carries tw:scroll-mt-18 tw:sm:scroll-mt-20. User screenshot feedback showed that matching the header height alone (scroll-mt-14/sm:scroll-mt-16) still left the banner flush under the header; the 16px gap version renders the full banner.
- Deleted the dead strips from ScheduleOverlapTimeGrid.vue, ScheduleOverlapDaysOnlyGrid.vue, and ScheduleOverlapMobileOverlay.vue and every showHintText, hintTextShown, hintText, and closeHint view-model field, action, contract, listener, defineExpose entry, test util, and component-test reference; ScheduleOverlapInstance exposes only hintText.

Evidence per acceptance criterion:
1. Top placement and per-state/viewport copy: Event.test.ts top-placement test plus useScheduleOverlapUI.test.ts eight copy cases; firefox-desktop and chromium-mobile e2e pass.
2. No dismiss control and no dismissal state in Event.vue: Event.test.ts no-button/clear-on-end test; firefox e2e asserts zero buttons in the strip; grep shows no hintClosed/closeHint/showHintText references remain.
3. No closedHintText* reads or writes: grep over frontend/src and e2e returns no matches; firefox e2e seeds closedHintTextedit_availability and sees the strip.
4. Active-state-only behavior: Event.test.ts clears the strip when hintText becomes empty; firefox e2e cancels editing and sees the strip gone, then re-enters editing and sees it again; the legacy-key test exercises a fresh page load.
5. Phone in-view: Event.test.ts scroll test asserts scrollIntoView({block:"start",behavior:"smooth"}) and the scroll-margin classes; chromium-mobile e2e asserts toBeInViewport({ratio:1}), proving the full banner is clear of the fixed header.
6. Dead machinery removed: grep is clean, typecheck passes, and the full unit suite passes.
7. FR-133 updated and formatted: root format:markdown, format:markdown:check, and lint:markdown pass.
8. Unit tests: frontend npm run test:unit passes with 154 files and 1236 tests.
9. E2E: firefox-desktop timed-event-add-availability-hint-firefox.spec.ts 4 passed; chromium-mobile event-mobile-editing-options.spec.ts 5 passed; e2e lint and typecheck pass.

Checks: frontend lint, fmt:check, typecheck, build, and test:unit pass. graphify update . completed with only the pre-existing zero-node JSON warnings. Swagger, scripts/prettier, and contract documents are untouched, so DoD items 5, 7, and 8 are not applicable. The user asked to skip the remaining broader e2e run after the required focused specs passed.

The change is staged in the working tree but not committed; backlog/backlog.md is left dirty and unstaged as required.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The event-page grid instruction is no longer dismissible, no longer consults browser-stored dismissal state, and the dead embedded hint strips were deleted.

- docs/requirements/functional/fr/FR-133.md records that the instruction is not dismissible, renders only while the applicable editing or scheduling state is active, ignores browser-stored dismissal state from earlier versions, and is brought into view when Availability Editing starts on a phone; matching acceptance criteria were added without changing the README index row or proposed status.
- frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts keeps hintText as the only copy source and removes hintState, hintStateLocalStorageKey, hintClosed, hintTextShown, closeHint, and the showHintText option, so no closedHintText* localStorage key is read or written.
- frontend/src/views/Event.vue drops closable and @click:close from the top alert, removes the ScheduleOverlap show-hint-text="false" prop, makes scheduleOverlapHintTextShown a non-empty hintText check, and adds a watcher that scrolls the alert into view on the phone isEditing transition; tw:scroll-mt-18 tw:sm:scroll-mt-20 keeps the full banner clear of the fixed h-14/h-16 app header.
- The three sticky strips are gone from ScheduleOverlapTimeGrid.vue, ScheduleOverlapDaysOnlyGrid.vue, and ScheduleOverlapMobileOverlay.vue, along with every related prop, view-model field/action/contract, listener, defineExpose entry, test utility, and component-test reference; ScheduleOverlapInstance exposes only hintText.
- Tests: Event.test.ts covers top placement, no dismiss control, clearing when editing ends, rendering with a stored legacy key, and the phone scroll with scroll-margin; useScheduleOverlapUI.test.ts keeps its eight per-state/viewport copy cases; the firefox-desktop and chromium-mobile specs cover the no-close control, cancel and re-enter visibility, the legacy key, and full-banner phone viewport visibility.

## Verification

- frontend: lint, fmt:check, typecheck, build, and test:unit (154 files, 1236 tests) pass.
- e2e: firefox-desktop timed-event-add-availability-hint-firefox.spec.ts 4 passed; chromium-mobile event-mobile-editing-options.spec.ts 5 passed with the mobile hint asserted toBeInViewport({ratio:1}); e2e lint and typecheck pass. The broader e2e run was stopped at the user's request after the required focused specs passed.
- docs: root format:markdown, format:markdown:check, and lint:markdown pass.
- graphify update . completed with only pre-existing zero-node JSON warnings.
- Swagger, scripts/prettier, and contract documents are unchanged.

## State

All task changes are staged in the working tree but not committed; backlog/backlog.md is left dirty and unstaged as the task required. No push, PR, or origin-branch action was taken.
<!-- SECTION:FINAL_SUMMARY:END -->
