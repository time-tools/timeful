---
id: TASK-0256
title: Define a PR Definition of Done with PR-time backlog checks
status: Done
assignee: []
created_date: '2026-09-17 16:40'
updated_date: '2026-09-17 16:46'
labels:
  - developer-experience
  - github
dependencies: []
type: docs
ordinal: 256000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Define a pull request-level Definition of Done in `.github/pull_request_template.md` and document the PR-preparation rule in `BACKLOG_WORKFLOW.md`. A pull request is the merge unit, and its completion criteria differ from the MCP-managed task Definition of Done that governs task finalization. The template's Backlog tasks section becomes references-only; a new PR Definition of Done section requires that every task the PR marks Done is committed in place with its `status: Done` record in `backlog/tasks/` and is not moved to `backlog/completed/` in the PR; Validation is reframed as local runs plus deviations from CI-covered checks; the workflow document gains a Pull Request Preparation section and states that moving Done tasks to `backlog/completed/` is always a separate periodic-cleanup step. No CI guard or other mechanical blocker is added.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `.github/pull_request_template.md` contains a `## PR Definition of Done` section with merge-gating checkboxes, including that every task the PR marks Done is committed in place with its `status: Done` record in `backlog/tasks/` and is not moved to `backlog/completed/` in the PR
- [x] #2 The Backlog tasks section is references-only and carries no probe or move instruction
- [x] #3 The Validation section records local runs and not-run reasons, and lists only checks CI does not cover: swagger regeneration, `graphify update .`, and root `npm run fmt:check`
- [x] #4 `BACKLOG_WORKFLOW.md` defines the PR Definition of Done as distinct from the task Definition of Done, and states that Done tasks stay in `backlog/tasks/` until periodic cleanup because moving them to `backlog/completed/` is always a separate step, never part of finalization or the completing PR
- [x] #5 `npm run format:markdown:check` and `npm run lint:markdown` pass from the repository root
- [x] #6 No runtime code, tests, build configuration, or CI workflow files are modified
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
1. Edit `.github/pull_request_template.md`: make the Backlog tasks section references-only; add the `## PR Definition of Done` section with the in-place Done record requirement; reframe `## Validation` as local runs plus not-run reasons and CI-uncovered checks.
2. Edit `BACKLOG_WORKFLOW.md`: add the Pull Request Preparation section and reconcile the Finalizing Tasks wording with in-place Done records and the separate periodic-cleanup move.
3. Run `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` from the repository root; run `graphify update .`.
4. Finalize: check acceptance criteria with evidence, record the final summary, and keep the task Done in `backlog/tasks/` with the changes.
5. A later periodic cleanup moves this record to `backlog/completed/` as a separate step, not this PR.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Final policy: finalization and the completing PR keep the task file in `backlog/tasks/` with `status: Done`, committed with the changes; moving records to `backlog/completed/` is always a separate periodic-cleanup step.

An earlier iteration implemented a PR-time `grep` probe and a requirement to move tasks closed by the PR to `backlog/completed/`; both were corrected before landing, and no CI guard or workflow change was added.

The earlier framing of `TASK-0253` and `TASK-0254` as stranded was wrong under this policy: Done tasks correctly remain in `backlog/tasks/` until periodic cleanup.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a `## PR Definition of Done` section to `.github/pull_request_template.md` with merge-gating checkboxes: every task the PR marks Done is committed in the PR with its `status: Done` record in `backlog/tasks/` and is not moved to `backlog/completed/` here, Validation records local runs and not-run reasons, contract-affecting changes update their documents, and CI is green. The Backlog tasks section is references-only. Validation records local runs, not-run reasons, and only CI-uncovered checks (swagger regeneration, `graphify update .`, root `npm run fmt:check`). `BACKLOG_WORKFLOW.md` gains a Pull Request Preparation section that defines the PR Definition of Done as distinct from the task Definition of Done and requires Done records to be committed in place, and its Finalizing Tasks wording states that Done tasks stay in `backlog/tasks/` until periodic cleanup and that moving them to `backlog/completed/` is always a separate step.

Evidence: `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all pass from the repository root; `graphify update .` completed; `git status` shows only `.github/pull_request_template.md`, `BACKLOG_WORKFLOW.md`, and this task record, with no runtime code, tests, build configuration, or CI workflow changes. Documentation-only change, so unit and e2e tests are exempt per BACKLOG_WORKFLOW.md.
<!-- SECTION:FINAL_SUMMARY:END -->
