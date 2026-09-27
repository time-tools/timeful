---
id: TASK-0279
title: Record the Access Transfer numbered steps with live status as FR-128
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 20:56'
updated_date: '2026-09-19 20:58'
labels: []
dependencies: []
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
  - docs/design/architecture/adr/ADR-010.md
  - docs/terminology/README.md
priority: medium
type: docs
ordinal: 279000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When the Manage access dialog and the target browser's transfer page were reworked into the numbered-step flow (TASK-0274 and TASK-0275), the behavior landed with unit and Firefox e2e coverage but without a durable requirement record. FR-126 and FR-127 already record the target-browser identification and the six-digit matching code; the step presentation and the live status it shows in both browsers are the remaining gap. `docs/requirements/` is the canonical home for durable product behavior, and the repository inbox carries the request as "FR for the access transfer steps".
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/requirements/functional/fr/FR-128.md` records the source-browser step presentation: three always-visible numbered steps in the order create and copy a transfer link, open the link in the other browser, approve the displayed matching code; the dialog marks the step the transfer has reached; and the matching-code input and approval control are available whenever a transfer is pending, including before the other browser's request has been observed.
- [x] #2 The requirement records the source-browser live status: while a transfer is pending the dialog states whether the other browser has opened the link, points to the approval step once a code is displayed there, and after source approval directs the user to finish in the other browser.
- [x] #3 The requirement records the target-browser page behavior: it presents its part of the transfer as the same numbered steps, waits while the source has not approved, reaches its final step and announces completion without a reload after approval, and ends the wait with an unusable-link state and no approval-dependent action when the transfer becomes cancelled, expired, or otherwise unusable while waiting.
- [x] #4 The FR uses the requirements format: id FR-128, `type: functional`, components `frontend` and `backend`, status `proposed`, one sentence per physical line, and controlled terms linked per the terminology rules.
- [x] #5 The functional requirements index in `docs/requirements/README.md` gains an FR-128 row with a stable relative link and canonical term links.
- [x] #6 `docs/design/architecture/adr/ADR-010.md` lists FR-128 in its `addresses` section.
- [x] #7 The `backlog/backlog.md` inbox item "FR for the access transfer steps" is checked off.
- [x] #8 Root `npm run format:markdown:check` and `npm run lint:markdown` pass for the changed Markdown.
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

- `docs/requirements/README.md` defines the FR format: front matter with `id`, `title`, `type`, `components`, `status`; one sentence per physical line; controlled terms linked on first occurrence per paragraph and bolded on repeats.
- `docs/requirements/functional/fr/FR-126.md` and `FR-127.md` are the recent transfer FRs and show the expected granularity, glossary link depth (`../../../terminology/glossary.md#...`), and status `proposed`.
- `docs/design/architecture/adr/ADR-010.md` `addresses.enables` lists FR-126 and FR-127; FR-128 belongs there for consistency.
- The source behavior lives in `frontend/src/components/event/EventAccessTransfer.vue` and the target behavior in `frontend/src/views/AccessTransfer.vue` (TASK-0274/TASK-0275); the durable outcomes are already exercised by `e2e/specs/timed-event-access-transfer-firefox.spec.ts`.
- `backlog/backlog.md:113` carries the unchecked inbox item `FR for the access transfer steps`.

## Plan

1. Add `docs/requirements/functional/fr/FR-128.md` with front matter (id FR-128, type functional, components frontend and backend, status proposed) and the agreed clauses: three always-visible source steps in order; active step marked; code input and approval available while pending including before the target opening is observed; pending status distinguishes whether the target opened the link and points to the approval step; approved status directs the user to finish in the target browser; the target page mirrors the steps, waits, announces approval without a reload, and ends the wait with an unusable-link state when the transfer is cancelled, expired, or otherwise unusable.
2. Add the FR-128 row to the functional requirements index in `docs/requirements/README.md` after FR-127.
3. Add `FR-128` to `docs/design/architecture/adr/ADR-010.md` `addresses.enables`.
4. Check off the `FR for the access transfer steps` inbox item in `backlog/backlog.md`.
5. Run `npm run format:markdown` from the repo root, then verify with `npm run format:markdown:check` and `npm run lint:markdown`.
6. Verify links and anchors manually; documentation-only, so unit, e2e, swagger, and `graphify update` checks are not required.

## Risks

- Terminology linking is strict in this repo; every paragraph that introduces a controlled term must link or bold it correctly, and table rows must stay on one physical line.
- README table alignment must survive `format:markdown`; run the formatter before the checks.

## Scope note

ADR-010 `updated_date` is already 2026-09-19; only the `addresses` list changes. The glossary `Access Transfer` entry keeps its current authoritative context because this FR records presentation, not the process definition.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Documentation-only change completed. Added FR-128 with the agreed clauses, linked it in the requirements index and ADR-010 addresses, and checked off the inbox item. Evidence: `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all pass from the repo root; both glossary links resolve to the `### Access Transfer` heading (docs/terminology/glossary.md:486); the README index row points at `functional/fr/FR-128.md`; and `backlog/backlog.md:113` is checked. No runtime code, tests, swagger, contracts, or `scripts/`/`prettier/` were touched, so the conditional DoD items for those are not applicable.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

- Added `docs/requirements/functional/fr/FR-128.md`, recording the Access Transfer step-flow scenario: the source browser's Manage access dialog presents three always-visible numbered steps (create/copy, open in target, approve code), marks the reached step, keeps the code input and approval available while pending, reports whether the target browser opened the link, points to the approval step once a code is displayed, and directs the user to finish in the target browser after approval; the target browser's page mirrors the steps, waits before approval, announces completion without a reload, and ends the wait with an unusable-link state when the transfer is cancelled, expired, or otherwise unusable.
- Added the FR-128 row to the functional requirements index in `docs/requirements/README.md`, linked to the new file with canonical term links.
- Listed FR-128 in `docs/design/architecture/adr/ADR-010.md` `addresses.enables`.
- Checked off the `FR for the access transfer steps` item in `backlog/backlog.md`.

## Why

TASK-0274 and TASK-0275 implemented the numbered-step transfer flow with unit and Firefox e2e coverage, but the durable behavior had no requirement record; FR-126 and FR-127 already cover target-browser identification and the six-digit code, leaving step presentation and its live status as the gap. The repository backlog inbox requested the FR explicitly.

## Verification

- `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass from the repo root (documentation-only change).
- Both glossary links in FR-128 point at the `### Access Transfer` glossary heading (docs/terminology/glossary.md:486), and the README row resolves to the new file.
- Front matter matches the requirements format (`id`, `type: functional`, components `frontend` and `backend`, `status: proposed`), and every sentence stays on one physical line.

## Notes

- Documentation-only: no runtime code, unit/e2e tests, swagger, contract documents, `scripts/`, or `prettier/` changed.
- Deliberately out of scope: reload restoration of an active transfer (TASK-0270) and changing the glossary `Access Transfer` entry, which defines the process rather than the presentation.
<!-- SECTION:FINAL_SUMMARY:END -->
