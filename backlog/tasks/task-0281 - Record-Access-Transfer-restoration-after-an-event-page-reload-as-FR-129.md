---
id: TASK-0281
title: Record Access Transfer restoration after an event-page reload as FR-129
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-19 21:04'
updated_date: '2026-09-19 21:05'
labels: []
dependencies: []
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
  - docs/design/architecture/adr/ADR-010.md
  - docs/terminology/README.md
priority: medium
type: docs
ordinal: 281000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0270 fixed a source-side defect where reloading the event page collapsed Manage access to the compact create form even though a saved transfer was still pending or approved, losing access to its link, status, matching-code approval, and cancellation. The fix landed with unit and Firefox e2e coverage, but the durable reload-restoration behavior has no requirement record; FR-128 now covers the numbered-step flow, so reload restoration is the remaining gap. `docs/requirements/` is the canonical home for durable product behavior.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/requirements/functional/fr/FR-129.md` records that after the source browser that created an Access Transfer reloads the event page, opening the Manage access dialog restores the most recently created saved transfer that is still pending or approved as the current transfer.
- [x] #2 The requirement records that a restored pending transfer shows its transfer link, status, matching-code input, approval control, and cancel control, while a restored approved transfer shows its status and cancel control without the matching-code input.
- [x] #3 The requirement records that a saved transfer that has ended or is unavailable is not restored as the current transfer, that the dialog presents the create-link form while keeping Granted EVCC history revocable, and that the most recently created active transfer is restored when several are saved.
- [x] #4 The FR uses the requirements format: id FR-129, `type: functional`, components `frontend`, status `proposed`, one sentence per physical line, and controlled terms linked per the terminology rules.
- [x] #5 The functional requirements index in `docs/requirements/README.md` gains an FR-129 row with a stable relative link and canonical term links.
- [x] #6 `docs/design/architecture/adr/ADR-010.md` lists FR-129 in its `addresses` section.
- [x] #7 Root `npm run format:markdown:check` and `npm run lint:markdown` pass for the changed Markdown.
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

- FR-128 (TASK-0279) records the numbered-step flow; this FR adds the reload-restoration behavior TASK-0270 fixed, which remains unrecorded.
- TASK-0270 evidence establishes the scope: source browser only, event-page reload, pending and approved restores, most-recently-created selection among several active transfers, terminal/unavailable transfers not restored, grant history still listed; unit tests plus a Firefox e2e reload assertion cover it.
- `docs/requirements/README.md` format and TASK-0279's FR-128 file are the direct precedents for front matter, sentence-per-line, and glossary link depth.
- `docs/design/architecture/adr/ADR-010.md` `addresses.enables` now ends at FR-128; add FR-129 there for consistency.

## Plan

1. Add `docs/requirements/functional/fr/FR-129.md` with front matter (id FR-129, type functional, components frontend, status proposed) and the reload-restoration clauses: source-browser restoration of the most recent pending or approved saved transfer; restored pending vs approved control sets; terminal/unavailable transfers not restored while Granted EVCC history stays revocable; most recent active transfer wins among several.
2. Add the FR-129 row to `docs/requirements/README.md` after FR-128 with canonical term links.
3. Add `FR-129` to `docs/design/architecture/adr/ADR-010.md` `addresses.enables`.
4. Run `npm run format:markdown` from the repo root, then verify with `npm run format:markdown:check` and `npm run lint:markdown`.
5. Verify links and anchors manually; documentation-only, so unit, e2e, swagger, and `graphify update` checks are not required.

## Risks

- `frontend` is the only component because TASK-0270 restored browser-side dialog state using existing status responses; the server enforces no reload-specific behavior.
- Scope stays at the event-page reload the source task established; no cross-session or target-side reload behavior is claimed beyond TASK-0270 evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Documentation-only change completed. Added FR-129 for source-browser reload restoration, linked it in the functional requirements index (docs/requirements/README.md:242) and in ADR-010 `addresses.enables` (docs/design/architecture/adr/ADR-010.md:20). Evidence: `npm run format:markdown` (reformatted the README table), `npm run format:markdown:check`, and `npm run lint:markdown` all pass from the repo root; both glossary links resolve to `### Access Transfer` (docs/terminology/glossary.md:486) and `### Granted Event Visitor Control Credential (Granted EVCC)` (docs/terminology/glossary.md:470); front matter and one-sentence-per-line verified by reading the file. No runtime code, tests, swagger, contracts, or `scripts/`/`prettier/` were touched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

- Added `docs/requirements/functional/fr/FR-129.md`, recording source-browser reload restoration: after the source browser reloads the event page, Manage access restores the most recently created saved transfer that is still pending or approved; a restored pending transfer shows its link, status, code input, approval, and cancel controls, while an approved one shows status and cancel without the code input; ended or unavailable transfers are not restored, the create-link form stays with Granted EVCC history revocable, and the most recent active transfer wins when several are saved.
- Added the FR-129 row to the functional requirements index in `docs/requirements/README.md`.
- Listed FR-129 in `docs/design/architecture/adr/ADR-010.md` `addresses.enables`.

## Why

TASK-0270 fixed the source-side defect where a reload collapsed Manage access to the compact create form even though a saved transfer was still active, and added unit plus Firefox e2e coverage, but the durable behavior had no requirement record. FR-128 covers the numbered-step flow, leaving reload restoration as the gap.

## Verification

- `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass from the repo root (documentation-only change).
- Both glossary links in FR-129 point at their headings (`### Access Transfer` at docs/terminology/glossary.md:486 and `### Granted Event Visitor Control Credential (Granted EVCC)` at :470), and the README row resolves to the new file.
- Front matter matches the requirements format (`id`, `type: functional`, component `frontend`, `status: proposed`), and every sentence stays on one physical line.

## Notes

- Documentation-only: no runtime code, unit/e2e tests, swagger, contract documents, `scripts/`, or `prettier/` changed.
- Scope stays at the event-page reload TASK-0270 established; no cross-session or target-side reload behavior is claimed, and no inbox item existed for this FR.
<!-- SECTION:FINAL_SUMMARY:END -->
