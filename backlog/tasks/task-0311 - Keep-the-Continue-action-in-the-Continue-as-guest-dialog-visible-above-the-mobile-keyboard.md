---
id: TASK-0311
title: >-
  Keep the Continue action in the Continue-as-guest dialog visible above the
  mobile keyboard
status: Done
assignee:
  - opencode
created_date: '2026-09-21 21:04'
updated_date: '2026-09-21 21:31'
labels: []
dependencies: []
references:
  - frontend/src/components/GuestDialog.vue
priority: high
type: bug
ordinal: 312000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On mobile, opening the Continue as guest dialog and focusing the Guest name field raises the soft keyboard over the lower part of the dialog, hiding the Continue action and blocking the guest from finishing their Event Response. The dialog must adapt to the visible viewport so the primary action stays on screen while the keyboard is open. This is the frontend Continue-as-guest flow rendered by frontend/src/components/GuestDialog.vue (Vue 3 + Vuetify); the dialog is opened from the event page when a viewer chooses to respond as a guest.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a phone-sized viewport with the soft keyboard open and focus in the Guest name field, the Continue action in the Continue as guest dialog is fully visible and tappable
- [x] #2 When the dialog content is taller than the visible area, the content area scrolls while the Continue action remains reachable
- [x] #3 The guest name, optional email, and allow-others-to-edit controls remain visible and usable while the keyboard is open
- [x] #4 Closing the keyboard or resizing the viewport restores the dialog to its normal layout without clipping or missing content
- [x] #5 Desktop layout and behavior are unchanged
- [x] #6 Validation and submit behavior are unchanged: an invalid or empty guest name does not submit, and a valid name submits the same payload as before
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

- `frontend/src/components/GuestDialog.vue` renders a Vuetify 4 `v-dialog` (width 400, static location strategy), so the card is flex-centered inside the `.v-overlay` root, which is `position: fixed; inset: 0` and spans the whole layout viewport.
- On mobile Chrome the soft keyboard resizes only the visual viewport (`interactive-widget=resizes-visual` default); the layout viewport stays tall, so the centered card ends up behind the keyboard even though `window.visualViewport` already reports the visible area.
- Vuetify 4.2 `VDialog` forwards its `style` prop to the `.v-overlay` root, and `.v-overlay__content` is already capped by `max-height: calc(100% - 48px)` with an internally scrollable card. Resizing the overlay root is therefore enough to move centering and the existing height cap into the keyboard-visible area.
- The app has no existing visual-viewport handling (`visualViewport` appears nowhere under `frontend/src`).

## Plan

1. Add `frontend/src/composables/useVisualViewport.ts`: tracks `window.visualViewport` as `{ top, height }`, updates on the viewport's `resize` and `scroll` events, cleans up on scope dispose, and returns `null` when the API is unavailable.
2. In `GuestDialog.vue`, bind that rect to `v-dialog` with `:style` (`top`, `height`, `bottom: auto`) so the existing Vuetify centering and max-height cap operate inside the visible viewport. Keep markup, validation, and submit behavior unchanged; when `visualViewport` is missing, apply no style and preserve today's behavior.
3. Regression tests first (they must fail on the current code):
   - Unit test for the composable: initial read, resize/scroll updates, dispose cleanup, unsupported fallback.
   - Component test for `GuestDialog`: mocked `visualViewport` plus an attribute-forwarding `v-dialog` stub; assert the overlay style tracks the visible-viewport rect.
   - New `e2e/specs/guest-dialog-mobile-keyboard.spec.ts` (chromium-mobile): override `window.visualViewport` before app scripts to simulate an open keyboard, open the Continue-as-guest dialog through the mobile Save flow, assert the guest name field, checkbox, and Continue button are inside the simulated visible area, then restore the full height and assert normal layout resumes.
4. Run frontend required checks and the focused e2e spec; record evidence, then finalize the task.

Final structure: the chromium-mobile e2e splits into two tests (reported no-email case, and the email-collecting case plus scroll reachability and submit), sharing a keyboard simulation helper; the artificial taller-keyboard middle step was folded into the email case, which naturally overflows the visible area.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the fix after recording the plan: added frontend/src/composables/useVisualViewport.ts and bound its rect to the v-dialog overlay root in GuestDialog.vue. Vuetify 4.2 VDialog forwards its style prop to the .v-overlay root, and the existing .v-overlay__content max-height plus card overflow keep the dialog scrollable when space is short.

Regression checks were added first and observed failing on the unfixed code: the composable suite failed to resolve the missing module, the GuestDialog overlay-style test failed on an undefined style, and the chromium-mobile e2e failed with Expected <= 365 / Received 389 for a control inside the simulated keyboard region.

Pass-after evidence: frontend lint, fmt:check, typecheck, build, and test:unit (158 files, 1278 tests) pass; e2e chromium-mobile specs/guest-dialog-mobile-keyboard.spec.ts 2/2 passes; e2e firefox-desktop specs/timed-event-add-availability-hint-firefox.spec.ts 4/4 passes; e2e lint and typecheck pass.

Playwright cannot drive a real soft keyboard, so the e2e spec replaces window.visualViewport before app scripts run and leaves the layout viewport untouched, reproducing mobile Chrome's resizes-visual behavior. The first version of the simulation read window.innerHeight before the mobile viewport meta applied (1369px); it now uses live getters and a controller to change the simulated inset.

Observed environmental e2e flake during development: repeated ERR_NETWORK_CHANGED failures on Vite module loads when each run cold-started the stack. Running with TEST_DB_PERSIST=true and a warm stack made runs stable and the failure did not reproduce in the final runs. Not caused by the change.

Pre-existing unrelated failure: e2e npm run fmt:check flags specs/timed-event-access-transfer-firefox.spec.ts, a file this task did not modify.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Mobile browsers resize only the visual viewport when the soft keyboard opens, while the layout viewport stays tall. The Continue-as-guest Vuetify dialog was centered in the layout viewport, so the raised keyboard covered the Continue action. The dialog now positions its overlay inside `window.visualViewport`, so Vuetify's existing centering and height cap operate in the keyboard-visible area.

## Changes

- Added `frontend/src/composables/useVisualViewport.ts`: tracks the visible viewport rect (`offsetTop`, `height`), updates on the API's `resize`/`scroll` events, cleans up on scope dispose, and returns `null` when `window.visualViewport` is unavailable (fallback preserves current behavior).
- `frontend/src/components/GuestDialog.vue`: binds the visible viewport rect to `v-dialog` with `:style` (`top`, `height`, `bottom: auto`). Vuetify forwards `style` to the `.v-overlay` root, so the dialog centers and caps itself inside the visible area; no markup, validation, or submit changes.

## Tests

- `frontend/src/composables/useVisualViewport.test.ts`: initial read, resize/scroll updates, dispose cleanup, unsupported fallback.
- `frontend/src/components/GuestDialog.test.ts`: overlay style tracks the mocked visible viewport rect and updates when the keyboard closes; no style when the API is unavailable.
- `e2e/specs/guest-dialog-mobile-keyboard.spec.ts` (chromium-mobile, 2 tests): simulates the keyboard by overriding `window.visualViewport` before app scripts; verifies the no-email dialog keeps name, checkbox, and Continue inside the visible viewport and recenters when the keyboard closes; verifies the email-collecting dialog keeps its fields visible, its action reachable by scrolling, and still submits the guest response.

## Verification

- Fail-before (e2e): with the overlay binding removed, the controls check failed with `Expected: <= 365`, `Received: 389` (Continue inside the simulated keyboard area).
- Fail-before (unit): composable module missing and the dialog overlay style assertion failed before implementation.
- Pass-after: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit` (158 files, 1278 tests) in `frontend/`; chromium-mobile e2e 2/2; firefox-desktop `timed-event-add-availability-hint-firefox.spec.ts` 4/4 (guest dialog submit flow); `e2e` lint and typecheck.
- `codebase-memory-mcp cli index_repository --repo-path .` refreshed; root `npm run format:markdown:check` passes.

## Notes and risks

- Desktop is a no-op: `visualViewport` matches the window viewport, so the computed style does not move the dialog.
- The dialog card still scrolls internally when its content exceeds the visible area, which keeps the action reachable without pinning it.
- Pre-existing, unrelated: `e2e` `npm run fmt:check` reports `specs/timed-event-access-transfer-firefox.spec.ts`; that file is untouched by this task.
<!-- SECTION:FINAL_SUMMARY:END -->
