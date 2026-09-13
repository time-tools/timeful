---
id: TASK-0220
title: >-
  Cap the sign-up event name field at 100 characters and surface server
  rejection (FR-119)
status: To Do
assignee: []
created_date: '2026-09-13 18:09'
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
- [ ] #1 The sign-up event form (NewSignUp.vue, create and edit modes) prevents entering or submitting an event name longer than 100 characters, per FR-119.
- [ ] #2 A name rejected by the API with the event-name-too-long token shows the specific too-long message rather than the generic create/edit failure message.
- [ ] #3 Unit tests cover the sign-up name cap in create and edit modes and the server-rejection message mapping.
- [ ] #4 npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit pass in frontend/.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
