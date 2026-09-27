---
id: TASK-0277
title: Show target browser details in the Manage access dialog
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 19:12'
updated_date: '2026-09-19 19:21'
labels:
  - frontend
  - backend
milestone: Manage access
dependencies: []
references:
  - frontend/src/components/event/EventAccessTransfer.vue
  - frontend/src/composables/transfer/transferBoundary.ts
  - frontend/src/views/AccessTransfer.vue
  - server/routes/transfers.go
  - server/postgres/transfers.go
  - server/migrations/20260912000000_baseline_schema.sql
documentation:
  - docs/requirements/functional/fr/FR-081.md
  - docs/requirements/functional/fr/FR-082.md
  - docs/requirements/functional/fr/FR-083.md
  - docs/requirements/functional/fr/FR-117.md
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: feature
ordinal: 277000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Manage access dialog's granted-access section lists rows as "Granted access N", where N is a browser-local counter assigned when the transfer was created. The label says nothing about which browser received access, so a user with several grants cannot tell them apart before revoking one, and the counter also numbers transfers that were cancelled or expired. While a transfer is pending, the status line likewise says only "The other browser is showing a code" with no device context.

The target browser's User-Agent is already presented to the backend when it opens the transfer link, but it is discarded. Capturing it and describing the browser and operating system gives the source browser a concrete way to recognize the target it approved.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A granted-access row identifies the target browser by browser name and operating system, for example "Granted access 2 · Firefox on Linux", and falls back to "Granted access 2" when the target browser cannot be identified.
- [x] #2 While an access transfer is pending and the target browser is displaying a matching code, the Manage access status text names the target browser, and falls back to the existing message when it cannot be identified.
- [x] #3 The description is derived from the User-Agent the target browser presented when it opened the transfer link, and the raw User-Agent string never appears in the dialog.
- [x] #4 A new functional requirement FR-126 documents the behavior, docs/requirements/README.md lists it, and ADR-010 records that it enables it.
- [x] #5 Unit tests cover browser parsing (Firefox, Chrome, Edge, Safari, desktop, mobile, empty, and malformed User-Agents), transport decoding, labeled and fallback granted-access rows, and the pending status hint.
- [x] #6 Backend tests prove the User-Agent presented at link open is stored per target request and returned in transfer status for the approved request.
- [x] #7 Firefox browser E2E proves a redeemed transfer displays the target browser description.
- [x] #8 Existing Manage access behavior is unchanged: create, copy, approve, cancel, revoke, two-second polling, and restore-after-reload.
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

- Granted-access rows are `Granted access {number}` plus Revoke access in `frontend/src/components/event/EventAccessTransfer.vue` (section at `:161-181`). `number` is a browser-local counter from `localStorage` (`SavedTransfer` in `frontend/src/composables/transfer/transferBoundary.ts:75-78`, `rememberTransfer` at `:152-160`); the server contributes only `state: "redeemed"` and `revocable: true` from the `status` action.
- `server/routes/transfers.go` `open` creates an `access_transfer_requests` row per target browser (`:246-260`) and never records the request's User-Agent. `status` returns `state`, `revocable`, and `requests` (`:198-207`). `TransferRequest` in `server/postgres/transfers.go` has `ID`, `Code`, `TargetHash`; hash is `json:"-"`.
- Schema: `access_transfer_requests` (baseline `server/migrations/20260912000000_baseline_schema.sql:147-153`) has `id`, `transfer_id`, `target_hash`, `code`; goose migrations are dated files with `-- +goose Up`/`Down`; `TestMigrationReplayMatchesAppliedSchema` replays every migration into temp tables, so an `ALTER TABLE ... ADD COLUMN` must work through that harness.
- `frontend/src/types/transport.ts:676-685` hand-maintains `RawAccessTransfer`; generated API types come from `npm run gen:api`.
- `frontend/src/views/AccessTransfer.vue` target page polls `open`; the proof cookie guarantees the redeeming browser is the one that opened the request.
- The repo root `LICENSE` is AGPL-3.0, so `ua-parser-js` v2 (`AGPL-3.0-or-later`, latest 2.0.10, ESM + types) is license-compatible. User selected uaparser.js and raw-UA-in-DB with parsing at the frontend boundary.

## Plan

1. Requirements docs first: create `docs/requirements/functional/fr/FR-126.md` (frontend, backend, status proposed) describing target-browser identification in the Manage access dialog for granted-access rows and the pending matching-code status, with a non-device fallback; add its row to the functional index in `docs/requirements/README.md`; add `FR-126` to `addresses.enables` in `docs/design/architecture/adr/ADR-010.md` and bump `updated_date`.
2. Backend: migration `server/migrations/20260919000000_transfer_request_user_agent.sql` adding `user_agent TEXT NOT NULL DEFAULT ''` to `access_transfer_requests`; carry `UserAgent` through `TransferRequest`, `CreateTransferRequest`, and `ListTransferRequests`; capture `c.Request.UserAgent()` (trimmed, length-capped) when `open` creates a request; in `status`, return `targetUserAgent` from the request named by `approved_request_id` alongside the existing `requests[].userAgent`; update Swag annotations, run `swag init`, then `npm run gen:api`.
3. Backend tests: UA capture and status assertions in `server/routes/transfers_test.go`, roundtrip in `server/postgres/transfers_test.go`, migration replay/fidelity covered by existing harness.
4. Frontend: add `ua-parser-js@^2.0.10`; new pure `frontend/src/utils/userAgent.ts` (`describeTargetBrowser` returning `"Firefox on Linux"`-style labels with partial/empty fallbacks); extend `RawAccessTransfer` (`targetUserAgent`, `requests[].userAgent`) and decode into an internal browser label in `transferBoundary.ts`; update `EventAccessTransfer.vue` to label granted-access rows (fallback to `Granted access N`) and name the browser in the pending status hint.
5. Frontend tests: parser unit tests over real UA strings; decode tests in `transferBoundary.test.ts`; labeled/fallback history rows and pending hint in `EventAccessTransfer.test.ts`; keep existing behavior assertions green.
6. E2E: extend the recorded Firefox access-transfer journey to assert the device label on a redeemed transfer.
7. Verification: frontend lint, fmt:check, typecheck, build, test:unit; backend suite via the isolated Compose test overlay; swagger regen; `graphify update .`; Markdown formatting via root `npm run format:markdown`.

## Risks

- UA strings are reduced/frozen by vendors; labels are best-effort and version-less. Empty or unparseable UAs must keep today's text.
- The replay-fidelity harness must absorb the `ALTER TABLE`; verify with the backend suite.
- E2E device label must match Playwright Firefox on Linux; use a partial label assertion if needed.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented requirements docs: FR-126 created, requirements index row added, ADR-010 enables list updated.

Backend: migration 20260919000000_transfer_request_user_agent.sql; TransferRequest carries UserAgent through create/list; open captures transferUserAgent(c.Request.UserAgent()) (trim, 512-rune cap); status returns targetUserAgent for the approved request; swagger regenerated and frontend API types regenerated.

Frontend: ua-parser-js@^2.0.10 added; describeTargetBrowser in src/utils/userAgent.ts; decodeTransfer maps targetUserAgent/requests[].userAgent to browser labels; granted rows show 'Granted access N · Firefox on Linux' with fallback; pending status names a single target browser.

Verification so far: focused vitest 47 passed; full frontend unit 150 files/1166 tests passed; lint, fmt:check, typecheck, build passed; backend suite via compose test overlay passed after rebuilding the migrator image; Markdown formatted with root npm run format:markdown.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

Granted-access rows and the pending matching-code status in the Manage access dialog now identify the target browser by browser name and operating system, for example `Granted access 2 · Firefox on Linux`, instead of showing only the browser-local counter. FR-126 documents the behavior.

### Backend
- Migration `server/migrations/20260919000000_transfer_request_user_agent.sql` adds `user_agent TEXT NOT NULL DEFAULT ''` to `access_transfer_requests`.
- `open` stores the target browser's User-Agent (trimmed, capped at 512 runes by `transferUserAgent`); `TransferRequest` carries `userAgent` through create/list; `status` returns `targetUserAgent` for the request named by `approved_request_id`, so redeemed grants carry the description.
- Swag annotations updated, `swag init` rerun, and `npm run gen:api` regenerated `frontend/src/types/api.ts`.

### Frontend
- `ua-parser-js@^2.0.10` added (AGPL-3.0-or-later, compatible with the repo license).
- New pure `frontend/src/utils/userAgent.ts` (`describeTargetBrowser`) maps real User-Agents to labels such as `Firefox on Linux`, `Chrome on Windows`, `Edge on Windows`, `Safari on macOS`, `Chrome on Android`, `Safari on iOS`, with partial and empty fallbacks.
- `transferBoundary.ts` decodes `targetUserAgent`/`requests[].userAgent` into browser labels at the boundary; views never see raw User-Agent strings.
- `EventAccessTransfer.vue` renders `Granted access N · Firefox on Linux` with a fallback to `Granted access N`, and names a single pending target in the status line (`Firefox on Linux is showing a code — enter it in step 3.`); multiple pending browsers keep the generic message.

### Requirements and tests
- `docs/requirements/functional/fr/FR-126.md` created, listed in the functional index, and ADR-010 records it under `enables`.
- Unit: 150 files / 1166 tests passed, including new parser, decode, labeled/fallback history-row, and pending-hint coverage.
- Backend: full suite passed through the isolated Compose overlay (route UA capture/status assertions, postgres roundtrip, migration replay fidelity).
- E2E: Firefox desktop access-transfer spec passed 8/8 with the new `Granted access 1 · Firefox on ...` assertion (`/tmp/opencode/task-0277-e2e-firefox.log`).
- `lint`, `fmt:check`, `typecheck`, `build`, root `fmt:check`, and `format:markdown:check` passed; `graphify update .` rebuilt the graph.

## Risks and notes
- Browser-vendor User-Agent reduction makes descriptions approximate and version-less (Chrome on Windows 11 still reports Windows 10, macOS versions are frozen); unknown values fall back to the numbered label.
- The description is informational only; no authorization decision depends on it.
- Raw User-Agents persist with redeemed transfers, which are retained as revocation anchors and visible only to the source browser.
<!-- SECTION:FINAL_SUMMARY:END -->
