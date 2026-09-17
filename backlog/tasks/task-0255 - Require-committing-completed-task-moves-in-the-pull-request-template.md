---
id: TASK-0255
title: Require committing completed-task moves in the pull request template
status: Done
assignee:
  - opencode
created_date: '2026-09-17 15:48'
updated_date: '2026-09-17 15:57'
labels:
  - developer-experience
  - github
dependencies: []
type: chore
ordinal: 255000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Completing a Backlog task moves its file from `backlog/tasks/` to `backlog/completed/`, and the backlog configuration does not auto-commit. A PR can therefore merge while the completed task records stay behind, leaving the merged backlog state inconsistent with the tasks the PR closed. The pull request template should make moving completed tasks to `backlog/completed/` and committing that move an explicit, checked requirement. Task finalization itself still leaves the task in `backlog/tasks/`; the move happens when the completing PR is prepared.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `.github/pull_request_template.md` includes a visible requirement that every Backlog task completed by the PR is moved to `backlog/completed/` and that this move is committed in the PR
- [x] #2 The requirement is visible in the rendered PR body and placed with the task references in the Backlog tasks section
- [x] #3 `npm run format:markdown:check` and `npm run lint:markdown` pass from the repository root
- [x] #4 No runtime code, tests, build configuration, or CI workflow files are modified
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
1. Read the current `.github/pull_request_template.md` and choose placement: a visible checkbox in the Backlog tasks section, next to the task references.
2. Add the requirement that every task completed by the PR is moved to `backlog/completed/` and that the move is committed.
3. Run `npm run format:markdown:check` and `npm run lint:markdown` from the repository root.
4. Confirm `git status` shows only `.github/pull_request_template.md` plus the Backlog task record, with no runtime code, tests, build configuration, or CI workflow files.
5. Check acceptance criteria with evidence, record the final summary, mark Done, and leave the task in `backlog/tasks/` per user instruction (no immediate move to `backlog/completed/`).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Placed the requirement as a visible checkbox in the Backlog tasks section, matching the existing section structure; section guidance stays in HTML comments.

Scope was briefly expanded with user approval to amend BACKLOG_WORKFLOW.md, then reverted at the user's direction: the task must not move to backlog/completed/ immediately at finalization, so the workflow policy stays unchanged and this task remains in backlog/tasks/ with status Done.

The new template requirement applies when preparing the completing PR, not at task finalization.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a visible checklist requirement to the Backlog tasks section of `.github/pull_request_template.md`: every task the PR marks Done is moved to `backlog/completed/`, and that change is committed in the PR. Motivation: completing a Backlog task moves its file out of `backlog/tasks/`, and the backlog configuration does not auto-commit, so a merged PR can leave the backlog state inconsistent with the tasks it closed. The task stays in `backlog/tasks/` when finalized, per user direction; the move happens when the completing PR is prepared. Evidence: `git diff` shows the single checkbox line in the Backlog tasks section; `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass from the repository root; `git status` shows only the template plus this task record, with no runtime code, tests, build configuration, or CI workflow changes. Documentation-only change, so unit and e2e tests are exempt per BACKLOG_WORKFLOW.md. A parallel amendment to `BACKLOG_WORKFLOW.md` was reverted at the user's request; finalization keeps tasks in `backlog/tasks/`.
<!-- SECTION:FINAL_SUMMARY:END -->
