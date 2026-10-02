---
id: TASK-0339
title: >-
  Make the GALA translation roster a rendering of typed data so its verdicts
  cannot drift from the tool that reads them
status: To Do
assignee:
  - '@opencode'
created_date: '2026-09-30 20:30'
updated_date: '2026-09-30 20:33'
labels:
  - gala
  - documentation
  - tooling
milestone: gala-rewrite-consolidation
dependencies: []
references:
  - docs/gala-translation.md
  - server/GALA.md
  - docs/gala-rewrite-playbook.md
  - .agents/skills/gala-from-go/SKILL.md
  - server/scripts/20260923_gala_translation_probes/inventory/roster.go
  - server/scripts/20260923_gala_translation_probes/inventory/main.go
  - server/scripts/20260923_gala_translation_probes/audit.sh
  - >-
    backlog/tasks/task-0330 -
    Make-the-GALA-translation-roster-executable-by-an-agent.md
  - >-
    backlog/tasks/task-0338 -
    Re-derive-the-import-resolution-rows-that-GALA-0.84.1-contradicts.md
documentation:
  - docs/gala-translation.md
  - server/GALA.md
  - docs/gala-rewrite-playbook.md
  - .agents/skills/gala-from-go/references/constructs.md
priority: medium
type: task
ordinal: 344000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

The GALA translation roster is hand-maintained Markdown, and three separate programs read it as prose: `inventory/roster.go` parses the construct roster, `inventory/main.go` parses the census, and `audit.sh` parses both plus the mechanical-rewrite catalog. The failure mode is not staleness. It is that **a decision which steers a gate is stored as English.**

`roster.go` reduces a verdict cell to rewrite, split, or keep-handwritten by substring-matching four hand-written phrases, and `readLadder` reduces the decision-ladder rung labels the same way. The classification is correct today by lexical coincidence, and `roster_test.go` pins the exact English rather than the intent. Rewording a verdict cell to `leave the file Go` silently reclassifies it as a full rewrite, and the rung mapping depends on rung 2's label happening to contain the word the `split` test looks for before the word the `rewrite` test looks for.

An advisory preflight is exactly the place where a silent reclassification is most expensive, because the tool's contract is that a family it does not classify is blocking. That property currently depends on a phrase matching.

## Outcome

The construct roster becomes a rendering of typed data rather than a hand-maintained table, the tool that reads it reads the same typed data, and the project-independent half of the knowledge moves into the skill that is staged for upstream.

This refines TASK-0330's constraint that the roster page stays the single source of truth for construct verdicts. That constraint stands; what changes is that the page stops being the *authoring* surface and becomes the *published* one, so a second competing document is still impossible and a drifting edit is no longer expressible.

## Decisions taken

- **D1, page shape.** The narrative sections stay hand-written and the tables become generated. Hand-written: sources and scope, the gap-classification scheme, compiler history including the `delta_*` record, the genuine-gaps prose, and when-to-revisit. Generated: the census table, the construct roster, workarounds and contested verdicts, the probe tables, replaced-by-analog-or-workaround, and the mechanical-rewrite catalog. Generating the whole page would discard the compiler history, which is this repository's own record of a specific release transition and has no place in a version-agnostic skill.
- **D2, formatter.** Emitter output must survive `npm run format:markdown:check` unchanged, and a required check runs the emitter, then the formatter check, then a diff, so a disagreement between Go output and Prettier fails loudly instead of silently. CI configuration is not changed. If this proves unworkable in practice, exempting the page from the Markdown pipeline is the acceptable alternative and the reason must be recorded in the task that does it.
- **D3, twin inventory.** Not generated. The twin table in `server/README.md` stays hand-written and becomes the only copy; `server/GALA.md` points at it. Fifteen stable rows with editorial columns, and the regeneration command column is not derivable from a walk of the tree.

## Scope

Delivered as ordered subtasks, each independently reviewable and each landing its own artifact:

1. Extract the pin-admissibility method into the skill and make the skill genuinely project-independent. Independent of the rest and may land first.
2. Replace the tool's prose parsers with a typed construct-verdict table and a typed decision ladder.
3. Make the tool emit the generated regions of the roster page.
4. Move the audit script's roster checks into Go tests and reduce it to genuinely cross-file checks.
5. Reduce `server/GALA.md` and the rewrite playbook to project facts and point them at the skill.

## Sequencing

TASK-0338 is in flight and edits `SKILL.md`, `constructs.md`, `server/GALA.md`, the roster page, `run.sh`, and `audit.sh`, and requires the inventory tool and its unit tests to be green. Every subtask below depends on it, so this work is designed against its result rather than racing it.

Subtask 1 does not depend on subtask 2 and can proceed once 0338 lands. Subtasks 2, 3, 4, and 5 are sequential.

TASK-0330.05 runs the rewrite playbook end to end on a real server file and should run after subtask 5, because subtask 5 changes the playbook it exercises. That is a coordination note, not a change to 0330.05.

TASK-0323 remains the version-gated corpus run and is not duplicated here.

## Constraints

- `gala` 0.84.1 and the go1.26 series stay the pinned toolchains, and the vendored runtime under `server/third_party/gala/` is not modified.
- The committed twins and the corpus expectations are not re-baselined as a side effect of this work.
- The probe corpus stays where it is. Nothing under `server/scripts/20260923_gala_translation_probes/probes/` moves, and the runner is not replaced.
- `audit.sh` stays read-only and stays free of a GALA and a Go toolchain for the checks that remain in it.
- The preflight exit-code contract is preserved: `0` for every verdict including the one that says stop, and `2`, `3`, `4` meaning the tool could not answer rather than the file being blocked. Exit `3` changes meaning from a restructured page to an incomplete table, and that change is stated wherever the status is documented.
- The skill's existing rule that construct rows are version-agnostic and name a check rather than a release is not relaxed. A version-pair delta such as the `delta_*` probes stays in this repository's compiler history and does not move into the skill.
- Root Markdown follows the sentence-per-line and table-row rules in `docs/AGENTS.md`, and terminology follows `docs/terminology/README.md`.
- No new upstream issue or PR. The work is local.

## Evidence required

Each subtask records its own verification in its task notes and final summary.
This task is finalized once all five subtasks are Done and their artifacts are cross-linked.
<!-- SECTION:DESCRIPTION:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->
