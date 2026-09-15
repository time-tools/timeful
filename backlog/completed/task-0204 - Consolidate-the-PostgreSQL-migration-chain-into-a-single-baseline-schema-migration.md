---
id: TASK-0204
title: >-
  Consolidate the PostgreSQL migration chain into a single baseline schema
  migration
status: Done
assignee:
  - OpenCode
created_date: '2026-09-11 20:25'
updated_date: '2026-09-11 21:24'
labels: []
dependencies: []
references:
  - >-
    backlog/backlog.md open item: remove server/migrations from mongo,
    consolidate into a single init script for postgres
  - server/migrations/20260910170000_migration_ledger.sql
  - server/migrations/20260911120000_drop_folder_legacy_event_id.sql
documentation:
  - docs/postgres-operations.md
  - docs/environments.md
  - server/docs/postgres-data-boundaries.md
modified_files:
  - server/migrations
  - server/postgres
  - docs/postgres-operations.md
  - docs/environments.md
  - compose.yaml
  - compose.test.yaml
  - server/Dockerfile
priority: medium
type: chore
ordinal: 243000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`server/migrations/` has accumulated 15 goose migrations (from anonymous-event compatibility through the folder legacy-column cleanup) that encode the MongoDB-to-PostgreSQL transition, Mongo backfill tooling, and intermediate schema revisions. MongoDB is fully retired, there is no Mongo data to migrate, and existing local, staging, and production PostgreSQL databases may be recreated, so no in-place upgrade path is required.

Replace the chain with one authoritative goose baseline migration that creates the complete schema the current server requires, so a fresh database initializes in one step and future schema changes continue as new incremental goose migrations on top.

The resulting schema must match the schema the current chain produces, except that the retired Mongo migration tooling tables (`migration_ledger`, `migration_quarantine`) are no longer created. Goose stays the migration tool, and the migrator image, Compose migration service, CI migration caches, and isolated test and E2E stacks must keep working. Historical ADRs and applied-migration records are not rewritten.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 server/migrations/ contains a single baseline goose migration that creates the complete current schema, and no other schema migration files remain.
- [x] #2 Applying the baseline to an empty PostgreSQL 18 database succeeds and produces the same schema the current migration chain produces, with the only intended difference being that migration_ledger and migration_quarantine are not created (verified by a schema comparison, not by inspection alone).
- [x] #3 Runtime compatibility columns and behavior still used by the current server are preserved in the baseline.
- [x] #4 PostgreSQL-backed test harnesses initialize their schema from the baseline instead of individual migration filenames, and the isolated backend test suite passes.
- [x] #5 A fresh development/test Compose stack reaches a healthy server using the baseline migration, and the Firefox browser E2E suite passes.
- [x] #6 Documentation that describes the migration history, the forward-only compatibility statement, and the migration_ledger/migration_quarantine retention cleanup reflects the baseline model.
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
1. Create `server/migrations/20260912000000_baseline_schema.sql` as the single goose baseline that creates the final schema produced by the existing 15-migration chain, excluding `migration_ledger` and `migration_quarantine`; preserve column order, constraint names, and index definitions exactly.
2. Verify with a two-database schema comparison in a disposable PostgreSQL 18 container: apply the concatenated Up sections of the old chain to one database and the baseline to another, then diff `pg_dump --schema-only` output; iterate until the only difference is the two retired tooling tables.
3. Refactor the `server/postgres` test harnesses (`accounts_test.go`, `calendar_test.go`, `signup_test.go`, `otp_test.go`) to initialize temp-table schemas by reading and applying the `server/migrations/` directory in lexical order (currently only the baseline) instead of listing individual filenames; convert `identity_migration_test.go` from backfill-chain coverage to baseline constraint coverage.
4. Delete the 15 superseded migration files so only the baseline remains.
5. Update `docs/postgres-operations.md` (retention cleanup) and `docs/environments.md` (forward-only statement) for the baseline model; leave historical ADRs and applied-migration history untouched.
6. Run required checks: `go build ./...`, `go vet ./...`, isolated `server-route-test` suite, fresh Compose stack healthy, Firefox E2E suite, and root markdown formatting.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

`server/migrations/` now contains the single goose baseline `20260912000000_baseline_schema.sql` that creates the complete current schema in one step. The 15-migration MongoDB transition chain is deleted, and the retired `migration_ledger`/`migration_quarantine` tables are no longer created. Future schema changes continue as incremental goose migrations on top of the baseline.

### Changes

- `server/migrations/20260912000000_baseline_schema.sql` — final-state schema for all 23 tables with the chain's exact column order, constraint names, and indexes; Down refuses reversal with a guard (same pattern as the retired visitor-identity migration).
- `server/postgres/migration_helpers_test.go` — new shared harness that applies every `server/migrations/*.sql` Up section in version order into temp tables, so the four harnesses no longer list filenames and future migrations stay covered automatically.
- `server/postgres/{accounts,calendar,signup,otp}_test.go` — delegate to the shared harness; unused imports removed. `attendees_test.go` now reuses it directly.
- `server/postgres/identity_migration_test.go` deleted; `server/postgres/baseline_schema_test.go` replaces it with baseline constraint coverage (composite event/visitor FKs, event-scoped owner relation, base-vs-granted owner constraint, compatibility response columns).
- `docs/environments.md`, `docs/postgres-operations.md`, `server/docs/postgres-anonymous-event-compatibility.md` — baseline model, forward-only statement for incremental migrations, recreation of pre-baseline databases, and removal of stale migration/retention text. No changes were required in `compose.yaml`, `compose.test.yaml`, or `server/Dockerfile`; the migrator image, migration service, and CI cache keys (`server/migrations/**`) keep working unchanged.

### Verification evidence

- AC1: `ls server/migrations/` shows only `20260912000000_baseline_schema.sql`.
- AC2: In a disposable `postgres:18.6-bookworm` container, the concatenated Up sections of all 15 deleted migrations were applied to database `chain`, and the baseline to `baseline`. `pg_dump --schema-only` diff hunks were exactly `5c5`/`1116c1044` (`\restrict`/`\unrestrict` nonces), `339,378d338` (`migration_ledger`, `migration_quarantine` tables), and `678,709d637` (their constraints); no other difference. The rebuilt migrator then applied the baseline with goose to an empty test database: `OK 20260912000000_baseline_schema.sql (77.25ms)` and `successfully migrated database to version: 20260912000000`.
- AC3: Exact schema equality preserves every runtime column; `TestBaselineVisitorIdentityAndOwnerConstraints` additionally exercises `respondent_kind`, `account_user_id`, `guest_edit_token`, the owner relation, and the granted-credential constraint.
- AC4: Isolated `server-route-test` suite passed all packages, including `timeful/server/postgres` and `timeful/server/routes`, against `timeful-test-postgres-baseline`.
- AC5: The E2E Playwright stack built and reached healthy on a fresh database, then `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop` passed 56 tests with 1 pre-existing skip and 0 failures (6.3m).
- AC6: Docs updated as above; historical ADRs and applied-migration records were left untouched.

### Checks run

- `cd server && go build ./... && go vet ./...` (pre-existing gofmt findings in `routes/auth_otp_test.go` and `utils/response_utils.go` are untouched by this change)
- Isolated backend suite via `server-route-test` (all packages ok)
- `E2E_FRONTEND=bundled npm run test:e2e -- --project=firefox-desktop` (56 passed, 1 skipped)
- `npm run format:markdown` and `npm run format:markdown:check`
- `graphify update .` (generated output is gitignored)
<!-- SECTION:FINAL_SUMMARY:END -->
