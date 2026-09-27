---
id: TASK-0309
title: Make the desktop event-header details and controls independent columns
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 18:48'
updated_date: '2026-09-21 20:04'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - e2e/specs/event-description-header-layout.spec.ts
  - e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts
  - e2e/specs/event-toolbar-mobile-layout.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-139.md
  - docs/requirements/README.md
modified_files:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - e2e/specs/event-page-no-responses-layout.spec.ts
  - e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts
  - docs/requirements/functional/fr/FR-139.md
  - docs/requirements/README.md
priority: medium
ordinal: 310000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The desktop event header currently stacks three full-width rows, and each row pairs a details item on the left with a control on the right. Because each row is one flex row, the two sides share row height and vertical position: a two-line title, a saved occurrence span, or a wrapped metadata action cluster changes where the right-side controls sit, and left-side details move when right-side controls wrap or grow. The header therefore reads as a single grid even though the left details and right controls are independent.

The work order is to lay the desktop header out as two independent columns, one for event details (title and saved occurrence span, metadata actions, description) and one for event controls (availability actions, options, and schedule or delete-availability controls), while keeping the same vertical gaps on both sides so the current aligned look holds when content heights match. Mobile order and layout must not change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a desktop viewport, the event header lays out event details (title and saved occurrence span, metadata actions, description) in one column and event controls (availability actions, options, and schedule or delete-availability controls) in a second column, with both columns starting at the same top edge.
- [x] #2 Both columns use the same vertical gap between stacked groups, so with a single-line title and no saved occurrence span the first group in each column starts at the same top edge; when the availability actions render as a single row, the title aligns with the primary availability control, the metadata action row aligns with the options row, and the description row starts on the same line as the schedule or delete-availability control.
- [x] #3 Changing the height of the details column, such as a wrapped title or a multiline description, does not change the vertical position of the controls column, and changing a control's height does not move the details column.
- [x] #4 A new regression test covers the details and controls columns as independent stacks.
- [x] #5 The mobile event header keeps its current single-column order and layout in both visual and source (focus) order, including availability-group actions between the title and the metadata actions.
- [x] #6 The controls column keeps its 22rem width and shared right-edge alignment, and the no-response desktop header keeps the half-width single-column controls behavior.
- [x] #7 The desktop header column behavior is recorded as a functional requirement with an index row.
- [x] #8 Frontend lint, fmt:check, typecheck, build, and test:unit pass, and the affected e2e specs pass on chromium-desktop, chromium-mobile, and firefox-desktop where their projects run them.
- [x] #9 In the desktop controls column, the availability actions render above the event options, and the event options render above the schedule or delete-availability action.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Requirement: create `docs/requirements/functional/fr/FR-139.md` for the desktop event-header two-column layout with independent vertical flow and matching gaps, and add its row to `docs/requirements/README.md`.

2. `frontend/src/views/Event.vue` desktop header restructure:
   - `#event-header` keeps mobile `tw:flex tw:flex-col tw:gap-3` and adds `tw:sm:flex-row tw:sm:items-start tw:sm:gap-4`.
   - Add a details column wrapper: `tw:contents tw:sm:flex tw:sm:min-w-0 tw:sm:flex-1 tw:sm:flex-col tw:sm:gap-3` containing the title group (with saved occurrence span), `#event-header-meta-row`, and the description group.
   - Add a controls column wrapper: `tw:contents tw:sm:flex tw:sm:min-w-0 tw:sm:flex-col tw:sm:gap-3` containing the availability actions group, the options groups, and the schedule/delete-availability group.
   - Give the six groups mobile order classes (`tw:order-1` ... `tw:order-6`) so the promoted `display: contents` children keep the current mobile order: title, availability/group actions, metadata actions, description.
   - Drop the now redundant `tw:flex-1`/row wrappers from left-side groups so they stack in a column; keep `#event-header-meta-row`, `#event-header-button-row`, `.event-header-row`, `.desktop-event-header-actions`, and the existing control classes where tests and styles rely on them.
   - Keep the 22rem `.desktop-event-header-actions` control column; give the metadata row a desktop min-height so its 28px buttons stay centered against the 40px options controls.
   - Add stable ids for e2e selectors (`event-header-title`, details/controls column ids).

3. Unit tests in `frontend/src/views/Event.test.ts`: update the `.event-header-row` structure test and the editing source-string assertions for the new nesting; add a regression test that asserts details and controls are separate column stacks with matching gap classes and preserved mobile order.

4. E2E: replace the brittle `#event-header > .event-header-row:first-child > ...` title selectors in `e2e/specs/event-page-no-responses-layout.spec.ts` and `e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts` with the stable title id; adjust the alignment assertions to the new column geometry. Add an independence assertion (for example a multiline description or wrapped title leaves the controls column y positions unchanged).

5. Verify: frontend lint, fmt:check, typecheck, build, test:unit; targeted e2e on chromium-desktop plus mobile/firefox coverage where the projects run the specs; `npm run format:markdown`; `graphify update .`.

Final approach after review follow-up: the order classes and the display:contents mobile reordering were dropped. Phone group events render a `#event-header-mobile-group-actions` copy between the title and the metadata actions, the desktop controls availability group is desktop-only (`!isPhone`), and the desktop column tests assert DOM order instead of order classes. The first e2e test is renamed to column wording, `.event-header-row` is removed from `#event-header-meta-row`, the dead `sm:mb-2` is removed with `tw:sm:min-h-10` added to `#event-header-title`, and the independence e2e covers a seeded description plus a taller-control-group perturbation.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
User clarification: the controls column order is availability actions first, then the event options (Show best times / More options), then the schedule action. The first implementation draft gave the availability group a mobile order of 2 while the options and schedule groups kept the default order, which sorted the availability group last on desktop; the options and schedule groups now carry explicit orders greater than the availability group's order.

PROGRESS (implementation complete, verification paused before e2e).

Done:
- `docs/requirements/functional/fr/FR-139.md` created (proposed) and indexed in `docs/requirements/README.md`. It states the two independent stacks, the matching gaps, the controls order (availability above options above schedule/delete-availability), and the mobile exclusion.
- `frontend/src/views/Event.vue`: `#event-header` is mobile `flex-col gap-3` plus desktop `sm:flex-row sm:items-start sm:gap-4`. Added `#event-header-details-column` (`tw:contents sm:flex sm:flex-1 sm:flex-col sm:gap-3`) holding `#event-header-title`, `#event-header-meta-row`, and `#event-header-description-row`; added `#event-header-controls-column` (`tw:contents sm:flex sm:flex-col sm:gap-3`) holding the availability group, the options groups, and the schedule/delete groups. Mobile order is preserved with order classes on the promoted `display: contents` children: title 1, availability 2, meta 3, description 4; options are 5 and schedule/delete are 6 so the controls column renders availability, options, schedule top to bottom on desktop. The metadata row keeps `.event-header-row` and adds `tw:sm:min-h-10` so its 28px buttons stay centered against the 40px options controls. Stable ids added: `#event-header-title`, `#event-header-details-column`, `#event-header-controls-column`, `#event-header-description-row`.
- `frontend/src/views/Event.test.ts`: `EventTestState` gained `description?: string`; updated the metadata-row source assertion to the new class string; reworked the no-response test to look for the collapse switch in `#event-header-controls-column` (and not in the details column or `#event-header-actions`); replaced the `.event-header-row` count test with `renders desktop header details and controls as independent column stacks`, which asserts the column classes, group membership, and the controls order classes (order-2/order-5/order-6); updated the editing options source assertion to `class="desktop-event-header-actions tw:order-5"`. `npx vitest run src/views/Event.test.ts` passes (95 tests).
- E2E: both brittle title selectors now use `#event-header-title` (`e2e/specs/event-page-no-responses-layout.spec.ts`, `e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts`). A new chromium-desktop test `desktop header details and controls stack independently` in `event-page-no-responses-layout.spec.ts` seeds a short-name and a long-name event, asserts matching gaps for the paired groups on the short event, and asserts a wrapped title grows the details column while the controls column offsets and height stay unchanged. Not executed yet.
- Checks already green: frontend `fmt:check`, `lint` (only the 2 pre-existing NewSignUp.test.ts warnings), `typecheck`, `build`; full `test:unit` passed 1271 tests before the final order-class edit, and `Event.test.ts` passed after it.

Remaining work:
- Run `npm run format:markdown` and `npm run fmt:check` from the repo root, then `graphify update .` (DoD items).
- Re-run the full frontend `npm run test:unit` after the final order-class edits.
- Run the focused e2e on chromium-desktop: from `e2e/`, `npm run test:e2e -- --project=chromium-desktop specs/event-page-no-responses-layout.spec.ts`. Then run the desktop description spec and, for the archived banner selector change, a firefox run that matches `timed-event-*-firefox.spec.ts` (the firefox-desktop project only matches `timed-event-.*firefox` specs, so `event-page-no-responses-layout.spec.ts` does not run there). Mobile order coverage: `npm run test:e2e -- --project=chromium-mobile specs/event-page-no-responses-layout.spec.ts` plus a group-event mobile flow if a suitable spec/seed exists.
- Verify each acceptance criterion with evidence, then finalize per the task-finalization workflow.

Environment notes:
- The first e2e attempt was aborted by the user and left a Vite process on 4174 and the `timeful-test` compose stack running. Both are cleaned up now (killed the Vite pid, ran `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml down` from the repo root). The external Go cache volumes `timeful-test-go-build-cache` and `timeful-test-go-mod-cache` still exist.
- The user mid-review clarified the controls order (availability first, then options, then schedule); that is implemented and recorded in FR-139 and the task AC #9. The first draft had the availability group order-2-only, which sorted it last on desktop because the options/schedule groups defaulted to order 0.
- `backlog/backlog.md` shows as modified in git; it is Backlog-managed, not part of the change.

REVIEW FINDINGS (staged work, 2026-09-21).

Verification run during review, all green:
- e2e chromium-desktop specs/event-page-no-responses-layout.spec.ts: 4/4 pass, including `desktop header details and controls stack independently` and the pre-existing center-alignment test.
- e2e chromium-mobile on the same spec: all 4 tests skipped, so that spec provides no mobile coverage.
- e2e firefox-desktop specs/timed-event-archived-banner-alignment-firefox.spec.ts: 2/2 pass.
- e2e chromium-desktop specs/event-description-header-layout.spec.ts and specs/event-page-days-only-layout.spec.ts: 12/12 pass (adjacent header specs not modified by this task).
- frontend fmt:check, lint (only the 2 pre-existing NewSignUp.test.ts warnings), typecheck, build: pass.
- frontend test:unit: 1271 tests pass after the final order-class edits.
- The fresh production CSS contains tw:contents, tw:order-2, tw:order-5, tw:order-6, tw:sm:min-h-10; it contains no rule for the unprefixed sm:mb-2.

The earlier note that e2e was not executed is superseded by the results above.

REVIEW FINDINGS TO RESOLVE BEFORE FINALIZING:

1. FR-139 AC #1 and task AC #2 are overbroad. docs/requirements/functional/fr/FR-139.md:26 promises the title aligns with the primary availability action and the metadata row aligns with the options row. For availability groups the controls stack starts with the Today/Refresh row (frontend/src/views/Event.vue:287), so the primary action sits one row lower and the metadata row lands roughly 44-48px above the options row; the old row layout kept those aligned. Fix: scope the alignment clause to availability groups that render as a single row, or restate it as the first group of each stack starting at the same top edge; add a group-event e2e only if that alignment is wanted.

2. AC #5 has no regression coverage and mobile focus order changed. The new unit test asserts only the controls order classes order-2/5/6 (frontend/src/views/Event.test.ts:3546-3552); the details children order-1/3/4 and the full six-item mobile order are unasserted, and every mobile test in event-page-no-responses-layout.spec.ts is skipped. Because display: contents plus order reorders visually without changing DOM order, phone group events now read/tab to Today/Refresh after Edit/Copy/Manage access. Fix: assert all six order classes or the visual order, and decide and document the focus-order behavior (only responsive markup duplication preserves both DOM and visual order).

3. First-row alignment silently depends on a dead class. frontend/src/views/Event.vue:176 still carries sm:mb-2, which generates no CSS rule (Tailwind is prefixed tw:, Vuetify responsive spacing is mb-sm-2); the title group is 40px only because that class does nothing. If it is ever corrected to tw:sm:mb-2, the title group becomes 48px and both e2e alignment tests fail. Fix: delete it or replace it with tw:sm:mb-0, and add tw:sm:min-h-10 to #event-header-title so the matching first-row height is intentional.

4. The new e2e test does not cover its ACs' interesting halves. Both seeds have no description, so the description-starts-with-schedule clause of AC #2 is never measured; the control-height direction of AC #3 is untested; and the schedule check at e2e/specs/event-page-no-responses-layout.spec.ts:419-426 is a gap equation rather than a paired-offset check. Fix: give one seed a description and compare #event-header-description-row y with #desktop-schedule-event-btn y; optionally change a control height and re-measure the details offsets.

5. Stale naming. The first e2e test is still `pairs each header row with one action column` with `header-row` failure messages (e2e/specs/event-page-no-responses-layout.spec.ts:12, 101-103), and .event-header-row on #event-header-meta-row (frontend/src/views/Event.vue:220) is now only a test hook. Fix: rename/update the messages and drop the class or comment that tests rely on it.

6. Finalization hygiene: npm run format:markdown (DoD #4) and graphify update . (DoD #6) are still pending; update these notes with the resolution of each finding, then tick the acceptance criteria at finalization.

Cleared in review: the #event-header-title locator change still aligns centers because the wrapper equals the inner title row height; the schedule group right edge still equals #event-header right; isPhone (Vuetify xs, threshold 640 in frontend/src/plugins/vuetify.ts) matches the Tailwind sm breakpoint, so the responsive switch points agree.

REVIEW RESOLUTION (2026-09-21): FR-139 AC #1 and task AC #2 were scoped to availability groups that render as a single row, with first-group top-edge alignment stated separately. The independence e2e seeds a non-group no-response event, which is the single-row availability case.

REVIEW RESOLUTION: phone DOM/focus order is preserved. Added the phone-only `#event-header-mobile-group-actions` copy between the title and the metadata actions inside the details column, made the desktop controls availability group desktop-only (`!isPhone`), and removed every `tw:order-*` class so DOM order equals visual order on both viewports. New unit test `keeps the phone header order with group actions above the metadata actions` asserts the four mobile groups and the empty controls column; new chromium-mobile e2e test `mobile group header keeps availability actions between the title and the metadata actions` asserts the geometry and the absence of duplicate actions.

REVIEW RESOLUTION: removed the dead `sm:mb-2` from the title group and added `tw:sm:min-h-10` to `#event-header-title` so the 40px first-row height is intentional. Removed the now unused `.event-header-row` class from `#event-header-meta-row` and updated its two unit assertions. Renamed the first e2e test to column wording and updated its failure message.

REVIEW RESOLUTION: the independence e2e now seeds a saved description, compares the `#event-header-description-row` y offset with `#desktop-schedule-event-btn`, and adds a taller-control-group step that grows the first controls group via min-height and asserts the details offsets and height are unchanged while the options and schedule rows shift down.

FINAL VERIFICATION: frontend lint (only the 2 pre-existing NewSignUp.test.ts warnings), fmt:check, typecheck, build, and test:unit (1272 tests) pass. E2E: chromium-desktop event-page-no-responses-layout 4/4; chromium-mobile new group header test plus event-toolbar-mobile-layout mobile tests pass; event-description-header-layout 3/3 desktop and 3/3 mobile; firefox-desktop timed-event-archived-banner-alignment 2/2. Combined project run: 18 passed, 8 skipped by project match.

DoD #6 `graphify update .` was skipped on explicit user instruction; the worktree is concurrently migrating code intelligence from Graphify to codebase-memory-mcp. `npm run format:markdown` ran and reformatted the FR-139 index row in docs/requirements/README.md. A shell restart dropped the devShell PLAYWRIGHT_BROWSERS_PATH, so the final e2e run set the nix store browsers path explicitly.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

The desktop event header in `frontend/src/views/Event.vue` now lays out as two independent columns instead of three coupled full-width rows. `#event-header-details-column` stacks the title (with the optional saved occurrence span), the metadata actions, and the description; `#event-header-controls-column` stacks the availability actions, the event options, and the schedule or delete-availability action. Both wrappers use the same `tw:sm:gap-3` stack gap and start at the same top edge. On phones both wrappers are `display: contents`, so the header stays a single column.

## Why

In the old row layout the two sides of each row shared row height, so a wrapped title or saved span moved the controls and a taller control moved the details. FR-139 records the two-column behavior and its index row was added to `docs/requirements/README.md`.

## Behavior

- With a single-line title and no saved occurrence span, the paired groups align; the metadata row keeps `tw:sm:min-h-10` so its 28px buttons center against the 40px options controls.
- A wrapped title or multiline description moves only the details column; a taller control group moves only the controls column (covered by an e2e perturbation step).
- Phone group events keep their previous source and visual order: a phone-only `#event-header-mobile-group-actions` row renders between the title and the metadata actions, the desktop controls availability group is desktop-only (`!isPhone`), and no CSS `order` classes are used, so DOM, focus, and visual order agree on both viewports.
- Availability renders above options, which render above schedule/delete-availability; the controls column keeps its 22rem width and right edge, and the no-response header keeps the half-width single-column behavior.

## Files

- `frontend/src/views/Event.vue`
- `frontend/src/views/Event.test.ts`
- `e2e/specs/event-page-no-responses-layout.spec.ts`
- `e2e/specs/timed-event-archived-banner-alignment-firefox.spec.ts`
- `docs/requirements/functional/fr/FR-139.md`
- `docs/requirements/README.md`

## Verification

- Frontend: `lint` (only the 2 pre-existing NewSignUp.test.ts warnings), `fmt:check`, `typecheck`, `build`, and `test:unit` (1272 tests) pass.
- E2E combined run across chromium-desktop, chromium-mobile, and firefox-desktop for the four affected specs: 18 passed, 8 skipped by project match. This includes the reworked `desktop header details and controls stack independently` test (description/schedule alignment plus taller-control-group independence) and the new chromium-mobile `mobile group header keeps availability actions between the title and the metadata actions` test.
- `npm run format:markdown` ran.

## Exempt or skipped DoD items

- DoD #6 `graphify update .` was skipped on explicit user instruction; the worktree is concurrently migrating code intelligence from Graphify to codebase-memory-mcp.
- DoD #5 and #7: no Swagger annotations and no `scripts/` or `prettier/` changes.
- DoD #8: no contract-affecting changes beyond the requirement document and its index row.
<!-- SECTION:FINAL_SUMMARY:END -->

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
