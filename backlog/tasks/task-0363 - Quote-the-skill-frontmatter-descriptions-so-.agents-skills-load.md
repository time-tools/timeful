---
id: TASK-0363
title: Quote the skill frontmatter descriptions so .agents skills load
status: Done
assignee: []
created_date: '2026-10-24 13:41'
updated_date: '2026-10-24 13:42'
labels: []
dependencies: []
references:
  - .agents/skills/commit/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/handoff/SKILL.md
  - .agents/skills/review-task/SKILL.md
modified_files:
  - .agents/skills/commit/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/handoff/SKILL.md
  - .agents/skills/review-task/SKILL.md
priority: medium
type: bug
ordinal: 367300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
All five `SKILL.md` files under `.agents/skills/` carried an unquoted YAML `description` value containing a colon followed by a space, for example `Create a git commit: use when the user explicitly asks to commit changes.`. YAML reads that `: ` as a nested mapping entry, so every frontmatter block failed to parse and none of the skills loaded. Quote each description as a double-quoted YAML scalar without changing its text, then confirm all five frontmatter blocks parse with a `name` matching the directory and the same `description` text. This is an agent-instruction-only change: no runtime code, tests, generated artifacts, build configuration, or deployment files change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every `SKILL.md` under `.agents/skills/` has YAML frontmatter that parses successfully, with a string `name` equal to its directory name and a non-empty string `description`.
- [x] #2 Each description's wording is unchanged apart from the YAML quoting; the original sentence, including its colon-space clause, is preserved.
- [x] #3 Only the `description` lines in `.agents/skills/*/SKILL.md` change; no runtime code, tests, generated artifacts, build configuration, or deployment files change.
- [x] #4 The fail-before and pass-after frontmatter parse evidence is recorded in the task notes and final summary, and the session skill catalog lists all five skills after the fix.
<!-- AC:END -->

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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Plan

1. Quote each `description` value in `.agents/skills/*/SKILL.md` as a double-quoted YAML scalar, preserving its text.
2. Reproduce the frontmatter parse failure with a YAML parser before the change, then re-run the identical check after the change and confirm all five blocks parse with `name` matching the directory and the same description text.
3. Confirm the diff is limited to the five `description` lines and record the fail-before and pass-after evidence in the task notes and final summary.

The work was requested and completed in one pass: the fail-before was observed on the unmodified files and the pass-after on the fixed files.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-10-24)

Quoted all five `description` values under `.agents/skills/` with double quotes. No description text changed; the colon-space clause that broke the YAML is preserved inside the quoted scalar.

## Bug Fix Protocol evidence

- Fail-before: parsing each frontmatter block with `js-yaml` (resolved from `frontend/node_modules` in a Node one-off) failed for all five files with `bad indentation of a mapping entry`, because the unquoted `: ` made the description a nested mapping entry.
- Pass-after: the identical check parsed `commit` (73 description characters), `gala-from-go` (262), `gala-loop` (211), `handoff` (150), and `review-task` (311), with each `name` equal to its directory and a non-empty string `description`.
- The session skill catalog lists all five skills after the fix.
- `git diff --stat -- .agents` reports 5 files changed, 5 insertions, and 5 deletions, and `git diff` shows only the `description` line in each file.

## Checks

- `.agents/**` is excluded by `scripts/markdown.mjs`, so `npm run format:markdown` does not cover these files and no formatting change was introduced.
- No runtime code, tests, generated artifacts, Swagger annotations, `scripts/` or `prettier/` files, or contract documents changed; the documentation-only exemption applies.
- No commit was made.

- `npm run format:markdown:check` passes; the pipeline excludes `.agents/**` and `backlog/**` from its file set. The created task file is the only new untracked path.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

All five `.agents/skills/*/SKILL.md` frontmatter blocks now parse: each `description` is a double-quoted YAML scalar, and its wording is unchanged.

## Verification

- Fail-before: `js-yaml` rejected all five unquoted descriptions with `bad indentation of a mapping entry`.
- Pass-after: the same parser accepts all five, with each `name` matching its directory and the original description text preserved; the session skill catalog lists all five.
- Diff is limited to the five `description` lines, a documentation-only change, so unit and e2e tests, Swagger regeneration, the code index, and `fmt:check` do not apply. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
