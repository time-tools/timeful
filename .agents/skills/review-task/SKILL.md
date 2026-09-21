---
name: review-task
description: Review an existing Backlog task against the current codebase, surface stale or invalid assumptions, weak acceptance criteria, dependency and overlap problems, then ask clarifying questions before any approved task updates: use when the user asks to review or sanity-check a Backlog task or invokes /review-task.
---

# Review Task

Review one existing Backlog task against the current codebase before or during execution.
Report concrete problems with evidence, then ask the user the questions that resolve open details and decisions.
This is a review, not execution: do not implement the task, and do not change its status.

## Boundaries

- Read `BACKLOG_WORKFLOW.md` before touching Backlog records, and use Backlog MCP tools for all reads and writes.
- Never edit Backlog-generated Markdown files directly.
- Do not write code, tests, or documentation as part of a review.
- Do not change the task status, priority, assignee, or milestone on your own.
- Ask before any Backlog write, and record only what the user approves.

## Step 1: Resolve The Task

- Take the task ID from the invocation, for example `/review-task TASK-0219`.
- If no task ID was provided, ask the user for one with the `question` tool before any research.
- Read the full task with `backlog_task_view`: description, acceptance criteria, Definition of Done, dependencies, references, documentation, implementation plan, notes, and comments.

## Step 2: Research The Current Codebase

Research before judging, and verify the task's claims rather than trusting their wording.

- Check the repository state: current branch, uncommitted changes, and commits that touch the named files since the task was created or last updated.
- If the codebase-memory-mcp graph is available, query it first (for example `search_graph` or `query_graph`) and use the returned subgraph as the map.
- If the graph is missing or stale, run `codebase-memory-mcp cli index_repository --repo-path .` first, then query it.
- Otherwise read the referenced files, routes, components, composables, scripts, and tests directly.
- For every concrete path, symbol, route, command, or requirement ID the task names, confirm it still exists and record any that do not.
- When the task touches product behavior, read the relevant `docs/requirements/` records and `docs/terminology/glossary.md`, and follow the repository terminology linking rules.
- Check dependency tasks with `backlog_task_view` and confirm they are Done or clearly provide what this task needs.
- Search active tasks with `backlog_task_list` or `backlog_task_search` for overlap with the reviewed task's files or outcome.

## Step 3: Review

Check each dimension and keep evidence for every finding.

- Relevance: is the described problem still present, or has the codebase already solved it?
- References: do the named files, symbols, routes, scripts, packages, and requirement IDs exist?
- Scope: can this be delivered and reviewed as one focused change, or does it hide multiple deliverables?
- Acceptance criteria: are they specific, testable, independent, and free of implementation details? Are negative and edge cases covered? Are test and documentation expectations included in the same task?
- Dependencies: are recorded dependencies required, satisfied, and complete? Are ordering constraints stated?
- Definition of Done: do the project defaults fit the task, or does it qualify for the documentation-only exemption?
- Overlap: does another active task already cover the same work, or conflict with it?
- Requirements and terminology: does the task contradict accepted requirements, ADRs, or canonical glossary terms?

## Step 4: Report Findings

- Group findings under Blocker, Important, Minor, and Question.
- Give every finding a concrete consequence and evidence, citing `file:line`, a commit, or a Backlog record.
- State when a dimension is clean instead of padding the report.
- Do not invent verification evidence or claim behavior you did not check.

## Step 5: Ask Clarifying Questions

- Ask only what the codebase cannot answer: intent, priority, scope, and ambiguous acceptance criteria.
- Use the `question` tool, batch related questions, and give each question concrete options with a recommended one.
- Do not ask the user to confirm facts you can verify yourself.

## Step 6: Record The Outcome

- Present findings and answers before writing anything.
- With explicit user approval, append a review comment with `backlog_task_edit` (`commentsAppend`) that captures findings, decisions, and any unresolved questions.
- With explicit user approval, update the description, acceptance criteria, references, or dependencies to match the agreed outcome, using `acceptanceCriteriaSet`, `acceptanceCriteriaAdd`, and related fields rather than editing Markdown.
- Leave status, priority, assignee, milestone, and final summary unchanged unless the user explicitly asks otherwise.
- Stop before implementation and hand the task back to the user.
