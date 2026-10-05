---
id: TASK-0350
title: >-
  Hand off the models group GALA wall after #528 closed and the runtime-bound
  opaque type
status: To Do
assignee: []
created_date: '2026-10-05 09:26'
updated_date: '2026-10-05 09:26'
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
- [ ] #1 The four probes under `server/scripts/gala/probes/models-*` are reviewed and their classifications confirmed against `.agents/skills/gala-from-go/references/gaps.md`.
- [ ] #2 A decision is recorded: either comment upstream on #528/#621 with the runtime-free `opaque type` evidence, or record the wall as permanent in `server/GALA.md` with no upstream action.
- [ ] #3 If an upstream comment is made, `server/GALA.md`'s report index and the method-on-a-defined-scalar row are updated with the outcome; if not, the row's probe path and classification stay as the recorded reason.
- [ ] #4 No new gap report is filed for struct tags, fixed-size arrays, or `struct{}`, which `gaps.md` marks answered.
<!-- AC:END -->

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
