---
id: TASK-0226
title: Restore automated backup verification and align the operations docs
status: To Do
assignee: []
created_date: '2026-09-13 19:36'
labels:
  - postgres
  - ops
  - docs
dependencies: []
references:
  - docs/postgres-operations.md
  - docs/environments.md
  - compose.test.yaml
  - server/Dockerfile
priority: medium
type: task
ordinal: 228000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
docs/environments.md and docs/postgres-operations.md state that a restored backup is verified by reconciling per-table row counts and key-content digests, but the only implementation (TestBackupRestoreRehearsal) was deleted with the migration tooling. The test stack still builds testbase-pg and passes POSTGRES_BOOTSTRAP_URI and POSTGRES_BACKUP_URI, and nothing references those variables. PostgreSQL is now the only store, so restore verification is the sole documented recovery check.

Outcome: the repository provides a repeatable restore-reconciliation check on the isolated stack, and the operations documentation makes only true claims. This task covers backup verification, not backup automation, off-host replication, or RPO/RTO; those remain separate operational work.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A repeatable check dumps the isolated test database using the backup role and restores it into a scratch database, then compares per-table row counts and a key-content digest for representative relations.
- [ ] #2 The check fails when a relation's count or digest differs, demonstrated with a deliberately mismatched or missing relation.
- [ ] #3 POSTGRES_BOOTSTRAP_URI, POSTGRES_BACKUP_URI, and the testbase-pg image wiring are either used by the check or removed from compose, Dockerfile, CI, and docs.
- [ ] #4 docs/environments.md and docs/postgres-operations.md describe only restore-verification capabilities that the repository actually provides.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
