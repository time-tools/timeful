---
id: TASK-0225
title: Add schema-integrity regression coverage for the PostgreSQL baseline
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-13 19:35'
updated_date: '2026-09-14 18:03'
labels:
  - postgres
  - tests
dependencies: []
references:
  - server/postgres/migration_helpers_test.go
  - server/postgres/baseline_schema_test.go
  - server/migrations/20260912000000_baseline_schema.sql
modified_files:
  - server/postgres/
priority: medium
type: task
ordinal: 227000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The baseline schema is the single authority for data integrity, but several of its constraints have no regression test, event hard-delete cascade behavior is untested, and the server/postgres test harness replays migration files into TEMP tables with a literal \"CREATE TABLE \" string replacement that can silently exercise the wrong schema. Outcome: tests that fail when a baseline invariant or the harness's schema fidelity regresses.

Exclusions: the FR-119 event-name constraint lifecycle and its convalidated assertion are owned by TASK-0224.02; coordinate rather than duplicating that coverage.

Useful references: server/migrations/20260912000000_baseline_schema.sql, server/postgres/baseline_schema_test.go, server/postgres/migration_helpers_test.go, server/postgres/supporting_indexes_test.go, compose.test.yaml.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Tests assert the currently untested baseline CHECK constraints, at least: postgres_events payload object type and short-id format, access_transfers source/platform-identity XOR, otp_challenges empty-value and attempts checks, postgres_event_responses payload object type, and calendar_preferences constraint behavior.
- [x] #2 A test deletes a postgres_events row and asserts every dependent row cascades: visitor identities, visitor credentials, responses, signup blocks, signup responses, signup response memberships, attendees, folder_events, and access transfers.
- [x] #3 A test asserts the application role can perform DML but not DDL, or the equivalent privilege separation is verified against the isolated stack.
- [x] #4 The temp-table migration replay used by server/postgres tests is checked against the goose-applied schema, or the harness is changed so any divergence fails a test.
- [x] #5 All new tests run in the isolated Compose overlay and pass.
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
1. Add server/postgres/baseline_constraints_test.go covering the untested baseline CHECK constraints: postgres_events short-id format and payload object, postgres_event_responses payload object, access_transfers source/platform-identity XOR, otp_challenges empty-value and attempts checks, and calendar_preferences checks.
2. Add a hard-delete cascade test seeding every dependent row class (visitor identities, credentials, responses, signup blocks, signup responses, signup response memberships, attendees, folder_events, access transfers) and asserting each is gone after DELETE FROM postgres_events.
3. Add an application-role privilege test: runtime DML succeeds and CREATE TABLE/ALTER TABLE are denied with SQLSTATE 42501 against the isolated stack.
4. Add a replay-fidelity test comparing the temp-schema migration replay against the goose-applied public schema (tables, columns, constraints, indexes) so any divergence fails.
5. Run the backend suite through the isolated Compose overlay, gofmt, and finalize the task.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-14 18:03
---
Implementation and verification (2026-09-14).

New tests under server/postgres/ (no production code changed):
- baseline_constraints_test.go: postgres_events short_id format and payload object, postgres_event_responses payload object, access_transfers source/platform-identity XOR plus source_hash and state, otp_challenges empty email/code_hash and negative attempts, calendar_preferences primary_account_key and token_origin, and the remaining untested baseline CHECKs (num_responses, owner token hash length, credential hash length/kind, transfer request target hash length, accounts counter, daily log position, sub-calendar id).
- baseline_cascade_test.go: deletes a postgres_events row and asserts visitor identities, visitor credentials, responses, signup blocks, signup responses, signup response block memberships, attendees, folder_events, access transfers, and access transfer requests are all gone while the account-owned folder survives.
- application_role_test.go: asserts the isolated stack's application role is not a superuser, has no CREATE on public, can INSERT/UPDATE/DELETE, and gets SQLSTATE 42501 for CREATE TABLE and ALTER TABLE.
- migration_replay_fidelity_test.go: reads the replay session's temp schema and a separate session's goose-applied public schema and compares every table's columns, constraints, and indexes; index definitions are normalized only for the schema qualifier pg_get_indexdef always emits.

Evidence:
- Focused new-test run: /tmp/opencode/task-0225-focused.log (all baseline, cascade, application-role tests pass).
- Fidelity sensitivity probe: temporarily added a column to the baseline migration Up section; TestMigrationReplayMatchesAppliedSchema failed naming `postgres_events columns: temp replay has harness_fidelity_probe|text|null|, goose-applied does not` (/tmp/opencode/task-0225-fidelity-probe.log); the migration file was restored and `git diff server/migrations/` is empty.
- Full backend suite in the isolated Compose overlay: `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` -> all packages ok, including timeful/server/postgres (7.435s) and timeful/server/routes (3.559s); log at /tmp/opencode/task-0225-backend-tests.log.
- Focused Firefox e2e `specs/timed-event-create-firefox.spec.ts` -> 6 passed (1.8m); log at /tmp/opencode/task-0225-e2e-create-firefox.log.
- `gofmt -l server/postgres/` reports nothing; `npm run format:markdown:check` passes.

TASK-0224.02's FR-119 event-name constraint lifecycle is untouched by this coverage.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added schema-integrity regression coverage for the PostgreSQL baseline in four new test files under server/postgres, no production code changed.

- baseline_constraints_test.go asserts the listed untested baseline CHECK invariants plus the remaining untested ones: postgres_events short-id format and payload object, postgres_event_responses payload object, access_transfers source/platform-identity XOR (and source_hash/state), otp_challenges empty email/code_hash and negative attempts (with zero default), calendar_preferences primary_account_key/token_origin, and the num_responses, owner token hash length, credential hash length/kind, transfer request target hash length, accounts counter, daily log position, and sub-calendar id checks. The FR-119 event-name constraint remains owned by TASK-0224.02.
- baseline_cascade_test.go deletes an event and asserts visitor identities, credentials, responses, signup blocks, signup responses, signup response block memberships, attendees, folder_events, access transfers, and access transfer requests all cascade away, while the account-owned folder survives.
- application_role_test.go proves the isolated stack separates DML from DDL: the application role is non-superuser, has no CREATE on public, can INSERT/UPDATE/DELETE, and receives SQLSTATE 42501 for CREATE TABLE and ALTER TABLE.
- migration_replay_fidelity_test.go reads the temp replay schema and the goose-applied public schema from separate sessions and fails on any table, column, constraint, or index difference. A temporary probe column in the baseline migration made it fail with the exact divergent column, proving sensitivity.

Verification: focused tests pass; full backend suite passes via the isolated Compose overlay (timeful/server/postgres 7.435s, timeful/server/routes 3.559s, all packages ok); focused Firefox e2e specs/timed-event-create-firefox.spec.ts passes 6/6; gofmt clean; npm run format:markdown:check passes.
<!-- SECTION:FINAL_SUMMARY:END -->
