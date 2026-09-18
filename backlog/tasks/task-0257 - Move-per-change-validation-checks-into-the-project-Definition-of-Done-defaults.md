---
id: TASK-0257
title: Move per-change validation checks into the project Definition of Done defaults
status: Done
assignee:
  - opencode
created_date: '2026-09-17 17:04'
updated_date: '2026-09-17 17:05'
labels:
  - developer-experience
  - github
dependencies: []
type: docs
ordinal: 257000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The pull request template restates per-change checks (swagger regeneration, `graphify update .`, root `npm run fmt:check`) and contract-document updates that are finalization hygiene rather than merge-unit criteria. They overlap the Definition of Done that governs task finalization, so every pull request carries the same instructions twice, and a pull request can claim them before the corresponding task records them. The merge-unit template should gate only what belongs to the merge unit, and the project Definition of Done defaults should enumerate the conditional per-change checks so task finalization cannot omit them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Project Definition of Done defaults include conditional items for swagger regeneration, `graphify update .`, root `npm run fmt:check`, and contract-affecting document updates
- [x] #2 `.github/pull_request_template.md` no longer lists or gates those checks in its PR Definition of Done or Validation sections
- [x] #3 The PR template Validation section records local runs and not-run reasons as non-checkbox lines
- [x] #4 `BACKLOG_WORKFLOW.md` states that per-change hygiene checks are enumerated in the project Definition of Done defaults
- [x] #5 Root `npm run format:markdown:check` and `npm run lint:markdown` pass
- [x] #6 No runtime code, tests, build configuration, or CI workflow files are modified
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
1. Upsert project Definition of Done defaults with conditional per-change items: swagger regeneration, `graphify update .`, root `npm run fmt:check`, and contract-affecting document updates. Create TASK-0257 after the upsert so it inherits the new defaults (done).
2. `.github/pull_request_template.md`: reduce `## PR Definition of Done` to merge-unit criteria (in-place `status: Done` records not moved to `backlog/completed/`, CI green); remove the contract-documents bullet and the duplicate validation-record bullet; in `## Validation`, replace the swagger/graphify/root-fmt checkboxes with non-checkbox `Local runs:` and `Not run / not covered (reason):` lines and keep the CI-backstop comment.
3. `BACKLOG_WORKFLOW.md`: state in the Definition Of Done section that per-change hygiene checks are enumerated in the project Definition of Done defaults.
4. Checks: from the repository root run `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown`; run `graphify update .`.
5. Finalize TASK-0257 with evidence, mark it Done, and leave its record in `backlog/tasks/` per BACKLOG_WORKFLOW.md.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The Definition of Done defaults upsert had to precede task creation so this task inherited the new eight-item DoD and exercised it.

The Backlog MCP schema forbids commas in default items, so the contract-documents item uses semicolons to separate examples.

Validation keeps non-checkbox bullet labels because plain consecutive lines would render as one paragraph on GitHub.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved per-change hygiene checks from the pull request template into the project Definition of Done defaults. The defaults in `backlog/config.yml` (updated through Backlog MCP) gain four conditional items: swagger regeneration (`swag init` from `server/` plus `npm run gen:api` from `frontend/`), `graphify update .` when code changed, root `npm run fmt:check` when `scripts/` or `prettier/` changed, and contract-affecting document updates (`docs/environments.md`, `PLUGIN_API_README.md`, migration and rollout notes). `.github/pull_request_template.md` now gates only merge-unit criteria: in-place `status: Done` task records and CI green. Its Validation section keeps a non-checkbox record of local runs and not-run reasons. `BACKLOG_WORKFLOW.md` states that the project Definition of Done defaults enumerate the per-change hygiene checks. This supersedes the PR-time validation bullets added by TASK-0256.

Evidence: the eight project DoD defaults were upserted before TASK-0257 was created, and this task inherited all of them; `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all pass from the repository root; `graphify update .` completed with no code-graph topology changes; `git status` shows only `.github/pull_request_template.md`, `BACKLOG_WORKFLOW.md`, `backlog/config.yml`, and this task record, with no runtime code, tests, build configuration, or CI workflow changes. Documentation-only change, so unit and e2e tests are exempt per BACKLOG_WORKFLOW.md.
<!-- SECTION:FINAL_SUMMARY:END -->
