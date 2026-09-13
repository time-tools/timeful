---
id: TASK-0220
title: >-
  Cap the sign-up event name field at 100 characters and surface server
  rejection (FR-119)
status: Done
assignee:
  - opencode
created_date: '2026-09-13 18:09'
updated_date: '2026-09-13 18:49'
labels:
  - frontend
dependencies: []
references:
  - docs/requirements/functional/fr/FR-119.md
  - TASK-0137
priority: medium
type: enhancement
ordinal: 223000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up to TASK-0136 and TASK-0137. TASK-0136 capped only the NewEvent.vue name field; the sign-up event form in frontend/src/components/NewSignUp.vue creates and edits events through the same POST /events and PUT /events/:id routes (frontend/src/components/NewSignUp.vue:476, :500) but its name field uses the shared nameRules from frontend/src/composables/event/useEventEditorState.ts:229, which require only a non-empty value.

Implement the FR-119 frontend cap for the sign-up event form:

- Cap the name field at 100 characters in both create and edit modes, reusing frontend/src/utils/eventName.ts (EVENT_NAME_MAX_LENGTH, validateEventName) as NewEvent.vue does.
- Show the existing specific too-long validation message when the API rejects the name with event-name-too-long, instead of the generic "There was a problem creating that event!" / "...editing..." error.
- Keep the shared nameRules behavior for sign-up response name flows unchanged.
- Add unit coverage for create and edit modes; update affected e2e selectors if the name field markup or placeholder changes.
- Scope is frontend only; the server-side rejection is delivered by TASK-0137.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The sign-up event form (NewSignUp.vue, create and edit modes) prevents entering or submitting an event name longer than 100 characters, per FR-119.
- [x] #2 A name rejected by the API with the event-name-too-long token shows the specific too-long message rather than the generic create/edit failure message.
- [x] #3 Unit tests cover the sign-up name cap in create and edit modes and the server-rejection message mapping.
- [x] #4 npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass in frontend/.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. NewSignUp.vue: add `:maxlength="EVENT_NAME_MAX_LENGTH"` to the name field and replace `:rules="nameRules"` with a local `eventNameRules` computed that mirrors NewEvent.vue (`getEventNameValidationMessage(validateEventName(value).code) ?? true`); leave the shared `nameRules` in useEventEditorState.ts untouched.
2. Add a local error-token mapping in NewSignUp.vue so a `parsed.error === "event-name-too-long"` rejection from POST /events or PUT /events/:id shows `getEventNameValidationMessage("tooLong")` via mainStore.showError instead of the generic create/edit failure text.
3. NewSignUp.test.ts: add a capturing v-text-field stub, cover the 100-character cap in create and edit modes, and cover the event-name-too-long mapping plus the generic fallback for both create and edit paths.
4. Run `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit` in frontend/.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation notes (2026-09-13):
- frontend/src/components/NewSignUp.vue: name field gained :maxlength="EVENT_NAME_MAX_LENGTH" and :rules="eventNameRules"; eventNameRules mirrors NewEvent.vue via getEventNameValidationMessage(validateEventName(value).code) ?? true. Dropped the shared nameRules destructure; useEventEditorState.ts was not modified.
- frontend/src/components/NewSignUp.vue: added getSubmitErrorMessage with the event-name-too-long token constant; both create and edit catches now show getEventNameValidationMessage("tooLong") for that token and the original generic text otherwise.
- frontend/src/components/NewSignUp.test.ts: hoisted showErrorMock so snackbar calls are assertable, added a capturing VTextField stub, submitSignUp helper, and four new tests (cap create, cap edit, create rejection mapping + generic fallback, edit rejection mapping + generic fallback).

Evidence:
- Focused: npx vitest run src/components/NewSignUp.test.ts -> 19 tests / 19 pass.
- Full: npm run test:unit -> 146 files / 1097 tests pass.
- npm run lint: 0 errors (2 vue/one-component-per-file warnings from the local test stubs; NewEvent.test.ts has the same pattern).
- npm run fmt:check, npm run typecheck, npm run build: pass.
- No e2e selectors affected and no e2e spec exercises the NewSignUp.vue form; sign-up e2e specs seed events via the API, so no e2e run was required. No Markdown files changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Capped the sign-up event form name field at 100 characters and surfaced the server's event-name-too-long rejection with the specific message.

Changes:
- NewSignUp.vue: the name field now binds :maxlength="EVENT_NAME_MAX_LENGTH" (100) and uses a local eventNameRules built from validateEventName/getEventNameValidationMessage, mirroring NewEvent.vue. The shared nameRules in useEventEditorState.ts is untouched.
- NewSignUp.vue: added getSubmitErrorMessage, used by both the create and edit catch blocks, so a parsed error token of event-name-too-long shows "Event name must be 100 characters or fewer" via mainStore.showError instead of the generic create/edit failure text; other errors keep the generic text.
- NewSignUp.test.ts: added a capturing VTextField stub and tests for the 100-character cap in create and edit modes plus the too-long message mapping (and generic fallback) for both POST /events and PUT /events/:id.

Validation:
- Focused: npx vitest run src/components/NewSignUp.test.ts -> 19 tests pass.
- npm run test:unit -> 146 files / 1097 tests pass.
- npm run lint -> 0 errors (2 vue/one-component-per-file warnings from local test stubs, matching the NewEvent.test.ts pattern); npm run fmt:check, npm run typecheck, and npm run build pass.
- No e2e selectors changed (placeholder and markup unchanged) and no e2e spec drives the NewSignUp.vue form (sign-up specs seed events through the API), so no e2e run was required. No Markdown files changed.
<!-- SECTION:FINAL_SUMMARY:END -->
