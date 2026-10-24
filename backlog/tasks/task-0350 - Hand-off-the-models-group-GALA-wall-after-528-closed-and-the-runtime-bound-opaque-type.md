---
id: TASK-0350
title: >-
  Hand off the models group GALA wall after #528 closed and the runtime-bound
  opaque type
status: Done
assignee: []
created_date: '2026-10-05 09:26'
updated_date: '2026-10-05 09:37'
labels: []
dependencies: []
references:
  - server/GALA.md
  - server/scripts/gala/probes/models-defined-scalar-methods/notes.md
  - server/scripts/gala/probes/models-struct-tag/notes.md
  - server/scripts/gala/probes/models-fixed-array/notes.md
  - server/scripts/gala/probes/models-empty-struct/notes.md
  - >-
    backlog/tasks/task-0349 -
    Advance-the-GALA-translation-loop-cursor-and-land-its-runtime-free-twins.md
documentation:
  - server/GALA.md
  - server/scripts/gala/probes/models-defined-scalar-methods/notes.md
  - .agents/skills/gala-from-go/references/gaps.md
  - .agents/skills/gala-from-go/references/constructs.md
modified_files:
  - server/GALA.md
  - server/scripts/gala/probes/models-defined-scalar-methods/notes.md
priority: medium
type: task
ordinal: 354005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Context

TASK-0349's iteration 4 reached the `models/` cursor entry on the pinned compiler (`GALA version 0.85.0`, flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`) and confirmed that the whole group stays handwritten. The re-check added four committed probes under `server/scripts/gala/probes/`:

- `models-defined-scalar-methods` — `models/datetime.go` / `models/uuid.go`
- `models-struct-tag` — `models/location.go` / `models/event.go`
- `models-fixed-array` — `models/uuid.go`
- `models-empty-struct` — `models/set.go`

## What changed upstream

`#528` (Triage language limitations) is **closed** upstream as of 2026-10-04; `server/GALA.md` still recorded it open, and its report-index row is now corrected to "closed 2026-10-04; tags, fixed-size arrays, and `struct{}` stay blocked (re-verified 2026-10-05)". No specific follow-up issue exists for the three constructs (searched 2026-10-05).

`#621` (No newtype for non-struct types) is closed, and its fix introduces `opaque type`. The re-check found the substitute is runtime-bound:

### Method on a defined scalar type

`type DateTime int64` plus a method is refused with `GALA-E0048`; the hint names `opaque type`. `opaque type DateTime int64` transpiles, but the emitted Go unconditionally imports `martianoff/gala/std` and synthesizes `Hash`/`Compare`:

```go
package probe

import "martianoff/gala/std"
import "time"

type DateTime int64

func (s DateTime) Hash() uint32 {
	return std.HashInt(int64(s))
}
func (s DateTime) Compare(other DateTime) int {
	return std.CompareInt(int64(s), int64(other))
}
```

**Classification:** documented answer per [`references/gaps.md`](.agents/skills/gala-from-go/references/gaps.md), not an upstream gap. Every committed twin in this repository is runtime-free (`server/GALA.md`), so the substitute is unavailable and `models/datetime.go` / `models/uuid.go` stay handwritten.

### Struct tags, fixed-size arrays, `struct{}`

Re-confirmed blocked on 0.85.0 with parse errors; `gaps.md` lists all three under "What Not To File", so no duplicate upstream report is proposed here.

## Decision needed

1. Comment on the closed `#528` / `#621` with the runtime-free `opaque type` evidence, asking whether `Hash`/`Compare` can be emitted only when used (which would reopen `models/datetime.go` and `models/uuid.go`), or
2. Record the wall as permanent in `server/GALA.md` and close this task with no upstream action.

Do not file a new gap report for the three constructs `gaps.md` marks answered.

## Handoff report template (for option 1)

**Construct:** a method on a defined scalar type whose generated Go must stay runtime-free.
**What I tried:** `type DateTime int64` + method (refused `GALA-E0048`); `opaque type DateTime int64` + method (transpiles, emits `martianoff/gala/std`).
**Diagnostic / emitted artifact:** `GALA-E0048` hint and the emitted import block above; full evidence in `server/scripts/gala/probes/models-defined-scalar-methods/notes.md`.
**Minimal repro:** the probe's `main.gala`.
**Compiler:** `GALA version 0.85.0`, extraction `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`.
**Searched:** `gala explain GALA-E0048`, `gala explain --list`, upstream tracker on 2026-10-05 (#528 closed, #621 closed, no follow-up on the construct).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The four probes under `server/scripts/gala/probes/models-*` are reviewed and their classifications confirmed against `.agents/skills/gala-from-go/references/gaps.md`.
- [x] #2 A decision is recorded: either comment upstream on #528/#621 with the runtime-free `opaque type` evidence, or record the wall as permanent in `server/GALA.md` with no upstream action.
- [x] #3 If an upstream comment is made, `server/GALA.md`'s report index and the method-on-a-defined-scalar row are updated with the outcome; if not, the row's probe path and classification stay as the recorded reason.
- [x] #4 No new gap report is filed for struct tags, fixed-size arrays, or `struct{}`, which `gaps.md` marks answered.
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

1. Re-run the probe on the pinned compiler (`GALA version 0.85.0`, flake rev `a888e824`, extraction `7a42d3c6`) and test the upstream suppression path: an `opaque type` plus a same-package `.go` sibling declaring `Hash()` and `Compare()`.
2. Correct `server/scripts/gala/probes/models-defined-scalar-methods/notes.md` to record both outputs: unsuppressed (synthesized std-backed methods) and suppressed (runtime-free, builds and vets), with the sibling listing, the exact invocation, and the date.
3. Correct `server/GALA.md`: the method-on-a-defined-scalar finding row, the `#621` report-index row, the prose below the findings table, and the cursor entry for `models/`.
4. Make no upstream comment and file no report; confirm the other three constructs stay under `gaps.md`'s "What Not To File".
5. Format changed Markdown with `npm run format:markdown` and finalize the task.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Resolution — no upstream action; suppression path recorded (2026-10-05)

**Decision:** no comment on #528/#621 and no new report. A method on a defined scalar is a documented answer (`GALA-E0048` names `opaque type`) and `gaps.md` lists it under "What Not To File". The probe's absolute claim that the substitute "cannot enter a runtime-free twin" was incomplete and is corrected.

**Suppression evidence** (pinned compiler: `GALA version 0.85.0`, flake rev `a888e824ff53adb653bbd48dcc83f67788eeced4`, extraction `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`): upstream PR #665 (the #621 fix) synthesizes `Hash`/`Compare`, each skipped when GALA or a same-package `.go` file already declares it. Re-verified with the probe's `opaque type DateTime int64` plus a same-package `.go` sibling declaring `Hash() uint32` and `Compare(DateTime) int`: the emitted Go names no `martianoff/gala` package, emits only `Time`/`IsZero` plus the sibling methods, and the generated file plus sibling pass `go build ./...` and `go vet ./...` in a scratch module (`/tmp/opencode`, not committed). The same suppression holds for `opaque type UUID string`.

**Cost and repo decision:** the runtime-free path adds two exported methods (`Hash`, `Compare`) the handwritten Go API did not carry, so the wall was not declared permanent and no upstream action was taken; whether to pay that cost is a repo-local decision for the TASK-0349 models iteration. The other three probes were reviewed and stay non-reportable under `gaps.md`: struct tags and fixed-size arrays are `answered` rows, `struct{}`/anonymous structs are `workaround` rows, and #528 already carried their triage.

**Files changed:** `server/GALA.md` (finding row, #621 report-index row, prose below the findings table, cursor entry 1) and `server/scripts/gala/probes/models-defined-scalar-methods/notes.md` (default output, suppression path with the sibling listing, corrected classification).

**Checks:** `npm run format:markdown` reformatted `server/GALA.md`; the change is documentation-only, so unit/e2e tests, Swagger, `fmt:check`, and the code knowledge-graph refresh are not applicable. No generated twin, test, signature, tag, or wire format changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
TASK-0350 resolved with no upstream action: the `opaque type` substitute for a method on a defined scalar can stay runtime-free when a same-package `.go` sibling declares `Hash`/`Compare`, which upstream PR #665 documents, so no comment on #528/#621 and no new report is warranted. Corrected `server/GALA.md` (method-on-a-defined-scalar row, #621 report-index row, prose below the findings table, and cursor entry 1) and `server/scripts/gala/probes/models-defined-scalar-methods/notes.md` with the verified suppression output and the two-added-methods cost. Struct tags, fixed-size arrays, and `struct{}` remain non-reportable under `gaps.md`'s "What Not To File". Documentation-only; Markdown formatted with `npm run format:markdown`.
<!-- SECTION:FINAL_SUMMARY:END -->
