---
id: TASK-0340
title: >-
  Pin the GALA type-position import hole, where a bare GALA type name transpiles
  and only fails go build
status: To Do
assignee:
  - Danila Danko
created_date: '2026-09-30 21:19'
updated_date: '2026-10-02 19:55'
labels: []
dependencies: []
references:
  - >-
    backlog/tasks/task-0338 -
    Re-derive-the-import-resolution-rows-that-GALA-0.84.1-contradicts.md
  - 'https://github.com/martianoff/gala/issues/648'
documentation:
  - .agents/skills/gala-from-go/references/constructs.md
  - docs/gala-translation.md
  - server/GALA.md
  - server/scripts/20260923_gala_translation_probes/run.sh
priority: medium
type: bug
ordinal: 343300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

TASK-0338 re-derived the import-resolution rows and found that `gala` 0.84.1 refuses a name in **value position** that the file does not import, with `GALA-E0023` and, for a GALA-runtime name, a hint naming the declaring package.
The same check does not cover **type position**, so the per-file rule has a hole that produces the exact failure the skill documents as impossible: a clean transpile followed by a `go build` error blamed on the `.gala` file through its `//line` directive.

This was found while TASK-0338 reproduced the value-position rule and is recorded there as out of its scope.

## What was reproduced, on 0.84.1 / go1.26.7

Each of these transpiles cleanly, with no import of `martianoff/gala/concurrent` anywhere in the package:

```gala
struct Holder(F Future[int])
```

```gala
var holder Future[int]
```

```gala
func Await2(f Future[int]) int = 1
```

The emitted Go carries the bare name through unchanged, so `go build` is where it fails:

```go
//line c.gala:3
func Await2(f Future[int]) int {
	return 1
}
```

The same three shapes with `Array[string]` from `martianoff/gala/collection_immutable` behave identically.
The struct case is the loudest, because a non-`var` field is wrapped and the wrapper is emitted over the unresolved name:

```go
type Holder struct {
	F std.Immutable[Future[int]]
}
```

Note what is *not* wrong here: the wrapper, the `Copy`/`Equal`/`Unapply` synthesis, and the private discriminant are all correct. Only the unresolved type name is carried.

`gala explain GALA-E0023` documents the gap in its own text: the check covers identifiers in value position, and in type position it checks only the package *qualifier*, leaving an unqualified type name to the scope rule instead.
`gala explain GALA-E0025` is the check that is supposed to cover this case, and it did not fire in any shape measured, including the two-file repro its own page documents.

## Why it matters beyond curiosity

`.agents/skills/gala-from-go/SKILL.md` and `docs/gala-translation.md` now both state that a missing import is refused at transpile time, with no branch where it survives into the generated Go.
That is true for value positions and false for type positions, so the corrected documents are correct as far as they go and silent about the case that still bites.
The runtime-enabled twins are the ones exposed to it, because a `concurrent.Future`, `collection_immutable.Array`, or `resource` type in a signature is exactly the shape that needs an import.

## Constraints

- `gala` 0.84.1 and the go1.26 series stay the pinned toolchains.
- `server/third_party/gala/` is not modified.
- The documents corrected by TASK-0338 must not be reverted to the pre-0.84.1 claim in the course of recording this.
- Deciding whether to file upstream is part of this task, because the reproduction is now a minimal one and the existing reports (#613 through #621) do not cover a type-position import hole.
- The change follows `docs/AGENTS.md`: one sentence per line, and the Prettier sentences-per-line pipeline rather than oxfmt.

## Where to look

The three shapes above, `gala explain GALA-E0023` ("Not covered" and "Scope") and `gala explain GALA-E0025`, the probe corpus under `server/scripts/20260923_gala_translation_probes/`, and the import rows in `.agents/skills/gala-from-go/references/constructs.md` that TASK-0338 left scoped to value positions.

## Relationship to TASK-0338

TASK-0338 owns the value-position rule and its three probes: `blocked_import_omitted_gala_sibling`, `blocked_import_omitted_go_sibling`, and `blocked_runtime_name_omitted_import`.
This task owns the type-position hole and must not re-litigate the value-position verdict.
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Filed upstream as #648 on 2026-09-30: "An unqualified GALA type name in a type position is not import-checked: the transpile succeeds and `go build` reports `undefined`". The report states the build provenance honestly (`gala version` reports 0.84.1 but the binary was built from master HEAD, so it is not the release tag), and it contrasts the three cases that already work so the maintainer can see the hole is narrow rather than "imports are unenforced".

No duplicate exists: #613 through #620 are all closed, and the only adjacent issue, #616, is the opposite direction (the import is present and an imported qualifier wins over a package-local declaration). #528 and #621 are the only open issues and neither covers import resolution.
<!-- SECTION:NOTES:END -->
