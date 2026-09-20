---
id: TASK-0284
title: Make the event page title non-interactive text
status: Done
assignee:
  - opencode
created_date: '2026-09-19 21:53'
updated_date: '2026-09-19 21:58'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - docs/requirements/migration/backlog-fr-inventory-candidates/CAND-184.md
priority: medium
type: enhancement
ordinal: 284000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the Event Page, the Event name currently acts as an edit control: for a viewer who can edit Event settings it shows a pointer cursor and hover highlight and clicking it opens the event editor. The page already has an explicit Edit event action below the title, so the interactive title presents a second, redundant edit affordance and makes plain text look like a button. The title should read as non-interactive text for every viewer and state.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Event Page title is rendered as non-interactive text: no click handler, no pointer cursor, and no hover highlight, for every viewer and state.
- [x] #2 Clicking the Event Page title does not open the event editor and does not trigger any other action.
- [x] #3 The explicit Edit event action continues to open the event editor for viewers who can edit Event settings and stays disabled while scheduling.
- [x] #4 Unit tests assert that clicking the title does not call the edit action, and any existing test that depends on title click behavior is updated or replaced.
- [x] #5 No other Event Page header behavior or layout changes.
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
## Research

- `frontend/src/views/Event.vue` renders the title once, in the first `#event-header` row. The title `div` carries the conditional classes `tw:-mx-2 tw:-my-1 tw:cursor-pointer tw:rounded tw:px-2 tw:py-1 tw:transition-all tw:hover:bg-light-gray` when `canEditMetadata` is true and calls `editEvent()` on click when not scheduling.
- No e2e spec or helper clicks the title; the e2e edit flow uses the explicit `#edit-event-btn` (`e2e/helpers/timed-event-helpers.ts`).
- `Event.test.ts` clicks `.tw\:text-xl` in the scheduling test only to confirm edits stay disabled; that assertion loses meaning once the title is not clickable.
- CAND-184 already classifies edit-action placement as an implementation detail, so no functional requirement change is expected.

## Steps

1. Remove the `@click` handler and the `canEditMetadata` class binding from the Event Page title `div`, leaving the text sizing classes.
2. Update `frontend/src/views/Event.test.ts`: replace the title-click assertion in the scheduling test and add coverage that clicking the title does not call `editEvent`, the title has no pointer/hover affordance, and the explicit Edit event button still opens the editor.
3. Run the frontend required checks (`lint`, `fmt:check`, `typecheck`, `build`, `test:unit`) and `graphify update .`.

## Risks

- The title `div` must keep its layout classes so header spacing is unchanged.
- Existing tests that rely on title interactivity must be found and updated; only the scheduling test currently clicks it.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Removed the `canEditMetadata` conditional classes (`tw:cursor-pointer`, hover highlight, padding) and the `@click` edit handler from the Event Page title in `frontend/src/views/Event.vue`; the title keeps its `tw:text-xl tw:sm:text-3xl tw:sm:leading-10` sizing.

Added unit test `renders the event title as non-interactive text` (asserts no pointer/hover classes, click does not call `editEvent`, Edit event button still calls it) and removed the obsolete title-click assertion from the rescheduling test.

Checks: lint (0 errors), fmt:check, typecheck, build, full test:unit (151 files / 1191 tests) passed; `graphify update .` run. No formattable Markdown changed (Backlog files are excluded from format:markdown).

E2E not required for this change: no spec exercises the title click, and the concurrent session's in-progress TASK-0283 edits were present in the working tree, so an isolated e2e run was deferred.

Concurrency: this session's edits are staged in the index while the other session's Event.vue/Event.test.ts edits are unstaged. Do not commit without reviewing the combined diff.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The Event Page title is now plain, non-interactive text. Previously, when the viewer could edit Event settings, the title showed a pointer cursor and hover highlight and clicking it opened the event editor, duplicating the explicit Edit event action below it.

## Changes

- `frontend/src/views/Event.vue`: removed the `canEditMetadata` conditional classes and the `@click="canEditMetadata && !isScheduling && editEvent()"` handler from the title element; only the text sizing classes remain. No other header markup or layout changed.
- `frontend/src/views/Event.test.ts`: added `renders the event title as non-interactive text`, which asserts the title has no pointer/hover affordance, clicking it does not call `editEvent`, and the explicit Edit event button still opens the editor. Removed the obsolete title-click assertion from `disables desktop editing actions while rescheduling`, which still covers the disabled Edit event button while scheduling.

## Verification

- `npm run lint` (0 errors; 2 pre-existing warnings in `NewSignUp.test.ts`)
- `npm run fmt:check`
- `npm run typecheck`
- `npm run build`
- `npm run test:unit` — 151 files, 1191 tests passed
- `graphify update .`

E2E was not run: no e2e spec clicks the title and this interaction removal has DOM-level unit coverage. The working tree also contained in-progress TASK-0283 changes from a concurrent session, so an isolated e2e run was deferred.

## Notes

- No functional requirement update: CAND-184 already classifies edit-action placement on the Event Page as an implementation detail.
- Concurrency: this session's edits are staged in the index; the concurrent session's `Event.vue`/`Event.test.ts` changes are unstaged on top. Do not commit without reviewing the combined diff.
<!-- SECTION:FINAL_SUMMARY:END -->
