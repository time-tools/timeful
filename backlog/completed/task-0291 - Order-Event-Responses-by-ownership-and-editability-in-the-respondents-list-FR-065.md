---
id: TASK-0291
title: >-
  Order Event Responses by ownership and editability in the respondents list
  (FR-065)
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-20 21:08'
updated_date: '2026-09-20 21:11'
labels: []
dependencies: []
references:
  - frontend/src/components/schedule_overlap/useRespondentsListState.ts
  - frontend/src/components/schedule_overlap/RespondentsList.vue
  - frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts
documentation:
  - docs/requirements/functional/fr/FR-065.md
  - docs/requirements/README.md
priority: medium
type: bug
ordinal: 291000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The respondents list on availability editing pages never implemented the FR-065 ordering acceptance criterion: before any response subset is selected, responses owned by the acting Event Guest must come first, then other Open Event Responses, then other Protected Event Responses. Today the shared list only moves selected responses to the top and otherwise sorts alphabetically by first name, so a visitor's own editable response can be buried below responses they cannot edit. The behavior is shared by timed, dates-only, and group availability editing on desktop and phone through `useRespondentsListState`, and there is no existing regression coverage for list ordering.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Before any response subset is selected, the list orders Event Responses owned by the acting visitor's Event Guest first, then other Open Event Responses, then other Protected Event Responses
- [x] #2 When a subset is selected, selected Event Responses still move before unselected ones and keep the existing click-time order among themselves; unselected responses keep the ownership tier order
- [x] #3 Canonical responses count as owned when the server exposes publicId with canEdit true; legacy and token responses count as owned when their lookup key is in ownedGuestResponseLookupKeys; other Open Event Responses come from guestEditPolicy open; everything else is protected
- [x] #4 Within each ownership tier, responses stay ordered alphabetically by first name
- [x] #5 Unit tests in useRespondentsListState.test.ts cover the tier order, the canonical canEdit case, and selected-first preservation
- [x] #6 Frontend lint, fmt:check, typecheck, build, and test:unit pass
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
1. Add an exported `responseOrderTier(response, ownedGuestResponseLookupKeys)` helper next to `canGuestEditResponse` in `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts`: returns 0 for a response owned by the acting visitor, 1 for another Open Event Response, 2 for protected/unknown. Canonical responses (`publicId` present) are owned only when `canEdit === true`; legacy and token responses use the existing ownership checks (token: `guestId`; legacy: `user._id`) against `ownedGuestResponseLookupKeys`; a remaining guest response with `guestEditPolicy === "open"` is tier 1; everything else is tier 2. Keep `canGuestEditResponse` unchanged so the pencil/lock rendering is untouched.

2. Extend `frontend/src/components/schedule_overlap/useRespondentsListState.ts`: add an `ownedGuestResponseLookupKeys: ComputedRef<Set<string>>` option, import `responseOrderTier`, and update the `orderedRespondents` comparator. Selected-vs-selected keeps the existing click-time comparison and selected-vs-unselected keeps the existing precedence; the unselected fallback changes from `firstName` compare to ownership tier compare, then `firstName` compare. This preserves FR-066 behavior while adding FR-065 ordering.

3. Wire the existing `ownedGuestResponseLookupKeys` prop in `frontend/src/components/schedule_overlap/RespondentsList.vue` into the `useRespondentsListState` call as a computed `Set`.

4. Extend `frontend/src/components/schedule_overlap/useRespondentsListState.test.ts`: make the state factory support several respondents with parsed-response fields (`publicId`, `canEdit`, `guest`, `guestId`, `guestEditPolicy`, `guestOwnershipMode`) and owned lookup keys; add an ordering describe block covering canonical `canEdit` owned-first, legacy/token owned → open → protected, alphabetical order within a tier, and selected-first preservation with click-time order among selected rows.

5. Known limitation to record in the task notes: canonical responses strip `guestEditPolicy` and `guestId` in `server/routes/event_routes.go`, so the "other Open" tier is only observable for legacy/token responses; canonical sorting effectively collapses to owned (`canEdit`) versus the rest. The mobile respondents scroll e2e spec seeds a browser visitor that owns no response, so its alphabetical-order assumption is unaffected.

6. Run `cd frontend && npm run lint && npm run fmt:check && npm run typecheck && npm run build && npm run test:unit`, then `graphify update .`, then finalize the task per the Backlog finalization workflow. No server or requirement changes are expected; FR-065 stays `proposed` and needs no edit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented `responseOrderTier` next to `canGuestEditResponse` and applied it only to the unselected fallback in `orderedRespondents`, so FR-066 selected-first and click-time behavior is untouched.

Known limitation recorded: canonical responses strip `guestEditPolicy`/`guestId` in `server/routes/event_routes.go`, so the other-Open tier is only observable on token/legacy responses; canonical sorting collapses to owned (`canEdit`) versus protected.

Checks: focused vitest 15/15, full `test:unit` 154 files / 1231 tests, `lint` 0 errors (2 pre-existing warnings in NewSignUp.test.ts), `fmt:check`, `typecheck`, `build` pass; `graphify update .` run.

Working tree also contains an uncommitted `status: accepted` edit to FR-065 made outside this task; left untouched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Implemented the FR-065 ordering acceptance criterion for the shared respondents list used by timed, dates-only, and group availability editing on desktop and phone. Before any response subset is selected, the list now orders responses owned by the acting Event Guest first, then other Open Event Responses, then other Protected Event Responses, with the existing alphabetical order preserved inside each tier. Selected-first ordering from FR-066 is unchanged, including the existing click-time order among selected rows.

## Changes

- `frontend/src/composables/schedule_overlap/useScheduleOverlapUI.ts`: added exported `responseOrderTier(response, ownedGuestResponseLookupKeys)` returning 0 (owned), 1 (other Open), or 2 (protected). Canonical responses with `publicId` count as owned only when `canEdit` is true; token and legacy responses count as owned through `ownedGuestResponseLookupKeys` (`guestId` for token, `user._id` for legacy); a remaining guest response with `guestEditPolicy === "open"` is tier 1; everything else is tier 2. `canGuestEditResponse` is unchanged.
- `frontend/src/components/schedule_overlap/useRespondentsListState.ts`: added the `ownedGuestResponseLookupKeys` option and applied tier-then-name ordering to the unselected fallback of `orderedRespondents`.
- `frontend/src/components/schedule_overlap/RespondentsList.vue`: passed the existing prop into `useRespondentsListState` as a computed `Set`.
- `frontend/src/components/schedule_overlap/useRespondentsListState.test.ts`: added six ordering tests covering owned → open → protected, canonical `canEdit`, legacy ownership keys, alphabetical order within a tier, selected-first preservation with tier-ordered unselected rows, and click-time order among selected rows.

## Verification

- Focused run: `npx vitest run src/components/schedule_overlap/useRespondentsListState.test.ts` — 15 passed.
- `npm run test:unit` — 154 files, 1231 tests passed.
- `npm run lint` — 0 errors (two pre-existing `vue/one-component-per-file` warnings in `NewSignUp.test.ts`).
- `npm run fmt:check`, `npm run typecheck`, `npm run build` — all pass.
- `graphify update .` — graph rebuilt.

## Risks and follow-ups

- Canonical responses strip `guestEditPolicy` and `guestId` server-side in `server/routes/event_routes.go`, so the "other Open" tier is only observable for legacy/token responses; canonical events collapse to owned (`canEdit`) versus the rest. This matches the current server contract and needs no client change.
- No e2e spec was added or required per the approved unit-only coverage decision. The mobile respondents scroll spec is unaffected because its browser visitor owns no response.
- Unrelated working-tree note: `docs/requirements/functional/fr/FR-065.md` has an uncommitted `status: accepted` edit that was not made by this task.
<!-- SECTION:FINAL_SUMMARY:END -->
