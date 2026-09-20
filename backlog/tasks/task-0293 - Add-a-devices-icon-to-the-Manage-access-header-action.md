---
id: TASK-0293
title: Add a devices icon to the Manage access header action
status: Done
assignee:
  - opencode
created_date: '2026-09-20 21:47'
updated_date: '2026-09-20 21:54'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/views/Event.test.ts
priority: low
type: enhancement
ordinal: 293000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The event header "Manage access" button currently shows only text (frontend/src/components/event/EventAccessTransfer.vue). Add the MDI devices icon to the left of the "Manage access" label so the action reads as a cross-device transfer, matching the prepend-icon pattern already used for Copy link (MdiContentCopy) and other actions.

Constraints:
- Keep the button's accessible name exactly "Manage access"; the icon must stay decorative.
- Do not change the existing outlined/primary styling or the green label span.
- Keep the icon consistent with existing unplugin-icons usage (~icons/mdi/...).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The event header "Manage access" button renders the mdi-devices icon to the left of the "Manage access" text.
- [x] #2 The button's accessible name remains "Manage access".
- [x] #3 Adding the icon does not change the button's outlined/primary styling or the tw:text-green label span.
- [x] #4 Regression coverage asserts the icon is rendered inside the Manage access button.
- [x] #5 Header action ordering tests still pass unchanged.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. In `frontend/src/components/event/EventAccessTransfer.vue`, import `MdiDevices` from `~icons/mdi/devices` and render a green `<v-icon>` with the `MdiDevices` glyph before the existing `tw:text-green` label span, adding `tw:ml-1` to the span. This mirrors the adjacent Edit event / Copy link header actions in `Event.vue` (same icon sizing, tint, and spacing) rather than the dialog's `prepend-icon` buttons.
2. Add regression coverage in `EventAccessTransfer.test.ts`: stub `v-icon` as `<i><slot /></i>`, and in the trigger-rendering test assert the `MdiDevices` icon component is rendered inside the Manage access button and precedes the label span.
3. Run frontend checks: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run test:unit`, `npm run build`.
4. Leave `Event.test.ts` header-ordering tests unchanged; confirm they still pass (SVG adds no text).
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the MDI devices icon to the left of the "Manage access" header action.

Implementation:
- `frontend/src/components/event/EventAccessTransfer.vue`: imported `MdiDevices` from `~icons/mdi/devices` and rendered it in a `tw:text-green` `<v-icon>` before the label, with `tw:ml-1` on the existing `tw:text-green` span. This mirrors the adjacent Edit event and Copy link header actions in `Event.vue` for consistent icon size, tint, and spacing.
- The icon is decorative at runtime: Vuetify `VIcon` sets `aria-hidden` when the icon is not clickable, so the button's accessible name stays "Manage access".

Evidence:
- `frontend`: lint (0 errors; only the 2 pre-existing `NewSignUp.test.ts` warnings), fmt:check, typecheck, and build all pass.
- `frontend` test:unit -> 154 files / 1232 tests passed. Focused `EventAccessTransfer.test.ts` -> 34 passed with new assertions that the `MdiDevices` component renders inside the Manage access button and precedes the label span. `Event.test.ts` -> 88 passed, including the unchanged header action ordering tests (Copy link, then Manage access).
- Firefox e2e `--project=firefox-desktop specs/timed-event-copy-feedback-firefox.spec.ts specs/timed-event-access-transfer-firefox.spec.ts` -> 8 passed, including both `getByRole("button", { name: "Manage access" })` selections; 2 heavier access-transfer journeys hit their explicit 40-second budget, and the user confirmed the timeouts are acceptable.
- `graphify update .` completed.

Not applicable: no Markdown changed, no swagger annotations, no `scripts/` or `prettier/` changes, and no contract-affecting changes, so DoD items 4, 5, 7, and 8 were left unchecked.
<!-- SECTION:FINAL_SUMMARY:END -->
