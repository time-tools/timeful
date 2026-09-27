---
id: TASK-0275
title: End the target transfer page when a waiting link is cancelled or expired
status: Done
assignee:
  - opencode
created_date: '2026-09-19 18:14'
updated_date: '2026-09-19 18:21'
labels:
  - frontend
dependencies: []
references:
  - frontend/src/views/AccessTransfer.vue
  - frontend/src/views/AccessTransfer.test.ts
  - e2e/specs/timed-event-access-transfer-firefox.spec.ts
documentation:
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 275000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0274 added a two-second poll of the `open` action on the target `Continue on this device` page so it highlights step 3 as soon as the source approves. That poll deliberately treats every failure as transient. When the source cancels the transfer or the five-minute window expires while the target is still waiting, `open` starts returning 403/404, the poll keeps ignoring it, and the page keeps showing "Waiting for approval in the other browser" with an active Continue after approval control. The target user only discovers the transfer is unusable when they click Continue and an error replaces the waiting state. The waiting page should instead end the wait with a clear unavailable message and stop offering Continue, while genuine transient failures still keep waiting.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 When a waiting target page receives a definitive 403 or 404 from a status poll, it stops polling and replaces the waiting message with a clear unavailable message that names the link as expired, cancelled, or unavailable and points to asking the source browser for a new link.
- [x] #2 Once the transfer is known unavailable, the Continue after approval control is no longer offered and step 3 no longer claims approval can still arrive.
- [x] #3 A transient poll failure such as a 500 or network error keeps the page waiting silently with no error alert, and polling continues.
- [x] #4 The existing initial-open failure behavior and the pending-to-approved highlight behavior remain unchanged.
- [x] #5 Unit tests cover cancellation and expiry detection during polling, the transient-versus-definitive distinction, polling stopping after detection, and removal of the Continue control.
- [x] #6 Firefox access-transfer e2e coverage for cancelled and expired links exercises the target waiting page ending without a Continue click.
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

- `frontend/src/views/AccessTransfer.vue` (target page) polls `transferAction(..., "open")` every 2s while a code is shown and the transfer is not approved; its `catch` currently ignores every failure (added by TASK-0274).
- `frontend/src/composables/transfer/transferBoundary.ts` already exports `isTransferUnavailable(error)`, which maps `FetchError` 403/404; `EventAccessTransfer.vue` uses it to turn failed `status` polls into a terminal "unavailable" state.
- `server/routes/transfers.go:208-234` returns 403 for cancelled/expired/redeemed transfers and for a non-approved request once another request is approved, so 403/404 is definitive for the target.
- Existing unit tests cover transient network failure keeping the wait; the Firefox e2e spec currently expects a manual Continue click on cancelled/expired waiting pages (and on the "other browser" after another request is approved).

## Plan

1. `frontend/src/views/AccessTransfer.vue`: add an `unavailable` state; in `checkApproval`'s catch, treat `isTransferUnavailable(cause)` as terminal by setting the unavailable state, reusing the existing unavailable alert copy, and clearing the poll timer; step 3 renders an unavailable status instead of the waiting text, and `Continue after approval` renders only while `code` exists and the transfer is not unavailable. The initial-open failure path and the pending-to-approved highlight path stay unchanged.
2. `frontend/src/views/AccessTransfer.test.ts`: add fake-timer cases for 403 and 404 poll detection (unavailable alert plus status, Continue removed, polling stopped, no approval claim), and extend transient coverage with a 500 `FetchError` that keeps waiting silently and keeps polling.
3. `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: after another request is approved, assert the other browser's waiting page ends on its own without a Continue click; rework the cancelled/expired loop to assert the waiting target page ends automatically (status names the link unavailable, Continue gone, no click), keeping the reload fallback check and the no-grant assertion.
4. Verification: frontend lint, fmt:check, typecheck, build, test:unit; Firefox desktop access-transfer e2e with streamed output; `graphify update .`. No server, swagger, Markdown, or contract changes.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the recorded plan. `checkApproval` now classifies poll failures with the existing `isTransferUnavailable` helper (FetchError 403/404); definitive failures set an `unavailable` flag, apply the initial-open failure alert copy, and clear the poll timer. Step 3 renders an ended-wait status and `Continue after approval` renders only while `code` exists and the transfer is not unavailable. The initial-open failure path is intentionally untouched so its behavior stays unchanged.

Deviation from the plan wording: the step-3 ended-wait message is "This link is expired, cancelled, or unavailable. Ask the source browser for a new link." while the alert keeps "This transfer is expired, cancelled, or unavailable...". The distinct wording keeps e2e text matchers unambiguous under strict mode; the alert shares the `UNAVAILABLE_MESSAGE` constant with the initial-open failure path.

Verification: focused AccessTransfer suite 14 passed; full frontend unit 149 files/1155 tests passed; frontend lint (0 errors; 2 pre-existing NewSignUp warnings), fmt:check, typecheck, build passed; e2e package lint/fmt/typecheck passed; Firefox desktop access-transfer spec with E2E_FRONTEND=bundled passed 8/8 (1.8m, log /tmp/opencode/task-0275-e2e-firefox.log); graphify update . run.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The target `Continue on this device` page now ends its wait when the transfer becomes definitively unusable, instead of hiding that state until a Continue click.

### Target page (`frontend/src/views/AccessTransfer.vue`)
- The two-second `open` poll now classifies failures: a definitive 403/404 (`isTransferUnavailable`) sets an `unavailable` state, surfaces the existing unavailable alert copy through a shared `UNAVAILABLE_MESSAGE`, and clears the poll timer. Other failures such as a 500 or network error stay silent and polling continues.
- Step 3 now shows "This link is expired, cancelled, or unavailable. Ask the source browser for a new link." once the wait ends, and `Continue after approval` renders only while a code exists and the transfer is not unavailable. The step-status wording deliberately differs from the alert wording so text matchers stay unambiguous.
- The initial-open failure rendering and the pending-to-approved highlight path are unchanged.

### Tests
- `frontend/src/views/AccessTransfer.test.ts`: new `it.each([403, 404])` covers cancellation/expiry detection, the unavailable alert and status copy, removal of the Continue control, and polling stopping; a new 500 `FetchError` case covers transient failure keeping the wait with no alert while polling continues, alongside the existing network-error case.
- `e2e/specs/timed-event-access-transfer-firefox.spec.ts`: in the multi-target journey the other browser's wait now ends after another request is approved, without a Continue click; the cancelled/expired loop asserts the waiting page ends on its own and keeps the reload-based initial-open failure check and no-grant cookie assertion.

## Verification
- Focused `AccessTransfer` suite: 14 passed.
- Full frontend unit: 149 files, 1155 tests passed.
- `npm run lint` (0 errors; 2 pre-existing `NewSignUp.test.ts` warnings), `fmt:check`, `typecheck`, and `build` passed.
- e2e package `lint`, `fmt:check`, and `typecheck` passed.
- `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop specs/timed-event-access-transfer-firefox.spec.ts`: 8/8 passed (1.8m), log at `/tmp/opencode/task-0275-e2e-firefox.log`.
- `graphify update .` run.

## Notes
- A 403 also signals "another request was approved" to a non-selected target; that browser now ends its wait with the unavailable message, which is accurate from its perspective.
- No server, swagger, Markdown, `scripts/`, or contract changes; the behavior matches ADR-010 and FR-117. No follow-ups identified.
<!-- SECTION:FINAL_SUMMARY:END -->
