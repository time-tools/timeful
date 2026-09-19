---
id: TASK-0285
title: Remove the anonymous-created note banner from the Edit event form
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 22:13'
updated_date: '2026-09-19 22:16'
labels: []
dependencies: []
references:
  - frontend/src/components/NewEvent.vue
  - frontend/src/components/NewEvent.test.ts
  - frontend/src/views/Event.test.ts
documentation:
  - frontend/AGENTS.md
modified_files:
  - frontend/src/components/NewEvent.vue
  - frontend/src/components/NewEvent.test.ts
priority: low
type: enhancement
ordinal: 285000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Edit event form shows an AlertText banner reading "Anybody can edit this event because it was created while not signed in" whenever the event was created while not signed in. The product no longer wants this note presented in the Edit event form; the edit-access behavior it described stays unchanged. The banner adds noise to the editor and is not needed for the owner to understand or use the form.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Edit event form no longer renders the "Anybody can edit this event because it was created while not signed in" note/banner for an event created while not signed in, or for any other event.
- [x] #2 Unchanged behavior: an event created while not signed in remains editable through the Edit event form after the note is removed.
- [x] #3 No user-facing occurrence of that message remains in the frontend source.
- [x] #4 Regression coverage asserts the note is absent when editing an anonymous-created event.
- [x] #5 Required frontend checks pass: npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit.
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
1. In `frontend/src/components/NewEvent.vue`, delete the `<AlertText v-if="edit && guestEvent">` banner and its now-unused `AlertText` import. Keep the `guestEvent` computed and the `edit` prop behavior: they still gate the sign-in-required controls, and `AlertText` stays in use by `ScheduleOverlapSidebar.vue`.
2. Add regression coverage in `frontend/src/components/NewEvent.test.ts`: mount `NewEvent` with `edit: true` and an event whose `ownerId` is `guestUserId`, then assert the banner component is absent and the copy no longer appears; also assert the editor still renders its form so the anonymous-created event stays editable.
3. Leave `frontend/src/components/AlertText.test.ts` untouched: its sample slot copy is a generic component fixture, not user-facing product copy.
4. Run the required frontend checks (`npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`) and `graphify update .`.
Risk: avoid touching files already modified by the in-progress TASK-0283 worktree changes (`Event.vue`, `Event.test.ts`).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Removed the `AlertText` guest-owner banner and its unused import from `NewEvent.vue`; `guestEvent` and `edit` remain in use for the sign-in-gated controls, and `AlertText` stays in use by `ScheduleOverlapSidebar.vue`.

Added a regression test in `NewEvent.test.ts`; confirmed it fails before the removal (`findComponent(AlertText).exists()` was true) and passes after.

Checks: lint (0 errors; 2 pre-existing warnings in NewSignUp.test.ts), fmt:check, typecheck, build, test:unit (151 files, 1192 tests) all pass; `graphify update .` run.

No e2e test referenced the banner (grep over `e2e/` found no matches), so no e2e coverage was required. `AlertText.test.ts` keeps the sentence only as generic slot fixture copy, not user-facing product copy, and was left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Removed the "Anybody can edit this event because it was created while not signed in" AlertText banner from the Edit event form in `frontend/src/components/NewEvent.vue` and deleted the now-unused `AlertText` import. The `guestEvent` computed and `edit` prop remain in use for the sign-in-gated controls, and `AlertText` stays in use by `ScheduleOverlapSidebar.vue`.

Added a regression test in `frontend/src/components/NewEvent.test.ts` that mounts the editor in edit mode for an event owned by `guestUserId` and asserts the banner component is absent, the copy is gone from the editor source, and the form still renders. Verified the test fails before the removal and passes after.

Verification: `npm run lint` (0 errors, 2 pre-existing warnings in `NewSignUp.test.ts`), `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (151 files, 1192 tests) all pass; `graphify update .` run. No e2e reference to the banner existed, and none is required for this copy removal. No Markdown, Swagger, or contract documents were affected.

Risk: none identified; edit permissions for anonymous-created events are unchanged, as covered by the existing Event view unit tests.
<!-- SECTION:FINAL_SUMMARY:END -->
