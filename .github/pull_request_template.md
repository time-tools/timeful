<!-- PR title: use conventional commit style, e.g. "feat(frontend): ..." or "fix(server): ...". -->

## Summary

<!-- What changed and why? State the motivation or root cause, not just the diff, and link the motivating issue. -->

## Changes

<!-- Key changes as concise bullets, grouped per Backlog task or per area. For multi-task PRs, use one subsection per group. -->

## Non-goals

<!-- What is deliberately deferred or out of scope? Name the follow-up tasks. Remove this section if empty. -->

## Backlog tasks

<!-- Reference the tracked tasks, e.g. TASK-0165, TASK-0182. Remove this section if none. -->

## PR Definition of Done

<!-- Merge-unit criteria, distinct from the MCP-managed task Definition of Done that governs task finalization.
     CI is the backstop for lint, formatting, typecheck, build, unit tests, Markdown, actionlint, backend tests, and all E2E projects. -->

- [ ] Every task this PR marks Done is committed in this PR with its `status: Done` record in `backlog/tasks/`, and is not moved to `backlog/completed/` here.
- [ ] CI is green.

## Validation

<!-- CI enforces the repository checks for frontend, Markdown, workflows, backend, and E2E.
     Record local runs done for faster feedback, and list anything you could not run with the reason. -->

- Local runs:
- Not run / not covered (reason):

## Risk and impact

<!-- Breaking changes, env var contract changes (`docs/environments.md`), DB migrations (backfill behavior, downgrade policy, rollout notes in `docs/postgres-operations.md`), browser-plugin `window.postMessage` payload changes (`PLUGIN_API_README.md`), stale docs to update, and follow-ups. Remove this section if empty. -->

## Reviewer notes

<!-- Point reviewers at the riskiest files or decisions and anything you want challenged. Remove this section if empty. -->
