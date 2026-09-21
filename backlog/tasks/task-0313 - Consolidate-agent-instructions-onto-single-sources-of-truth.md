---
id: TASK-0313
title: Consolidate agent instructions onto single sources of truth
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-21 21:37'
updated_date: '2026-09-21 21:53'
labels: []
dependencies: []
references:
  - AGENTS.md
  - BACKLOG_WORKFLOW.md
  - docs/environments.md
  - docs/terminology/README.md
  - docs/codebase-memory.md
  - .github/pull_request_template.md
modified_files:
  - AGENTS.md
  - frontend/AGENTS.md
  - e2e/AGENTS.md
  - docs/AGENTS.md
  - docs/requirements/README.md
  - docs/requirements/migration/README.md
  - docs/environments.md
  - docs/codebase-memory.md
  - BACKLOG_WORKFLOW.md
  - server/README.md
  - frontend/README.md
  - .github/pull_request_template.md
  - .github/workflows/backlog-weekly-cleanup.yml
  - .agents/skills/handoff/SKILL.md
  - .agents/skills/review-task/SKILL.md
  - scripts/handoff/create-handoff.sh
  - .opencode/command/handoff.md
priority: medium
ordinal: 313000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Why: the repository instruction surface (root AGENTS.md, nested AGENTS.md files, skills, READMEs, and policy docs) restates the same rules in up to five places and carries five contradictions. Agents load root plus nested files, so duplicated rules double the instruction budget and drift; contradictory copies make behavior depend on which file an agent read.

Outcome: every topic has one normative owner, all other files carry pointers, and the contradictions are reconciled so agents receive consistent guidance.

Scope: agent-instruction Markdown, human READMEs, the PR template, the cleanup workflow body, and the handoff script. No runtime code, tests, or build configuration changes.

Confirmed constraints:
- Documentation-only change, so the Backlog Definition of Done exempts unit and e2e tests; Markdown checks still run.
- Each rule lives at the level that owns its trigger: root AGENTS.md routes to nested files, nested files own scoped detail.
- Firefox is the canonical local E2E project; chromium remains CI matrix detail.
- The handoff script keeps the canonical section headings; the skill keeps content guidance, with cross-reference comments in both.
- Add .opencode/command/handoff.md for parity with the commit and review-task commands.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root AGENTS.md replaces the restated Cross-Cutting Frontend Rules, Documentation Authoring, Server Test Workflow, Local Firefox E2E Verification, Local Frontend Debug details, Required Checks, codebase-memory, and Swagger sections with pointers to their canonical owners, and keeps only repo-wide rules plus the Bug Fix Protocol.
- [x] #2 Backend test guidance is consistent: server/README.md includes the docker volume create step and the retained-state default, and docs/environments.md carries the isolation semantics without a duplicate command block or the redundant POSTGRES_TEST_DATABASE override.
- [x] #3 frontend/README.md points to the frontend/AGENTS.md Required Checks instead of listing an incomplete check set that omits fmt:check.
- [x] #4 CI worker wording in docs/environments.md matches playwright.config.ts and the e2e-ci.yml overrides.
- [x] #5 e2e/AGENTS.md names firefox-desktop as the canonical local project, resolving the conflict with AGENTS.md.
- [x] #6 Terminology rules live only in docs/terminology/README.md, docs/AGENTS.md owns the repo-wide Markdown sentence and table rules, and root AGENTS.md, docs/requirements/README.md, docs/requirements/AGENTS.md, and frontend/AGENTS.md point to them.
- [x] #7 Root AGENTS.md, the review-task skill, .github/pull_request_template.md, and the cleanup workflow body reference BACKLOG_WORKFLOW.md instead of restating Backlog no-edit and Done-record rules.
- [x] #8 /handoff resolves to a command wrapper, and scripts/handoff/create-handoff.sh and .agents/skills/handoff/SKILL.md cross-reference each other for the section list.
- [x] #9 npm run lint:markdown, npm run format:markdown:check, and npm run test:markdown-rules pass, and every added pointer path resolves.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Create this task, mark it In Progress, and record the plan.
2. Reconcile contradictions in server/README.md, docs/environments.md, frontend/README.md, and e2e/AGENTS.md.
3. Slim root AGENTS.md to pointers, keeping the Bug Fix Protocol and the Backlog trigger.
4. Consolidate terminology, Markdown authoring, codebase-memory, Swagger, and Backlog policy onto single owners across the child files.
5. Add .opencode/command/handoff.md and sync comments in the handoff script and skill.
6. Run the Markdown checks and verify every pointer target.
7. Record the final summary and mark the task Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the consolidation. Root AGENTS.md now routes to the nested owners and keeps only repo-wide rules plus the Bug Fix Protocol. server/README.md owns the backend test command sequence, including the required external-volume creation and the retained-state default, while docs/environments.md keeps the isolation semantics and the corrected CI worker wording. frontend/README.md points at the frontend required checks. e2e/AGENTS.md names firefox-desktop as the canonical local project and documents project selection by testMatch because the firefox projects match only timed-event Firefox specs. Terminology restatements were reduced to pointers in docs/AGENTS.md, docs/requirements/README.md, and docs/requirements/migration/README.md. The review-task skill, pull request template, and cleanup workflow body now reference BACKLOG_WORKFLOW.md. The handoff script and skill carry sync notes, and .opencode/command/handoff.md was added. The planned replacement of the BACKLOG_WORKFLOW bug-fix sentence was unnecessary because that sentence is not present in the current revision.

Verification: npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-rules (30 tests), root npm run fmt:check, and nix develop actionlint all pass. Every pointer target and anchor introduced by this change resolves. Pre-existing broken links in docs/requirements/migration/README.md (relative-source-file.md and backlog/backlog.md anchors) are unchanged and out of scope.

Merge resolution on branch improve-agent-instructions: resolved the carried-over conflicts preferring the consolidation version. AGENTS.md keeps one Bug Fix Protocol section and drops the layout-fixes bullet duplicated in frontend/AGENTS.md. e2e/AGENTS.md keeps the Firefox canonical and project-selection bullets and also retains the branch regression-check bullet. BACKLOG_WORKFLOW.md kept its newly added bug-fix sentence as a pointer to the root Bug Fix Protocol. No conflict markers remain, and Markdown checks still pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Consolidated the repository agent-instruction surface onto single owners and removed the five contradictions. Root AGENTS.md is now a routing file that points to frontend/AGENTS.md, e2e/AGENTS.md, docs/AGENTS.md, docs/terminology/README.md, docs/codebase-memory.md, server/README.md, docs/environments.md, server/routes/README.md, and BACKLOG_WORKFLOW.md for scoped detail. The backend test workflow is consistent across server/README.md and docs/environments.md, the frontend check list is single-sourced, the CI worker wording matches playwright.config.ts and e2e-ci.yml, Firefox desktop is named the canonical local E2E project with selection by testMatch, and /handoff resolves to a command wrapper. Documentation-only change; Markdown format, lint, rule tests, root fmt:check, and actionlint pass.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
