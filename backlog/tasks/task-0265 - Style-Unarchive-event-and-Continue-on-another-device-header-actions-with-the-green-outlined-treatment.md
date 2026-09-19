---
id: TASK-0265
title: >-
  Style Unarchive event and Continue on another device header actions with the
  green outlined treatment
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 13:30'
updated_date: '2026-09-19 13:33'
labels:
  - frontend
dependencies: []
modified_files:
  - frontend/src/components/event/EventOwnerActions.vue
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventOwnerActions.test.ts
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/views/Event.test.ts
priority: medium
type: enhancement
ordinal: 265000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the event page header action row, the Unarchive event and Continue on another device actions currently render as neutral outlined buttons while the neighboring secondary actions Edit event and Copy link use the green outlined treatment. Apply the same green treatment: outlined border plus green label; the existing Unarchive event archive icon also turns green. Per the user decision, Continue on another device keeps its text-only content and gets no new icon. Match the existing header action color pattern (variant="outlined", color="primary", green slotted content) without changing action order, visibility rules, button sizes, or behavior.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Unarchive event header action renders with an outlined border, a green label, and its existing archive icon in green
- [x] #2 The Continue on another device header action renders with an outlined border and a green label, and gains no new icon
- [x] #3 Action order, visibility rules, and click behavior stay unchanged for both actions
- [x] #4 Unit coverage asserts the outlined primary color treatment for both actions and the frontend required checks pass
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
1. Add color="primary" to the outlined Unarchive event v-btn in frontend/src/components/event/EventOwnerActions.vue and apply tw:text-green to its icon and label.
2. Add color="primary" to the outlined trigger v-btn in frontend/src/components/event/EventAccessTransfer.vue and wrap its label in tw:text-green; add no icon.
3. Extend unit coverage: EventOwnerActions.test.ts and EventAccessTransfer.test.ts assert outlined/primary attributes and green slotted classes; extend the Event.test.ts archived header action test with data-variant/data-color assertions for both actions.
4. Run the frontend required checks (lint, fmt:check, typecheck, build, test:unit) and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Applied the existing header action pattern from `frontend/src/views/Event.vue` (`variant="outlined"`, `color="primary"`, `tw:text-green` slotted content) to both requested actions. Deliberately did not add the `event-metadata-action-button` height/padding class so button sizing stays out of scope.

Coverage added at two layers: component tests assert the outlined/primary attrs and green classes directly, and the Event view test asserts the rendered header row treatment through the semantic button stub with real `EventOwnerActions` and `EventAccessTransfer`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Restyled the event header Unarchive event and Continue on another device actions to match the green outlined treatment used by Edit event and Copy link.

## Change
- `frontend/src/components/event/EventOwnerActions.vue`: added `color="primary"` to the outlined Unarchive event button and applied `tw:text-green` to the archived icon and label.
- `frontend/src/components/event/EventAccessTransfer.vue`: added `color="primary"` to the outlined trigger and wrapped its label in a `tw:text-green` span; no icon was added per the user decision.
- Action order, visibility rules, button sizes, and behavior are unchanged.

## Tests
- `frontend/src/components/event/EventOwnerActions.test.ts`: added a green-outlined treatment test and switched the `v-icon` stub to one forwarding attrs so the icon class is assertable.
- `frontend/src/components/event/EventAccessTransfer.test.ts`: added a trigger treatment test.
- `frontend/src/views/Event.test.ts`: added an integration test asserting `data-variant="outlined"`, `data-color="primary"`, and green labels for both buttons in the archived header action row.
- Verified: focused suites 81 passed; `npm run lint` (0 errors, 2 pre-existing warnings), `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (149 files / 1123 tests, +3 new) all pass; `graphify update .` run.
- No e2e coverage required for a styling-only change that keeps accessible names and behavior; no Markdown, swagger, `scripts/`, `prettier/`, or contract-affecting changes.
<!-- SECTION:FINAL_SUMMARY:END -->
