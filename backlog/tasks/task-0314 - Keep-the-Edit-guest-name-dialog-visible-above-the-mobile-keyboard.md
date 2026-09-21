---
id: TASK-0314
title: Keep the Edit guest name dialog visible above the mobile keyboard
status: Done
assignee:
  - opencode
created_date: '2026-09-21 21:39'
updated_date: '2026-09-21 21:46'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue
  - frontend/src/composables/useVisualViewport.ts
  - frontend/src/components/GuestDialog.vue
  - e2e/specs/guest-dialog-mobile-keyboard.spec.ts
priority: high
type: bug
ordinal: 314000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On mobile, tapping the guest-name chip in the Editing availability indicator opens the Edit guest name dialog. Focusing the Guest name field raises the soft keyboard over the lower part of the dialog, hiding the Save and Cancel actions and blocking the guest from renaming their Event Response. The dialog must adapt to the visible viewport so its actions stay on screen while the keyboard is open. This is the shared frontend Edit guest name dialog rendered by frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue (Vue 3 + Vuetify); it appears in both the mobile overlay chip variant and the desktop/sidebar sentence variant, and it reuses the visual-viewport pattern introduced for the Continue-as-guest dialog in TASK-0311.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a phone-sized viewport with the soft keyboard open and focus in the Guest name field, the Save and Cancel actions in the Edit guest name dialog are fully visible and tappable
- [x] #2 When the dialog content is taller than the visible area, the content area scrolls while the Save action remains reachable
- [x] #3 The Guest name field remains visible and usable while the keyboard is open
- [x] #4 Closing the keyboard or resizing the viewport restores the dialog to its normal layout without clipping or missing content
- [x] #5 Desktop layout and behavior are unchanged
- [x] #6 Validation and save behavior are unchanged: an invalid or empty guest name does not save, and a valid name saves the same payload as before
- [x] #7 Regression coverage is added at the appropriate layer, or the task records a concrete justification for why the behavior cannot be covered
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
## Research findings

- `frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue` renders a Vuetify `v-dialog` (width 400) without any visible-viewport binding, so the centered card sits in the layout viewport and ends up behind the soft keyboard on mobile, exactly like the Continue-as-guest dialog before TASK-0311.
- TASK-0311 added `frontend/src/composables/useVisualViewport.ts` and bound its `{ top, height }` rect to the `v-dialog` overlay root in `frontend/src/components/GuestDialog.vue` via `:style` (`top`, `height`, `bottom: auto`). Vuetify 4 forwards the style prop to `.v-overlay`, and the existing `.v-overlay__content` max-height plus internal card scroll keep actions reachable.
- The Edit guest name dialog is shared by the mobile overlay chip variant (`ScheduleOverlapMobileOverlay.vue`) and the sidebar/sentence variant (`ScheduleOverlapSidebar.vue`), so one change in `EditingAvailabilityAs.vue` covers both.
- Guest e2e journeys can seed a response through `page.request.post(/api/events/:id/response)` (the page request context keeps the HttpOnly visitor credential), then enter editing from `#mobile-primary-availability-btn`; the chip then exposes `openEditGuestNameDialog`.
- Playwright cannot drive a real soft keyboard; TASK-0311's `e2e/specs/guest-dialog-mobile-keyboard.spec.ts` replaces `window.visualViewport` before app scripts run and changes its inset through a page-global controller.

## Plan

1. Add regression checks first (they must fail on the current code):
   - Extend `frontend/src/components/schedule_overlap/EditingAvailabilityAs.test.ts` with an attribute-forwarding `v-dialog` stub: assert the overlay style tracks a mocked `window.visualViewport` rect, updates on its `resize` event, and is absent when the API is unavailable.
   - Add a chromium-mobile e2e spec that simulates the open keyboard, seeds a guest response, enters editing, opens the Edit guest name dialog from the chip, and asserts the field and Save/Cancel actions stay inside the simulated visible viewport, then restores the full height and asserts normal layout.
2. In `EditingAvailabilityAs.vue`, bind the `useVisualViewport()` rect to the `v-dialog` with `:style` exactly as `GuestDialog.vue` does. Keep markup, validation, and save behavior unchanged; missing API keeps today's behavior.
3. Extract the shared visual-keyboard simulation from the TASK-0311 spec into an e2e helper if the new spec can reuse it without behavior changes.
4. Run frontend required checks and the focused e2e specs; record fail-before and pass-after evidence, then finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added regression coverage before the fix and observed the failures: the unit overlay-style assertion failed on an undefined style, and the chromium-mobile e2e failed with Expected <= 365 / Received 421 for a control inside the simulated keyboard region.

Pass-after evidence: frontend lint, fmt:check, typecheck, build, and test:unit (158 files, 1280 tests) pass; e2e chromium-mobile edit-guest-name-dialog-mobile-keyboard.spec.ts and guest-dialog-mobile-keyboard.spec.ts 3/3 pass; e2e lint and typecheck pass; root format:markdown:check passes.

Extracted the visual-keyboard simulation from the TASK-0311 spec into e2e/helpers/visual-keyboard.ts so both specs share it. Added an optional minimum-visible-height parameter; the new spec uses 160px to force the edit dialog taller than the visible area and prove the card cap and scroll reachability. Default behavior for the TASK-0311 spec is unchanged (240px floor).

Pre-existing unrelated failure: e2e npm run fmt:check flags specs/timed-event-access-transfer-firefox.spec.ts, a file this task did not modify. Same pre-existing item noted by TASK-0311.

Left concurrent task-0313 documentation edits in the worktree untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Mobile browsers resize only the visual viewport when the soft keyboard opens, while the layout viewport stays tall. The Edit guest name Vuetify dialog was centered in the layout viewport, so the raised keyboard covered the Save and Cancel actions. The dialog now positions its overlay inside `window.visualViewport`, so Vuetify's existing centering and height cap operate in the keyboard-visible area. This reuses the pattern introduced for the Continue-as-guest dialog in TASK-0311.

## Changes

- `frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue`: binds the visible-viewport rect from `useVisualViewport()` to `v-dialog` with `:style` (`top`, `height`, `bottom: auto`). Vuetify forwards the style to the `.v-overlay` root, so the dialog centers and caps itself inside the visible area; no markup, validation, or save behavior changes. Covers both the mobile chip variant and the desktop/sidebar sentence variant.
- `e2e/helpers/visual-keyboard.ts`: extracted the TASK-0311 spec's visual-keyboard simulation (init script, inset controller, visible-rect read, in-viewport assertion) into a shared helper. Added an optional minimum-visible-height parameter so a spec can force dialog overflow; the default keeps the previous 240px floor and TASK-0311 behavior.
- `e2e/specs/guest-dialog-mobile-keyboard.spec.ts`: now consumes the shared helper; behavior unchanged.
- `e2e/specs/edit-guest-name-dialog-mobile-keyboard.spec.ts`: new chromium-mobile regression spec.

## Tests

- `frontend/src/components/schedule_overlap/EditingAvailabilityAs.test.ts`: new tests assert the overlay style tracks the mocked visible viewport and updates on the API's `resize` event, and that no style is applied when `window.visualViewport` is unavailable.
- `e2e/specs/edit-guest-name-dialog-mobile-keyboard.spec.ts` (chromium-mobile): simulates the keyboard by overriding `window.visualViewport` before app scripts, seeds an owned guest response, opens the Edit guest name dialog from the editing chip, verifies the field, Save, and Cancel stay inside the visible viewport, verifies the dialog card is capped and its action reachable by scrolling when the visible area is shorter than the dialog, verifies the normal centered layout returns when the keyboard closes, and verifies the rename still saves.

## Verification

- Fail-before (unit): the overlay-style assertion failed on an undefined `style` before the binding.
- Fail-before (e2e): the controls check failed with `Expected: <= 365`, `Received: 421` (Save inside the simulated keyboard area).
- Pass-after: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit` (158 files, 1280 tests) in `frontend/`; e2e chromium-mobile `edit-guest-name-dialog-mobile-keyboard.spec.ts` plus both `guest-dialog-mobile-keyboard.spec.ts` tests 3/3 pass; e2e `npm run lint` and `npm run typecheck` pass; root `npm run format:markdown:check` passes.
- `codebase-memory-mcp` index refreshed.

## Notes and risks

- Desktop is a no-op: `visualViewport` matches the window viewport, so the computed style does not move the dialog.
- The dialog card still scrolls internally when content exceeds the visible area, keeping the actions reachable without pinning them.
- Pre-existing, unrelated: e2e `npm run fmt:check` reports `specs/timed-event-access-transfer-firefox.spec.ts`; that file is untouched by this task.
- The worktree also contains concurrent task-0313 documentation edits; they were left untouched.
<!-- SECTION:FINAL_SUMMARY:END -->
