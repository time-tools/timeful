---
id: TASK-0215
title: >-
  Rewrite the PostgreSQL event API and data-boundaries contracts for the
  PostgreSQL-only system
status: Done
assignee:
  - opencode
created_date: '2026-09-12 22:08'
updated_date: '2026-09-12 22:14'
labels:
  - docs
  - postgres
dependencies: []
references:
  - server/docs/postgres-anonymous-event-compatibility.md
  - server/docs/postgres-data-boundaries.md
  - server/models/event.go
  - server/migrations/20260912000000_baseline_schema.sql
  - docs/postgres-operations.md
priority: medium
type: docs
ordinal: 220000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Both contract documents under server/docs predate the completed MongoDB retirement and still describe migration-era state.

`server/docs/postgres-anonymous-event-compatibility.md` is the API-behavior contract for PostgreSQL event storage, but its anonymous/compatibility framing, its claim that migrated external account IDs are plain strings without a cross-database foreign key, its legacy schedule/shared-field wording, its guest-edit-token sentence, its Delivered section headings, and its policy-transition sentence all describe things that no longer exist.

`server/docs/postgres-data-boundaries.md` is the durable backend contract for store ownership, account identity, calendar identity and keys, OTP handling, credential encryption, and reporting reads, but it still speaks of retained record kinds, fresh identities, preserved behavior, migrated events, and a retired friend-request row and baseline-recreate history that other documents own.

Rewrite both documents as current-state contracts:
- Rename the event contract to `server/docs/postgres-event-api-contract.md` with H1 `PostgreSQL Event API Contract`.
- Update the inbound links in the inactive ADR-011 and ADR-014.
- Remove every MongoDB, legacy, migration, and compatibility reference that does not describe current behavior.
- Keep only verified current behavior, canonical glossary terminology, one sentence per physical line, and current cross-document links.
- Run Markdown lint and format checks and refresh the graphify graph.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `server/docs/postgres-event-api-contract.md` replaces `postgres-anonymous-event-compatibility.md` and describes the current PostgreSQL event API contract without anonymous/compatibility framing
- [x] #2 Neither document mentions MongoDB, a second store, retained record kinds, migrated external identifiers, guest edit tokens, legacy schedule columns, or the retired friend-request storage
- [x] #3 `server/docs/postgres-data-boundaries.md` describes only current store ownership, account identity, calendar identity and key rules, OTP rules, credential encryption, and reporting reads
- [x] #4 Baseline-recreate and pre-cutover tombstone history appears only in `docs/postgres-operations.md`, not in the boundaries contract
- [x] #5 Inbound links in ADR-011 and ADR-014 resolve to the renamed contract, and the renamed contract links to the boundaries contract
- [x] #6 `npm run lint:markdown` and `npm run format:markdown:check` pass from the repository root
- [x] #7 The rewritten documents follow `docs/AGENTS.md` sentence-per-line and canonical terminology rules
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
1. Rename `server/docs/postgres-anonymous-event-compatibility.md` to `server/docs/postgres-event-api-contract.md` with git mv.\n2. Rewrite the event API contract as current-state sections: scope, identifier exposure, stored relations, JSONB payloads, request semantics, authority credentials, response mutation, event owner authority, access transfers, transactions.\n3. Rewrite `server/docs/postgres-data-boundaries.md` in place with current store ownership, account identity, calendar identities and keys, OTP, daily logs, encryption, and reporting sections.\n4. Update the renamed-contract links in ADR-011 and ADR-014 and the boundaries contract cross-link.\n5. Verify with git grep, `npm run lint:markdown`, `npm run format:markdown:check`, and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Renamed `server/docs/postgres-anonymous-event-compatibility.md` to `server/docs/postgres-event-api-contract.md` and rewrote both PostgreSQL contracts as current-state documents.

The event API contract now covers scope, identifier exposure, stored relations, JSONB payloads, request semantics, Event Visitor Identity credentials, response mutation, Event Owner authority, Access Transfers, and transactions without anonymous/compatibility framing, MongoDB or migration references, cross-database foreign keys, legacy schedule wording, or guest-edit-token statements. It names the persisted days-only schedule columns directly and links the boundaries contract.

The boundaries contract now states PostgreSQL as the single store with no second read path, drops the Friend-request row and all retained/migrated/fresh-identity wording, states account identity from `platform_identities.id`, describes calendar key resolution positively, keeps the OTP and AES-256-GCM rules, and leaves baseline-recreate history to `docs/postgres-operations.md`. The two inactive ADR links in ADR-011 and ADR-014 point at the renamed contract, with ADR-014's claim aligned to the current authority model.

Verification: `git grep` for the old filename and migration wording returns nothing outside backlog history and the graph artifact; `npm run lint:markdown` and `npm run format:markdown:check` pass; `graphify update .` ran. The AST-only update does not re-extract Markdown, so a `/graphify --update` semantic refresh will repoint the local ignored graph artifact from the old filename. A follow-up task, TASK-0216, tracks the stale anonymous-event-compatibility route test naming and comments.
<!-- SECTION:FINAL_SUMMARY:END -->
