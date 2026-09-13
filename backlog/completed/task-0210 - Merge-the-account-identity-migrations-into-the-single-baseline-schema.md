---
id: TASK-0210
title: Merge the account identity migrations into the single baseline schema
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-12 20:51'
updated_date: '2026-09-12 21:06'
labels: []
dependencies: []
modified_files:
  - server/migrations/20260912000000_baseline_schema.sql
  - server/postgres/baseline_schema_test.go
  - server/docs/postgres-data-boundaries.md
  - server/scripts/20260912_export_legacy_deletion_tombstones/export.sh
priority: medium
type: chore
ordinal: 211000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
server/migrations/ now holds the 20260912000000 baseline plus the two account-identity cutover migrations (20260912120000_account_identity_platform_uuid.sql and 20260912130000_rewrite_legacy_account_payloads.sql). The cutover chain and its data-only payload rewrite exist only to upgrade databases that predate ADR-021; a fresh database can be created directly in the consolidated final shape, so the baseline should absorb the consolidation and the other migration files should be deleted.

The repository policy is that a database created before the baseline is recreated from it instead of upgraded in place, so no in-place upgrade path for the three-migration chain needs to be preserved. Goose stays the migration tool, and future schema changes remain incremental migrations on top of the single baseline.

Merging the payload-rewrite migration is safe because a fresh database holds no stored legacy payloads; the runtime re-projects account identity from the consolidated uuid columns.

Acceptance criteria:
1. server/migrations/ contains only 20260912000000_baseline_schema.sql, and that baseline creates the consolidated final schema directly: platform_identity_id references on postgres_events, postgres_event_responses, event_signup_responses, event_attendees, folders, folder_events, access_transfers, and daily_user_log_members; account_deletion_tombstones keyed by platform_identity_id without a foreign key; and no external_user_id, account_user_id, or owner_external_id legacy columns.
2. Applying the baseline to an empty PostgreSQL 18 database produces the same final schema as the retired three-migration chain, verified with a pg_dump --schema-only comparison (not by inspection alone); any intentional difference is documented.
3. The obsolete consolidation rehearsal tests are removed, the migrations-directory harness keeps working, and the isolated backend test suite passes.
4. Documentation and operator guidance no longer reference the retired migration versions, and the legacy deletion tombstone export path remains correctly described for a pre-baseline database.
5. go build ./... and go vet ./... pass for the server.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 server/migrations/ contains only the baseline, and it creates the consolidated final schema directly with no legacy identity columns
- [x] #2 A pg_dump --schema-only comparison shows the baseline matches the retired three-migration chain's final schema, with any intentional difference documented
- [x] #3 The consolidation rehearsal test is removed, the migrations-directory harness still works, and the isolated backend suite passes
- [x] #4 Docs and operator scripts no longer reference the retired migration versions, and the tombstone export path stays correctly described
- [x] #5 go build ./... and go vet ./... pass for the server
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
1. Capture the retired chain's final schema in a disposable postgres:18.6 container: apply 20260912000000 then 20260912120000 then 20260912130000, and pg_dump --schema-only.
2. Rewrite 20260912000000_baseline_schema.sql to create the consolidated final schema directly, preserving final constraint and index names (access_transfers_source_xor_platform_identity, daily_user_log_members_log_platform_identity_key, event_signup_responses_identity, postgres_response_platform_identity_idx, event_attendees_platform_identity_id_idx, folders_platform_identity_id_idx, daily_user_log_members_platform_identity_id_idx, folder_events_event_unique_idx) and keeping the baseline comment style.
3. Delete 20260912120000_account_identity_platform_uuid.sql and 20260912130000_rewrite_legacy_account_payloads.sql.
4. Delete server/postgres/account_identity_migration_test.go and preserve any still-relevant schema-shape coverage in baseline_schema_test.go.
5. Update server/docs/postgres-data-boundaries.md and server/scripts/20260912_export_legacy_deletion_tombstones/export.sh for the single-baseline model.
6. Verify: pg_dump diff between chain and new baseline, go build/vet, isolated server-route-test suite, graphify update.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: rewrote the baseline to the cutover chain's exact final schema (verified with two pg_dump comparisons: direct psql application and goose application), removed the two cutover migration files and the obsolete rehearsal test, and updated the baseline shape test, data-boundaries doc, and tombstone export script comments.

Retired migration files and the deleted test were moved aside to /tmp/opencode/retired-migrations/ before removal from the worktree; the temporary verification container and test-stack volume were cleaned up afterward.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

`server/migrations/` now contains only `20260912000000_baseline_schema.sql`, which creates the consolidated account identity schema directly. The two cutover migrations (`20260912120000_account_identity_platform_uuid.sql`, `20260912130000_rewrite_legacy_account_payloads.sql`) are deleted; there is no fresh-database work left for them because a fresh database holds no legacy rows and the baseline already carries the platform identity uuid columns.

### Changes

- `server/migrations/20260912000000_baseline_schema.sql` — absorbs the cutover: `platform_identities.external_user_id`, `postgres_events.owner_external_id`, and every `account_user_id`/`external_user_id` reference are gone; `platform_identity_id` is the account reference on `postgres_event_responses`, `event_signup_responses`, `event_attendees`, `folders`, `folder_events`, `access_transfers`, and `daily_user_log_members`; `account_deletion_tombstones` is keyed by `platform_identity_id` with no foreign key. Final object names are preserved (`access_transfers_source_xor_platform_identity`, `daily_user_log_members_log_platform_identity_key`, `event_signup_responses_identity`, `postgres_response_platform_identity_idx`, `event_attendees_platform_identity_id_idx`, `folders_platform_identity_id_idx`, `daily_user_log_members_platform_identity_id_idx`, `folder_events_event_unique_idx`), and physical column order matches the retired chain's final schema.
- `server/migrations/20260912120000_account_identity_platform_uuid.sql`, `server/migrations/20260912130000_rewrite_legacy_account_payloads.sql` — deleted. The payload rewrite is data-only and unnecessary on a fresh baseline.
- `server/postgres/account_identity_migration_test.go` — deleted; its rehearsal of a migration that no longer exists is obsolete. `server/postgres/baseline_schema_test.go` now asserts the baseline shape directly: the ten legacy columns must be absent and the nine platform identity columns must exist.
- `server/docs/postgres-data-boundaries.md` — the tombstone sentence now describes the single-baseline recreate policy instead of naming a deleted migration version.
- `server/scripts/20260912_export_legacy_deletion_tombstones/export.sh` — comments now describe exporting from a pre-baseline database before recreating it from the baseline.

### Verification evidence

- AC1: `ls -1 server/migrations/` returns only `20260912000000_baseline_schema.sql`; `TestBaselineConsolidatesAccountIdentity` passed in the isolated suite.
- AC2: In a disposable `postgres:18.6-bookworm` container the retired chain (baseline + 20260912120000 + 20260912130000) was applied to database `chain`, and the merged baseline to database `baseline`; `pg_dump --schema-only --no-owner --no-comments` output was identical after filtering only the `\restrict`/`\unrestrict` nonces. The merged baseline was then applied with the rebuilt goose migrator to a fresh `timeful-test-baseline-merge` database: `OK 20260912000000_baseline_schema.sql (73.72ms)`, only version `20260912000000` recorded, and its dump (excluding the `goose_db_version` bookkeeping table) was identical to the chain dump as well.
- AC3: The obsolete test file is gone; `go test ./...` in the isolated `server-route-test` stack passed all packages (`ok timeful/server/postgres`, `ok timeful/server/routes`, and the rest).
- AC4: `grep` across `server/`, `docs/`, `frontend/`, and `e2e/` finds no live reference to `20260912120000`/`20260912130000`; only historical ADRs and completed Backlog records name them. The export script and data-boundaries doc describe the pre-baseline export path.
- AC5: `go build ./...` and `go vet ./...` both pass.
- E2E: `--project=firefox-desktop` at the default dev-server configuration passed 53 tests with 1 skip; the three heavy access-transfer approval journeys (`Source approves the exact target code for guest/owner/signed-in access`) hit their 30-second budget (`Test timeout of 30000ms exceeded`), which matches the documented two-worker dev-server timing behavior; the user accepted these timeouts. The schema is byte-identical to the chain, so the failures are not migration-related.

### Checks run

- `cd server && go build ./... && go vet ./...`
- `docker compose --env-file .env.test -f compose.yaml -f compose.test.yaml run --rm server-route-test` (all packages ok)
- Fresh goose migration via the rebuilt migrator image on `timeful-test-baseline-merge`
- `npm run format:markdown:check` (clean)
- `graphify update .` (graph regenerated)
<!-- SECTION:FINAL_SUMMARY:END -->
