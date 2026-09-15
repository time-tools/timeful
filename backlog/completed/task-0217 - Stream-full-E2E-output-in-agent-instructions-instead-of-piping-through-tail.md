---
id: TASK-0217
title: Stream full E2E output in agent instructions instead of piping through tail
status: Done
assignee:
  - opencode
created_date: '2026-09-13 08:22'
updated_date: '2026-09-13 08:24'
labels: []
dependencies: []
priority: medium
type: docs
ordinal: 222000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Agents run browser E2E suites as `npm run test:e2e -- --project=firefox-desktop 2>&1 | tail -70`. Piping through tail hides all progress until the run ends and can mask the suite's exit status. No repository instruction currently covers E2E output handling, so the behavior is agent improvisation.

Codify an explicit rule: E2E runs must stream their full output; never pipe a run through `tail` or `head`; when a persistent full log is needed, use `2>&1 | tee /tmp/opencode/<name>.log`. The default convention stays the plain `npm run test:e2e -- --project=firefox-desktop` invocation, with `tee` reserved for runs whose complete log must be kept. Update the root AGENTS.md Local Firefox E2E Verification section and the e2e/AGENTS.md run instructions, then close the matching inbox note in backlog/backlog.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The root AGENTS.md Local Firefox E2E Verification section instructs agents to run E2E with full streamed output, forbids piping through tail or head, and names `2>&1 | tee /tmp/opencode/<name>.log` as the way to retain a full log.
- [x] #2 e2e/AGENTS.md carries the same streaming rule in its failure-diagnosis or run instructions, consistent with the root AGENTS.md wording.
- [x] #3 The backlog/backlog.md inbox item 'Always show logs when running e2e tests...' is checked off.
- [x] #4 `npm run format:markdown:check` and `npm run lint:markdown` pass from the repo root.
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
1. Root AGENTS.md: add one bullet to the Local Firefox E2E Verification section requiring full streamed output and forbidding `tail`/`head` pipes, with `2>&1 | tee /tmp/opencode/<name>.log` as the persistent-log option.
2. e2e/AGENTS.md: add the equivalent rule to the Failure Diagnosis Loop, wording aligned with root AGENTS.md.
3. backlog/backlog.md: check off the matching inbox item.
4. Run `npm run format:markdown:check` and `npm run lint:markdown` from the repo root; run `npm run format:markdown` only if the checks require it.
5. Review `git diff`, verify each acceptance criterion with objective evidence, record the final summary, and finalize per the task-finalization guide.

Research notes: the run instruction agents see is root AGENTS.md:124-131 (Local Firefox E2E Verification); frontend/AGENTS.md:77-85 delegates to e2e/AGENTS.md; the canonical run and diagnosis rules are e2e/AGENTS.md:7-24 and :41-73. No file prescribes output handling; the only related record is the backlog/backlog.md inbox note. Documentation-only change, so Backlog workflow exempts unit and e2e suites; Markdown checks are the required verification.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented as planned. Research confirmed the run instruction lives in root AGENTS.md:124-131 and e2e/AGENTS.md:7-24/41-73, with frontend/AGENTS.md delegating to the e2e package; the only prior record of the output problem was the backlog/backlog.md inbox note.

Verification: `npm run format:markdown:check` and `npm run lint:markdown` from the repo root exited clean; `git diff --cached` and grep confirmed the added rules and the checked-off inbox item. Backlog workflow exempts documentation-only changes from unit and e2e suites; none were run.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Documentation-only change to agent instructions for E2E runs.

What changed:
- Root `AGENTS.md` (Local Firefox E2E Verification): added a bullet requiring full streamed output, forbidding `tail`/`head` pipes, and naming `2>&1 | tee /tmp/opencode/<name>.log` for a persistent full log.
- `e2e/AGENTS.md` (Failure Diagnosis Loop): added the matching rule with the same command guidance.
- `backlog/backlog.md`: checked off the inbox item "Always show logs when running e2e tests. Also write to a file if needed, never use just tail".

Why: agents improvised `npm run test:e2e ... 2>&1 | tail -70`, which hides all progress until the run ends and can mask the suite's exit status; no repository instruction covered output handling.

Verification (documentation-only, so unit and e2e suites are exempt per BACKLOG_WORKFLOW.md):
- `npm run format:markdown:check` and `npm run lint:markdown` pass from the repo root.
- Diff inspection confirms all three edits at `AGENTS.md:129`, `e2e/AGENTS.md:10-11`, and `backlog/backlog.md:48`.
- `graphify update .` run after the doc changes.

No risks: instruction-only change, no runtime code, tests, or configuration touched.
<!-- SECTION:FINAL_SUMMARY:END -->
