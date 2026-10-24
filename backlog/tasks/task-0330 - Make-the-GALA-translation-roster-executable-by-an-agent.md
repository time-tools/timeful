---
id: TASK-0330
title: Make the GALA translation roster executable by an agent
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-26 13:35'
updated_date: '2026-10-04 13:23'
labels: []
dependencies: []
references:
  - 'https://github.com/martianoff/gala/issues/528'
  - 'https://github.com/martianoff/gala/pull/529'
  - docs/gala-translation.md
  - server/GALA.md
  - server/README.md
  - server/scripts/20260923_gala_translation_probes/
  - >-
    backlog/tasks/task-0322 -
    Spike-build-a-Go-to-GALA-translation-roster-for-the-server.md
  - >-
    backlog/tasks/task-0323 -
    Add-a-version-gated-check-for-the-GALA-translation-probe-corpus.md
  - >-
    backlog/tasks/task-0326 -
    Classify-genuine-GALA-gaps-and-gate-the-translation-roster-with-a-structural-audit.md
documentation:
  - docs/gala-translation.md
  - server/GALA.md
priority: medium
type: task
ordinal: 330000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Make `docs/gala-translation.md` usable as an executable work order for rewriting a Go server file in GALA, rather than only as a lookup table.

TASK-0322 built the per-construct roster, TASK-0324 made its verdicts deterministic, TASK-0325 linked every probe, and TASK-0326 classified genuine gaps and gated the page with a read-only audit.
Together they answer "what does GALA do with this Go construct" completely and auditably.
What they do not answer is "given this file, what do I do".
There is no entry procedure, no way to triage a single file mechanically, no directly appliable rewrite catalog, no per-twin verification loop, and no stated condition under which an agent must stop and leave a file handwritten.
An agent handed the roster today has to infer all of that, and the inferred procedure is unverified.

Without a stop rule an agent attempts files the roster classifies as untranslatable and either fails repeatedly or silently reshapes a wire-facing Go API.
With one, the existing knowledge becomes usable.

The language limitations behind these gaps were reported upstream in https://github.com/martianoff/gala/issues/528, which is where the triage recorded in `server/GALA.md` originated.
The goal here is local: make the knowledge already in this repository actionable.
No upstream issue or PR is produced.

## Scope

Delivered as ordered subtasks, each independently reviewable and each landing its own artifact:

1. A single-file preflight mode for the committed inventory tool, so the roster can be asked about one Go file instead of only a whole module.
2. A mechanical rewrite catalog inside the roster, giving one directly appliable rule per row instead of prose spread across an eleven-column table.
3. An agent-facing rewrite playbook: preconditions, the ordered procedure, the per-twin verification loop, abort rules, and a definition of done for one file.
4. A thin invokable skill that routes to the playbook and the preflight command.
5. A pilot that runs the finished procedure end to end on a real server file and records what the procedure failed to say.

## Constraints

- `docs/gala-translation.md` stays the single source of truth for construct verdicts.
  A second competing document is the drift risk this initiative exists to remove, so nothing here may fork it.
- No new upstream issue or PR.
  The work is local to this repository.
- GALA 0.84.1 and the Go 1.26 series stay the pinned toolchains, and the vendored runtime under `server/third_party/gala/` is not modified.
- The committed twins and the corpus expectations are not re-baselined as a side effect of this work.
- Root Markdown follows the sentence-per-line and table-row rules in `docs/AGENTS.md`, and terminology follows `docs/terminology/README.md`.
- Discovery for a new page in `docs/` is by inbound link from `server/GALA.md` and `server/README.md`; the repository has no docs index.
- TASK-0323 remains the version-gated corpus run and is not duplicated here.
  It will run the tooling this initiative changes, so new checks must stay runnable on a Go 1.26 toolchain, and the audit script must remain free of a GALA and Go toolchain.

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

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-10-04 12:36
---
TASK-0330.05 was superseded by TASK-0341 and has been archived. The pilot it described was never run; its function is now the `gala-loop` skill, and the first executed loop iteration is the type-position import re-check recorded as TASK-0342. The carried facts live in TASK-0341 and `server/GALA.md`.
---

created: 2026-10-04 13:23
---
DoD annotation (TASK-0343, AC #6): TASK-0330 defines no acceptance criteria of its own; its four delivered subtasks carried their own DoD items, and subtask .05 was superseded by TASK-0341 and archived. #1: no task-level ACs to satisfy; .01-.04 are Done and .05's function moved to the gala-loop skill. #2: the inventory tool's unit tests were green at delivery under .01/.02; both test files and the tool itself were deleted by TASK-0341, so they cannot be re-run today. #3: no e2e tests existed for this documentation and tooling work. #4: the roster and playbook Markdown was formatted before TASK-0341 deleted those files; the current tree passes `npm run format:markdown:check` under TASK-0343. #5: no swagger annotations changed. #6: the code graph was refreshed by TASK-0341 after the tool was deleted; TASK-0343 refreshes it again. #7: no root `scripts/` or `prettier/` files changed; TASK-0343 runs root `npm run fmt:check` to confirm. #8: no environment, plugin-API, migration, or rollout documents changed. The boxes stay unchecked because the artifacts they would certify no longer exist; this comment is the annotation.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
TASK-0330 delivered the preflight, the mechanical rewrite catalog, the playbook, and the rewrite skill through subtasks .01-.04, all Done. Subtask .05 was superseded by TASK-0341, which replaced the playbook and the corpus with the `server/GALA.md` ledger, the `gala-loop` skill, and `server/scripts/gala/verify.sh`, and archived the subtask with its durable facts carried into TASK-0341. This task is finalized Done rather than archived so its four delivered subtasks remain recorded.
<!-- SECTION:FINAL_SUMMARY:END -->
