---
id: TASK-0294
title: Render the compact timezone selector label in black
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 22:02'
updated_date: '2026-09-20 22:03'
labels: []
dependencies: []
modified_files:
  - frontend/src/components/schedule_overlap/TimezoneSelector.vue
  - frontend/src/components/schedule_overlap/TimezoneSelector.test.ts
priority: low
type: bug
ordinal: 294000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The compact timezone selector shows its selected offset label (for example `+3:00`) in Vuetify's muted on-surface color (`color(srgb 0 0 0 / 0.87)`) instead of solid black, so the label reads gray next to the black reset icon and the black `Timezone`/`Shown in` labels. The selector has a rule intended to paint this label black, but the rule targets a class the selection slot never renders, so Vuetify's own solo-field color wins. Make the compact label solid black without changing the other selector variants.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The compact timezone selector's selected-value label renders in solid black (rgb(0, 0, 0)) in the toolbar and the new-event dialog
- [x] #2 The underlined timezone selector variants keep their current label color
- [x] #3 A regression test ties the compact label color rule to the class the selection slot actually renders
- [x] #4 Required frontend checks pass: npm run lint, npm run fmt:check, npm run typecheck, npm run build, and npm run test:unit
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
Research (2026-09-21):
- The compact selector renders its value through the `#selection` slot as `div.timezone-select__selection-text.v-select__selection.v-select__selection--comma`; its computed color was `color(srgb 0 0 0 / 0.87)` inherited from `.v-field--variant-solo` in `@layer vuetify-components`.
- The scoped black rule anchored `:deep(.v-select__selection-text)`, but no element rendered that class (the owned class is `timezone-select__selection-text`), so the rule never matched while its sibling `:deep(.v-field__input)` and `:deep(.v-field)` rules did.

Plan:
1. Retarget the compact-button color rule to the class the selection slot renders.
2. Update the existing source assertion and add a unit regression test that extracts the class targeted by the color rule and asserts the rendered selection element carries it.
3. Verify the computed color in a real browser against a seeded compact selector.
4. Run required frontend checks and a focused e2e for the mobile timezone control.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-21: Root cause confirmed in the dev app: the selection div carries `timezone-select__selection-text` (added in 5dafba4f for truncation) while the compact-button color rule added later in 2f4c55ec targeted Vuetify's `v-select__selection-text`, so the rule was dead and the solo-field `color(srgb 0 0 0 / 0.87)` won. Fix retargets the rule at TimezoneSelector.vue:278.

Evidence: browser check on the dev server (new-event dialog, compact-button solo selector) shows the selection element computed color changed from `color(srgb 0 0 0 / 0.87)` to `rgb(0, 0, 0)`. Focused unit file 25/25 passed; full test:unit 154 files / 1233 tests passed; lint 0 errors (2 pre-existing NewSignUp.test.ts warnings); fmt:check clean; typecheck passed; build passed; focused chromium-mobile e2e passed (1 passed, 46.1s); graphify update completed.

Note: three `.compact-inline-select` rules at TimezoneSelector.vue:321, 330, and 375 still reference the stale `.v-select__selection-text`; they were left untouched because enabling them would change the underlined variant layout beyond this task's scope.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Fixed the compact timezone selector's selected-value label color, per request.

Why: the label rendered in Vuetify's muted solo-field color `color(srgb 0 0 0 / 0.87)` (sampled as gray) because the black rule anchored `:deep(.v-select__selection-text)`, a class the `#selection` slot never renders; the slot div carries the owned `timezone-select__selection-text` class instead.

Changes:
- frontend/src/components/schedule_overlap/TimezoneSelector.vue: retargeted the compact-button color rule to `.timezone-select__selection-text`.
- frontend/src/components/schedule_overlap/TimezoneSelector.test.ts: updated the source assertion and added a regression test that extracts the class targeted by the color rule and asserts the rendered selection element carries it.

Evidence: browser check on the dev server showed the selection element's computed color change from `color(srgb 0 0 0 / 0.87)` to `rgb(0, 0, 0)`; focused unit file 25/25 passed; full test:unit 154 files / 1233 tests passed; lint 0 errors (2 pre-existing NewSignUp.test.ts warnings); fmt:check clean; typecheck passed; build passed; focused chromium-mobile e2e `mobile timezone control keeps its fixed width when the reset button appears` passed; graphify update completed.

Notes: three `.compact-inline-select` rules still reference the stale `.v-select__selection-text` class (TimezoneSelector.vue lines 321, 330, 375); left untouched because enabling them would change the underlined variant layout beyond this task's scope.
<!-- SECTION:FINAL_SUMMARY:END -->
