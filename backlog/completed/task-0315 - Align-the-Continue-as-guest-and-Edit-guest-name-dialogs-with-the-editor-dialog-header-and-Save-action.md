---
id: TASK-0315
title: >-
  Align the Continue-as-guest and Edit-guest-name dialogs with the editor-dialog
  header and Save action
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 22:22'
updated_date: '2026-09-21 22:34'
labels:
  - frontend
dependencies: []
references:
  - frontend/src/components/GuestDialog.vue
  - frontend/src/components/GuestDialog.test.ts
  - frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue
  - frontend/src/components/schedule_overlap/EditingAvailabilityAs.test.ts
  - frontend/src/components/EditorDialogHeader.vue
  - frontend/src/components/NewDialog.vue
  - e2e/specs/guest-dialog-mobile-keyboard.spec.ts
  - e2e/specs/edit-guest-name-dialog-mobile-keyboard.spec.ts
  - e2e/specs/timed-event-add-availability-hint-firefox.spec.ts
  - e2e/specs/timed-event-visitor-identities-firefox.spec.ts
priority: medium
type: enhancement
ordinal: 315000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Continue as guest and Edit guest name dialogs render tighter top spacing than the Edit event form and use off-pattern actions: Continue as guest has a lone green Continue button, and Edit guest name has text-variant Cancel/Save actions. They therefore read as a different surface family, and the dismissal affordances vary between the two. Align both with the editor-dialog family: the shared editor header treatment (including the top-right cross), matching card top padding, and one brand-green solid Save primary action, with the cross as the only dismissal control.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Both the Continue as guest and Edit guest name dialog cards render the same top padding above the title as the Edit event dialog card.
- [x] #2 Both dialogs render their title through the same header treatment as the Edit event form, including the top-right cross close control with an accessible name.
- [x] #3 Neither dialog renders a Cancel action or any second dismiss control; dismissal is only the header cross.
- [x] #4 The Continue as guest primary action is a brand-green solid button labeled Save (replacing Continue) that keeps its current disabled and validation behavior.
- [x] #5 The Edit guest name bottom actions become a single brand-green solid Save button; the text Cancel and text-variant Save are gone.
- [x] #6 Existing behavior is preserved: guest name validation (required, taken, length), email collection and validation, Enter-key submit, autofocus, and visible-viewport sizing above the mobile keyboard.
- [x] #7 Unit tests cover the shared header, the named cross close control, the top padding, and the solid green Save in both dialogs, and assert no Cancel control remains.
- [x] #8 E2E specs that target the old Continue button or the Edit guest name Cancel button are updated and pass.
- [x] #9 Requirement or design documents are updated if they capture these dialog labels or dismissal controls; otherwise the final summary records that no requirement change was needed.
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
## Researched context (current system)

- `frontend/src/components/GuestDialog.vue`: `v-card` without top padding, local `v-card-title` with title plus inline icon close, `v-card-text` form with a right-aligned flat green `Continue` submit button; visible-viewport overlay sizing; submit gated by `canSubmit` and form validate.
- `frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue`: `v-dialog` to a `v-card` without top padding, plain `v-card-title` (no close control), name field, then `v-card-actions` with text-variant Cancel and primary text Save; `saveIfValid` gates non-required messages behind blur/save.
- `frontend/src/components/EditorDialogHeader.vue`: shared editor header (`v-card-title tw:mb-2 tw:flex tw:gap-2 tw:px-4 tw:sm:px-8`, title div `tw:mb-1`, spacer, top-right icon close with `aria-label="Close"` emitting `close`); requires `subtitle`, `helpHeader`, `dialog`, `showHelp`, `hideDialogActions`.
- Reference card: `NewDialog.vue` wraps editor forms in `<v-card class="tw:pt-4">`; editor body uses `tw:px-4 tw:py-1 tw:sm:px-8`.
- Tests: `GuestDialog.test.ts` stubs VBtn and takes the submit as `findAll("button").at(1)` (close is currently first); `EditingAvailabilityAs.test.ts` clicks a Cancel text button for dismissal and uses `getDialogButton(..., "Save")`; e2e `guest-dialog-mobile-keyboard.spec.ts` (two Continue locators), `edit-guest-name-dialog-mobile-keyboard.spec.ts` (Cancel and Save locators), `timed-event-add-availability-hint-firefox.spec.ts:141` and `timed-event-visitor-identities-firefox.spec.ts:316` (Continue locators).
- `HelpDialog` renders inside EditorDialogHeader whenever `dialog && !hideDialogActions`, so its Ok button joins the stubbed button list in unit tests.

## Plan

1. `GuestDialog.vue`: replace the local title with `EditorDialogHeader` (empty subtitle/help header, `:dialog="true"`, `:show-help="false"`, `:hide-dialog-actions="false"`, `@close` emits `update:modelValue` false); add `tw:pt-4` to the card and `tw:px-4 tw:sm:px-8` to `v-card-text` for header/body alignment; rename the flat green submit to `Save` keeping `timeful-flat-button tw:bg-green tw:text-white` and all validation/disabled behavior.
2. `EditingAvailabilityAs.vue`: swap the plain title for the same `EditorDialogHeader` (`@close` emits `update:editGuestNameDialog` false); add card `tw:pt-4` and text `tw:px-4 tw:sm:px-8`; replace Cancel and text Save with a single right-aligned flat green solid `Save`.
3. Unit tests: `GuestDialog.test.ts` locates Save by text, stubs HelpDialog, and asserts the header title, named cross, card top padding, solid green Save, no Cancel, and cross dismissal; `EditingAvailabilityAs.test.ts` replaces the Cancel relay test with cross dismissal and adds header/padding/solid-Save assertions while keeping save and validation coverage.
4. E2E: rename Continue to Save in the two keyboard specs and the two Firefox specs, and update the Edit guest name keyboard spec to drop Cancel and target Save plus the named cross while keeping the keyboard geometry intent.
5. Verify with frontend lint/fmt/typecheck/build/test:unit and the focused e2e specs; refresh the code knowledge graph.
6. Docs: no requirement changes expected; the final summary records that finding if confirmed.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: both dialogs use EditorDialogHeader (empty subtitle/help header, dialog mode, showHelp false) with card tw:pt-4 and body tw:px-4 tw:sm:px-8; GuestDialog submit renamed Continue to Save; the Edit guest name actions row replaced Cancel plus text Save with one right-aligned timeful-flat-button tw:bg-green tw:text-white Save in a tw:px-4 tw:sm:px-8 row; removed the GuestDialog inline close and MdiClose import.

Test adjustments: GuestDialog getSubmitButton now finds the Save button by text because HelpDialog's Ok button joins the stubbed button list; EditingAvailabilityAs.test.ts gained a slot-rendering v-card-title stub and the chip icon assertion was scoped to the chip row so the header cross no longer inflates it.

E2E adjustments: guest Save clicks are scoped to a dialog locator in the two Firefox specs to avoid a strict-mode collision with the desktop editing Save button; the Edit guest name keyboard spec now checks Save plus the named Close cross instead of Cancel.

Evidence: frontend lint (0 errors, 2 pre-existing NewSignUp warnings), fmt:check, typecheck, build, test:unit 158 files/1282 tests; chromium-mobile 3/3 then edit-guest-name 1/1; firefox-desktop 8/8; e2e lint/typecheck pass; e2e fmt:check flags only the pre-existing untouched timed-event-access-transfer-firefox.spec.ts; codebase-memory-mcp index refreshed.

Docs: searched docs/requirements, docs/design, and docs/terminology; no document references the guest dialog labels or dismissal controls, so no requirement change was needed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The Continue as guest and Edit guest name dialogs now match the Edit event dialog family and use one named dismiss plus one solid green Save.

### Continue as guest (`frontend/src/components/GuestDialog.vue`)
- Replaced the local title and inline close icon with the shared `EditorDialogHeader` (dialog mode, no help), so the title treatment and named top-right cross match the editor dialogs; removed the now-unused `MdiClose` import.
- Added `tw:pt-4` to the card to match the Edit event dialog card top padding, and `tw:px-4 tw:sm:px-8` to the body so fields align with the header.
- Renamed the flat green submit from `Continue` to `Save`, keeping `timeful-flat-button tw:bg-green tw:text-white` and all disabled, validation, and Enter-submit behavior.

### Edit guest name (`frontend/src/components/schedule_overlap/EditingAvailabilityAs.vue`)
- Adopted the same `EditorDialogHeader` treatment with a `tw:pt-4` card and `tw:px-4 tw:sm:px-8` body plus actions row.
- Replaced the text-variant Cancel and Save with a single right-aligned `timeful-flat-button tw:bg-green tw:text-white` Save; the header cross is now the only dismissal.

### Tests
- `GuestDialog.test.ts`: locates the submit by its `Save` text (the header's HelpDialog button joins the stub tree), and asserts the shared header title, named cross dismissal, `tw:pt-4`, solid green flat Save, and no Cancel.
- `EditingAvailabilityAs.test.ts`: replaces the Cancel relay test with named-cross dismissal, adds header/padding/green-Save/no-Cancel assertions, adds a slot-rendering `v-card-title` stub, and scopes the chip icon assertion to the chip row so the header cross does not inflate it.
- E2E: renamed Continue to Save in `guest-dialog-mobile-keyboard.spec.ts`; scoped the guest Save clicks to a dialog locator in `timed-event-add-availability-hint-firefox.spec.ts` and `timed-event-visitor-identities-firefox.spec.ts` to avoid a strict-mode collision with the desktop editing Save button; `edit-guest-name-dialog-mobile-keyboard.spec.ts` now checks Save plus the named Close cross.

## Verification
- frontend `npm run lint` (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (158 files, 1282 tests) pass.
- Focused e2e: chromium-mobile guest dialog specs 3/3 passed; edit-guest-name keyboard spec re-run 1/1 passed after the actions padding; firefox-desktop add-availability-hint plus visitor-identities specs 8/8 passed.
- e2e `npm run lint` and `npm run typecheck` pass; e2e `npm run fmt:check` flags only the pre-existing, untouched `specs/timed-event-access-transfer-firefox.spec.ts`.
- `codebase-memory-mcp cli index_repository --repo-path .` refreshed the graph.

## Notes
- No requirement, design, or glossary document references the dialogs' labels or dismissal controls, so AC9 is satisfied with no documentation change.
- No server, swagger, scripts, or contract changes.
<!-- SECTION:FINAL_SUMMARY:END -->
