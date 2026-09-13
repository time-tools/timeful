---
id: TASK-0219
title: Add a review-task skill for reviewing Backlog tasks against the codebase
status: Done
assignee:
  - opencode
created_date: '2026-09-13 17:58'
updated_date: '2026-09-13 17:59'
labels:
  - tooling
  - backlog
dependencies: []
references:
  - BACKLOG_WORKFLOW.md
  - opencode.json
  - .opencode/opencode.jsonc
  - .agents/skills/commit/SKILL.md
  - .agents/skills/handoff/SKILL.md
  - 'https://opencode.ai/docs/skills/'
modified_files:
  - .agents/skills/review-task/SKILL.md
  - .opencode/command/review-task.md
priority: medium
type: task
ordinal: 222000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Give agents a reusable way to review an existing active Backlog task against the current codebase before or during execution. The review should surface concrete problems such as stale assumptions, invalid references, weak or untestable acceptance criteria, missing dependencies, and overlap with other active work, then ask the user clarifying questions that resolve the open details and decisions. Findings and agreed task updates must be recorded through Backlog MCP rather than by editing task Markdown. The result is a /review-task skill plus a matching slash command that make task readiness review repeatable instead of ad hoc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The skill exists at .agents/skills/review-task/SKILL.md with valid frontmatter whose name matches its directory and whose description states what it does and when to trigger it
- [x] #2 The /review-task command is exposed from .opencode/command/review-task.md and forwards its arguments to the review-task skill
- [x] #3 The skill procedure resolves a Backlog task through Backlog MCP, researches the current codebase, and reports findings grouped by severity with file:line evidence
- [x] #4 The review checks cover already-fixed or stale work, invalid references, scope and deliverability, acceptance-criteria specificity and testability, dependency correctness, Definition of Done fit, overlap with other active tasks, and requirements or terminology alignment
- [x] #5 The skill asks clarifying questions with the question tool before any Backlog write, writes only user-approved comments or description, acceptance-criteria, and dependency updates, never changes task status, and never implements code
- [x] #6 When invoked without a task ID, the skill asks for one before starting research
- [x] #7 Changed Markdown files are formatted with npm run format:markdown from the repository root
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
1. Write .agents/skills/review-task/SKILL.md with frontmatter name: review-task and a trigger-rich description, encoding the procedure: resolve the task ID and ask when absent, read the full task through Backlog MCP, research the codebase read-only (git history since creation or last modification, graphify query when graphify-out/graph.json exists, verify referenced files/routes/symbols, check requirements and terminology alignment, check dependency statuses and overlapping active tasks), report findings grouped by severity with file:line evidence, ask clarifying questions with the question tool, and record only user-approved comments or description/acceptance-criteria/dependency updates while never changing task status or implementing code.
2. Add .opencode/command/review-task.md forwarding $ARGUMENTS to the review-task skill.
3. Format changed Markdown with npm run format:markdown and verify with npm run format:markdown:check.
4. Validate frontmatter name\directory match and command file shape, and smoke-check each review checklist item against an existing active task.
5. Finalize TASK-0219 with evidence, then mark it Done.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the review-task agent skill and /review-task command.

.agents/skills/review-task/SKILL.md (frontmatter name review-task matches the directory; programmatically validated against the skill name regex and the 1-1024 character description limit) defines a six-step review procedure: resolve the task ID and ask when absent, read the full task through Backlog MCP, research the codebase read-only (git history since creation or last update, graphify query when graphify-out/graph.json exists, reference/requirements/terminology checks, dependency status and active-task overlap scans), report findings grouped by Blocker/Important/Minor/Question with file:line evidence, ask clarifying questions with the question tool, and record only user-approved comments or description/acceptance-criteria/dependency updates while never changing task status or implementing code.

.opencode/command/review-task.md exposes /review-task and forwards $ARGUMENTS so the command loads the skill for the given task.

Verification: frontmatter and command shape validated with a Node check; the procedure was dry-run against active task TASK-0139, confirming backlog_task_view, graphify query, referenced-file existence checks (scripts/markdown.mjs, prettier/markdown/sentences-per-line.js, .oxfmtrc.json), and git log all work; npm run format:markdown and format:markdown:check pass, and root markdown lint passes. This is a documentation-only change, so unit and e2e tests are exempt per the project Definition of Done. opencode must be restarted to load the new skill and command.
<!-- SECTION:FINAL_SUMMARY:END -->
