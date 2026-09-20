---
id: TASK-0273
title: Scope Manage access action disabled state to the running action
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 17:41'
updated_date: '2026-09-19 17:46'
labels: []
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
  - >-
    backlog/tasks/task-0272 -
    Stop-Manage-access-dialog-flicker-when-creating-or-copying-transfer-links.md
modified_files:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
priority: medium
type: bug
ordinal: 273000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After TASK-0272 stopped the Manage access dialog from collapsing while a replacement transfer is created, the remaining flicker became visible: Cancel transfer dims while Create new transfer link is pending. Cancel transfer, Approve matching code, and Revoke access all bind their disabled state to the single shared busy flag, so any action visually disables every action control in the dialog for the duration of its network round trips. The card is now stable, so the dimming is the visible artifact.

Each action control should reflect only its own pending operation, matching the treatment Copy link received in TASK-0272, while the shared guard continues to ignore overlapping action clicks.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Clicking Create new transfer link does not disable, dim, or otherwise visually change Cancel transfer, Approve matching code, or Revoke access while the create is pending.
- [x] #2 Each action control reflects only its own pending operation: Cancel transfer during cancel, Approve matching code during approve, Revoke access during revoke, and Create new transfer link during create.
- [x] #3 The existing reentrancy guard still ignores an action click while a different action is in flight.
- [x] #4 Existing behavior, including refresh on dialog open, two-second status polling, cancel-before-create replacement, 403/404 redeem handling, and failure messages, is unchanged.
- [x] #5 Unit tests cover that the other action controls stay enabled while a replacement create is pending.
- [x] #6 The Firefox access-transfer e2e spec still passes.
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

- `frontend/src/components/event/EventAccessTransfer.vue` binds `:disabled="busy"` on Create new transfer link, Approve matching code (`busy || !code`), Cancel transfer, and Revoke access. `busy` is set by `run()` for every action, so any single pending action dims every action control.
- TASK-0272 already decoupled Copy link from `busy` with a local `copying` guard and scoped Create's disabled to `busy`; the shared guard in `run()` still ignores overlapping action clicks.
- `run(refresh, ...)` is also used by `openDialog` for the background status refresh, which should not visually disable any control.
- Unit tests stub `VBtn` with attribute fallthrough, so per-button `disabled` is observable.

## Plan

1. Add a `busyAction` ref alongside `busy`; `run(action, work, failureMessage, clearError)` sets both and clears them in `finally`. Add an explicit action union covering `refresh`, `start`, `approve`, `cancel`, and `revoke`.
2. Update call sites: `openDialog` refresh runs as `refresh`, `start` as `start`, `approve` as `approve`, `cancel` as `cancel`, `revoke` as `revoke`.
3. Template: bind Create to `busyAction === 'start'`, Approve to `busyAction === 'approve' || !code`, Cancel to `busyAction === 'cancel'`, Revoke to `busyAction === 'revoke'`. Copy stays on its local `copying` guard.
4. Add a regression unit test that holds a replacement create pending and asserts Create is disabled for its own action while Cancel transfer and (with a code entered) Approve matching code stay enabled; confirm it fails before the fix.
5. Verify: focused unit file, full unit suite, lint, fmt:check, typecheck, build, focused Firefox access-transfer e2e, `graphify update .`.

## Risks

- Buttons from other actions stay visually enabled while a different action is in flight but remain inert through the `run()` guard; this matches the Copy link treatment from TASK-0272.
- The `!code` condition must remain on Approve so an empty matching code still disables it.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented on `fix-frontend`.

- Added `busyAction` (`refresh | start | approve | cancel | revoke`) next to the shared `busy` guard. `run()` now takes the action name, sets `busyAction` for the duration, and clears it in `finally`; the reentrancy guard is unchanged.
- Bound each control to its own action: Create to `busyAction === 'start'`, Approve to `busyAction === 'approve' || !code`, Cancel to `busyAction === 'cancel'`, and Revoke to `busyAction === 'revoke'`. Copy keeps its local `copying` guard. The dialog-open status refresh runs as `refresh` and disables nothing.
- Added a regression test that holds a replacement create pending and asserts Create is disabled for its own action while Cancel transfer and Approve matching code (with a code entered) stay enabled. It failed against the old implementation with Cancel carrying `disabled=""`.

Focused unit file: 21/21 pass. Full suite: 149 files, 1138 tests pass. lint/fmt:check/typecheck/build pass. Firefox access-transfer e2e: 8/8 pass in 2.2m. `graphify update .` ran.

Browser instrumentation during a replacement create: Cancel transfer now stays `disabled=false` with color `0.87` throughout, versus flipping to disabled and `0.26` before; only Create dims for its own pending operation.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

Every Manage access action control now reflects only its own pending operation instead of the shared busy flag.

## Why

After TASK-0272 kept the dialog card stable during a replacement create, the remaining artifact was visible: Cancel transfer (and Approve matching code / Revoke access) dimmed whenever any other action was running, because all of them bound `:disabled="busy"`.

## Behavior

- `run()` takes an action name and tracks it in a new `busyAction` ref (`refresh | start | approve | cancel | revoke`) while the `busy` reentrancy guard stays unchanged.
- Create binds to `busyAction === 'start'`, Cancel to `'cancel'`, Revoke to `'revoke'`, and Approve to `'approve' || !code`. Copy keeps its local `copying` guard from TASK-0272, and the dialog-open status refresh disables nothing.
- Clicking Create no longer changes the disabled or computed color of Cancel transfer, Approve matching code, or Revoke access.

## Files

- `frontend/src/components/event/EventAccessTransfer.vue`
- `frontend/src/components/event/EventAccessTransfer.test.ts`

## Verification

- Added a regression test holding a replacement create pending: Create is disabled for its own action while Cancel transfer and Approve matching code (with a code entered) stay enabled. It failed against the old implementation.
- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` pass; `npm run test:unit` passes with 149 files and 1138 tests.
- Firefox access-transfer e2e: 8/8 pass in 2.2m.
- Firefox DOM instrumentation during a replacement create shows Cancel transfer stays enabled with color `0.87` throughout, versus flipping to disabled and `0.26` before; only Create dims for its own operation.
- `graphify update .` ran.

## Exempt DoD items

No authored Markdown, Swagger annotations, `scripts/` or `prettier/` files, or contract documents were changed; `backlog/**` is excluded from `format:markdown`.
<!-- SECTION:FINAL_SUMMARY:END -->
