---
id: TASK-0312
title: Codify regression-first bug fixes in agent instructions
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-21 21:20'
updated_date: '2026-09-21 21:21'
labels: []
dependencies: []
references:
  - AGENTS.md
  - BACKLOG_WORKFLOW.md
  - frontend/AGENTS.md
  - e2e/AGENTS.md
modified_files:
  - AGENTS.md
  - BACKLOG_WORKFLOW.md
  - frontend/AGENTS.md
  - e2e/AGENTS.md
  - backlog/backlog.md
priority: medium
ordinal: 313000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Task plans and final summaries already report fail-before/pass-after regression evidence, but no standing instruction requires it, and a scratch note in backlog/backlog.md asks for the agent to follow TDD. Codify the agreed approach: a repo-wide bug-fix protocol that requires observing the failure before the fix and the pass after, scoped to bug fixes only, with a documented scripted or manual reproduction fallback when no test layer can capture the behavior.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 AGENTS.md defines a repo-wide bug-fix protocol: add or update the regression check, observe it fail before the fix, confirm the identical check passes after
- [x] #2 AGENTS.md scopes the protocol to bug fixes and allows a scripted or documented manual reproduction when no test layer can capture the behavior
- [x] #3 BACKLOG_WORKFLOW.md Definition Of Done requires bug-fix tasks to record fail-before/pass-after evidence in notes or the final summary, or document why no layer can cover it
- [x] #4 frontend/AGENTS.md and e2e/AGENTS.md point to the protocol instead of duplicating it
- [x] #5 The scratch note about asking the agent to follow TDD is removed from backlog/backlog.md
- [x] #6 No agent-instruction file still presents a conflicting or duplicated regression-test rule
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add a Bug Fix Protocol section to AGENTS.md and remove the now-redundant Working Defaults regression bullets.
2. Add the bug-fix evidence requirement to the Definition Of Done section of BACKLOG_WORKFLOW.md.
3. Align frontend/AGENTS.md and e2e/AGENTS.md with one-line references to the protocol.
4. Remove the codified scratch note from backlog/backlog.md.
5. Run npm run format:markdown and verify no conflicting wording remains.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added the Bug Fix Protocol section to AGENTS.md, removed the two now-redundant Working Defaults bullets, added the bug-fix evidence requirement to the Definition Of Done section of BACKLOG_WORKFLOW.md, and pointed frontend/AGENTS.md and e2e/AGENTS.md at the protocol. Removed the codified scratch note from backlog/backlog.md.
Validation: node scripts/markdown.mjs format:check (the script behind npm run format:markdown) reports no unformatted files; a search across AGENTS.md, BACKLOG_WORKFLOW.md, frontend/AGENTS.md, e2e/AGENTS.md, e2e/inspect/AGENTS.md, docs/AGENTS.md, .opencode, and .agents found no conflicting regression-test rule. Unit and e2e tests are exempt because this is a documentation-only change; no code, scripts, prettier, swagger, or contract documents changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Codified regression-first bug fixes: AGENTS.md now defines a Bug Fix Protocol requiring an observed failing regression check before the fix and an observed pass after, BACKLOG_WORKFLOW.md requires the fail-before/pass-after evidence in bug-fix task records, and frontend and e2e instructions point at the protocol. Verified with node scripts/markdown.mjs format:check and a repo-wide search for conflicting rules.
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
