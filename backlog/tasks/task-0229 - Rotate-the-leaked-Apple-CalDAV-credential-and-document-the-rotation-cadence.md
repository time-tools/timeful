---
id: TASK-0229
title: Rotate the leaked Apple CalDAV credential and document the rotation cadence
status: To Do
assignee: []
created_date: '2026-09-13 19:44'
labels:
  - security
  - docs
dependencies: []
references:
  - server/scripts/20240721_apple_calendar_test/main.go
documentation:
  - docs/environments.md
priority: high
type: task
ordinal: 231000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0223 removed the hardcoded iCloud account email and app-specific password from `server/scripts/20240721_apple_calendar_test/main.go`, but the old literal still exists in git history, so the credential remains exposed until it is revoked at the provider. TASK-0223 is only the repository cleanup; this task owns the rotation and the durable rotation guidance.

This work requires the credential owner to act in the Apple account: it cannot be completed by a repository-only change and should not be assigned to an agent. The old value is discoverable in git history for the script's previous revisions; do not copy the old or new secret into any tracked file, including this task.

Outcome: the leaked app-specific password is revoked and replaced by its owner, the rotation is recorded in a task comment, and a rotation cadence or expectations note for repository-adjacent credentials is durable in an operator-facing document.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The leaked app-specific password for the iCloud account previously embedded in server/scripts/20240721_apple_calendar_test/main.go is revoked and replaced by the credential owner, and the rotation is recorded in a task comment without recording either secret value.
- [ ] #2 The task notes that git history still contains the old credential and that rotation is the mitigation for the historical exposure.
- [ ] #3 A credential-rotation cadence or expectations note for repository-adjacent credentials is documented in docs/environments.md or an equally durable operator document, and this task links to it.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
