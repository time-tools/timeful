---
id: TASK-0283
title: >-
  Let non-owner Event Guests manage multiple Event Responses in Blind
  Availability Mode
status: Done
assignee:
  - opencode
created_date: '2026-09-19 21:50'
updated_date: '2026-09-19 22:33'
labels: []
dependencies: []
references:
  - frontend/src/views/Event.vue
  - server/routes/event_routes.go
  - e2e/specs/timed-event-visitor-identities-firefox.spec.ts
documentation:
  - docs/requirements/functional/fr/FR-084.md
  - docs/requirements/functional/fr/FR-061.md
  - docs/requirements/functional/fr/FR-059.md
  - PLUGIN_API_README.md
modified_files:
  - frontend/src/views/Event.vue
  - frontend/src/views/Event.test.ts
  - >-
    frontend/src/components/schedule_overlap/useScheduleOverlapPreferences.test.ts
  - server/routes/identity_test.go
  - e2e/specs/timed-event-visitor-identities-firefox.spec.ts
  - e2e/specs/timed-event-plugin-firefox.spec.ts
  - PLUGIN_API_README.md
priority: high
type: feature
ordinal: 283000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A non-owner Event Guest on a Blind Availability Mode event cannot add a second Event Response: the Event Page hides the secondary add action for non-owners in blind mode and the browser-plugin flow refuses guest availability for non-owners, so the FR-084 capability (a non-owner Event Guest can create and manage multiple Event Responses the guest is authorized to manage) is unreachable through the product. The privacy clauses must keep holding: the guest must only ever see and manage responses under the guest's own Event Visitor Identity, while the Event Owner still sees every response. Discovered while fixing TASK-0282, which restored additional responses for non-blind events and deliberately left blind-mode gating unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a Blind Availability Mode Event Page, a non-owner Event Guest who already owns one Event Response can use the secondary Add availability action on desktop and mobile to create another Event Response owned by the guest's Event Visitor Identity.
- [x] #2 A non-owner Event Guest can select, edit, and delete each Event Response the guest owns in Blind Availability Mode, and cannot access another guest's Event Response.
- [x] #3 In Blind Availability Mode, the page continues to expose to a non-owner only the Event Responses the guest is authorized to manage; other responses, their availability, identities, and count do not appear in payloads, Schedule Overlap views, or response selection.
- [x] #4 The browser-plugin availability flow permits a non-owner to create and manage the guest's own Event Responses in Blind Availability Mode without changing window.postMessage payload shapes, and PLUGIN_API_README.md reflects the behavior.
- [x] #5 Event Owner visibility of every Event Response and all non-blind behavior stay unchanged.
- [x] #6 Regression coverage includes unit tests for the blind-mode non-owner action areas and owned-response selection, server route coverage for blind-mode response authorization, and a Firefox e2e journey where a non-owner creates and manages a second Event Response in blind mode.
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
1. Add failing frontend unit coverage in `frontend/src/views/Event.test.ts`: a blind-mode non-owner Event Guest state (`eventVisitorId` set, `canManageEvent: false`, `blindAvailabilityEnabled: true`, PostgreSQL `publicId` + `canEdit` responses) where one owned response renders the desktop and mobile secondary Add availability action and starts the add flow, and two owned responses stay selectable through the owned-response chooser.
2. Fix `frontend/src/views/Event.vue` `showSecondaryAddAvailabilityAction`: the owner-only blind-availability gate must no longer hide the secondary action from a non-owner who owns editable responses, because the PostgreSQL server proves ownership per Event Visitor Identity and already filters blind payloads.
3. Fix the plugin `setSlots` visitor-identity target resolution and gating: a supplied `guestName` must select an owned response by normalized name or create a new response under the calling Event Visitor Identity instead of reusing the selected response, and the blanket blind non-owner refusal must not block visitor-identity events.
4. Add server route coverage for blind-mode response authorization: a non-owner creates multiple responses, edits and deletes each owned response, cannot read or mutate another visitor's response, the owner sees every response, and the non-owner payload omits `numResponses` and other responses.
5. Extend Firefox e2e: a blind event journey where a non-owner creates a second Event Response through the Event Page, selects and edits both owned responses, and cannot see another visitor's response or the response count; a plugin journey where a non-owner creates and manages responses through `set-slots` in blind mode without payload shape changes.
6. Update `PLUGIN_API_README.md` to document blind-mode non-owner behavior.
7. Run required checks: frontend lint, fmt:check, typecheck, build, test:unit; server route tests through the isolated Compose stack; focused Firefox e2e; `graphify update .`; `npm run format:markdown` for changed Markdown; regenerate Swagger only if annotations change.
Risks: the plugin target-resolution change must not regress non-blind plugin edits; blind payload filtering must stay server-authoritative; the e2e grid-save journey must name the new response through the guest dialog.

Review follow-up: keep the blind/non-blind parity in `showSecondaryAddAvailabilityAction` (deliberately broader than the original owner-of-a-response wording), correct the README authority wording, fix the mislabeled preferences test, strengthen the chooser unit test to fail pre-fix, and add signed-in owner plugin e2e coverage for unmatched `guestName` creation.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: `frontend/src/views/Event.vue` now computes `showSecondaryAddAvailabilityAction` from viewer authority alone (signed-in viewer or owned editable response), so blind and non-blind modes share one rule; the old owner-only blind gate is gone. The plugin `setSlots` gate only refuses guest-named calls for events without an Event Visitor Identity, and in blind mode a supplied `guestName` resolves to the owned response with that normalized name or creates a new response under the caller's Event Visitor Identity; non-blind resolution keeps the old selected-response fallback.

Coverage: three blind-mode non-owner tests in `frontend/src/views/Event.test.ts` (desktop add, mobile add, chooser selection; all three fail against pre-fix `Event.vue`), owned-response filtering in `useScheduleOverlapPreferences.test.ts`, `/responses` schedule-endpoint filtering in `server/routes/identity_test.go`, Firefox e2e for the Event Page journey and the non-owner plugin journey, plus a signed-in owner plugin test for unmatched `guestName` creation.

Review round fixes: corrected the README authority wording to include account-associated responses; renamed the mislabeled preferences test; strengthened the chooser test so it fails pre-fix; added the owner plugin e2e test. Kept the broad secondary-action condition because it matches non-blind behavior and FR-084; the task plan described only the response-owner case, so the broader parity is recorded here.

Checks: frontend lint (2 pre-existing NewSignUp warnings), fmt:check, typecheck, build, test:unit (151 files, 1192 tests); full server route suite via the isolated Compose stack; `e2e --project=firefox-desktop` on the two touched specs (7 passed); root `npm run format:markdown`; `graphify update .`. Swagger unchanged because no annotations changed.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-19 22:33
---
Review round findings (2026-09-20):
1. `showSecondaryAddAvailabilityAction` removed the owner-only blind gate entirely, which is broader than the plan's "non-owner who owns editable responses" wording: signed-in non-owners with no owned response also get the secondary add action. Kept intentionally for blind/non-blind parity; recorded in the implementation notes and final summary.
2. The owner plugin path in blind mode was untested and undocumented. It was in fact unreachable before the gate removal (the old `isSignedInOwner` check is always false for PostgreSQL events because the API emits the zero UUID owner sentinel, so every `guestName` call was refused). Added a signed-in owner Firefox e2e test that fails pre-fix and passes post-fix.
3. `PLUGIN_API_README.md` said response authority comes only from the browser's Event Visitor Identity; account-associated responses are also manageable. Wording corrected.
4. `useScheduleOverlapPreferences.test.ts` was mislabeled as blind-mode though the fixture cannot carry the blind flag; renamed to describe the `publicId` + `canEdit` filtering it actually verifies.
5. The chooser unit test passed pre-fix and was not regression coverage; added a secondary-action assertion so all three blind-mode unit tests fail against pre-fix `Event.vue`.
6. Task record finalization was pending; completed now.
7. Out of scope, pre-existing: `e2e/specs/timed-event-access-transfer-firefox.spec.ts` fails the `e2e` oxfmt check at HEAD; untouched by this task.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A non-owner Event Guest on a Blind Availability Mode event can now add and manage multiple Event Responses. The Event Page no longer applies the owner-only blind gate to the secondary add action (`frontend/src/views/Event.vue`), so it follows the same rule as non-blind events: a signed-in viewer or a guest owning an editable response sees it. The plugin `setSlots` flow no longer refuses blind non-owners that pass a `guestName`; it resolves a supplied name to the owned response with that normalized name or creates a new response under the caller's Event Visitor Identity, while non-blind target resolution and owner visibility are unchanged.

Blind privacy stays server-authoritative: `eventResponses` filters blind payloads by visitor authorization and deletes `numResponses`, and `/events/{id}/responses` applies the same filter for Schedule Overlap and plugin `get-slots`.

Tests: three blind-mode non-owner tests in `frontend/src/views/Event.test.ts` (desktop add, mobile add, owned-response chooser; all three fail against pre-fix `Event.vue`), owned-response filtering in `useScheduleOverlapPreferences.test.ts`, `/responses` filtering in `server/routes/identity_test.go` (alongside the existing create/edit/delete/owner/attacker coverage), Firefox e2e for the Event Page and non-owner plugin journeys, and a new signed-in owner plugin e2e test for unmatched `guestName` creation (fails pre-fix, passes post-fix).

Verification: frontend lint (2 pre-existing NewSignUp warnings), fmt:check, typecheck, build, test:unit (151 files, 1192 tests); full server route suite through the isolated Compose stack; `e2e --project=firefox-desktop` on `timed-event-plugin-firefox.spec.ts` and `timed-event-visitor-identities-firefox.spec.ts` (7 passed); root `npm run format:markdown`; `graphify update .`. Swagger unchanged (no annotation changes).

Notes: removing the gate also exposes the secondary add action to signed-in non-owners without owned responses, matching non-blind behavior and FR-084. Pre-existing and unrelated to this task: `e2e/specs/timed-event-access-transfer-firefox.spec.ts` fails the `e2e` oxfmt check at HEAD.
<!-- SECTION:FINAL_SUMMARY:END -->
