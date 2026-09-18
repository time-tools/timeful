---
id: TASK-0259
title: Document the Backlog Weekly Cleanup configuration and cleanup token setup
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-18 07:27'
updated_date: '2026-09-18 07:32'
labels: []
dependencies: []
references:
  - docs/ci.md
  - BACKLOG_WORKFLOW.md
  - .github/workflows/backlog-weekly-cleanup.yml
  - >-
    backlog/tasks/task-0258 -
    Run-the-completed-task-Backlog-cleanup-weekly-in-GitHub-Actions.md
modified_files:
  - docs/ci.md
  - BACKLOG_WORKFLOW.md
type: docs
ordinal: 259000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0258 added the scheduled Backlog Weekly Cleanup workflow, but the repository configuration it needs was never written down. Document, in `docs/ci.md`, the Actions setting and secret that make it work, and name the workflow from `BACKLOG_WORKFLOW.md`.

Required content:
- The `Backlog Weekly Cleanup` row in the docs/ci.md Workflows table.
- What the workflow does: weekly Sunday 00:00 UTC and manual dispatch, moves tasks in the terminal Done status to `backlog/completed/` with `backlog task complete`, opens or refreshes one `chore/backlog-weekly-completed` pull request, and no-op behavior when no Done task exists.
- Enabling "Allow GitHub Actions to create and approve pull requests" at the organization and repository levels, explained as required for the `github.token` fallback while the PAT path is not gated by it.
- Creating `BACKLOG_CLEANUP_TOKEN` as a fine-grained personal access token with `Contents: Read and write` and `Pull requests: Read and write` (`Metadata: Read-only` automatic), plus storing and rotating it as a repository Actions secret.
- UI navigation steps written as ordered lists, without arrow notation.

Out of scope: changing `.github/workflows/backlog-weekly-cleanup.yml`, the pull request template, or any runtime code or tests.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/ci.md lists backlog-weekly-cleanup.yml in the Workflows table and describes its schedule, moves, single pull request, and no-op behavior
- [x] #2 docs/ci.md documents enabling "Allow GitHub Actions to create and approve pull requests" at the organization and repository levels and why the github.token path needs it
- [x] #3 docs/ci.md documents creating BACKLOG_CLEANUP_TOKEN as a fine-grained personal access token with Contents: Read and write and Pull requests: Read and write, and storing it as an Actions secret
- [x] #4 BACKLOG_WORKFLOW.md names the Backlog Weekly Cleanup workflow and links docs/ci.md
- [x] #5 Changed Markdown passes root npm run format:markdown:check and npm run lint:markdown
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `graphify update .`
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add the Backlog Weekly Cleanup row to the docs/ci.md Workflows table.
2. Add the "Backlog weekly cleanup", "Enabling pull request creation", and "Backlog cleanup token" sections with ordered-list UI steps and no arrow notation.
3. Add a naming and documentation pointer for the workflow in BACKLOG_WORKFLOW.md.
4. Run root npm run format:markdown, npm run format:markdown:check, and npm run lint:markdown.
5. Finalize: check acceptance criteria with evidence, record the final summary, and leave the Done record in backlog/tasks/.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-18 07:32
---
Follow-up after finalization: added an "Actions secrets" section and table to docs/ci.md indexing `BACKLOG_CLEANUP_TOKEN` (used by `backlog-weekly-cleanup.yml`; fine-grained personal access token with Contents: Read and write and Pull requests: Read and write; the workflow falls back to `github.token` when it is unset). It is currently the only secret any workflow reads. `npm run format:markdown` reformatted docs/ci.md, and `npm run format:markdown:check` and `npm run lint:markdown` pass. Modified file remains docs/ci.md.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Documented the repository configuration the Backlog Weekly Cleanup workflow needs, as requested after TASK-0258.

What changed:
- docs/ci.md now lists `backlog-weekly-cleanup.yml` in the Workflows table and adds a "Backlog weekly cleanup" section covering the Sunday 00:00 UTC schedule, manual dispatch, `backlog task complete` moves to `backlog/completed/`, the single `chore/backlog-weekly-completed` pull request, and the no-op path when no Done task exists.
- docs/ci.md adds "Enabling pull request creation" with ordered-list organization and repository steps for "Allow GitHub Actions to create and approve pull requests", explaining that the `github.token` fallback cannot create the cleanup pull request without it while the PAT path is not gated.
- docs/ci.md adds "Backlog cleanup token" with ordered-list steps to create the fine-grained personal access token and store it as the `BACKLOG_CLEANUP_TOKEN` Actions secret, plus rotation guidance.
- BACKLOG_WORKFLOW.md's "Finalizing Tasks" now names the scheduled `Backlog Weekly Cleanup` workflow and links docs/ci.md, closing TASK-0258's follow-up candidate.

Evidence:
- The documented permission set is corrected from the original draft: the workflow pushes and force-updates `chore/backlog-weekly-completed`, and GitHub's fine-grained PAT permission mapping places create/update git refs under Contents write, so the docs require Contents: Read and write and Pull requests: Read and write (Metadata: Read-only is automatic).
- `npm run format:markdown` reformatted docs/ci.md (table alignment); `npm run format:markdown:check` and `npm run lint:markdown` pass.
- No runtime code, tests, or workflow files changed, so swagger, graphify, and root fmt:check triggers did not apply.

Acceptance criteria: the workflow is listed and described (AC1); both Actions setting levels and the fallback rationale are documented (AC2); the PAT permissions, secret storage, and rotation are documented (AC3); BACKLOG_WORKFLOW.md names the workflow and links the doc (AC4); changed Markdown passes the root checks (AC5).
<!-- SECTION:FINAL_SUMMARY:END -->
