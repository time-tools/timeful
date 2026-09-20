---
id: TASK-0258
title: Run the completed-task Backlog cleanup weekly in GitHub Actions
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-17 17:11'
updated_date: '2026-09-17 17:14'
labels:
  - github
  - backlog-md
  - tooling
dependencies: []
references:
  - flake.nix
  - BACKLOG_WORKFLOW.md
  - .github/pull_request_template.md
  - .github/workflows/markdown-ci.yml
priority: medium
type: chore
ordinal: 258000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add a scheduled GitHub Actions workflow that performs the periodic Backlog cleanup described in `BACKLOG_WORKFLOW.md`: move tasks in the terminal Done status from `backlog/tasks/` to `backlog/completed/`, and open a pull request for the moves so a human merges them.

Constraints:
- The workflow runs weekly on Sundays and supports manual dispatch.
- It uses the Backlog CLI's `task complete` cleanup command rather than moving files directly.
- The pull request body follows `.github/pull_request_template.md` and records the moved task IDs.
- The workflow must pass `actionlint` in Markdown CI.
- The workflow is idempotent: no Done tasks means no commit and no pull request; a re-run with an open cleanup pull request refreshes it instead of opening a duplicate.
- Pull requests opened with the repository `GITHUB_TOKEN` do not trigger pull-request workflows, so the workflow prefers a `BACKLOG_CLEANUP_TOKEN` secret when one is configured and falls back to `github.token`.

Out of scope: changing `BACKLOG_WORKFLOW.md`, the pull request template, or any other workflow.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 .github/workflows/backlog-weekly-cleanup.yml exists, runs on a weekly Sunday schedule plus workflow_dispatch, and requests contents: write and pull-requests: write
- [x] #2 A run installs the pinned Backlog CLI, moves every task in the terminal Done status from backlog/tasks/ to backlog/completed/ with `backlog task complete`, and exits successfully with no commit or pull request when no Done task exists
- [x] #3 A run with Done tasks commits the moves on the workflow-owned chore/backlog-weekly-completed branch and opens a pull request against main whose body follows .github/pull_request_template.md and lists the moved task IDs
- [x] #4 A re-run with an open cleanup pull request updates that pull request instead of opening a duplicate
- [x] #5 The workflow uses secrets.BACKLOG_CLEANUP_TOKEN when configured and falls back to github.token
- [x] #6 The workflow file passes actionlint 1.7.12, the version Markdown CI runs
- [x] #7 No runtime code, tests, or other workflow files are modified
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
1. Add `.github/workflows/backlog-weekly-cleanup.yml` scheduled for Sundays and manual dispatch.
2. Install the pinned Backlog CLI (`backlog.md@1.52.0`), list Done tasks as JSON, and run `backlog task complete` for each.
3. Commit the moves on the workflow-owned `chore/backlog-weekly-completed` branch, open or refresh the pull request, and skip everything when no Done task exists.
4. Verify the workflow with `actionlint` 1.7.12 and dry-run the cleanup logic against a shallow clone.
5. Finalize: check acceptance criteria with evidence, record the final summary, and keep the Done record in `backlog/tasks/`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verification used a shallow clone plus stubbed `gh` and a local bare remote; the real repository worktree was left with only the new workflow and this task record.

The workflow deliberately uses `npm install --global backlog.md@1.52.0` rather than the Nix flake input: the flake pins a commit two commits ahead of the v1.52.0 release, and neither commit changes `task complete`.

Pull requests created with the repository `GITHUB_TOKEN` do not start pull-request workflows, so the PR body's Reviewer notes tell maintainers to add a `BACKLOG_CLEANUP_TOKEN` secret to get CI on generated cleanup pull requests.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added `.github/workflows/backlog-weekly-cleanup.yml`. The workflow runs on Sundays at 00:00 UTC and on manual dispatch with `contents: write` and `pull-requests: write`, installs the pinned Backlog CLI (`backlog.md@1.52.0`), reads Done task IDs from `backlog task list --status Done --json`, and moves each with `backlog task complete`. With no Done tasks it exits successfully without a commit or pull request.

With Done tasks it commits the moves on the workflow-owned `chore/backlog-weekly-completed` branch, force-with-leases a refresh of an unmerged cleanup pull request, and opens a pull request against `main` with a body that follows the PR template and lists the moved task IDs; a re-run with an open cleanup pull request refreshes it with `gh pr edit` instead of opening a duplicate. `GH_TOKEN` prefers `secrets.BACKLOG_CLEANUP_TOKEN` and falls back to `github.token`, and is scoped to the git-configuration, push, and pull-request steps.

Evidence: `actionlint` 1.7.12, the version Markdown CI runs, passes on the new workflow both with and without shellcheck 0.11.0 in `PATH`. An end-to-end simulation in a shallow clone moved TASK-0253 through TASK-0257 as pure renames (5 files changed, 0 insertions, 0 deletions), the empty-Done path produced `count=0` with a clean worktree, the first and refresh `git push --force-with-lease` paths were exercised against a local bare remote, and the `gh pr create` and `gh pr edit` paths were exercised with stubbed `gh`. Root `npm run test:markdown-rules` and `npm run test:markdown-format` pass (60 tests), and `graphify update .` ran with no topology changes. No runtime code, tests, or other workflow files changed; no Markdown, swagger, `scripts/`, or contract documents were affected.

Follow-up candidate, not in scope: `BACKLOG_WORKFLOW.md` still describes periodic cleanup without naming this workflow.
<!-- SECTION:FINAL_SUMMARY:END -->
