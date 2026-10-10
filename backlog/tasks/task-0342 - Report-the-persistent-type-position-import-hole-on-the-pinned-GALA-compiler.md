---
id: TASK-0342
title: Report the persistent type-position import hole on the pinned GALA compiler
status: Done
assignee: []
created_date: '2026-10-04 12:34'
updated_date: '2026-10-04 13:29'
labels:
  - gala
  - tooling
dependencies: []
references:
  - server/scripts/gala/probes/type-position-import/
  - 'https://github.com/martianoff/gala/issues/648'
  - server/GALA.md
documentation:
  - server/scripts/gala/probes/type-position-import/notes.md
  - server/GALA.md
priority: medium
type: bug
ordinal: 346005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Origin

The first GALA loop iteration of TASK-0341 re-checked the type-position import hole on the synced compiler and it persists.
Upstream fixed it in [#648](https://github.com/martianoff/gala/issues/648) (closed 2026-10-03), but the pinned flake rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f` (locked 2026-10-02) predates the fix, so the compiler on `PATH` still has the hole.

The loop does not file or comment on upstream issues itself; this task is the handoff record.

## Filled report

### A bare GALA runtime type name in a type position is not import-checked

A GALA type name in a *type position* that the file never imports transpiles cleanly, and `go build` then reports `undefined` against the generated Go. Value positions are already refused at transpile time with `GALA-E0023`.

**Classification:** defect, not a gap; the construct is accepted and lowered wrongly.

**What I tried**

`main.gala` names `Future[int]` in a struct field, a package variable, and a parameter without importing `martianoff/gala/concurrent`.
The same two shapes with `Array[string]` from `martianoff/gala/collection_immutable` behave identically.

**Diagnostic**

Transpile: none, exit `0`.
`go build`: exit `1` with `main.gala:4: undefined: Future`, `main.gala:5: undefined: Future`, `main.gala:7: undefined: Future`, `main.gala:13: undefined: Future`, and `main.gala:20: undefined: Future`.
The generated Go carries `Future[int]` in `std.Immutable[Future[int]]`, in `var holder Future[int]`, and in the parameter, and imports only `martianoff/gala/std`.

**Minimal repro**

`server/scripts/gala/probes/type-position-import/`, with the exact invocation in its `notes.md`.
The probe `main.gala` is the three shapes above, and the build failure is reproduced in a scratch module that replaces `martianoff/gala` with `server/third_party/gala`.

**What a caller would need**

A transpile-time refusal in the shape the value-position check already produces with `GALA-E0023`, or a diagnostic naming the package that declares the type and the import the file is missing.
The workaround is to declare the import in every file that names the type, which is exactly what the missing check would enforce; the hole is that the transpiler does not ask.

**Compiler**

`GALA version 0.84.1` (the flake-locked commit's version string), rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f`, checked 2026-10-04.

**Environment**

go1.26.7; the generated program builds and runs once the import is added.

**Searched**

`gala explain GALA-E0023` documents that the check leaves an unqualified type name to the scope rule; `gala explain GALA-E0025` is the check that should cover it and does not fire.
Upstream [#648](https://github.com/martianoff/gala/issues/648) was filed 2026-09-30 and closed 2026-10-03; #613-#621 are closed; the only open issue is #528 (triage).

## Resolution

The fix is a flake bump to a rev that includes the upstream change.
After the bump, run the sync iteration, re-run the probe, and if it stops reproducing, retire the finding in `server/GALA.md` and this task.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The filled report and the committed probe at `server/scripts/gala/probes/type-position-import/` are recorded
- [x] #2 The open-findings row in `server/GALA.md` names this report, the probe, and upstream #648
- [x] #3 The flake-bump follow-up task (TASK-0344) is filed to carry the retirement of the probe and finding on a rev that includes the upstream fix
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

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-10-04 12:57
---
Review decision (2026-10-04): closed as the recorded handoff for the type-position import hole. Acceptance criteria #1 and #2 are satisfied by the artifacts TASK-0341 staged: the report in this description, the committed probe at `server/scripts/gala/probes/type-position-import/`, and the `server/GALA.md` finding row that names TASK-0342 and links the probe and upstream #648. The row names this task in prose rather than as a hyperlink; TASK-0344 owns repointing it when the finding retires. The flake-bump follow-up is TASK-0344, which depends on TASK-0343 for the corrected re-vendor procedure. Per `BACKLOG_WORKFLOW.md`, Done tasks stay in the Done state and are not moved to the completed folder.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Closed as the recorded handoff for the persistent type-position import hole (#648). The filled report and the committed minimal repro at `server/scripts/gala/probes/type-position-import/` are recorded; the probe reproduces on the pinned compiler (`GALA version 0.84.1`, rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f`) with a clean transpile and `main.gala:4,5,7,13,20: undefined: Future` from `go build`. `server/GALA.md`'s open-findings row names this task and links the probe and upstream #648. The flake bump and sync iteration that retire the probe and finding are filed as TASK-0344, which depends on TASK-0343 for the corrected re-vendor procedure. No new code or documents changed under this task; the artifacts were delivered by TASK-0341. Limitation: the ledger row names TASK-0342 in prose rather than linking it, and TASK-0344 owns repointing the row when the finding retires.
<!-- SECTION:FINAL_SUMMARY:END -->
