---
id: TASK-0205
title: Reformulate ADRs for the PostgreSQL-only architecture
status: Done
assignee: []
created_date: '2026-09-11 20:37'
updated_date: '2026-09-11 20:45'
labels: []
dependencies: []
modified_files:
  - docs/design/architecture/adr/ADR-010.md
  - docs/design/architecture/adr/ADR-011.md
  - docs/design/architecture/adr/ADR-012.md
  - docs/design/architecture/adr/ADR-013.md
  - docs/design/architecture/adr/ADR-014.md
  - docs/design/architecture/adr/ADR-015.md
  - docs/design/architecture/adr/ADR-016.md
  - docs/design/architecture/adr/ADR-017.md
  - docs/design/architecture/adr/ADR-018.md
  - docs/design/README.md
  - server/docs/postgres-data-boundaries.md
priority: medium
type: docs
ordinal: 244000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reformulate the architecture decision records so accepted ADRs describe the PostgreSQL-only system without second-store or migration-era wording, consolidate the single-authoritative-store rule in a new ADR, and mark the transition-only records deprecated without deleting their history.

Scope:
- Reformulate ADR-010, ADR-012, ADR-015, and ADR-016 in place to describe current PostgreSQL-only decisions.
- Add ADR-018 "PostgreSQL Is The Single Authoritative Store" (satisfies FR-121), consolidating the surviving single-authority and account-identity rules.
- Mark ADR-013, ADR-014, and ADR-017 deprecated with dated notes; keep ADR-011 superseded by ADR-016.
- Update the ADR index in docs/design/README.md and the durable-decision links in server/docs/postgres-data-boundaries.md.

Constraints:
- Deprecated records keep their historical bodies and titles, so MongoDB wording remains in ADR-011/013/014/017 by decision.
- ADR-008/ADR-009 "legacy" wording refers to pre-existing frontend guest semantics and is out of scope.
- Documentation-only change per BACKLOG_WORKFLOW.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ADR-010, ADR-012, ADR-015, and ADR-016 contain no MongoDB or migration-era legacy wording and describe current PostgreSQL behavior
- [x] #2 ADR-018 exists with valid frontmatter, an index row, and records the single authoritative store rule and account resolution through platform_identities.external_user_id
- [x] #3 ADR-013, ADR-014, and ADR-017 have status deprecated and dated notes; ADR-011 remains superseded by ADR-016 with its surviving principles linked from ADR-016 and ADR-018
- [x] #4 docs/design/README.md rows match the reformulated titles and include ADR-018; server/docs/postgres-data-boundaries.md no longer lists deprecated ADR-017 as a durable decision
- [x] #5 npm run format:markdown, npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass; graphify update . refreshes the graph
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

1. Reformulate ADR-010 (drop the stale second-store sentence and redundant PostgreSQL qualifiers), ADR-012 (account resolution through platform_identities.external_user_id), ADR-015 (authenticated AES-256-GCM credential envelope), and ADR-016 (PostgreSQL data ownership boundaries and friend-request retirement).
2. Add ADR-018 "PostgreSQL Is The Single Authoritative Store" consolidating the single-authority and account-identity rules.
3. Mark ADR-013, ADR-014, and ADR-017 deprecated with dated notes; keep ADR-011 superseded by ADR-016 and link the surviving principles from ADR-016 and ADR-018.
4. Update docs/design/README.md and server/docs/postgres-data-boundaries.md.
5. Verify with no-Mongo greps over the active ADRs and the Markdown checks; run graphify update .
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-11)

Documentation-only change; no runtime code, tests, or configuration touched.

- ADR-010: dropped "This decision does not change MongoDB credentials, persistence, or transfer behavior." and the now-redundant "PostgreSQL events" / "anonymous PostgreSQL source" qualifiers; updated_date 2026-09-11.
- ADR-012: retitled and rewritten as account resolution through `platform_identities.external_user_id` (unique 24-character hexadecimal identifier, session-held, idempotent `FindOrCreatePlatformIdentity`, no permanent identifier table, equal emails not merged); addresses FR-121 instead of deprecated FR-122/QR-012.
- ADR-015: rewrote Context, the AES-CFB option, the decryption-failure sentence, and the two migration-related Bad consequences; the AES-256-GCM envelope decision and QR-013 traceability are unchanged.
- ADR-016: dropped the ADR-011 recital and the retained-data removal framing, retitled without "Retained", restated the single-authority principles and the PostgreSQL data ownership boundary, kept the `postgres-data-boundaries.md` link, and retained its supersession of ADR-011 and QR-013 constraint.
- ADR-018 (new): "PostgreSQL Is The Single Authoritative Store", accepted, satisfies FR-121; records one authority per record kind, no second store or dual write, account resolution through ADR-012, and friend requests having no store.
- ADR-011: second paragraph now links ADR-016 and ADR-018; link-only change, so updated_date was not bumped.
- ADR-013/ADR-014/ADR-017: status deprecated and a dated note added; historical bodies retained by owner decision.
- docs/design/README.md: ADR-012/ADR-016 rows updated and ADR-018 row added; `npm run format:markdown` realigned the table.
- server/docs/postgres-data-boundaries.md: durable-decision link now points at ADR-018 instead of deprecated ADR-017.

## Verification

- AC1: `rg -i 'mongo|legacy' docs/design/architecture/adr/` returns hits only in ADR-011/013/014/017 (deprecated history) and ADR-008/009 (pre-existing frontend guest semantics); the active ADRs 010, 012, 015, 016, and 018 are clean.
- AC2: ADR-018 has MADR frontmatter, an index row, and the single-store and `platform_identities.external_user_id` rules.
- AC3: ADR-013/014/017 declare `status: deprecated` with dated notes; ADR-011 keeps `superseded_by: ADR-016` and links ADR-018.
- AC4: index rows match the H1 titles; `postgres-data-boundaries.md` line 9 lists ADR-015, ADR-016, and ADR-018.
- AC5: `npm run format:markdown` (updated docs/design/README.md), `format:markdown:check`, `lint:markdown`, `test:markdown-format` (30 tests), and `fmt:check` pass; `graphify update .` rebuilt the graph (5501 nodes, 9721 edges, 409 communities).
- Unit and e2e tests are exempt: documentation-only change per BACKLOG_WORKFLOW.md and the task DoD.
- No commit was made. `backlog/backlog.md` was already dirty with the owner's inbox edits before this task.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-11 20:45
---
Approach revised by TASK-0206: ADR-012, ADR-015, and ADR-016 were restored to their original bodies and marked superseded by ADR-019, ADR-020, and ADR-018 instead of being rewritten in place. The rest of this task's delivery stands: the ADR-010 stale-sentence cleanup, the ADR-013/ADR-014/ADR-017 deprecation edits, the new ADR-018, and the index and contract link updates.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Reformulated the ADRs so every accepted record describes the PostgreSQL-only system, consolidated the single-authoritative-store rule in ADR-018, and marked the transition-only records deprecated without deleting their history.

Delivered:
- ADR-010 lost the stale second-store sentence and the redundant PostgreSQL qualifiers.
- ADR-012 is now "Resolve Accounts By External User Identifier": account references resolve through platform_identities.external_user_id with no compatibility lookup.
- ADR-015 keeps the authenticated AES-256-GCM decision and drops the retained-document, earlier AES-CFB, and backfill wording.
- ADR-016 is now "PostgreSQL Owns Calendar, OTP, And Daily-Log Data And Friend Requests Are Retired" as a current data-ownership decision that links ADR-018 and the PostgreSQL Data Boundaries contract, and it states its supersession of ADR-011.
- New ADR-018 "PostgreSQL Is The Single Authoritative Store" satisfies FR-121 and consolidates the single-store and account-identity rules.
- ADR-011 remains superseded by ADR-016 with its surviving principles now linked to ADR-016 and ADR-018; ADR-013/ADR-014/ADR-017 are deprecated with dated notes.
- docs/design/README.md index rows match the reformulated titles and include ADR-018; server/docs/postgres-data-boundaries.md lists ADR-018 instead of deprecated ADR-017.

Known residual (owner-approved): the deprecated ADR-011/013/014/017 bodies and their index titles keep their historical MongoDB and legacy wording by decision.

Evidence: active ADRs (010, 012, 015, 016, 018) contain no MongoDB or migration-era legacy wording; npm run format:markdown, format:markdown:check, lint:markdown, test:markdown-format (30 tests), and fmt:check all pass; graphify update . rebuilt the graph (5501 nodes, 9721 edges, 409 communities). Unit and e2e tests are exempt for this documentation-only change per BACKLOG_WORKFLOW.md and the task DoD. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
