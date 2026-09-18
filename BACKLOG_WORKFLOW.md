# Backlog Workflow

This document is the authoritative policy for using Backlog in this repository.
Backlog MCP is the transport for task, milestone, document, and Definition of Done operations.
Its workflow resources are not repository policy.

## Before Using Backlog

Read this document before creating, changing, executing, or finalizing a Backlog task.

Use Backlog to track work that requires investigation, design, sequencing, or other implementation decisions.
Do not create a task for questions, requests to explain existing work, or obvious mechanical changes.

## Creating Tasks

Before creating a task, search active Backlog tasks with a focused query and appropriate status filter.
Reuse an existing task when it covers the work.

Create one task for work that can be delivered and reviewed as one focused change.
Split independently deliverable work into separate tasks, or use a parent and subtasks for tightly coupled sequential work.

Write task descriptions and acceptance criteria as durable work orders for a future agent.
State the outcome and constraints, not a speculative implementation.
Acceptance criteria define the required behavior and scope; the Definition of Done defines completion hygiene.

## Executing Tasks

Before implementation, read the full task, confirm its dependencies and scope, mark it In Progress, assign it, research the current system, and record an implementation plan in the task.

Keep the task plan and implementation notes current as work proceeds.
If work outside the acceptance criteria is discovered, stop and ask whether to expand the task or create a follow-up.
Do not silently change scope.

## Finalizing Tasks

Verify each acceptance criterion with objective evidence.
Record a concise final summary, mark the task Done, and leave completed tasks in the Done state until periodic cleanup.
Moving Done tasks to `backlog/completed/` is always a separate step, never part of task finalization or the completing pull request.
The scheduled `Backlog Weekly Cleanup` workflow performs this periodic cleanup; see `docs/ci.md` for its configuration.
Do not use task completion archival as part of normal task finalization.

Use Backlog MCP tools for managed task, milestone, document, and project DoD records.
Do not directly edit their generated Markdown files.

## Pull Request Preparation

A pull request is the merge unit, and its merge-readiness criteria live in the PR Definition of Done in `.github/pull_request_template.md`, separate from the task Definition of Done that governs task finalization.

Before creating or updating a pull request body, confirm that every task the pull request closes is marked Done and committed in the pull request, with its record left in `backlog/tasks/`.
Do not move Done tasks to `backlog/completed/` as part of a pull request.

## Definition Of Done

Project Definition of Done defaults apply to new tasks unless a task has an exceptional, documented override, and they enumerate the per-change hygiene checks that apply when a task changes the relevant artifacts: swagger regeneration, `graphify update .`, root `npm run fmt:check`, and contract document updates.
Documentation-only changes are changes limited to documentation or agent instructions and that do not modify runtime code, tests, build or deployment configuration, generated artifacts, or runtime assets.

For documentation-only changes, unit and e2e tests are not required unless the user explicitly requests the respective test.
Run the relevant checks for all other work and report any checks that could not be run.
