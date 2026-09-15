---
id: TASK-0199
title: >-
  Retire MongoDB after moving retained integration and supporting data to
  PostgreSQL
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-10 19:56'
updated_date: '2026-09-11 15:02'
labels:
  - postgresql
  - migration
dependencies: []
references:
  - TASK-0190
  - TASK-0190.08
  - docs/design/architecture/adr/ADR-017.md
  - server/docs/postgres-data-boundaries.md
  - server/docs/postgres-anonymous-event-compatibility.md
  - docs/postgres-operations.md
documentation:
  - BACKLOG_WORKFLOW.md
  - docs/environments.md
priority: high
type: feature
ordinal: 220000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make PostgreSQL authoritative for the data TASK-0190 deliberately retained in MongoDB, then remove MongoDB entirely.

Move calendar connections, provider tokens, calendar preferences, and OAuth origin off the retained MongoDB `users` document and into PostgreSQL, encrypted at rest.
Move OTP challenges off `otpCodes`, and historical daily user logs off `dailyuserlogs`.
Retire the dormant `friendrequests` collection instead of migrating it.
Move event-creator analytics and active-user/num-users bot reporting onto PostgreSQL.
After the core-record cutover (TASK-0190.08) and the retained-data cutover both pass, delete the retained MongoDB account document, the MongoDB runtime and driver, Compose and environment configuration, and the collections.

Use one authoritative store per record and no permanent dual writes.
Preserve account identity and relationships through platform_identities.external_user_id while legacy source documents remain intact for rollback.
Implementation and isolated rehearsal are in scope; live deployments and production data migration are separately scheduled operational actions.
Preserve historical handoffs and completed task records.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Linked subtasks deliver independently verified calendar, OTP, daily-log, analytics, friend-request-retirement, retained-document-retirement, and MongoDB-removal stages.
- [x] #2 Calendar connections, encrypted credentials, sub-calendars, and preferences are served authoritatively from PostgreSQL with identity preserved through platform_identities.external_user_id.
- [x] #3 OTP challenges and historical daily user logs are served authoritatively from PostgreSQL with their existing expiry, lockout, and timezone-bucketing semantics preserved.
- [x] #4 Event-creator analytics and active-user/num-users reporting read PostgreSQL without double counting across stores.
- [x] #5 MongoDB runtime, driver, configuration, and collections are removed only after the core-record cutover and retained-data cutover both pass, with rollback boundaries recorded.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation Plan

All ten delivery subtasks (TASK-0199.01 through TASK-0199.10, including the .07 and .09 grandchildren) are Done. This session is the parent umbrella finalization: verify each acceptance criterion against the current repository state and recorded subtask evidence, refresh stale metadata, then close TASK-0199.

### Steps
1. Confirm every linked subtask is Done and collect its final-summary evidence for AC1.
2. Re-verify the current tree, independent of subtask summaries: no MongoDB runtime, driver, or `MONGODB_*` configuration remains; calendar, OTP, daily-log, analytics, and friend-request data paths are PostgreSQL-authoritative.
3. Map each parent acceptance criterion (#1-#5) to concrete file or recorded evidence, including the rollback boundaries for AC5.
4. Repoint the parent's stale references to the surviving documents (`server/docs/postgres-data-boundaries.md`, `docs/postgres-operations.md`) since the migration contracts and runbooks were intentionally removed.
5. Record the evidence in the parent's Implementation Notes, write the final summary, check every acceptance criterion and Definition of Done item, and mark TASK-0199 Done.

### Guardrails
- Record-only finalization: no runtime code, tests, configuration, or generated artifacts change.
- Use Backlog MCP for the task record; do not hand-edit generated Markdown.
- Cite the subtask evidence for the code-changing checks rather than claiming new test runs; run only cheap state-verification commands.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Umbrella finalization (2026-09-11)

All delivery subtasks are Done: TASK-0199.01 through TASK-0199.10, including .07.01-.03 and .09.01-.07. This session verified the current tree against each parent acceptance criterion; no runtime code, test, or configuration file was changed.

- AC1: every stage is an independently verified subtask record. Calendar TASK-0199.01/.02/.03/.07, OTP TASK-0199.04, daily logs TASK-0199.05/.10, analytics TASK-0199.06, friend-request retirement and retained-document removal TASK-0199.08, and MongoDB removal TASK-0199.09 with its seven subtasks. Each completed record carries its own test and verification evidence.
- AC2: `server/migrations/20260910180000_calendar_integrations.sql` (calendar_accounts, calendar_sub_calendars, calendar_account_credentials, calendar_preferences); `server/postgres/calendar.go:117` resolves owners only through `platform_identities.external_user_id`; `server/postgres/credentials.go` implements the AES-256-GCM credential codec; `server/accounts/calendar.go` is the runtime boundary; `e2e/specs/timed-event-calendar-integration-firefox.spec.ts` exercises connect, OAuth refresh, sub-calendar toggle, and remove.
- AC3: `server/postgres/otp.go` keeps ten-minute expiry, `otpMaxAttempts = 5` with the sixth attempt rejected, delete-on-success, replace/reset-before-insert, and an explicit expired-challenge sweep; `server/postgres/dailylogs.go` keeps timezone-adjusted account-local date bucketing, same-day idempotency, and first-seen order; TASK-0199.10 backfilled historical logs with ledger-based reconciliation and `missing-owner-account` quarantine.
- AC4: `server/postgres/analytics.go` reads only `postgres_events.creator_posthog_id` with the legacy half-open window and threshold semantics; the Slack/Discord active-user and num-users paths read PostgreSQL daily-log memberships. No reporting query joins stores, so records cannot be double counted.
- AC5: TASK-0190.08 (core cutover), TASK-0199.07/.08/.10 (retained cutovers), and TASK-0199.09.06 (PostgreSQL-only verification) all passed before removal. ADR-017 records the per-kind write freeze as the point of no return and the final-removal preconditions; `docs/postgres-operations.md` keeps the deployed-Mongo decommission procedure.
- Current-state checks: `server/db` is absent; `go.mongodb.org/mongo-driver` is absent from `server/go.mod` and `server/go.sum`; `MONGODB_*` appears in no env file, and rendered Compose has no mongo service. Remaining "Mongo" strings are historical comments or migration annotations only.
- Verification inherited from the last code-changing task (TASK-0198, HEAD f32f32c0): isolated backend suite every package ok; frontend lint/fmt:check/typecheck/build/unit (1,088 tests); Firefox desktop E2E 56 passed / 1 skipped / 0 failed. TASK-0199.09.06 recorded the same runs against the PostgreSQL-only stack. `npm run format:markdown:check` exits 0.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Retired MongoDB after moving every retained record kind to PostgreSQL. Delivered by linked subtasks TASK-0199.01 through TASK-0199.10 (including the .07 and .09 grandchildren), with TASK-0190.08 as the core-record cutover predecessor.

## Acceptance criteria evidence

- #1 Linked subtasks deliver every stage: calendar (TASK-0199.01/.02/.03/.07), OTP (TASK-0199.04), daily logs (TASK-0199.05/.10), analytics (TASK-0199.06), friend-request retirement and retained-document removal (TASK-0199.08), and MongoDB removal (TASK-0199.09 with .01-.07). Every subtask is Done with its own recorded verification and test evidence.
- #2 Calendar authority: `server/migrations/20260910180000_calendar_integrations.sql` defines `calendar_accounts`, `calendar_sub_calendars`, `calendar_account_credentials`, and `calendar_preferences`; `server/postgres/calendar.go:117` resolves owners only through `platform_identities.external_user_id`; `server/postgres/credentials.go` stores AES-256-GCM envelopes with fail-closed decryption; `server/accounts/calendar.go` is the runtime boundary, and `e2e/specs/timed-event-calendar-integration-firefox.spec.ts` covers connect, refresh, sub-calendar toggle, and remove.
- #3 OTP and daily logs: `server/postgres/otp.go` preserves ten-minute expiry, the five-attempt lockout (`otpMaxAttempts = 5`, sixth attempt rejected), delete-on-success, and replace/reset-before-insert with an explicit sweep replacing the TTL index; `server/postgres/dailylogs.go` preserves timezone-adjusted account-local date bucketing, same-day idempotency, and first-seen order. TASK-0199.10 backfilled historical `dailyuserlogs` with ledger reconciliation and `missing-owner-account` quarantine.
- #4 Reporting: `server/postgres/analytics.go` reads only `postgres_events.creator_posthog_id` with the legacy half-open window, and the Slack/Discord active-user and num-users commands read PostgreSQL daily-log memberships, so no query joins stores and double counting is impossible.
- #5 Gated removal with recorded rollback boundaries: TASK-0190.08 (core cutover), TASK-0199.07/.08/.10 (retained cutovers), and TASK-0199.09.06 (PostgreSQL-only verification) passed before removal. ADR-017 records that the per-kind write freeze is the point of no return and that final removal requires validated cutovers, an unexpired retention window, and verified backups; `docs/postgres-operations.md` keeps the deployed-Mongo decommission procedure.

## Verification

- Current tree: `server/db` is absent; `go.mongodb.org/mongo-driver` is absent from `server/go.mod`/`server/go.sum`; `MONGODB_*` appears in no env or Compose file; rendered Compose has no mongo service; remaining "Mongo" strings are historical comments and migration annotations.
- Last code-changing work (TASK-0198, HEAD commit f32f32c0) recorded the isolated backend suite with every package ok, frontend lint/fmt:check/typecheck/build/unit (1,088 tests), Firefox desktop E2E 56 passed/1 skipped/0 failed, and clean Markdown checks; TASK-0199.09.06 recorded the same runs against the PostgreSQL-only stack. No runtime code changed in this finalization.
- `npm run format:markdown:check` exits 0; Backlog-generated Markdown is managed through Backlog MCP.

## Notes

- Record-only finalization: no runtime code, tests, configuration, or generated artifacts were modified.
- Live deployments and production data migration remain separately scheduled operational actions, as scoped in the task description.
<!-- SECTION:FINAL_SUMMARY:END -->
