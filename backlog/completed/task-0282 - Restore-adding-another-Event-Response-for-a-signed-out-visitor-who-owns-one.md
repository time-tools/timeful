---
id: TASK-0282
title: Restore adding another Event Response for a signed-out visitor who owns one
status: Done
assignee:
  - opencode
created_date: '2026-09-19 21:39'
updated_date: '2026-09-19 21:48'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - frontend/src/components/schedule_overlap/useScheduleOverlapPreferences.ts
  - server/routes/event_routes.go
documentation:
  - docs/requirements/functional/fr/FR-001.md
  - docs/requirements/functional/fr/FR-059.md
  - docs/requirements/functional/fr/FR-084.md
priority: high
type: bug
ordinal: 282000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A signed-out Event Visitor who has already created one Event Response cannot create another on a PostgreSQL-backed event: once the first response exists, the primary action becomes Edit availability and the secondary Add availability / Add guest availability action disappears, so there is no way to reach the add flow again. The capability is required by FR-001 (an Event Guest can create and own multiple Event Responses) and reinforced by FR-059 and FR-084. The event payload now identifies response ownership through the server-proven publicId plus canEdit pair instead of the legacy guest ownership fields the secondary-action gate still reads, so the gate no longer recognizes responses the visitor already owns.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Given a PostgreSQL-backed, non-blind Event Page whose only Event Response is owned by the current signed-out Event Visitor, the page renders the secondary Add availability action (the signed-out copy) beside an enabled Edit availability primary action on desktop and mobile.
- [x] #2 Activating the secondary Add availability action starts the add-availability flow for the current Event Visitor instead of editing the existing Event Response.
- [x] #3 Existing action-area states render unchanged: zero-response events, owner and owner-capability viewers, signed-in respondents, Group Events, Sign-up Form Events, editing and scheduling states, archived read-only events, and blind-availability gating.
- [x] #4 Unit regression coverage represents PostgreSQL response payloads with publicId and canEdit and without legacy guest ownership fields, and asserts the secondary Add availability action for the owning signed-out Event Visitor on desktop and mobile.
- [x] #5 After the visitor adds a second Event Response, both owned Event Responses remain selectable and editable through the existing owned-response selection.
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
1. Add failing unit regression coverage in `frontend/src/views/Event.test.ts` using PostgreSQL response payloads (`publicId` + `canEdit`, no `guestOwnershipMode`/`guestId`/`user`) with the schedule-overlap stub keyed by response public ID: desktop secondary add action renders and starts the add flow, mobile secondary add action renders, and two owned PostgreSQL responses remain selectable through the owned-response chooser.
2. Fix `frontend/src/views/Event.vue`: the secondary-action gate must recognize server-proven editable ownership through the same `ownedGuestEditOptions` derivation that already matches responses by public ID, replacing the `guestAddedAvailability` checks that read stripped guest fields. Keep the existing disabled-Edit-primary early return, blind-availability gating, and all other state gates unchanged.
3. Run `npx vitest run src/views/Event.test.ts` and confirm the new tests fail before the fix and pass after it.
4. Add a focused Firefox e2e regression for a signed-out visitor with one API-seeded response if the isolated test stack is usable, and run it with `npm run test:e2e -- --project=firefox-desktop`.
5. Run the required frontend checks: `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build`, `npm run test:unit`, plus `graphify update .` from the repo root.
6. Finalize TASK-0282 with objective evidence per acceptance criterion.

Risks: the unit stub's `ownedGuestResponses` must mirror the real `useScheduleOverlapPreferences` derivation (responses with `publicId` and `canEdit`); the Postgres payload has no `user` object, so owned-response naming falls back to the response map key/name. Blind-availability non-owner behavior is intentionally out of scope.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reproduced at the view layer before the fix: a PostgreSQL-shaped event (`publicId` + `canEdit`, no `guestOwnershipMode`/`guestId`/`user`) with a visitor-owned response rendered an enabled Edit availability primary but no secondary add action. The new desktop and mobile unit tests failed before the fix and pass after it; the two-owner-response chooser test passed before and after because it uses the already-correct `ownedGuestEditOptions` path.

Root cause: the removed `guestAddedAvailability` computed matched `ownedGuestResponses` against response fields the server strips in `responseModel` (server/routes/event_routes.go), so it was always false for PostgreSQL payloads. The gate now uses `ownedGuestEditOptions`, the same public-ID ownership derivation the primary action already uses.

Verification: `npx vitest run src/views/Event.test.ts` 72/72; full unit suite 151 files / 1186 tests; `npm run lint`, `npm run fmt:check`, `npm run typecheck`, `npm run build` pass; e2e package `lint` and `typecheck` pass. Firefox e2e `timed-event-visitor-identities-firefox.spec.ts` new test passes with the fix and fails without it at the `Add availability` locator. Full `--project=firefox-desktop` run was intentionally skipped at the user's request and remains for CI. `graphify update .` completed.

Follow-up note: `npm run fmt:check` in the e2e package flags the pre-existing untouched `specs/timed-event-access-transfer-firefox.spec.ts`; out of scope for this task.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Restored the secondary Add availability action for a signed-out Event Visitor who already owns an Event Response on a PostgreSQL-backed Event Page, so the visitor can create the additional Event Responses FR-001 grants.

## Root cause

`Event.vue` gated the secondary action on `guestAddedAvailability`, which matched `ownedGuestResponses` against `guestOwnershipMode`, `guestId`, and `user._id`. The server strips those fields in `responseModel` (`server/routes/event_routes.go:72`) and exposes ownership only as `publicId` plus `canEdit`, so the flag was always false for PostgreSQL payloads. The gate then returned false in exactly the state where the visitor owned an editable response: primary Edit availability enabled and no way back into the add flow.

## Changes

- `frontend/src/views/Event.vue`: removed the `guestAddedAvailability` computed and gated `showSecondaryAddAvailabilityAction` on the existing `ownedGuestEditOptions` derivation, which matches responses by public ID. Disabled-Edit-primary, blind-availability, group, sign-up, editing, and scheduling gates are unchanged.
- `frontend/src/views/Event.test.ts`: added a PostgreSQL-shaped response helper and schedule-overlap stubs keyed by response public ID, exposed the `addAvailability` mock, and added three regression tests: desktop secondary action plus add-flow trigger, mobile secondary action plus add-flow trigger, and both owned PostgreSQL responses selectable through the primary chooser.
- `e2e/specs/timed-event-visitor-identities-firefox.spec.ts`: added a Firefox regression that seeds one visitor-owned response through the API, expects the secondary Add availability action, opens the manual choice dialog, and enters editing.

## Verification

- New desktop and mobile unit tests failed before the fix and pass after it; `npx vitest run src/views/Event.test.ts` 72/72.
- Full unit suite: 151 files, 1186 tests passed.
- Firefox e2e: `specs/timed-event-visitor-identities-firefox.spec.ts` 3/3 with the fix; the new test fails without the fix at the Add availability locator.
- `npm run lint`, `npm run fmt:check`, `npm run typecheck`, and `npm run build` pass in `frontend/`; e2e package `lint` and `typecheck` pass; `graphify update .` completed.

## Risks and follow-ups

- Full `--project=firefox-desktop` suite was deferred to CI at the user's request; the focused spec was run locally.
- Blind-availability gating for non-owner guests is intentionally unchanged. FR-084 expects non-owner guests to create multiple Event Responses in blind mode, which remains a separate requirement gap outside this bug.
- e2e `npm run fmt:check` flags the pre-existing, untouched `specs/timed-event-access-transfer-firefox.spec.ts`.
<!-- SECTION:FINAL_SUMMARY:END -->
