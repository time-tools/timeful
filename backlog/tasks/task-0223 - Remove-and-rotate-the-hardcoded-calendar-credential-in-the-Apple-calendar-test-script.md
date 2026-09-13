---
id: TASK-0223
title: Remove the hardcoded calendar credential in the Apple calendar test script
status: Done
assignee:
  - OpenCode
created_date: '2026-09-13 19:33'
updated_date: '2026-09-13 19:45'
labels:
  - security
  - cleanup
dependencies: []
references:
  - server/scripts/20240721_apple_calendar_test/main.go
modified_files:
  - server/scripts/20240721_apple_calendar_test/main.go
priority: high
type: bug
ordinal: 222000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
server/scripts/20240721_apple_calendar_test/main.go embeds an iCloud account email and app-specific password as string literals. The file is tracked in git, compiled by `go build ./...`, and was found during the post-migration quality audit. The credential is therefore exposed both in the working tree and in git history.

Outcome: the script is kept as a manual CalDAV debugging tool and reads `APPLE_CALDAV_EMAIL` and `APPLE_CALDAV_APP_PASSWORD` from the environment, so no credential literal remains in the repository and both variables are documented in docs/environments.md. Because git history still retains the old value, revoking and replacing the leaked app-specific password is tracked separately in TASK-0229.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 No hardcoded account email, password, or app-specific password remains in server/scripts/20240721_apple_calendar_test/; the kept script reads its values from environment variables and they appear in no tracked file.
- [x] #2 docs/environments.md documents APPLE_CALDAV_EMAIL and APPLE_CALDAV_APP_PASSWORD as manual script variables that must not be stored in a tracked file.
- [x] #3 go build ./... and go vet ./... pass.
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
1. Kept server/scripts/20240721_apple_calendar_test/main.go as a manual debugging tool, per user direction and the earlier decisions that preserved it (TASK-0199.09.01, TASK-0199.09.03, TASK-0211).
2. Replaced the hardcoded iCloud email and app-specific password with os.Getenv reads of APPLE_CALDAV_EMAIL and APPLE_CALDAV_APP_PASSWORD, failing fast with log.Fatal when either is unset or blank; all CalDAV logic is unchanged.
3. Documented both variables in docs/environments.md under the variable-ownership section as manual script variables that must not be stored in tracked files.
4. Verified with repo-wide git grep and hidden-file ripgrep searches for the old literals, gofmt, go build ./..., go vet ./..., go test ./scripts/..., and npm run format:markdown:check.
5. Scope change (user-approved): rotation of the leaked credential and the git-history note moved to TASK-0229; this task is limited to removing the literal and documenting the env vars.
Residual risk: git history still contains the old credential; TASK-0229 owns revocation and replacement.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementer: OpenCode. User-approved scope change during execution: this task now covers only removing the credential literal and documenting the environment variables; credential rotation and the git-history note moved to TASK-0229.

Implementation: server/scripts/20240721_apple_calendar_test/main.go keeps all CalDAV logic but reads APPLE_CALDAV_EMAIL and APPLE_CALDAV_APP_PASSWORD via os.Getenv and calls log.Fatal when either is unset or blank.

Docs: docs/environments.md lists both variables under the variable-ownership section as manual CalDAV debug script variables and states they must never be stored in a tracked file.

Verification evidence: repo-wide `git grep` and a hidden-file/untracked `rg` search for the old email and password literals return no match; the script directory contains only main.go; `gofmt -l` is clean; `go build ./...`, `go vet ./...`, and `go test ./scripts/...` (no test files) pass; `npm run format:markdown:check` passes.

Residual risk: git history retains the old literal, so the credential stays exposed until TASK-0229 revokes and replaces it.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

Removed the hardcoded iCloud account email and app-specific password from `server/scripts/20240721_apple_calendar_test/main.go`. The script is kept as a manual CalDAV debugging tool: it now reads `APPLE_CALDAV_EMAIL` and `APPLE_CALDAV_APP_PASSWORD` from the environment and exits with an error when either is unset or blank, with all CalDAV logic unchanged. `docs/environments.md` documents both variables as manual script variables that must never be stored in a tracked file.

## Scope change

The user approved re-scoping this task to the repository cleanup only. Revoking and replacing the leaked app-specific password, plus documenting a rotation cadence, is tracked separately in TASK-0229.

## Verification

- Repo-wide `git grep` and a hidden-file/untracked `rg` search for the old email and password literals return no match.
- The script directory contains only main.go.
- `gofmt -l` is clean for the changed script.
- `go build ./...`, `go vet ./...`, and `go test ./scripts/...` (no test files) pass from `server/`.
- `npm run format:markdown:check` passes after the docs change.

## Risks / follow-ups

- The old credential remains in git history; TASK-0229 owns revocation and replacement as the mitigation.
- No unit or e2e tests cover this manual script; the PostgreSQL-backed suites are unrelated to this change.
<!-- SECTION:FINAL_SUMMARY:END -->
