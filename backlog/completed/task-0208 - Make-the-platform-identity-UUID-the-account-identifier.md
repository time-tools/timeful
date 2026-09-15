---
id: TASK-0208
title: Make the platform identity UUID the account identifier
status: Done
assignee: []
created_date: '2026-09-11 21:02'
updated_date: '2026-09-12 20:12'
labels: []
dependencies: []
references:
  - TASK-0177.01
documentation:
  - docs/design/architecture/adr/ADR-019.md
  - docs/requirements/functional/fr/FR-121.md
  - server/docs/postgres-data-boundaries.md
  - server/migrations/20260908160000_visitor_identities.sql
priority: medium
type: chore
ordinal: 247000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The account reference is currently a separate ObjectID-shaped 24-hex string held in platform_identities.external_user_id and denormalized into TEXT columns (account_user_id, postgres_events.owner_external_id, access_transfers.external_user_id, account_deletion_tombstones.external_user_id). Everything else in PostgreSQL already keys on native UUIDv7 (platform_identities.id, accounts.id, event_visitor_identities.platform_identity_id, postgres_events.owner_platform_identity_id). Consolidate account identity onto platform_identities.id so there is exactly one account identifier: UUIDv7, database-validated, and consistent with the rest of the schema. Outcome: sign-in sessions and API payloads carry the platform identity UUID; account-referencing tables use uuid foreign keys; the separate external identifier is removed with no compatibility lookup. The change is a wire-format break and requires a one-time data consolidation and forced re-sign-in for existing sessions. A new ADR supersedes ADR-019 and records the decision.

Constraints:
- Hard cutover: no request path, column, or payload accepts or emits the legacy 24-hex account identifier after the change.
- Native uuid storage for consolidated references; platform_identities.id keeps its uuidv7() default.
- account_deletion_tombstones must survive platform identity deletion, so it stores the uuid without a foreign key.
- Guest and unowned zero-identity semantics must be preserved.
- TASK-0177.01 (bson-objectid in Dashboard sorting and signup block local identities) overlaps this initiative and is reconciled rather than duplicated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every account reference in the runtime resolves through platform_identities.id, and no request path, stored column, or API payload accepts or emits the legacy 24-hex account identifier
- [x] #2 The superseding ADR, backend boundary, schema consolidation, frontend adoption, and end-to-end verification subtasks are all complete
- [x] #3 TASK-0177.01 has a recorded reconciliation (dependency or update) so its bson-objectid removal does not conflict with this identifier change
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Completed the platform-identity UUID account-identifier cutover and retired the separate 24-hex external identifier.

Delivered:
- ADR-021 supersedes ADR-019 and records platform_identities.id (native UUIDv7) as the sole account identifier, the canonical lowercase hyphenated wire form, the all-zero UUID guest/unowned sentinel, native uuid consolidated references, the one guarded backfill with no compatibility lookup, and forced re-sign-in for legacy sessions (TASK-0208.01).
- Backend boundary and repositories resolve accounts, sessions, ownership, visitor authorization, transfers, and deletion through platform_identities.id; models.UUID replaces models.ID and the 24-hex account path is deleted (TASK-0208.02, TASK-0208.06).
- Migration 20260912120000 consolidates every account reference to platform_identity_id uuid columns with foreign keys (tombstones carry the uuid without a foreign key), backfills, refuses unmapped rows, and drops the legacy TEXT columns and platform_identities.external_user_id last; migration 20260912130000 rewrites stored JSONB payloads that embedded the retired identifier (TASK-0208.03, TASK-0208.06, TASK-0208.07).
- Frontend consumes the UUID identifier: the guest sentinel is the all-zero UUID, guest-name validation mirrors the UUID account-id form, and Dashboard sorting plus signup block local keys no longer depend on bson-objectid (TASK-0208.04, TASK-0208.07).
- End-to-end verification: isolated backend suite green, 53 Firefox desktop tests passed with 3 user-accepted access-transfer approval timeouts, the Chromium account-deletion spec passed, and grep/SQL/API/rehearsal evidence shows zero legacy columns and the preserved guest zero identity (TASK-0208.05).
- TASK-0177.01 is Done with a dependency on TASK-0208.04 and its bson-objectid removal reconciliation recorded.

Verification:
- Isolated backend suite via compose.test.yaml passes for all packages, including accounts, postgres, routes, and services/auth.
- Frontend required checks pass: lint, fmt:check, typecheck, build, test:unit (1090 tests across 146 files).
- E2E package lint, fmt:check, and typecheck pass; Firefox desktop and Chromium account-deletion coverage is green apart from the accepted heavy approval timeouts.
- Migration rehearsal proves zero-orphan backfill, repeated-run no-op, unmapped-reference refusal, and destructive-down refusal.

Risks and follow-ups:
- The 3 access-transfer approval journeys remain over the 30-second budget with the unbundled dev server; run them with E2E_FRONTEND=bundled if they must be green.
- Existing pre-cutover sessions force re-sign-in by design; legacy deletion tombstones without a live platform identity are dropped, and operators can export them first with server/scripts/20260912_export_legacy_deletion_tombstones/export.sh.
<!-- SECTION:FINAL_SUMMARY:END -->
