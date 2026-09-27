---
id: TASK-0290
title: Move the response selection checkbox to the row's name/action cluster (FR-134)
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 20:54'
updated_date: '2026-09-20 21:02'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - docs/requirements/functional/fr/FR-134.md
documentation:
  - docs/requirements/README.md
  - docs/requirements/functional/README.md
modified_files:
  - docs/requirements/functional/fr/FR-134.md
  - docs/requirements/README.md
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - frontend/src/components/schedule_overlap/RespondentsList.test.ts
  - e2e/specs/timed-event-respondent-selection-firefox.spec.ts
  - e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts
type: enhancement
ordinal: 290000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The responses list on the event page reserves a 16px checkbox slot after the avatar/availability-status indicator in every row, so response names start 20px to the right of the Legend Section item labels (56px vs 36px) and visibly fail to line up with them. Product decision confirmed with the user: move the response checkbox out of the left indicator cluster to the right of the response name, immediately before the row's edit/lock and delete actions.

Confirmed constraints:
- The avatar/availability-status indicator stays the accessible selection control: it keeps the existing button with aria-pressed and click handling; the moved checkbox is a visible pointer target with the same selection action, not a second tab stop.
- The checkbox must remain visible for a selected row on desktop even when the row's hover-revealed action group is hidden, so it cannot live inside the action group's opacity fade.
- Phone always-visible and desktop hover-or-selected reveal behavior, the neutral unchecked outline, the primary action green hovered/checked state, and the no-name-movement guarantee all stay.
- FR-134 is still proposed and tracks this control, so amend it in place rather than creating a superseding requirement.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 FR-134 records the new placement (checkbox after the response name and before the row edit/lock/delete actions, checked state visible for selected rows, checkbox click toggles selection, response names aligned with Legend Section item labels) and docs/requirements/README.md keeps its FR-134 row with an updated title
- [x] #2 Each row keeps the avatar/availability-status indicator in the left slot with the response name immediately after it, and renders the checkbox after the name and before the row action group
- [x] #3 Clicking the checkbox toggles selection exactly once, while the avatar/status button keeps aria-pressed and remains the keyboard selection control
- [x] #4 On a phone viewport the checkbox stays visible on every row; on a desktop viewport an unselected row reveals it only on hover and a selected row keeps it visible; hovered and checked states use the primary action green and the unchecked state uses the neutral outline
- [x] #5 Hovering, selecting, deselecting, and clicking the checkbox move neither the response name nor the row height
- [x] #6 RespondentsList unit tests cover the new DOM order, the reveals matrix, selected-row checkbox and status retention, and the checkbox click toggling selection once
- [x] #7 E2E coverage asserts the checkbox to the right of the name with the name aligned to the Legend item label text, and click-toggle on firefox-desktop, plus the phone order and toggle in firefox-touch
- [x] #8 Frontend lint, fmt:check, typecheck, build, and test:unit pass; changed Markdown is formatted; graphify update runs after the code change
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
1. Requirements: amend `docs/requirements/functional/fr/FR-134.md` in place (title, statement, acceptance criteria) for the checkbox after the response name and before the row actions, response names aligned with Legend Section item labels, the checkbox click toggling selection, and the retained phone/desktop reveal, border-color, and no-movement rules. Update the FR-134 row title in `docs/requirements/README.md`. Run root Markdown format and lint.

2. `frontend/src/components/schedule_overlap/RespondentsList.vue`: keep the left control container (`tw:ml-1 tw:mr-3 tw:flex tw:h-5 tw:shrink-0 tw:items-center`) holding only the `.respondent-control` button with the avatar/status, `aria-pressed`, `aria-label`, and `@click.stop`; remove the checkbox from the button and drop the button's unused `tw:gap-1`. Add a right-side cluster in the name/action row (`tw:flex tw:shrink-0 tw:items-center tw:gap-1`) containing the checkbox first and `.respondent-row-actions` second, so the checkbox stays visible outside the action group's opacity fade. Give the checkbox `aria-hidden="true"` and `@click.stop` emitting `clickRespondent`, and add a `respondent-row--selected` modifier on the row bound to `respondentSelected(...)`. Scope CSS: swap the `[aria-pressed="true"]` reveal and border selectors for `.respondent-row--selected`.

3. Unit tests `frontend/src/components/schedule_overlap/RespondentsList.test.ts`: update the fixed-slot/DOM-order, viewport reveal matrix, status-before-checkbox, and selected-row tests for the new structure; add a test that clicking the checkbox emits `clickRespondent` once; keep the no-absolute-position source guard.

4. E2E: update `e2e/specs/timed-event-respondent-selection-firefox.spec.ts` to assert the checkbox right of the name, name x unchanged across hover/selection and aligned with `.color-legend__indicator-slot + span` x, and checkbox click toggling selection. Update the phone test in `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` for the avatar/status, name, checkbox order and click toggle with the status retained. Confirm `e2e/inspect/src/dom-resolvers.ts` still resolves the checkbox.

5. Checks: frontend `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`; focused e2e with streaming output for the desktop and firefox-touch specs; root Markdown format/check and lint; `graphify update .`; then finalize the task.

Risks: unit tests select the old nested button/checkbox structure; jsdom cannot verify pixel alignment, so the legend alignment assertion belongs in the desktop e2e; the desktop spec needs the seeded response so the Legend response palette renders. Deviation from the creation-time assumption: the checkbox gets its own click handler with `.stop` (confirmed with the user) instead of remaining a purely visual indicator.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Moved the checkbox into a new right cluster with `.respondent-row-actions` instead of inside the action group, because the group fades at opacity 0 on unhovered desktop rows and would hide a selected row's checkbox.

Added `respondent-row--selected` bound to `respondentSelected(...)` and swapped the scoped `[aria-pressed="true"]` reveal/border selectors for it, since the checkbox is no longer a child of the control button.

The checkbox keeps `aria-hidden="true"` because the avatar/status button remains the announced, keyboard-focusable selection control; the checkbox is a redundant pointer target that emits `clickRespondent` with `@click.stop` so a click toggles once.

Legend alignment is verified in the firefox-desktop spec (`name.x` equals `.color-legend__indicator-slot + span` x within 1px), and the selected-not-hovered visibility is verified there with `page.mouse.move(0, 0)`.

Full unit suite: 154 files / 1225 tests pass. Focused firefox-desktop and firefox-touch specs pass. Frontend and e2e lint/typecheck/build pass; the only fmt:check issue is the pre-existing unmodified `specs/timed-event-access-transfer-firefox.spec.ts`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The event-page responses checkbox moved out of the row's left indicator cluster to the right of the response name, immediately before the edit/lock and delete actions, so response names line up with the Legend item labels again.

- `docs/requirements/functional/fr/FR-134.md` is amended in place: the left slot keeps the availability status/avatar with the name right after it, the checkbox sits after the name and before the row actions, names align with the Legend Section item labels, the checkbox click toggles selection once, and the phone/desktop reveal, border colors, and no-movement rules stay. The `docs/requirements/README.md` index row title is updated.
- `frontend/src/components/schedule_overlap/RespondentsList.vue`: the `.respondent-control` button now wraps only the avatar/status and keeps `aria-pressed` and the keyboard path; the checkbox lives in a new right cluster (`tw:flex tw:shrink-0 tw:items-center tw:gap-1`) ahead of `.respondent-row-actions`, so it stays visible outside the action group's opacity fade. The checkbox is `aria-hidden`, `tw:cursor-pointer`, and emits `clickRespondent` with `@click.stop`. A `respondent-row--selected` class drives the scoped reveal and green border selectors, replacing the old `[aria-pressed="true"]` selectors.
- Unit tests cover the avatar-only left slot, the checkbox after the name and before the actions, the reveal matrix, selected-row status retention, and a single `clickRespondent` emit from the checkbox click.
- `e2e/specs/timed-event-respondent-selection-firefox.spec.ts` asserts the name x matches `.color-legend__indicator-slot + span` x within 1px, the checkbox sits between the name and the edit action, the name never moves, and clicking the checkbox selects/deselects and stays visible with the green border while the row is not hovered. `e2e/specs/schedule-overlap-mobile-touch-firefox.spec.ts` asserts the phone order status → name → checkbox, the neutral unchecked border, and checkbox click selection with the status retained.

## Why

The reserved checkbox slot after the avatar pushed names 20px right of the Legend labels (56px vs 36px), so the names looked misaligned. Moving the checkbox to the right cluster makes the left offset exactly the Legend indicator slot offset.

## Verification

- Frontend: `npm run lint` (only the two pre-existing `NewSignUp.test.ts` vue/one-component-per-file warnings), `npm run fmt:check`, `npm run typecheck`, `npm run build`, and `npm run test:unit` (154 files / 1225 tests) pass.
- E2E: `--project=firefox-desktop specs/timed-event-respondent-selection-firefox.spec.ts` passes with the 1px legend alignment and selected-not-hovered assertions; `--project=firefox-touch specs/schedule-overlap-mobile-touch-firefox.spec.ts -g "phone respondent selection sits after the name"` passes.
- E2E package lint and typecheck pass; `fmt:check` still flags only the unmodified pre-existing `specs/timed-event-access-transfer-firefox.spec.ts`.
- Root `format:markdown:check` and `lint:markdown` pass; `graphify update .` completed.
- No swagger annotations, environment docs, plugin API shapes, or migration contracts changed.
<!-- SECTION:FINAL_SUMMARY:END -->
