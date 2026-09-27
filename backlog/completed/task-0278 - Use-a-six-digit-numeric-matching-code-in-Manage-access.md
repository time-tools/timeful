---
id: TASK-0278
title: Use a six-digit numeric matching code in Manage access
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-19 19:54'
updated_date: '2026-09-19 20:07'
labels: []
milestone: Manage access
dependencies: []
documentation:
  - docs/requirements/functional/fr/FR-082.md
  - docs/requirements/functional/fr/FR-083.md
  - docs/requirements/quality/qr/QR-011.md
  - docs/design/architecture/adr/ADR-010.md
priority: medium
type: enhancement
ordinal: 278000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The matching code shown on the target browser and typed into the source browser is currently eight Crockford base32 characters mixing digits and uppercase letters. Transcribing it is error-prone, especially on mobile keyboards, and wrong entries are hard to diagnose. Six decimal digits are easier to read and enter, and a digits-only input lets a wrong entry fail with the existing guidance instead of being ambiguous. The code stays a per-transfer match, not a bearer secret: approval still requires the source proof and the exact request code.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A target browser opening an Access Transfer link is shown a matching code of six decimal digits, including leading zeros.
- [x] #2 Each pending target request in one Access Transfer displays a matching code distinct from the other pending target requests' codes.
- [x] #3 The source browser accepts a matching code typed with optional spaces or hyphens between digits and approves only when the digits match one pending request exactly.
- [x] #4 The source browser's Approve matching code control stays disabled until a complete six-digit code is entered, and a non-matching code surfaces existing failure guidance.
- [x] #5 The target page presents the matching code grouped for readability and offers a Copy code control with copied and copy-failure feedback, including a screen-reader announcement.
- [x] #6 The Manage access numbered-step flow, two-second polling, cancel, revoke, restore-after-reload, account-switch, and transfer-link behavior are unchanged.
- [x] #7 FR-127 records the six-digit matching-code format, the requirements index links it, ADR-010 lists it under addresses.enables with an updated date, the Access Transfer glossary entry names it as authoritative context, and the PostgreSQL event API contract states the numeric code format.
- [x] #8 Unit tests cover numeric code generation and zero-padding, typed-code normalization, length-gated approval, target-page grouping and copy feedback, and the wrong-code path; backend route tests cover numeric format, per-transfer distinctness, and rejection of a wrong numeric code.
- [x] #9 The Firefox access-transfer e2e spec asserts six-digit numeric codes and exercises a wrong numeric code without regressing the successful typed approval flow.
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
1. Requirements docs first: create `docs/requirements/functional/fr/FR-127.md` (status proposed, components frontend + backend) stating the target displays a six-decimal-digit matching code including leading zeros, pending requests in one Access Transfer have distinct codes, and the source accepts spaces or hyphens between digits and compares only digits; add its row to `docs/requirements/README.md`; add `FR-127` to `addresses.enables` in `docs/design/architecture/adr/ADR-010.md` and bump `updated_date`; add FR-127 to the Access Transfer entry's authoritative context in `docs/terminology/glossary.md`; update the matching-code sentence in `server/docs/postgres-event-api-contract.md` to state the numeric format.
2. Backend: add `GenerateTransferCode()` in `server/postgres/repository.go` (crypto/rand, uniform, zero-padded six digits; leave `GenerateShortID()` for event short IDs); use it in `server/routes/transfers.go` in the `open` case, looping while the candidate equals an already-listed request code. Add unit coverage for format/zero-padding and route coverage for the numeric shape, per-transfer distinctness, and a wrong numeric code returning 403.
3. Frontend boundary: add `normalizeTransferCode()` (trim, strip non-digits) in `frontend/src/composables/transfer/transferBoundary.ts` and use it in `matchingRequest()`; cover spaces, hyphens, and leading zeros in `transferBoundary.test.ts`.
4. Source dialog: in `frontend/src/components/event/EventAccessTransfer.vue`, add `inputmode="numeric"`, `autocomplete="one-time-code"`, `maxlength="6"`, digits-only input filtering, and length-gated Approve; update the hint to six-digit wording while keeping existing labels.
5. Target page: in `frontend/src/views/AccessTransfer.vue`, render the digits grouped without whitespace text so `innerText` stays six digits, and add a Copy code button with copied and copy-failure feedback plus a polite screen-reader announcement.
6. Tests: update `EventAccessTransfer.test.ts`, `AccessTransfer.test.ts`, and the Firefox `e2e/specs/timed-event-access-transfer-firefox.spec.ts` (regexes to `^\d{6}$`, wrong-code case uses a computed non-matching six-digit value, add copy-control coverage where practical).
7. Checks: frontend lint, fmt:check, typecheck, build, test:unit; backend suite through the compose test overlay; Firefox e2e; `npm run format:markdown` for changed Markdown; `graphify update .`. No swagger regeneration because annotations and payload types are unchanged.
8. Rollout note: a pending 8-character transfer created before the deploy cannot be approved after the frontend filters to digits, but it expires within five minutes; accepted transient, documented in the task notes.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Requirements docs landed first: `docs/requirements/functional/fr/FR-127.md` (proposed, frontend + backend) records the six-decimal-digit format with leading zeros, per-transfer distinct codes, and separator-tolerant entry; the requirements index links it, `ADR-010` lists it under `addresses.enables`, the Access Transfer glossary entry names it as authoritative context, and `server/docs/postgres-event-api-contract.md` now states the numeric format.

Backend keeps `GenerateShortID()` for event short IDs and adds `GenerateTransferCode()` plus `formatTransferCode()`; the `open` action loops until the candidate differs from the transfer's loaded request codes, and the existing `UNIQUE (transfer_id, code)` constraint remains the backstop. No schema change or timing change was needed.

Rollout note: a transfer pending across the deploy still carries an 8-character code that the digits-only source input cannot submit; it expires within five minutes, so the transient was accepted rather than adding a legacy-code path.

Verification evidence: frontend lint, `fmt:check`, `typecheck`, `build`, and `test:unit` (150 files, 1168 tests) pass; the server suite passes through the isolated compose overlay, including `timeful/server/postgres` and `timeful/server/routes`; the Firefox `timed-event-access-transfer-firefox.spec.ts` suite passes 8/8 with `E2E_FRONTEND=bundled` at two workers, and the guest scenario also passes isolated at one worker after the default dev-server two-worker run timed out at its 40-second budget (known heavy-journey contention, not a behavioral failure); `npm run format:markdown`, `format:markdown:check`, and `lint:markdown` pass, and `graphify update .` ran.

Swagger regeneration and root `npm run fmt:check` were not applicable: handler annotations, payload types, `scripts/`, and `prettier/` are unchanged.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Manage access now uses a six-digit numeric matching code instead of an eight-character Crockford base32 code. The target browser displays the digits in two visual groups with a Copy code control, and the source browser accepts digits only, tolerates spaces or hyphens, and enables **Approve matching code** only when six digits are present. Exact request-id plus code approval, source-proof gating, and the five-minute single-use transfer semantics are unchanged.

## Changes

- **Backend**: added `GenerateTransferCode()` and `formatTransferCode()` in `server/postgres/repository.go` (crypto/rand, rejection sampling, zero-padded six digits); `server/routes/transfers.go` now loops until the candidate differs from the transfer's loaded request codes, with `UNIQUE (transfer_id, code)` as the backstop. `GenerateShortID()` still serves event short IDs.
- **Frontend**: `normalizeTransferCode()` in `transferBoundary.ts` strips non-digits for `matchingRequest()`; `EventAccessTransfer.vue` gains numeric inputmode, one-time-code autocomplete, six-digit maxlength, digit filtering, and length-gated approval; `AccessTransfer.vue` groups the displayed code and adds the Copy code control with copied/copy-failure feedback and a polite screen-reader announcement. The code and copy control hide once the link is unavailable.
- **Requirements/docs**: new `FR-127` (proposed, frontend + backend) with its index row, `ADR-010` `addresses.enables`, the Access Transfer glossary authoritative context, and the numeric-format sentence in `server/docs/postgres-event-api-contract.md`.

## Verification

- Frontend: `lint`, `fmt:check`, `typecheck`, `build`, and `test:unit` (150 files, 1168 tests) pass.
- Backend: full suite through the isolated compose overlay passes, including `timeful/server/postgres` and `timeful/server/routes` with the new format, distinctness, and wrong-numeric-code coverage.
- Firefox e2e: `timed-event-access-transfer-firefox.spec.ts` passes 8/8 with `E2E_FRONTEND=bundled` at two workers; the guest scenario also passes isolated at one worker (the default dev-server two-worker run hit the scenario's 40-second budget, consistent with documented heavy-journey contention).
- Markdown format/lint checks pass; `graphify update .` ran.

## Risks / notes

- Swagger regeneration and root `fmt:check` were not applicable (annotations, payload types, `scripts/`, and `prettier/` unchanged).
- Rollout transient: a transfer pending across the deploy still shows an 8-character code that the digits-only source input cannot submit; it expires within five minutes, so no legacy-code path was added.
<!-- SECTION:FINAL_SUMMARY:END -->
