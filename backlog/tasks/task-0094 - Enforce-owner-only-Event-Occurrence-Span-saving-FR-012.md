---
id: TASK-0094
title: Enforce owner-only Event Occurrence Span saving (FR-012)
status: In Progress
assignee:
  - opencode
created_date: '2026-08-28 20:35'
updated_date: '2026-09-13 17:28'
labels:
  - backend
  - authorization
  - postgres
  - frontend
dependencies: []
references:
  - server/routes/events.go
  - server/routes/postgres_event_routes.go
  - server/routes/postgres_owner.go
  - server/routes/postgres_owner_test.go
  - server/routes/event_schedule_test.go
  - server/routes/anonymous_event_compatibility_test.go
  - server/routes/postgres_signed_in_event_test.go
  - server/docs/postgres-event-api-contract.md
  - frontend/src/views/Event.vue
documentation:
  - docs/requirements/functional/fr/FR-012.md
  - docs/requirements/functional/fr/FR-018.md
  - docs/requirements/functional/fr/FR-083.md
priority: medium
type: bug
ordinal: 100000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
FR-012 (status: accepted; reworded in TASK-0092) restricts saving, replacing, or clearing the optional Event Occurrence Span to the Event Owner: "Other Event Visitors shall not be able to change it." The application does not enforce this today, and the frontend still offers the Schedule event controls to non-owners.

# Verified current behavior (re-verified 2026-09-13 after the MongoDB retirement)
- The PostgreSQL span handlers postgresSaveSchedule and postgresClearSchedule (server/routes/postgres_event_routes.go:771-772) delegate to postgresUpdateSchedule (same file, starting at line 774), which performs no owner authorization; postgresWritableEvent (server/routes/postgres_owner.go:122) only rejects deleted or archived events.
- The owner-authorization path already exists: postgresOwnerMutation (server/routes/postgres_owner.go:132) locks the event, authorizes through authorizePostgresOwner, and otherwise rejects with the owner-credential-required 403. TASK-0129 established it for settings edits, archive/unarchive, and deletion, and it already accepts an owner-issued Granted EVCC when present.
- The legacy MongoDB handlers saveTimefulSchedule and clearTimefulSchedule no longer exist: server/db/, the MongoDB runtime, and the MongoDB event store were removed by TASK-0199.
- Existing test coverage is misleading: postgres_owner_test.go:206-207 expects stranger schedule PUT/DELETE to return 403, but the event is archived at that point (archived at line 196, unarchived at line 208), so those assertions pass through archived-write rejection rather than owner authorization. No test rejects a non-owner on a writable event. postgres_signed_in_event_test.go:92-93 covers owner success.
- event_schedule_test.go and TestAnonymousTimedEventCompatibilityContract (anonymous_event_compatibility_test.go:200-221) save and clear spans through timedEventRequest, which does not carry cookies, so they mutate spans without proving ownership today.
- Frontend: showScheduleEventButton (frontend/src/views/Event.vue:1381-1382) gates only on editing, sign-up form, and archived state, so non-owners still see the desktop and mobile Schedule event controls, while settings, archive, and delete controls are already hidden through server-proven canEditSettings/canManageEvent capabilities.
- server/docs/postgres-event-api-contract.md:51 states "Public schedule save, replace, and clear remain supported while the event is not archived", which contradicts the intended owner-only authority.

# Scope
- Enforce owner-only Event Occurrence Span mutation for PostgreSQL events per FR-012 by routing the span handlers through the postgresOwnerMutation owner-authorization path established by TASK-0129 (server/routes/postgres_owner.go), following the FR-018 credential model (Event Owner Edit Token, the associated Platform Visitor Identity, or an owner-issued Granted EVCC; FR-018 and FR-083 remain proposed, but the mechanism is implemented and used by the existing owner actions).
- Preserve current body semantics: save/replace stores the ScheduledEvent snapshot with the event name as Summary, clear removes it, an end at or before start is still rejected, and archived events remain read-only.
- Hide or disable the Schedule event controls for visitors without owner authority, consistent with the existing server-proven capability gating for settings, archive, and delete.
- Update route tests so owner success and non-owner rejection are proven on a writable event, and adapt the existing span tests that currently mutate without owner credentials.
- Update server/docs/postgres-event-api-contract.md so the documented schedule authority matches the enforced behavior.

# Constraints
- Keep PostgreSQL access in server/postgres/; route handlers must not query the store directly (AGENTS.md backend conventions).
- Run route tests via the isolated Compose overlay: docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test, then run --rm server-route-test; the test database must be timeful-test or carry a timeful-test- prefix (AGENTS.md Server Test Workflow).
- Keep compatibility coverage intentional: timedEventRequest-based tests must either carry the creation owner cookie or assert owner-only rejection; do not weaken enforcement to keep them green.
- TestPublicTimefulScheduleRejectsEmptyRange calls the endpoint without owner credentials; decide the validation-versus-authorization order deliberately and update that test to match (owner-first authorization is acceptable).
- Add Swag annotations if route signatures or responses change, then regenerate swagger and the frontend API types per AGENTS.md.
- Frontend work follows the required checks and the layout/regression-test conventions in frontend/AGENTS.md.
- FR-012's accepted status means the intended behavior is already the requirement; this task closes the enforcement, UI, and contract gaps. This bug was discovered during TASK-0092's VERIFY step, which deliberately made no code change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Saving, replacing, or clearing an Event Occurrence Span on a PostgreSQL event succeeds only for the Event Owner and is rejected with 403 for every other Event Visitor, matching FR-012's owner-only authority; owner authorization follows the FR-018 credential model (Event Owner Edit Token, the associated Platform Visitor Identity, or an owner-issued Granted EVCC) as implemented by postgresOwnerMutation, and archived events remain read-only.
- [ ] #2 Route tests prove owner success and non-owner rejection on a writable (non-archived) event for both span endpoints and pass via the isolated Compose test stack (postgres-test, server-route-test per compose.test.yaml); existing span tests that mutated without owner credentials are updated to prove owner authority or assert rejection, and the archived-only stranger assertions in postgres_owner_test.go are no longer the only non-owner coverage.
- [ ] #3 The Schedule event controls are not offered to visitors without owner authority, with frontend regression coverage consistent with the existing server-proven capability gating and the frontend required checks.
- [ ] #4 server/docs/postgres-event-api-contract.md documents owner-only schedule save, replace, and clear, including archived read-only behavior.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Implementation Plan (recorded 2026-09-13)

## Decisions
- Body binding and the end-before-start check stay before owner authorization, matching postgresEditEvent (server/routes/postgres_event_routes.go:630). Non-owners with a valid range get 403; malformed or invalid-range requests keep 400. True owner-first ordering is not cleanly achievable because c.Bind with binding:"required" already emits 400 before authorization.
- Reuse canEditEventMetadata for the frontend gate (server-proven canEditSettings && !isArchived for PostgreSQL events, legacy owner fallback otherwise). No new capability or transport field is needed.
- No Swagger regeneration: route signatures and response bodies do not change, and the schedule routes carry no existing Swag annotations to update.

## Step 1 - Backend enforcement (server/routes/postgres_event_routes.go:771-824)
- Rewrite postgresUpdateSchedule(c, clear): bind/validate as today, then delegate the write to postgresOwnerMutation(c, false, mutate).
- mutate: postgresEventModel(event); set value.ScheduledEvent to nil (clear) or &models.CalendarEvent{Summary: event.Name, StartDate, EndDate}; zero Id/ShortId/OwnerId/NumResponses/ResponsesMap; marshal into event.Payload; tx.UpdateEvent.
- Owner authorization (token, associated Platform Visitor Identity, or owner-issued Granted EVCC) and deleted/archived rejection then come from postgresOwnerMutation + authorizePostgresOwner + postgresWritableEvent (404 deleted, 403 owner-credential-required, 403 event-archived).
- Preserve the 200 success status and the 400 scheduled-event-end-must-follow-start error.

## Step 2 - Backend tests
- postgres_owner_test.go (writable state, before the archive section around line 196):
  - tokenOnly client: PUT /schedule and DELETE /schedule expect 200 (Event Owner Edit Token alone).
  - in the `for _, who := range []*http.Client{baseOnly, stranger}` loop (line 111): PUT and DELETE /schedule expect 403 on the writable event (creator base EVCC is not owner proof; stranger is uncredentialed).
  - in the grant loop (lines 140-171): PUT and DELETE /schedule expect `status` (200 for a GrantsOwner Granted EVCC, 403 for a non-owner grant) to prove FR-018 credential-model parity for the span.
  - keep the archived-only stranger assertions at lines 206-207 as archived read-only coverage, no longer the only non-owner coverage.
- event_schedule_test.go:
  - TestPublicTimefulScheduleCanBeSavedReplacedAndCleared: wrap store.newRouter(t) in compatibilityOwnerBrowser so save, replace, and clear are proven as the creation owner; rename to drop the stale "Public" wording.
  - TestPublicTimefulScheduleRejectsEmptyRange: keep cookie-less to lock the deliberate validation-before-authorization order; rename and add a comment stating that order.
- anonymous_event_compatibility_test.go: TestAnonymousTimedEventCompatibilityContract (lines 137-224) mutates spans through cookie-less timedEventRequest; wrap its router in compatibilityOwnerBrowser so the save/clear carry the creation owner cookie. Blind availability is not enabled there, so owner response visibility does not change expectations.

## Step 3 - Frontend gating
- frontend/src/views/Event.vue (showScheduleEventButton, line 1381): add canEditMetadata.value so non-owners and archived events no longer render the desktop (#desktop-schedule-event-btn) or mobile Schedule controls.
- frontend/src/views/Event.test.ts:
  - extend EventTestState with eventVisitorId/canEditSettings/canManageEvent/isArchived and make createDefaultEventState a PostgreSQL owner (eventVisitorId set, canEditSettings true, canManageEvent true) so existing header/scheduling layout tests keep exercising an owner.
  - add regression tests: a non-owner visitor (eventVisitorId with canEditSettings/canManageEvent false) sees no desktop schedule button and no mobile button containing "Schedule"; an archived owner (isArchived true / canEditSettings false) sees none either.

## Step 4 - Contract doc (server/docs/postgres-event-api-contract.md)
- Line 51: replace "Public schedule save, replace, and clear remain supported while the event is not archived." with owner-only wording for the Event Occurrence Span, including archived rejection.
- Line 83 (Event Owner Edit Token authority) and line 92 (archived read-only) should include occurrence-span mutation alongside settings.
- One sentence per physical line; run npm run format:markdown from the repo root.

## Validation
- docker volume create timeful-test-go-build-cache timeful-test-go-mod-cache
- docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml up -d postgres-test
- docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test
- cd frontend && npm run lint && npm run fmt:check && npm run typecheck && npm run build && npm run test:unit
- npm run format:markdown from the repo root for changed Markdown
- Optional targeted e2e sanity: e2e/specs/event-page-no-responses-layout.spec.ts seeds through page.request, so the owner cookie survives; rerun with --project=chromium-desktop if the schedule gate changes rendered layout.

## Risks / watch items
- Any remaining cookie-less span mutation now returns 403; only event_schedule_test.go, anonymous_event_compatibility_test.go, and postgres_owner_test.go call the span endpoints.
- Wrapping the anonymous contract router changes response visibility to owner (all responses); that test reads only its own response and does not enable blind availability.
- Making the Event.test.ts default fixture PostgreSQL-shaped touches showSecondaryAddAvailabilityAction and plugin set-slots paths; canManageEvent true preserves prior isOwner behavior for authUser owner-1, and no existing test sends set-slots. If fixture fallout appears, fall back to setting authUserState per schedule test instead.
- postgresOwnerMutation sets 200 itself; do not also write a status in postgresUpdateSchedule.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-13 15:05
---
2026-09-13: refreshed for the MongoDB retirement (TASK-0199). Removed the dead Mongo handler references and the legacy-Mongo-endpoints decision, dropped the obsolete Mongo AC, pointed the scope at the existing postgresOwnerMutation path, and switched the test workflow to the postgres-test/server-route-test stack.
---

author: opencode
created: 2026-09-13 17:15
---
2026-09-13 review update: verified the enforcement gap still exists in postgresUpdateSchedule; re-verified the false-positive non-owner coverage in postgres_owner_test.go (archived-only), the cookie-less span tests in event_schedule_test.go and the anonymous compatibility contract, the stale public-schedule wording in server/docs/postgres-event-api-contract.md:51, and the non-owner-visible Schedule controls in frontend/src/views/Event.vue. Updated the credential model to the current FR-018 set (Event Owner Edit Token, associated Platform Visitor Identity, or owner-issued Granted EVCC) and, with user approval, expanded scope to include frontend gating of the Schedule controls plus the contract-doc update; added a fourth acceptance criterion and frontend label.
---
<!-- COMMENTS:END -->
