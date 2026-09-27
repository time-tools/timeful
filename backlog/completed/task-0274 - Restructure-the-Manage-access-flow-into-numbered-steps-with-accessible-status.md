---
id: TASK-0274
title: Restructure the Manage access flow into numbered steps with accessible status
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 18:00'
updated_date: '2026-09-19 18:12'
labels:
  - frontend
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/components/event/EventAccessTransfer.test.ts
  - frontend/src/views/AccessTransfer.vue
  - frontend/src/views/AccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
documentation:
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 274000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Manage access dialog explains a two-sided browser flow in one dense paragraph and then renders every control at once, so users cannot tell the sequence: which browser acts when, what to press first, and when a matching code is expected. The code input appears even before any other browser has opened the link, and the terse `Transfer status` line does not use the target-request information the dialog already polls, so the source browser never learns that the other browser is showing a code. The target page uses different framing, so the two-sided flow is hard to follow end to end.

Accessibility needs a pass as well: the dialog exposes `role="dialog"` with no accessible name, the card title is not a heading, step content has no heading structure or current-step indication, Copy success is only a button-label swap, and the dialog is not scrollable on short viewports.

The user wants the flow presented as three simple, always-visible steps on both sides, with the live status following what the other browser is doing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Manage access dialog presents three always-visible numbered steps with the active step indicated: 1) create and copy a transfer link, 2) open the link in the other browser, 3) enter its code and approve.
- [x] #2 Step 3 code input and approve control render only while a transfer is pending, and step controls are never gated on polled target-request data so no approve race is introduced.
- [x] #3 The dialog status reflects live transfer state: pending with no target request says the other browser has not opened the link yet, pending with a target request says the other browser is showing a code and points to step 3, approved says to finish in the other browser, and terminal states keep their current meaning.
- [x] #4 The dialog keeps the required flow facts: the link expires five minutes after creation, opening the link alone gives no access, the source approves the exact matching code, and granted access can be revoked.
- [x] #5 The dialog exposes an accessible name, uses ordered-list steps with heading elements and `aria-current` on the active step, announces status changes and Copy success politely, gives the code field a hint and character capitalization, and scrolls on short viewports.
- [x] #6 The target-side Continue on this device page is aligned to the same numbered step language while keeping the matching-code test hook, the approval status, the account-switch consent copy, and the Continue after approval label.
- [x] #7 Existing labels and behavior are unchanged: Manage access, Create new transfer link, Copy link, Approve matching code, Cancel transfer, Revoke access; replacement-create, cancel, revoke, two-second polling, and restore-after-reload still work.
- [x] #8 Unit tests cover step structure and order, active-step transitions for no-request, request, approved, and terminal states, contextual status copy, accessibility attributes, and target-page alignment; the Firefox access-transfer e2e spec passes with the new structure.
- [x] #9 The target page polls the existing open action while a matching code is waiting and, when the source approves, marks step 3 current and announces 'Approved — you can continue' without requiring a reload or an early Continue click; polling stops after approval, on unmount, or when the initial open failed, and never redeems without a click.
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

- `frontend/src/components/event/EventAccessTransfer.vue` renders a dense intro paragraph, one action row (Create new transfer link / Copy link), a readonly Transfer link field, a single `role="status"` line, a conditional code input + Approve matching code, Cancel transfer, and Granted access rows. `refresh()` polls status every two seconds and `decodeTransfer` already surfaces `requests` (target browser codes); `approve()` calls `refresh()` before matching, so the code input must not be gated on locally polled `requests`.
- `frontend/src/views/AccessTransfer.vue` is the target-side page: matching code via `data-testid="matching-code"`, `Approved — you can continue` status, `Continue after approval` button, account-switch consent dialog.
- Unit tests stub Vuetify components; the `VCard` stub in `EventAccessTransfer.test.ts` drops the `title` prop, so title markup placed in the default slot becomes visible to `wrapper.text()`.
- The Firefox access-transfer e2e spec targets labels and roles; `frontend/src/views/Event.test.ts` asserts header action labels that must not change.
- Shared semantic tokens live in `frontend/src/index.css` (`--timeful-selection-bg`, `--timeful-muted-foreground`, `--timeful-disabled-foreground`, `--timeful-outline-neutral`).

## Plan

1. `EventAccessTransfer.vue`
   - Replace the intro paragraph with a one-line lead plus an `<ol>` of three always-visible steps: 1) Create and copy a transfer link (action row, readonly link, status, Cancel transfer), 2) Open the link in the other browser (static guidance), 3) Enter its code and approve (outcome line, pending input + approve, or approved message).
   - Add `activeStep` computed (presentational only): no live transfer → 1; pending and no target request and not copied → 1; pending, copied without request → 2; pending with request or approved → 3.
   - Contextual status copy from `current.state` + `requests.length`: pending/no request → waiting for the other browser to open the link; pending/request → other browser is showing a code, enter it in step 3; approved → finish in the other browser; keep terminal labels.
   - Accessibility: `scrollable` dialog with `aria-labelledby` and a real `<h2>` title; `<h3>` step headings with `aria-hidden` number badges; `aria-current="step"` on the active `<li>`; one polite status region; sr-only polite Copy confirmation; code field hint plus `autocapitalize="characters"` and `spellcheck="false"`; "Granted access" section `<h3>`.
   - Keep all existing labels, handlers, and polling behavior unchanged.

2. `AccessTransfer.vue`: add matching `Step 2 — show this code` / `Step 3 — approve it in the browser that created the link` structure with heading semantics and a polite status, keeping `matching-code`, `Approved — you can continue`, `Continue after approval`, and all consent copy.

3. Tests
   - `EventAccessTransfer.test.ts`: update copy assertions; add step-order, active-step transitions (no request, request, approved, terminal), contextual status, accessibility attribute, and copy-announcement tests.
   - `AccessTransfer.test.ts`: update text expectations for the new step headings, keep behavior cases.
   - `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: add step/status assertions where stable; keep existing label/role locators working; avoid strict-mode collisions with new copy.

4. Verification: frontend lint, fmt:check, typecheck, build, test:unit; targeted Firefox `timed-event-access-transfer-firefox` spec with streamed output; `graphify update .`. No server, swagger, or contract changes; ADR-010 describes the security model and needs no edit.

## Risks
- Status copy assertions and e2e text matchers must be updated in lockstep.
- Active-step highlighting can lag one two-second poll behind the target browser opening the link; acceptable because controls are never gated on it.
- New step headings must not collide with strict Playwright locators (use exact names where needed).

## Scope addition approved during review: target-side approval detection

User confirmed the target browser should highlight step 3 automatically instead of learning approval only from a reload or an early Continue click, and chose to expand this task rather than split a follow-up.

- `frontend/src/views/AccessTransfer.vue`: poll `transferAction(..., 'open')` every 2s while `code` exists and `approved` is false; on `state === 'approved'` set `approved = true`, clear any stale early-continue error, and stop polling; ignore poll failures after the first successful open; clear the interval on unmount. The step-3 status becomes a persistent live region whose text flips from a waiting message (including 'You have no transferred access until then.') to 'Approved — you can continue' so the change is announced.
- Deliberate limit: polling ignores cancellation/expiry (403/404); those still surface on the Continue click as today.
- Tests: `AccessTransfer.test.ts` fake-timer coverage (pending to approved highlight, polling stops, initially-approved pages never poll, transient poll failure keeps waiting, early error clears on approval); e2e asserts the target reaches step 3 after source approval without a reload in the main journey.
- No server, swagger, ADR, or contract changes; `open` already serves approved-but-unredeemed transfers to the proof-holding target.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the approved plan plus the target-side approval-detection addition. Design decisions: the source dialog keeps control labels untouched so existing E2E and unit locators stay valid; the active-step highlight is presentational and never gates controls (the approve path refreshes before matching, so a stale local requests list cannot block approval); the target page reuses the existing idempotent `open` action for polling instead of adding a server endpoint, so no server or swagger changes were needed.

Verification evidence: focused `EventAccessTransfer` + `AccessTransfer` suites 39 passed; full `npm run test:unit` 149 files / 1152 tests passed; `npm run lint` 0 errors with the 2 pre-existing `NewSignUp.test.ts` one-component-per-file warnings; `fmt:check`, `typecheck`, and `build` passed; e2e package `lint`, `fmt:check`, and `typecheck` passed; `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop specs/timed-event-access-transfer-firefox.spec.ts` passed 8/8 in 1.9m with log at `/tmp/opencode/task-0274-e2e-firefox.log`; `graphify update .` rebuilt the graph.

Known limit: target polling ignores cancellation/expiry and keeps waiting; those states still surface on the Continue click. If we want the waiting page to end on its own, that would be a separate follow-up.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The Manage access dialog and the target `Continue on this device` page now present the two-sided access transfer as three numbered, always-visible steps instead of one dense paragraph plus a stacked control list.

### Source dialog (`frontend/src/components/event/EventAccessTransfer.vue`)
- Steps: 1) Create and copy a transfer link, 2) Open the link in the other browser, 3) Enter its code and approve; the active step is marked with `aria-current="step"` and token-based highlighting, and step 1 keeps the action row, readonly link, status, and Cancel transfer.
- Contextual status reuses the already-polled target requests: waiting for the other browser to open the link, "The other browser is showing a code — enter it in step 3.", approved, or the terminal labels.
- Accessibility: `scrollable` dialog named via `aria-labelledby` on a real `h2`, ordered-list steps with `h3` headings and sr-only step prefixes, persistent polite status region, sr-only polite copy announcement, code-field hint with `autocapitalize="characters"` and `spellcheck="false"`, and a "Granted access" section heading.
- Labels and behavior (replacement-create, cancel, revoke, 2s polling, reload restore) are unchanged.

### Target page (`frontend/src/views/AccessTransfer.vue`)
- Aligned to steps 2 and 3 with the same badge/active treatment; the step-3 status is now a persistent live region that flips from the waiting text to "Approved — you can continue".
- New: while a code is waiting, the page polls the existing `open` action every 2s; on approval it marks step 3 current, announces the change, clears any stale early-continue error, and stops polling. Poll failures after the first successful open are ignored, unmount clears the timer, and redeeming still requires an explicit click.

### Tests
- `EventAccessTransfer.test.ts`: step order, active-step transitions (no request, request, approved, terminal), contextual status, dialog/input accessibility, copy announcement, plus all existing behavior cases.
- `AccessTransfer.test.ts`: step alignment and five fake-timer cases for polling (highlight, stop, no poll when approved, transient failure, stale-error clearing).
- `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: step/structure assertions and a "Target sees approval without reloading" step.

## Verification
- Focused unit suites: 39 passed.
- Full frontend suite: 149 files, 1152 tests passed.
- `npm run lint` (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), `fmt:check`, `typecheck`, and `build` passed.
- e2e `lint`, `fmt:check`, and `typecheck` passed.
- Firefox desktop access-transfer spec with `E2E_FRONTEND=bundled`: 8/8 passed (1.9m).
- `graphify update .` run.

## Notes
- Polling deliberately ignores cancellation/expiry (403/404); those still surface on the Continue click as before.
- No server, swagger, ADR, contract, Markdown, or `scripts/` changes.
<!-- SECTION:FINAL_SUMMARY:END -->
