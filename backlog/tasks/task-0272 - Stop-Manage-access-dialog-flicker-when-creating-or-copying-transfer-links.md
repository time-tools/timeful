---
id: TASK-0272
title: Stop Manage access dialog flicker when creating or copying transfer links
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 17:35'
updated_date: '2026-09-19 17:41'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
documentation:
  - docs/design/architecture/adr/ADR-010.md
modified_files:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
priority: medium
type: bug
ordinal: 272000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Manage access dialog flickers during the two transfer actions shipped in TASK-0271. Clicking Copy link briefly disables Create new transfer link, so the create button visibly blinks while the clipboard write is pending. Clicking Create new transfer link while a live transfer exists first cancels that transfer, and the intermediate cancelled render unmounts the Copy link button plus the matching-code and Cancel transfer controls, collapsing the centered card and making the whole modal jump before the replacement appears. The jump is intermittent only in how long the intermediate state stays visible, not in whether it is rendered.

A Firefox DOM-level reproduction against the local dev stack ruled out Vite/HMR and dialog remounts: the overlay and card stay mounted, the scrim opacity is constant, and only card content, button widths, and disabled attributes change. The measured replacement path collapsed the card from 712px to 510px and moved its top position 94 to 195 and back within about 40ms.

The dialog should keep the live transfer continuously rendered while it is being replaced, and a copy action should not change the Create new transfer link button's visual state.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Clicking Copy link does not disable, dim, or otherwise visually change the Create new transfer link button while the clipboard write is pending.
- [x] #2 Creating a replacement while a pending or approved transfer is live keeps the existing transfer's status text, Copy link, matching-code input, and Cancel transfer controls continuously rendered until the replacement is created; no intermediate terminal state is painted.
- [x] #3 If the replacement create fails after the existing transfer was cancelled, the dialog reconciles to the cancelled state immediately and still shows the create failure message.
- [x] #4 At most one transfer stays live: the cancel-before-create order and the 403/404 refresh-and-recheck behavior are unchanged.
- [x] #5 Unit tests cover that Create new transfer link stays visually enabled while a copy is pending and that no terminal intermediate state is rendered between cancel and replacement create.
- [x] #6 Existing Manage access behavior, including two-second status polling, matching-code approval, cancel, revoke, reload restoration, and clipboard failure guidance, is unchanged.
- [x] #7 The Firefox access-transfer e2e spec still passes.
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

- `frontend/src/components/event/EventAccessTransfer.vue`

  - `run()` toggles a single `busy` ref for every action. The Create button binds `:disabled="busy"`, and the Copy button has no disabled binding, so starting a copy disables/dims Create.
  - `start()` cancels the live transfer, calls `updateTransfer`, which synchronously sets `current.value` to the cancelled state before awaiting `createTransfer`. That paints a terminal intermediate state, unmounting Copy link plus the matching-code/Cancel transfer controls and collapsing the card.
  - `copy()` routes through `run()`, even though it performs no server mutation and is purely local.

- Vuetify's disabled elevated `v-btn` changes text/icon to a 26%-alpha color. Reproduced in Firefox against the dev stack: the create button's computed color flips `rgba(0,0,0,0.87)` to `rgba(0,0,0,0.26)` during a copy; the card collapses `712px -> 510px` and top moves `94 -> 195 -> 94` during a replacement create. The dialog/overlay never remounts and the scrim opacity is constant, so this is not Vite/HMR.

- Unit tests in `frontend/src/components/event/EventAccessTransfer.test.ts` stub `VBtn` as a plain `button` with attribute fallthrough, so `disabled` is observable; `post` is mocked and can be made deferred.

- ADR-010 requires the cancel-before-create ordering so at most one transfer is live; the fix must not change that ordering or the 403/404 refresh-and-recheck path.

## Plan

1. `EventAccessTransfer.vue`
   - Add a local `copying` ref and make `copy()` own its clipboard work directly: guard re-entry, clear `error`, write the link, set `copied`, set the clipboard failure message on rejection, and reset `copying` in `finally`. It must no longer set the shared `busy`, so Create's disabled state is untouched by a copy.
   - Keep the Create button on `:disabled="busy"` (its own feedback). The Copy button may bind `:disabled="copying"` so only the clicked button reflects its pending state.
   - In `start()`, defer the cancel bookkeeping: capture the awaited cancel state plus its `SavedTransfer` entry, await `createTransfer`, then call `updateTransfer(entry, state)` and assign `current`/`currentId` in one synchronous block so Vue paints only the replacement state. If `createTransfer` rejects after a successful cancel, the still-tracked entry lets the next status refresh reconcile `current` to cancelled while retaining the create failure message.
2. Unit tests
   - Add a regression test that holds a pending `navigator.clipboard.writeText` promise and asserts Create new transfer link has no `disabled` attribute (and no visual change), then resolves and asserts `Copied`.
   - Add a regression test using deferred cancel and create responses that asserts the live transfer controls (Waiting for approval status, Copy link, matching-code input, Cancel transfer, old link value) stay rendered after the cancel resolves and before the create resolves; then resolves the create and asserts the replacement link.
   - Confirm the new tests fail against the current implementation before applying the fix.
3. Verification
   - `cd frontend && npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`.
   - Focused Firefox access-transfer e2e spec (`e2e`, `--project=firefox-desktop specs/timed-event-access-transfer-firefox.spec.ts`).
   - `graphify update .` afterwards.

## Risks

- Deferring cancel bookkeeping changes the failure path after a cancelled-but-not-replaced transfer; the next 2-second status poll must reconcile the UI to Cancelled. Cover with the existing error-path behavior and rely on `refresh()` since the entry stays tracked.
- Reordering must keep one live transfer at a time (cancel still precedes create) and preserve the 403/404 redeem-race handling.

Implementation deviation from the drafted risk note: the create-failure path reconciles the cancelled transfer immediately in `start()`'s catch through the existing `updateTransfer` bookkeeping instead of waiting for the 2-second status refresh. This keeps the failure UI accurate at the moment the error appears; the entry is also pruned from tracking as before.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the fix on `fix-frontend`.

- `EventAccessTransfer.vue`: `copy()` no longer routes through the shared `busy` flag. It owns its clipboard work with a local `copying` guard, clears the error, and sets the existing clipboard-failure guidance. The Create button's `:disabled="busy"` is therefore untouched by a copy.
- `start()` now captures the awaited cancel result plus its tracked entry in `retired` and applies `updateTransfer` together with the new `current`/`currentId` in one synchronous block after `createTransfer` resolves (or in the catch when it rejects). Vue paints once, so no terminal intermediate state is rendered; the failure path still reconciles to Cancelled and shows the create failure message.
- Added three unit tests in `EventAccessTransfer.test.ts`: live transfer stays visible while its replacement is created, cancelled transfer reconciles when the replacement create fails, and Create stays enabled while a copy is pending. The first two new tests failed against the old implementation before the fix (the copy one failed with `disabled=""`).

Focused unit run: 20/20 pass. Full unit suite: 149 files, 1137 tests pass. lint/fmt:check/typecheck/build pass.

Browser verification against the dev stack (Firefox, same DOM instrumentation used for the diagnosis): replacement create now holds the card at 712px height and top 94px with Copy link continuously present, versus 712 to 510 and top 94 to 195 before; Copy click produces no create-button disabled/color change. The dialog never remounts and the scrim stays constant.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The Manage access dialog no longer flickers while creating or copying transfer links.

## Why

Both actions shared one `busy` flag, so starting a local clipboard copy disabled and dimmed Create new transfer link, and replacing a live transfer rendered the intermediate cancelled state, which unmounted Copy link plus the matching-code and Cancel transfer controls and collapsed the centered card (measured 712px to 510px, top 94 to 195 and back).

## Behavior

- `copy()` now owns its clipboard write with a local `copying` guard instead of the shared `busy` flag, so the Create button's disabled state is untouched by a copy. The existing clipboard failure guidance and `Copied` feedback are preserved.
- `start()` captures the cancelled transfer plus its cancel state in `retired` and applies the bookkeeping and the new `current`/`currentId` in one synchronous block once `createTransfer` resolves, so Vue paints only the replacement state. If the create fails after a successful cancel, the same bookkeeping runs in the catch, so the dialog shows Cancelled immediately along with the create failure message.
- The cancel-before-create ordering, the 403/404 refresh-and-recheck redeem race handling, polling, approval, revocation, and reload restoration are unchanged.

## Files

- `frontend/src/components/event/EventAccessTransfer.vue`
- `frontend/src/components/event/EventAccessTransfer.test.ts`

## Verification

- Added three unit tests: live transfer stays visible while its replacement is created, cancelled transfer reconciles when the replacement create fails, and Create stays enabled while a copy is pending. The latter two failed against the old implementation before the fix.
- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` pass; `npm run test:unit` passes with 149 files and 1137 tests.
- Firefox access-transfer e2e: 8/8 pass in 2.4m.
- Firefox DOM instrumentation against the dev stack confirms the card stays at 712px height and top 94px with Copy link continuously present during a replacement create, and a copy produces no create-button disabled or color change.
- `graphify update .` ran.

## Exempt DoD items

No authored Markdown, Swagger annotations, `scripts/` or `prettier/` files, or contract documents were changed; `backlog/**` is excluded from `format:markdown`, and the task record is Backlog-managed.
<!-- SECTION:FINAL_SUMMARY:END -->
