---
id: TASK-0343
title: Correct the GALA ledger and loop issues found in the TASK-0341 review
status: Done
assignee:
  - '@opencode'
created_date: '2026-10-04 12:56'
updated_date: '2026-10-04 13:41'
labels:
  - gala
  - documentation
  - tooling
dependencies: []
references:
  - >-
    backlog/tasks/task-0341 -
    Consolidate-the-GALA-document-stack-into-a-ledger-and-a-translate-verify-loop.md
  - server/GALA.md
  - server/README.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/references/gaps.md
  - server/scripts/gala/verify.sh
  - server/third_party/gala/VENDORED_FROM
  - server/scripts/gala/probes/type-position-import/
documentation:
  - server/README.md
  - server/scripts/gala/probes/type-position-import/notes.md
  - BACKLOG_WORKFLOW.md
modified_files:
  - server/GALA.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/references/gaps.md
  - server/scripts/gala/verify.sh
  - server/third_party/gala/VENDORED_FROM
priority: medium
type: task
ordinal: 347005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The TASK-0341 consolidation is staged and its core claims verify, but a review found two stale or under-specified instructions in the new documents plus small robustness and record-keeping gaps. The re-vendor procedure is load-bearing: the loop trusts it whenever the flake rev moves, the compiler version string is pinned at 0.84.1, and TASK-0341's own motivating incident was a vendored runtime copied from an extraction that did not match the compiler. Left as-is, the next sync can silently re-vendor a stale runtime, and the next translation iteration is told to add a README row that no longer exists. The review evidence: `server/scripts/gala/verify.sh` passes with 15 twins, the type-position-import probe reproduces exactly, the vendored tree is byte-identical to `~/.gala/stdlib/v0.84.1` apart from `VENDORED_FROM` and the flattened module files, no deleted-artifact references survive outside `backlog/`, and both format checks pass. Problems to fix: both `.agents/skills/gala-loop/SKILL.md` step 4 and `server/GALA.md` instruct adding a "pointer row" to `server/README.md` even though the twin table was deliberately dropped, so the target does not exist; the re-vendor steps copy `~/.gala/stdlib/v<version>` without forcing or verifying that the extraction was produced by the pinned compiler, even though the extraction carries a `.stdlib-extracted` marker with a version and hash and `verify.sh` never checks the recorded `compiler:` line; `verify.sh` still exits 0 when `VENDORED_FROM` is missing; no check catches a committed generated twin whose `.gala` source was deleted; the `gaps.md` report template cannot express the `defect` classification that TASK-0342 uses; the #648 workaround wording in `server/GALA.md` is unclear; TASK-0330 is finalized Done with all DoD boxes unchecked; the archived superseded tasks carry no supersession note; and the `gala-rewrite-consolidation` milestone has no milestone file. Constraints: do not change the translation approach, keep `verify.sh` green on the current tree, follow the `docs/AGENTS.md` sentence-per-line and table-row rules for repo Markdown and the same style by hand under `.agents/`, and leave TASK-0341 and TASK-0342 as their own records. Backlog policy keeps Done tasks in the Done state; do not move them to the completed folder.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The loop registration step in `.agents/skills/gala-loop/SKILL.md` and the matching sentence in `server/GALA.md` describe only targets that exist: registering a twin adds its ledger registry row and passes `verify.sh`, with no instruction to add a row to `server/README.md`
- [x] #2 `server/GALA.md`'s re-vendor procedure states how the extraction at `~/.gala/stdlib/v<version>` is forced or verified to come from the pinned compiler before copying, records the extraction's `.stdlib-extracted` provenance in `server/third_party/gala/VENDORED_FROM`, and does not rely on the version string alone
- [x] #3 `server/scripts/gala/verify.sh` exits non-zero when `server/third_party/gala/VENDORED_FROM` is missing, and when `gala version` disagrees with the `compiler:` line it records
- [x] #4 `server/scripts/gala/verify.sh` detects a committed generated twin whose `.gala` source is gone, or the task records why that check is deliberately out of scope
- [x] #5 The `gaps.md` report template accepts a `defect` classification alongside `language gap` and `boundary gap`, and the #648 workaround wording in `server/GALA.md` no longer reads as unclear
- [x] #6 TASK-0330's applicable DoD items are checked or annotated, and every archived superseded task (TASK-0339 and its five subtasks, TASK-0323, TASK-0340) carries a supersession record or its absence is explained in TASK-0341's carried-facts section
- [x] #7 The `gala-rewrite-consolidation` milestone either has a milestone file or the value is removed from every task, and the decision is recorded
- [x] #8 Evidence is recorded in the task: `verify.sh` passes on the current tree with the new provenance checks exercised, and the Markdown and root format checks pass
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
Plan (2026-10-04), confirmed against the current tree.

1. `.agents/skills/gala-loop/SKILL.md` step 4 and `server/GALA.md`'s twin-registration sentence: drop the `server/README.md` pointer-row instruction; registration is the ledger registry row plus `verify.sh`.
2. `server/GALA.md` re-vendor procedure: add the extraction-provenance steps. Enter the dev shell so `gala` is the flake-locked compiler; delete `~/.gala/stdlib/v<version>/.stdlib-extracted` and run a transpile so the pinned compiler rewrites the marker; read the marker (version + embedded-snapshot fingerprint, from `snapshotFingerprint` in upstream `internal/build/stdlibver.go`); record it in `VENDORED_FROM`; then copy.
3. `server/third_party/gala/VENDORED_FROM`: add the `extraction:` line with the current marker contents.
4. `server/scripts/gala/verify.sh`: fail when `VENDORED_FROM` is missing; fail when `gala version` disagrees with `compiler:`; fail when the recorded extraction fingerprint disagrees with the on-disk marker after the transpile walk; detect committed generated `.go` twins (header `// Code generated by GALA transpiler. DO NOT EDIT.`) whose `.go`-sibling `.gala` source is gone.
5. `.agents/skills/gala-from-go/references/gaps.md`: template classification line accepts `defect`; clarify `server/GALA.md` #648 workaround cell.
6. Backlog bookkeeping: check/annotate TASK-0330's applicable DoD items; add a supersession comment to TASK-0339 and .01-.05, TASK-0323, TASK-0340 if the MCP can reach archived tasks, else explain the absence in TASK-0341's carried-facts section; create the `gala-rewrite-consolidation` milestone file and record the decision.
7. Verification: run `server/scripts/gala/verify.sh` (expect OK), exercise each new failure path by temporary mutation and restore, run `npm run format:markdown` and root `npm run fmt:check`, refresh the code graph, record evidence, finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation notes (2026-10-04)

Executed the plan with one recorded deviation (the milestone decision below).

### AC #1 - registration targets

- `.agents/skills/gala-loop/SKILL.md` step 4 now reads: "Register the twin: add its row to the twin registry in `server/GALA.md` and remove it from the cursor."
- `server/GALA.md` now reads: "Adding a twin means adding its registry row here and passing `verify.sh`."
- The `server/README.md` pointer-row instruction is gone from both; no "pointer row" text remains in either file.

### AC #2 - re-vendor provenance

- `server/GALA.md`'s procedure now has nine steps: enter `nix develop`; delete `~/.gala/stdlib/v<version>/.stdlib-extracted` and run one transpile so the pinned compiler rewrites the extraction; read and record the marker; copy; flatten; build; record provenance; verify.
- The marker is `<version> <snapshot fingerprint>` (upstream `internal/build/stdlibver.go` writes `snapshotFingerprint`, the normalized version plus a hash of the embedded stdlib). The fingerprint changes when the embedded snapshot changes, so the version string alone no longer proves provenance.
- Forced-refresh evidence: deleting the marker and transpiling `eventid.gala` (exit 0, 0.53s) rewrote the marker byte-identically as `0.84.1 4df6252b73c34f06e045502457f9d486ac2fb3078330280038e096967c4bebf2`, now recorded in `VENDORED_FROM` as `extraction: ...`.
- `diff -rq ~/.gala/stdlib/v0.84.1 server/third_party/gala` after the refresh showed only the expected flattening differences: per-package `go.mod`, `.stdlib-extracted`, `LICENSE`, `VENDORED_FROM`, and the rewritten root `go.mod`.

### AC #3 and #4 - verify.sh

- New failures: missing `VENDORED_FROM`; the `compiler:` line disagreeing with `gala version`; the `extraction:` fingerprint disagreeing with the on-disk marker after the transpile walk; a committed `.go` file whose first line is exactly `// Code generated by GALA transpiler. DO NOT EDIT.` and whose sibling `.gala` is missing.
- Bug Fix Protocol evidence. Fail-before with the pre-TASK-0343 script (reconstructed to `/tmp/opencode/verify-old.sh` from the version present at session start):
  - `VENDORED_FROM` moved aside plus an orphan `server/eventid/orphan_probe.go`: old exit 0, printed `vendored runtime: MISSING` and `verify.sh: OK (15 twins)`.
  - `compiler: GALA version 0.0.0`: old exit 0, `verify.sh: OK (15 twins)`.
- Pass-after with the new script on the same mutating conditions, each exit 1:
  - `PROVENANCE: .../VENDORED_FROM is missing; run a sync iteration and record it`
  - `PROVENANCE: vendored runtime records compiler 'GALA version 0.0.0' but gala reports 'GALA version 0.84.1'; run a sync iteration`
  - `PROVENANCE: vendored runtime records extraction '0.84.1 deadbeef' but ~/.gala/stdlib/v0.84.1/.stdlib-extracted is '0.84.1 4df6252b...'; run a sync iteration`
  - `ORPHAN: eventid/orphan_probe.go carries the generated header but eventid/orphan_probe.gala is missing`
- Restored tree: `server/scripts/gala/verify.sh` exit 0, `verify.sh: OK (15 twins)`.

### AC #5 - classification and wording

- `references/gaps.md` template now reads `**Classification:** language gap | boundary gap | defect`.
- The #648 row's workaround cell now reads "declare the type's package import in the file; the transpiler does not check or diagnose it".

### AC #6 - TASK-0330 and archived tasks

- TASK-0330's DoD boxes stay unchecked and carry an annotation comment naming each item's disposition; the artifacts the boxes would certify were deleted by TASK-0341, and the task defines no acceptance criteria of its own.
- Backlog MCP views archived tasks but refuses edits (`Task not found`), and the Backlog CLI behaves the same (`Task lookups read only the local working copy`). The acceptance criterion's alternative was taken: TASK-0341's implementation notes now carry an "Archived-task supersession records (TASK-0343, AC #6)" section naming TASK-0339, TASK-0339.01-.05, TASK-0323, and TASK-0340 as superseded by TASK-0341 and pointing at the carried-facts section.

### AC #7 - milestone

- Decision (per the user, 2026-10-04): no milestone. The active `gala-rewrite-consolidation` value was cleared from TASK-0341, TASK-0342, TASK-0343, and TASK-0344 through the MCP; `backlog_milestone_list` now reports no file and no task value.
- The six archived TASK-0339 files still carried the label after removal. `backlog milestone remove --task-handling clear` clears `0 local tasks` and cannot reach the archive, and the MCP cannot edit archived tasks. Removing one `milestone:` line from each of exactly those six archived files is a documented, minimal exception to the no-direct-edit rule, required by AC #7. `git diff` for the archive shows exactly six deleted lines and nothing else. A backup of the six files is at `/tmp/opencode/archived-milestone-backup/`.

### AC #8 - formatting and verification evidence

- `npm run format:markdown` rewrote `server/GALA.md` (table alignment and the new prose); `npm run format:markdown:check` exit 0; root `npm run fmt:check` printed "All matched files use the correct format."
- `codebase-memory-mcp cli index_repository --repo-path .`: status `indexed`, 11321 nodes, 46322 edges.
- `verify.sh` final run: `verify.sh: OK (15 twins)`.
- No runtime code changed; swagger and contract documents are unaffected. Unit and e2e tests are not required for this documentation and tooling change; `verify.sh` is the executable check and it passed.

### Deviation from plan

- Plan item 6 said to create the `gala-rewrite-consolidation` milestone file; the user decided no milestone is needed, so the removal branch of AC #7 was taken instead. The milestone artifact produced by the add/remove test was moved out of the repository to `/tmp/opencode/milestone-cleanup/`; no milestone file is tracked or untracked anywhere in `backlog/`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Corrected the TASK-0341 review findings in the GALA ledger and loop. Twin registration now points only at the ledger registry and `verify.sh`; the re-vendor procedure forces the pinned compiler to rewrite its `.stdlib-extracted` marker and records its fingerprint in `VENDORED_FROM`; `verify.sh` fails on a missing or mismatched provenance record and on a generated twin whose `.gala` source is gone; the gaps report template accepts `defect`; and the #648 workaround wording is clear. TASK-0330's DoD is annotated, the archived superseded tasks are recorded in TASK-0341's notes because the MCP cannot edit archived tasks, and the `gala-rewrite-consolidation` milestone was removed from every task per the user's decision. Evidence: the pre-change script exited 0 on the reported defects (missing `VENDORED_FROM`, compiler mismatch, orphan twin) and the new one fails on the injected conditions while the restored tree passes (`verify.sh: OK (15 twins)`); Markdown and root format checks pass; the code graph was refreshed (11321 nodes, 46322 edges).
<!-- SECTION:FINAL_SUMMARY:END -->
