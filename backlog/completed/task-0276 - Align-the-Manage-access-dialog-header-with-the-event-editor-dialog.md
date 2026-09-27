---
id: TASK-0276
title: Align the Manage access dialog header with the event editor dialog
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 18:28'
updated_date: '2026-09-19 18:40'
labels:
  - frontend
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/components/EditorDialogHeader.vue
  - frontend/src/components/EditorDialogHeader.test.ts
  - frontend/src/components/NewEvent.vue
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
priority: medium
type: enhancement
ordinal: 276000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Manage access dialog renders its title in a plain card title and offers only a bottom-right text Close action, while the Edit event dialog family (Edit event, New sign up, New group) uses the shared editor-dialog header with a top-right cross close control. The mismatch makes Manage access feel like a different surface and puts the dismiss action far from the title. Align the Manage access header with the editor-dialog header so the two surfaces read as one family.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Manage access dialog title uses the shared editor-dialog header treatment, matching the Edit event dialog title's typography, spacing, and alignment.
- [x] #2 The dialog no longer renders the bottom-right Close text action.
- [x] #3 The dialog renders a top-right close icon control, matching the Edit event dialog close control, that dismisses the dialog and carries an accessible name.
- [x] #4 The dialog keeps its accessible name via aria-labelledby pointing at the Manage access title, and the title keeps the manage-access-title id.
- [x] #5 The dialog body aligns horizontally with the header on compact and wide viewports.
- [x] #6 Unit tests cover the header structure, the title id, the close control's accessible name, close behavior, and the removal of the bottom action; existing Manage access behavior tests still pass.
- [x] #7 Existing step content, statuses, polling, and control labels are unchanged.
- [x] #8 The dialog card keeps the same top padding above the title as the Edit event dialog.
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
## Researched context (current system)

- `frontend/src/components/event/EventAccessTransfer.vue` renders its own `v-card-title` with `<h2 id="manage-access-title" class="tw:text-xl">` and a bottom `v-card-actions` text Close button; the dialog supplies `aria-labelledby="manage-access-title"` through `content-props`.
- `frontend/src/components/EditorDialogHeader.vue` is the shared editor header (NewEvent/NewSignUp/NewGroup): `v-card-title tw:mb-2 tw:flex tw:gap-2 tw:px-4 tw:sm:px-8`, title in a plain div with `tw:mb-1`, then a spacer and, in dialog mode with `showHelp=false`, a gray text-variant icon button emitting `close`.
- NewEvent's `v-card-text` uses `tw:px-4 tw:py-1 tw:sm:px-8`, so body content aligns with the header; EventAccessTransfer's `v-card-text` has no horizontal padding override.
- `EditorDialogHeader` has no title-id prop and its icon close button has no accessible name.
- Tailwind preflight is not enabled (only theme + utilities in `index.css`), so a native `h2` inside the card title would pick up UA bold/1.5em sizing and no longer match the editor title (Vuetify 4 `.v-card-title` is 1.375rem/400). Reusing the shared header is the only way to guarantee the requested match.
- `EventAccessTransfer.test.ts` mounts the real component with Vuetify stubs (VBtn renders `<button><slot /></button>` and inherits attrs), asserts `#manage-access-title` and the dialog's `aria-labelledby`, and clicks the bottom "Close" text action at line 624. `EditorDialogHeader.test.ts` shallow-mounts with VBtn/VIcon stubs that pass attrs through.

## Plan

1. `EditorDialogHeader.vue`
   - Add an optional `titleId?: string` prop and bind it to the title div so callers can keep an `aria-labelledby` target.
   - Give the close icon button an accessible name (`aria-label="Close"`).
2. `EventAccessTransfer.vue`
   - Replace the local title markup with `EditorDialogHeader` (`title="Manage access"`, `title-id="manage-access-title"`, empty subtitle/help header, `:dialog="true"`, `:show-help="false"`, `:hide-dialog-actions="false"`, `@close="dialog = false"`).
   - Delete the bottom `v-card-actions` Close button.
   - Add `tw:px-4 tw:sm:px-8` to the dialog's `v-card-text` so body content aligns with the shared header.
   - Keep `content-props` `aria-labelledby`, all step content, statuses, handlers, and two-second polling unchanged.
3. Tests
   - `EventAccessTransfer.test.ts`: stub `HelpDialog`; replace the bottom-Close click with the top-right `button[aria-label="Close"]`; add assertions for the shared header (title id text, named close control, no bottom Close action).
   - `EditorDialogHeader.test.ts`: cover `titleId` binding and the close button accessible name.
4. Verification
   - `cd frontend && npm run lint && npm run fmt:check && npm run typecheck && npm run build && npm run test:unit`.
   - Targeted Firefox e2e `npm run test:e2e -- --project=firefox-desktop specs/timed-event-access-transfer-firefox.spec.ts` from `e2e/` (streamed output) to confirm the real dialog still renders and behaves.
   - `graphify update .`.
   - No server, swagger, or contract changes expected.

## Risks
- Swapping the native `h2` for the shared header div removes the heading role; the dialog keeps its accessible name via `aria-labelledby`, and step `h3` headings are unchanged. This is the accepted trade-off for the requested editor-header consistency.
- The HelpDialog rendered by the shared header appears in unit tests; it must be stubbed to keep existing text assertions stable.

User follow-up during execution: the Edit event dialog card adds `tw:pt-4` above the header, so Manage access gets the same `tw:pt-4` on its dialog card for top-spacing parity.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: replaced the local Manage access title with the shared EditorDialogHeader, added an optional titleId plus an aria-label on the shared close button, removed the bottom v-card-actions Close, and matched editor body/card padding (`tw:px-4 tw:sm:px-8`, `tw:pt-4`).

Evidence: focused suites 32 passed; full frontend suite 149 files / 1157 tests passed; lint (0 errors, 2 pre-existing NewSignUp warnings), fmt:check, typecheck, and build passed; e2e package lint, fmt:check, typecheck passed; Firefox access-transfer spec 8/8 passed including a new step that checks title/body x alignment, clicks the named top-right close, and asserts the dialog hides (logs: /tmp/opencode/task-0276-e2e-firefox-final.log, /tmp/opencode/task-0276-e2e-close.log); graphify update . run.

Known trade-off: the shared header renders the title as a div, so the former native h2 heading role is gone; the dialog still has its accessible name via aria-labelledby and the step h3 headings are unchanged.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The Manage access dialog now uses the same header treatment as the Edit event dialog instead of a plain title plus a bottom-right text Close action.

### Header (`frontend/src/components/EditorDialogHeader.vue`)
- Added an optional `titleId` prop bound to the title element so callers can keep an `aria-labelledby` target.
- Gave the icon-only close button an accessible name (`aria-label="Close"`), which also covers the editor dialogs that share the component.

### Manage access dialog (`frontend/src/components/event/EventAccessTransfer.vue`)
- Replaced the local `v-card-title`/`h2` with `EditorDialogHeader` (`title="Manage access"`, `title-id="manage-access-title"`, `:dialog="true"`, `:show-help="false"`, `@close="dialog = false"`), giving the title the identical typography, spacing, and top-right cross control as Edit event.
- Removed the bottom `v-card-actions` Close button.
- Matched the editor body indent (`tw:px-4 tw:sm:px-8`) and added the editor card top padding (`tw:pt-4`) requested during execution.
- All step content, statuses, handlers, two-second polling, and labels are unchanged.

### Tests
- `EditorDialogHeader.test.ts`: covers `titleId` binding and the close control accessible name.
- `EventAccessTransfer.test.ts`: asserts the shared header, title id, card top padding, named close control, absence of a bottom Close text action, and closes through the top-right control in the existing polling test.
- `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: new step in the cancel journey asserting the title and body share the same x position, then dismissing the dialog through the named top-right close control.

## Verification
- Focused unit suites: 32 passed.
- Full frontend suite: 149 files / 1157 tests passed.
- `npm run lint` (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), `npm run fmt:check`, `npm run typecheck`, `npm run build` passed.
- e2e `lint`, `fmt:check`, `typecheck` passed.
- Firefox desktop access-transfer spec: 8/8 passed with the new browser assertion (2.3m, log `/tmp/opencode/task-0276-e2e-firefox-final.log`).
- `graphify update .` run.

## Notes
- The title is no longer a native `h2` because the shared editor header renders a div; the dialog keeps its accessible name through `aria-labelledby="manage-access-title"`, and the step `h3` headings are unchanged. This is the accepted trade-off for matching the Edit event header exactly.
- No server, swagger, ADR, or contract changes.
<!-- SECTION:FINAL_SUMMARY:END -->
