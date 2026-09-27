---
id: TASK-0271
title: Simplify the Manage access transfer controls into one action row
status: Done
assignee:
  - opencode
created_date: '2026-09-19 17:08'
updated_date: '2026-09-19 17:21'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
documentation:
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 271000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Manage access dialog currently stacks three controls that read as disconnected steps: a Create transfer link button, an always-visible readonly Transfer link field, and a Copy transfer link button. Users cannot tell what to press first, and the link field looks like an input they should fill in.

There is also a correctness gap behind the confusion: nothing stops the user from creating several pending transfers, and the dialog can only approve the newest one, so earlier links become dead ends even though they still open on the other browser.

Keep the security model unchanged (source-confirmed, single-use, five-minute link) while collapsing the controls into one coherent action row: create clearly means mint a fresh link, copy clearly means copy the current link.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Manage access dialog presents one action row containing Create new transfer link with a refresh icon and Copy link with a copy icon.
- [x] #2 Copy link is visible only while the current transfer is pending or approved, and clicking it copies the current transfer link without creating a server-side transfer.
- [x] #3 The readonly Transfer link field stays visible below the action row whenever a transfer exists, so a denied clipboard write still leaves the link selectable with the existing manual-copy guidance.
- [x] #4 Clicking Create new transfer link while the current transfer is pending or approved cancels that transfer before creating the new one, so at most one transfer is live at a time.
- [x] #5 Clicking Create new transfer link when the current transfer is expired, cancelled, redeemed, or unavailable creates a new transfer without cancelling anything.
- [x] #6 Transfer status, matching-code approval, Cancel transfer, granted-access revocation, two-second status polling, and restore-after-reload behavior are unchanged.
- [x] #7 Unit tests cover creating a replacement link, copy without a new transfer, the terminal-state create path, and Copy link visibility; the existing tests are updated to the new labels.
- [x] #8 The Firefox access-transfer e2e spec passes with the new control labels and still covers cancel and expired-link journeys.
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

- `frontend/src/components/event/EventAccessTransfer.vue`: `start()` creates and tracks a transfer; `copy()` is a local clipboard write; the template stacks button (line 23), readonly field (25), copy button (26-28). `refresh()` polls tracked transfers and, on dialog open, promotes the newest pending/approved saved transfer to `current`.
- `frontend/src/composables/transfer/transferBoundary.ts`: `transferAction(..., "cancel")` is source-proof protected; `isTransferUnavailable` maps 403/404; `rememberTransfer`/`forgetTransfer` persist numbering.
- `server/routes/transfers.go`: create mints a single-use five-minute transfer; cancel is allowed while pending or approved-but-unredeemed; terminal transitions return 403/404.
- Unit tests stub `VBtn`, so `prepend-icon` does not need a Vuetify plugin.
- The access-transfer e2e spec runs only in the `firefox-desktop` project; Playwright clipboard permissions are Chromium-only, so copy success is not asserted in Firefox e2e.

## Plan

1. `EventAccessTransfer.vue`:
   - Import `MdiRefresh from "~icons/mdi/refresh"` and `MdiContentCopy from "~icons/mdi/content-copy"`.
   - Template: one flex action row with Create new transfer link (`:prepend-icon="MdiRefresh"`, disabled while busy) and Copy link (`:prepend-icon="MdiContentCopy"`, rendered only when `canCopy`). Move the readonly Transfer link field directly below the row, still gated on `currentId`. Remove the old standalone copy button.
   - `canCopy` = current state is `pending` or `approved`.
   - `start()`: if `current` is `pending`/`approved`, cancel it via `transferAction` and `updateTransfer` before creating. If cancel fails with a 403/404 (`isTransferUnavailable`), `refresh()` and proceed only when the transfer is no longer pending/approved; rethrow otherwise. Cancel failures for other reasons abort the replacement. Then `createTransfer`, set `current`/`currentId`, `rememberTransfer`, reset `code`/`copied`.
   - `copy()` unchanged (local only, error guidance already names manual copy).
2. Unit tests `EventAccessTransfer.test.ts`: update labels (`Create new transfer link`, `Copy link`) and add coverage for replacement create (cancel then create, one tracked transfer), copy without a server call, create after a terminal transfer without cancel, and Copy link visibility for live vs terminal states.
3. E2E `timed-event-access-transfer-firefox.spec.ts`: update the two create-button locators; the readonly link field reads and reload-restore assertions stay valid. No copy click assertion in Firefox.
4. Run frontend required checks, the targeted Firefox access-transfer spec, and `graphify update .`.

Layout refinement during user review: the row centers Create when Copy is absent, and `tw:flex-1` gives both buttons equal width when Copy is visible.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the action row with `MdiRefresh` and `MdiContentCopy` passed through Vuetify's `:prepend-icon`; when Copy is absent the Create button centers, and when both are present each button takes half the row via `tw:flex-1`.

`start()` cancels the current pending/approved transfer before creating a replacement. If the cancel returns 403/404, it refreshes status and only proceeds when the old transfer is no longer live, so a redeem race cannot strand the click.

Verification: frontend lint, fmt:check, typecheck, build, and the full 1134-test unit suite pass; the Firefox access-transfer spec passes 8/8 after the layout tweak; `graphify update .` ran.

New unit tests: replacement create issues cancel then create and tracks one transfer; Copy does not POST; a terminal transfer creates without cancel; Copy link hides once the transfer is terminal.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The Manage access dialog no longer stacks a Create button, a readonly link field, and a Copy button as three disconnected steps. A single centered action row now contains **Create new transfer link** (refresh icon) and, while a transfer is pending or approved, **Copy link** (copy icon); both share the row width equally (`tw:flex-1`) and the readonly Transfer link field stays directly below.

## Why

The old layout read as an unclear sequence and let users mint several pending transfers, while the dialog could only ever approve the newest one, so earlier links became dead ends.

## Behavior

- `Create new transfer link` cancels the current pending/approved transfer before creating its replacement, so at most one transfer is live. A 403/404 cancel response refreshes status first and only proceeds when the old transfer is no longer live (e.g. redeemed in a race).
- `Copy link` is local-only (no server call) and is hidden once the transfer reaches a terminal state.
- Status, matching-code approval, Cancel transfer, revocation, polling, and reload restoration are unchanged.

## Files

- `frontend/src/components/event/EventAccessTransfer.vue`
- `frontend/src/components/event/EventAccessTransfer.test.ts`
- `e2e/specs/timed-event-access-transfer-firefox.spec.ts`

## Verification

- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` pass; the 2 pre-existing `NewSignUp.test.ts` lint warnings remain.
- `npm run test:unit`: 149 files, 1134 tests pass, including four new tests (replacement create, copy without a transfer, terminal-state create, Copy link visibility).
- Firefox access-transfer e2e: 8/8 pass in 2.2m after the final layout tweak, covering guest/owner/signed-in approval, reload restoration, cancel, and cancelled/expired links.
- `graphify update .` run.

## Exempt DoD items

No authored Markdown, Swagger annotations, `scripts/`/`prettier/` files, or contract documents were changed.
<!-- SECTION:FINAL_SUMMARY:END -->
