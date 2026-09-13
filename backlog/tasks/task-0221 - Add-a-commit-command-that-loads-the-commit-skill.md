---
id: TASK-0221
title: Add a /commit command that loads the commit skill
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-13 18:13'
updated_date: '2026-09-13 18:14'
labels:
  - tooling
  - agents
dependencies: []
references:
  - eca89bf727662c7b7bec67dafa800fc69de7de7b
  - .agents/skills/commit/SKILL.md
  - .opencode/command/review-task.md
priority: medium
type: task
ordinal: 224000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Give agents a slash-command entry point for the existing commit skill, mirroring the /review-task command added in eca89bf7. The skill at .agents/skills/commit/SKILL.md already defines the required conventional-commit format, staged-changes-only rule, and model-identity resolution; the missing piece is an opencode command that loads it. The result is .opencode/command/commit.md, so /commit becomes a repeatable invocation instead of an ad hoc request.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The /commit command is exposed from .opencode/command/commit.md with a frontmatter description stating that it creates a commit following the repository commit conventions
- [x] #2 The command instructs the agent to load the commit skill with the skill tool and follow it, treating $ARGUMENTS as optional user guidance that cannot override the skill's message format, staged-changes-only, and model-identity rules
- [x] #3 Changed Markdown files are formatted with npm run format:markdown and pass npm run format:markdown:check
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
1. Add .opencode/command/commit.md mirroring the .opencode/command/review-task.md shape: frontmatter description plus a body that loads the commit skill with the skill tool and forwards $ARGUMENTS as optional guidance, keeping the skill's format, staged-only, and model-identity rules authoritative.
2. Run npm run format:markdown from the repository root, then verify with npm run format:markdown:check and npm run lint:markdown.
3. Validate the command file's frontmatter and name/filename match with a scripted check.
4. Finalize TASK-0221 with evidence and mark it Done.

Research notes: opencode commands live in .opencode/command/<name>.md and expose /<name>. The existing review-task command forwards $ARGUMENTS and asks for required input. The commit skill needs no required argument, so arguments are optional guidance only.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented .opencode/command/commit.md following the .opencode/command/review-task.md shape. Kept arguments optional because the commit skill defines no required parameter; unlike /review-task, there is no missing-input branch. Verified frontmatter/body with a scripted check and ran the root Markdown format, format:check, and lint scripts; all pass. Discovered scripts/markdown.mjs excludes .opencode/**, .agents/**, and backlog/** from the Markdown pipeline, so the checks do not exercise this file directly.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the /commit opencode command that loads the existing commit skill.

.opencode/command/commit.md carries a frontmatter description ("Create a git commit following the repository commit conventions.") and instructs the agent to load the `commit` skill with the skill tool and follow it; a non-empty `$ARGUMENTS` is treated as optional guidance only, with the skill's message format, staged-changes-only rule, and model-identity rules remaining authoritative. This mirrors the /review-task command pattern from eca89bf7 and makes the commit workflow invocable as a slash command.

Verification: a Node check parsed the frontmatter (description present, filename/command name match), confirmed the body references the skill tool and $ARGUMENTS, and the referenced .agents/skills/commit/SKILL.md exists. `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all pass. Note that scripts/markdown.mjs intentionally excludes `.opencode/**`, so the Markdown pipeline passes without reformatting this path; the file was authored in the repository's one-sentence-per-line style. This is a documentation/agent-instructions-only change, so unit and e2e tests are exempt per the project Definition of Done. opencode must be restarted to load the new command. No commit was made; the worktree also contains unrelated in-flight changes from other work.
<!-- SECTION:FINAL_SUMMARY:END -->
