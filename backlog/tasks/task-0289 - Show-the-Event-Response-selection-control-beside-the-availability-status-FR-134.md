---
id: TASK-0289
title: >-
  Show the Event Response selection control beside the availability status
  (FR-134)
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 20:32'
updated_date: '2026-09-20 20:45'
labels: []
dependencies: []
references:
  - backlog/backlog.md
  - frontend/src/components/schedule_overlap/RespondentsList.vue
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
  - docs/requirements/functional/fr/FR-066.md
priority: medium
type: enhancement
ordinal: 289000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In the responses list on the event page and during Availability Editing, the response selection checkbox currently cross-fades on top of the row's availability status or avatar inside one fixed 20px slot (frontend/src/components/schedule_overlap/RespondentsList.vue template and scoped styles), so a selected row's availability status is hidden. That defeats the purpose of selecting responses to compare them, and it is the open inbox question in backlog/backlog.md ("How to show response selections so that the status is still visible for selected responses?").

Product decisions confirmed with the user:
- Keep the checkbox to the right of the availability status (or avatar when no slot is in context) and never cover the status.
- Reserve the checkbox space in every row so names never move and row height is stable; this supersedes an earlier "shift the name when the checkbox appears" idea.
- Phone viewport: show the checkbox on every row.
- Desktop viewport: reveal the checkbox on row hover or selection.
- Unchecked checkbox uses the neutral outline; hovered and checked states use the primary action green.
- Record the durable behavior as FR-134 and implement it in one reviewable change (same shape as TASK-0288).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/requirements/functional/fr/FR-134.md` records the behavior (checkbox beside the availability status or avatar, phone always visible, desktop hover-or-selected reveal, neutral unchecked border, primary action color on hover and checked, no name movement), and `docs/requirements/README.md` gains an FR-134 row with stable canonical links; changed Markdown is formatted
- [x] #2 Each row renders the checkbox after the availability status or avatar indicator without overlaying it, and a selected row shows both the availability status and the checked checkbox
- [x] #3 On a phone viewport the checkbox is visible on every row without hover; on a desktop viewport an unselected row shows it only while hovered and a selected row keeps it visible
- [x] #4 Hovering, selecting, or deselecting never moves the row name or changes the row height
- [x] #5 The unchecked checkbox uses the neutral outline, and hovered and checked states use the primary action green
- [x] #6 Unit tests cover the checkbox DOM order, the viewport reveal matrix, and selected-row availability-status retention, and existing RespondentsList control tests are updated
- [x] #7 E2E coverage asserts phone placement and toggling with the availability status visible in firefox-touch, and hover reveal, status retention, and unchanged name position in a firefox-desktop spec
- [x] #8 Frontend lint, fmt:check, typecheck, build, and test:unit pass, and `graphify update .` runs after the code change
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
1. Requirements: create `docs/requirements/functional/fr/FR-134.md` (id FR-134, type functional, component frontend, status proposed) recording that the response checkbox sits beside the availability status or avatar, is always visible on a phone viewport, reveals on hover or selection on a desktop viewport, uses the neutral outline unchecked and the primary action green on hover/checked, and never moves the row name or height. Add the FR-134 row to `docs/requirements/README.md` after FR-133 and run root `npm run format:markdown`.

2. `frontend/src/components/schedule_overlap/RespondentsList.vue`: reorder the control to avatar/status first and checkbox second inside a `tw:inline-flex tw:gap-1` button; keep the 20px avatar slot and give the checkbox a reserved 16px slot so the name position is stable. Drop the absolute-position overlay and cross-fade scoped rules; hide the checkbox with visibility/opacity by default, reveal on `.respondent-row:hover` and `[aria-pressed="true"]`, and force it visible on phone via an `isPhone` modifier class. Move the border color from the inline style to the scoped rules: neutral outline by default, primary action green on hover and checked.

3. Unit tests in `RespondentsList.test.ts`: extend `mountRespondentsList` with a `curRespondents` option; update the fixed-slot and hover-to-select tests for the new DOM order and classes; add a viewport reveal-matrix test using the mocked `isPhone`; add a selected-row test proving the status square stays rendered and the checkbox is checked and follows the status in DOM order.

4. E2E: extend `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` (firefox-touch) with a test that seeds a timed event plus guest responses, clicks a slot, and asserts the phone checkbox is visible unhovered, sits right of the status with the name after it, and toggling keeps the status visible. Add `e2e/specs/timed-event-respondent-selection-firefox.spec.ts` (firefox-desktop) asserting hover reveal, status retention, and unchanged name x before/after hover and after selection.

5. Checks: frontend lint, fmt:check, typecheck, build, test:unit; focused e2e with full streaming output; root `format:markdown`, `format:markdown:check`, `lint:markdown`; `graphify update .`. Then check off `backlog/backlog.md:186-187` and finalize the task.

Risks: existing unit tests select the old control-slot class list and overlay assumptions; the mobile 2-column grid loses name width to the reserved slot, so the touch test should use short names; the desktop e2e needs a slot in context for status squares (hover a timeslot first) or it asserts the avatar variant.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation (uncommitted worktree):
- Requirements: added docs/requirements/functional/fr/FR-134.md and its index row; root Markdown format/lint checks pass.
- RespondentsList.vue: control is now [avatar/status 20px slot][reserved 16px checkbox][name]; removed the absolute overlay and avatar/checkbox cross-fade; checkbox hidden via visibility/opacity, revealed on hover or aria-pressed, always visible on phone via the --always-visible modifier; border color moved from the inline style to scoped CSS (neutral outline default, primary green hover/checked).
- Unit tests: mount helper gained curRespondents; updated the control-slot and status tests; added viewport reveal matrix, checkbox-after-status DOM order, selected-row status retention, and a source guard against position:absolute. Full suite passes.
- E2E: new specs/timed-event-respondent-selection-firefox.spec.ts (desktop hover reveal, indicator retention, stable name position, green border) and a new phone test in schedule-overlap-mobile-touch-firefox.spec.ts (unhovered checkbox, neutral border, checkbox right of status, name right of checkbox, toggle keeps status visible and border green).
- backlog/backlog.md inbox item checked off with the FR-134 resolution.

Deviation from the plan: the phone e2e initially scoped rows to .schedule-overlap-mobile-overlay, but a short response list renders in-flow without the sticky overlay, so the test targets the visible .respondent-row directly.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The event-page and Availability Editing responses list now keeps a row's availability status visible while offering response selection.

- `docs/requirements/functional/fr/FR-134.md` records the durable behavior: the response checkbox sits beside the row's availability status or avatar, is always visible on a phone viewport, reveals on hover or selection on a desktop viewport, uses the neutral outline unchecked and the primary action green on hover or checked, and never moves the row name. The functional requirements index in `docs/requirements/README.md` gains the FR-134 row.
- `frontend/src/components/schedule_overlap/RespondentsList.vue` renders the fixed 20px avatar/status slot first and a reserved 16px checkbox after it inside an inline-flex button, so revealing the checkbox never moves the name or changes row height. The absolute-position overlay and the avatar/checkbox cross-fade are removed; the checkbox is hidden with visibility/opacity, revealed on row hover or `aria-pressed`, and forced visible on phone via `respondent-control__checkbox--always-visible`. Its border color moved from an inline style to scoped CSS: neutral outline by default, primary action green on hover and checked.
- `RespondentsList.test.ts` covers the viewport reveal matrix, checkbox-after-status DOM order, selected-row status retention with a checked box, and a source guard against the absolute overlay; the mount helper gained a `curRespondents` option.
- New `e2e/specs/timed-event-respondent-selection-firefox.spec.ts` asserts desktop hover reveal, indicator retention, unchanged name position, and the green hover border; the new phone test in `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` asserts the unhovered checkbox, neutral border, checkbox right of the status, name right of the checkbox, and that selecting keeps the status visible with a green checked border.
- `backlog/backlog.md` marks the inbox question resolved by FR-134.

## Why

The selection checkbox used to cross-fade on top of the row's status/avatar, hiding a selected response's availability status, which defeats comparing selected responses. The new layout keeps both signals visible and the name column stable.

## Verification

- Frontend: `npm run lint` (2 pre-existing warnings in `NewSignUp.test.ts`), `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (154 files / 1224 tests) pass.
- E2E: the new firefox-desktop spec passes; the full firefox-touch spec passes (8 tests); a combined firefox-desktop + firefox-touch run with the border-color assertions passes (2 tests).
- Root `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, and `npm run fmt:check` pass; `graphify update .` completed.
- E2E package lint and typecheck pass; `e2e npm run fmt:check` still flags only the unmodified pre-existing `specs/timed-event-access-transfer-firefox.spec.ts`.
- No swagger annotations, environment docs, plugin API shapes, or migration contracts changed.
<!-- SECTION:FINAL_SUMMARY:END -->
