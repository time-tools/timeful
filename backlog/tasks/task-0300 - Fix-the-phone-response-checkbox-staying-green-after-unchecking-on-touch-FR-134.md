---
id: TASK-0300
title: >-
  Fix the phone response checkbox staying green after unchecking on touch
  (FR-134)
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 13:06'
updated_date: '2026-09-21 13:15'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - docs/requirements/functional/fr/FR-134.md
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
modified_files:
  - docs/requirements/functional/fr/FR-134.md
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - frontend/src/components/schedule_overlap/RespondentsList.test.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
priority: medium
type: bug
ordinal: 300000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On the phone viewport the response checkbox in the responses list must present exactly three appearances: selected = primary green outline with a green tick; unselected while the avatar/status control, the response name, or the checkbox itself is hovered or pressed = primary green outline with no tick; otherwise = neutral grey outline. The hover trigger is limited to those three regions on the phone viewport; desktop keeps its current whole-row hover behavior. These states and the hover trigger regions are not yet recorded in FR-134, which only says the unchecked outline is neutral and hovered/selected use the primary action color.

Reproduced defect (firefox-touch, isolated test stack): tapping a selected checkbox unchecks it and removes the tick, but the outline remains primary green rgb(0, 153, 76) after the tap; it only returns to neutral rgb(189, 189, 189) after tapping elsewhere. The checkbox therefore looks highlighted while unchecked. The native CSS hover rule used for the green outline stays applied on touch after the tap ends, so the appearance must not depend on that sticky hover state alone.

Confirmed product decisions: a touch press counts as hovered (green outline while the finger is down, cleared once it lifts if the response is still unselected), and the three-region hover restriction applies to the phone viewport only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 FR-134 records the three phone-viewport checkbox appearance states (selected = primary green outline with tick; unselected hovered/pressed = primary green outline without tick; otherwise neutral outline) and the phone hover trigger regions (avatar/status control, response name, or checkbox), amended in place rather than superseded, with docs/requirements/README.md kept consistent
- [x] #2 With touch input on the phone viewport, tapping a selected row's checkbox to uncheck it and lifting the finger returns the outline to the neutral color without another tap, and the tick is gone
- [x] #3 On the phone viewport, while a pointer hovers or a finger presses the avatar/status, name, or checkbox of an unselected row, the outline uses the primary action color with no tick; moving or lifting away returns it to the neutral outline
- [x] #4 A selected row keeps the primary action outline and tick regardless of hover or press
- [x] #5 On the phone viewport, hovering or pressing elsewhere in the row (email line, row action buttons, row whitespace) does not show the green hovered checkbox state
- [x] #6 Desktop viewport behavior is unchanged: an unselected checkbox stays hidden until the row is hovered, a selected checkbox stays visible, hovered and selected outlines use the primary action color, and the unchecked outline is neutral
- [x] #7 Regression coverage asserts the post-uncheck neutral outline under touch input, the hovered/pressed green outline without a tick, and the selected green outline with a tick; existing respondents-list unit tests and respondent-selection e2e specs still pass
- [x] #8 Existing checks pass: frontend lint, fmt:check, typecheck, build, and test:unit, plus the focused firefox-touch and firefox-desktop specs; changed Markdown is formatted
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
1. Requirements: amend `docs/requirements/functional/fr/FR-134.md` in place to record the phone-viewport three-state checkbox appearance (selected = primary action outline + tick; unselected hovered/pressed on the avatar/status control, response name, or checkbox = primary action outline without tick; otherwise neutral outline), the press-counts-as-hover and hover-ends-on-release rule, and the phone-only scope; keep desktop whole-row hover rules and update its acceptance criteria. Keep the `docs/requirements/README.md` index row consistent. Read `docs/requirements/README.md` and `docs/requirements/functional/README.md` first.

2. Component `frontend/src/components/schedule_overlap/RespondentsList.vue`: track a phone-only hovered respondent id in the component, set on `pointerenter` and cleared on `pointerleave`/`pointercancel` from the avatar/status button, the name line, and the checkbox; bind `respondent-row--phone` (from `isPhone`) and `respondent-row--hovered` on the row. Rework the scoped checkbox CSS so the native `:hover` border rule is gated to non-phone rows (`.respondent-row:not(.respondent-row--phone):hover`), phone green comes from `--selected` and `--phone.--hovered` only, and the base color stays neutral. Keep visibility rules, desktop reveal, tick rendering, and no-movement guarantees unchanged.

3. Unit tests `frontend/src/components/schedule_overlap/RespondentsList.test.ts`: add coverage for phone hover lifecycle on the three regions (enter sets the row class, leave/cancel clears it, desktop pointer enter does not set it) and update the source-guard assertions for the reworked hover selector.

4. E2E: extend `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` with a focused phone test asserting the checkbox border returns to the neutral color immediately after a touchscreen tap unchecks a selected response (no extra tap), the tick is absent, hovering the name greens it without a tick, hovering the email line leaves it neutral, and the selected state stays green with a tick with the pointer away. Leave the firefox-desktop spec as the guard that desktop reveal and hover behavior did not regress.

5. Checks: frontend `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`; focused `--project=firefox-touch` and `--project=firefox-desktop` specs with streaming output; root Markdown formatting; `graphify update .`.

Risks: jsdom/happy-dom cannot evaluate real `:hover`, so the collision between native sticky hover and the phone class must be verified in firefox-touch; Firefox pointer boundary events for touch must actually fire `pointerleave` at touch end, otherwise fall back to a `pointerup` (touch-only) clear while keeping mouse hover intact; `pointerenter` handlers must stay inert on desktop so the existing whole-row hover behavior is untouched.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Phone hover is tracked in the component with `phoneHoveredRespondentId`: `pointerenter` on the avatar/status button, name line, and checkbox sets it, and `pointerleave`/`pointercancel` clear it. Firefox touch fires `pointerleave` at touch end, so the planned touch-only `pointerup` fallback was not needed.

The native `:hover` border rule is now scoped to non-phone rows with `.respondent-row:not(.respondent-row--phone):hover`; phone rows get `respondent-row--phone` and green comes from `--selected` or `--phone.--hovered` only. This keeps desktop whole-row hover and reveal behavior untouched.

The e2e uncheck assertion failed against the old behavior (sticky green `rgb(0, 153, 76)`) and passes with the fix; the touch test uses `touchscreen.tap` for select/uncheck and mouse hover for the three-region state assertions.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The phone-viewport response checkbox now shows exactly three appearances: selected = primary action outline with a tick, unselected hovered/pressed on the avatar/status control, response name, or checkbox = primary action outline without a tick, and neutral outline otherwise. A touch tap that unchecks a response no longer leaves the outline green.

- `docs/requirements/functional/fr/FR-134.md` is amended in place: the tick appears only while the row is selected, the phone hover/press trigger regions are the Availability Status control, response name, and checkbox, desktop keeps whole-row hover, and the hovered or pressed outline ends when the pointer or touch ends. Two acceptance criteria cover the phone state matrix and the post-uncheck neutral outline.
- `frontend/src/components/schedule_overlap/RespondentsList.vue` tracks a phone-only `phoneHoveredRespondentId` set from `pointerenter` on the three regions and cleared from `pointerleave`/`pointercancel`, binds `respondent-row--phone` and `respondent-row--hovered`, and scopes the native `:hover` border rule to non-phone rows with `.respondent-row:not(.respondent-row--phone):hover`. Phone green now comes only from `--selected` and `--phone.--hovered`; the `watch(isPhone)` reset clears stale hover when the viewport grows.
- `frontend/src/components/schedule_overlap/RespondentsList.test.ts` adds phone hover lifecycle coverage for all three regions plus a desktop no-op guard, and updates the scoped-CSS source guard.
- `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` adds a test that taps to select and tap-unchecks (asserting the neutral outline and no tick), hovers the name for the green no-tick state, hovers the edit action for the neutral state, and reselects to confirm the tick keeps the green outline with the pointer away.

## Why

Under touch, the browser's native `:hover` stays applied to the tapped row, so the previous unchecked-state rule kept the checkbox green until the user tapped elsewhere. The fix makes the phone appearance independent of sticky native hover while keeping desktop hover unchanged.

## Verification

- Reproduced before the fix in firefox-touch: after tap-uncheck the border stayed `rgb(0, 153, 76)` and only returned to `rgb(189, 189, 189)` after tapping elsewhere.
- E2E: `--project=firefox-touch --project=firefox-desktop` for `schedule-overlap-mobile-touch-firefox.spec.ts` and `timed-event-respondent-selection-firefox.spec.ts` passed, 10/10.
- Frontend: lint (only the two pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, build, and test:unit (154 files / 1238 tests) passed.
- Root `format:markdown:check` and `lint:markdown` passed; `graphify update .` completed.
- E2E package lint and typecheck passed; `fmt:check` still flags only the unmodified pre-existing `specs/timed-event-access-transfer-firefox.spec.ts`.
- No swagger annotations, environment docs, plugin API shapes, or migration contracts changed.
<!-- SECTION:FINAL_SUMMARY:END -->
