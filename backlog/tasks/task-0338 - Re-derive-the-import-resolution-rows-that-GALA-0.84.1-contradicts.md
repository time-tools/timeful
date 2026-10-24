---
id: TASK-0338
title: Re-derive the import-resolution rows that GALA 0.84.1 contradicts
status: Done
assignee:
  - Danila Danko
created_date: '2026-09-30 17:16'
updated_date: '2026-09-30 21:20'
labels: []
dependencies: []
references:
  - >-
    backlog/tasks/task-0337.01 -
    Repair-the-local-GALA-documents-left-inconsistent-by-the-TASK-0337-filing-pass.md
  - 'https://github.com/martianoff/gala/blob/master/docs/GALA.MD'
documentation:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
  - docs/gala-translation.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - server/scripts/20260923_gala_translation_probes/audit.sh
modified_files:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - docs/gala-translation.md
  - server/GALA.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_gala_sibling/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_gala_sibling/sibling.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_gala_sibling/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_go_sibling/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_go_sibling/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_runtime_name_omitted_import/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_runtime_name_omitted_import/sibling.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_runtime_name_omitted_import/expect
priority: medium
type: docs
ordinal: 343200
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Two documents under `.agents/skills/gala-from-go/` claim a GALA failure mode that `gala` 0.84.1 does not have, and the claim is the kind that costs an afternoon if a reader believes it.

The claim is that a qualified name resolves against the imports of *every* file in the package, so a `.gala` file that omits an import its sibling already has transpiles cleanly and then fails `go build` with `undefined: <pkg>`, blamed on the `.gala` file through its `//line` directive.
On 0.84.1 the transpile is refused instead, with a hint that names the package which declares the name, whether the sibling that imports it is a `.gala` file or a handwritten `.go` file.

This was found while TASK-0337.01 was writing a verified import bullet for `server/GALA.md`, and it is recorded there rather than fixed, because it predates TASK-0337 and is outside that task's criteria.

## What was reproduced, on 0.84.1 / go1.26.7

A package whose sibling imports `strings`, and a second file that calls `strings.ToLower` without importing it:

```
error[GALA-E0023]: undefined: strings
  --> b.gala:3:31
    = hint: check the spelling, add the import that introduces this name, or
      declare it — every identifier must resolve to a binding, a declaration in
      this package, or an imported symbol
```

The same diagnostic appears when the importing sibling is a handwritten `helper.go` instead of a `.gala` file, so the sibling's language is not what decides it.

An unqualified GALA-runtime name is the other shape: `Future(2)` in a file that does not import `martianoff/gala/concurrent` is refused with `GALA-E0023` and a hint that names that package and suggests `import . "martianoff/gala/concurrent"`.
`gala explain GALA-E0025` documents the same per-file rule for the unqualified case, gives a two-file repro, and says outright that sibling files' imports do not propagate.
`server/GALA.md` already quotes an `GALA-E0025` diagnostic from a concurrency probe, so the two documents disagree about the same rule.

## What is wrong where

- `SKILL.md`, the mixed-package bullet: "A qualified name resolves against the imports of *any* file in the package ... When no file in the package imports that package at all, the transpile is refused with `GALA-E0023` instead, so the diagnostic itself tells you which of the two cases you are in." The second branch is the only one that exists on 0.84.1.
- `SKILL.md`, the trap "A sibling file's import hides a missing import from the transpiler": the same claim, stated as a trap, so a reader is told to watch for a build failure the transpiler prevents.
- `constructs.md`, the "Packages And Files" row "an import the file does not declare": `Status` `direct`, `Code` `—`, and a `Check` that reads "omit it from one file of a pair, read the generated import block, then build".
- `docs/gala-translation.md`, the roster's `named-imports` row: `—` and "Works", with no per-file caveat; TASK-0337.01 added the caveat to `server/GALA.md` and named it as the limit that row does not carry, so the two documents do not agree until this is done.

## Constraints

- `gala` 0.84.1 and the go1.26 series stay the pinned toolchains, and no probe expectation is re-baselined beyond what this change moves.
- `server/third_party/gala/` is not modified.
- Filing is out of scope: a documentation defect about the transpiler's own diagnostic is not a new issue, and no existing issue is touched.
- The change follows `docs/AGENTS.md`: one sentence per line, and the Prettier sentences-per-line pipeline rather than oxfmt.
- The new corpus probe and every document edit must leave `run.sh` and `audit.sh` green, including `audit.sh`'s seventh check, which compares every Markdown table row in the five GALA documents against its own header.

## Where to look

The three rows above, the probe corpus under `server/scripts/20260923_gala_translation_probes/`, and `gala explain GALA-E0023` and `gala explain GALA-E0025`, which between them state the rule the documents contradict.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `SKILL.md` states what 0.84.1 does when a `.gala` file omits an import a sibling has, and no sentence claims a qualified name resolves against another file's imports
- [x] #2 `constructs.md`'s two import rows carry the Status and Code the pinned compiler produces, and each row's Check re-derives that behaviour rather than reading the generated import block
- [x] #3 A corpus probe pins the refusal, so neither claim can drift back without a failing expectation
- [x] #4 `server/GALA.md` and `docs/gala-translation.md` say the same thing about import resolution, including the per-file check the roster's `named-imports` row does not carry
- [x] #5 `run.sh`, `audit.sh` with all seven checks, `go run ./inventory -check`, the inventory unit tests, `npm run format:markdown:check`, `npm run test:markdown-rules` and root `npm run fmt:check` all pass
- [x] #6 No upstream issue is filed or edited, and `server/third_party/gala/` is unchanged
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Research done on the pinned toolchains (gala 0.84.1, go1.26.7)

Every shape below was transpiled in a scratch copy of the probe module, with and without a sibling that imports the package.

| Case | Result |
| --- | --- |
| `strings.ToLower`, no `strings` import, sibling `.gala` imports `strings` | refused, `GALA-E0023: undefined: strings`, generic hint |
| `strings.ToLower`, no `strings` import, sibling handwritten `.go` imports `strings` | identical refusal |
| `strings.ToLower`, nothing in the package imports `strings` | identical refusal |
| `go_interop.StringLength`, sibling `.gala` imports `go_interop` | identical refusal |
| `Future(2)`, sibling `.gala` dot-imports `concurrent` | refused, `GALA-E0023: undefined: Future`, hint names `martianoff/gala/concurrent` |
| `Future(2)`, sibling handwritten `.go` dot-imports `concurrent` | identical refusal |
| `SliceOf(1)`, sibling `.gala` dot-imports `go_interop` | identical refusal, hint names `martianoff/gala/go_interop` |

So the sibling's language never decides it, and there is no second branch: a missing import is always refused at transpile time.

Two further facts the documents must reflect:

1. `GALA-E0025` did not fire in any shape. `gala explain GALA-E0025` still states the per-file rule and says outright that sibling files' imports do not propagate, so it is the authority for the rule but not for the code the pinned compiler emits. `server/GALA.md` currently quotes it as the observed diagnostic; the user approved correcting that quote to `GALA-E0023`.
2. The per-file check covers value positions only. A bare GALA type name in a type position with no import (`struct Holder(F Future[int])`, `var holder Future[int]`, a parameter type) transpiles cleanly and emits a bare `Future[int]`, which fails `go build`. That is a separate defect, out of this task's criteria, and the user approved a follow-up task for it.

## Planned changes

1. `server/scripts/20260923_gala_translation_probes/probes/`: add `blocked_import_omitted_gala_sibling` (a sibling `.gala` file imports `strings`) and `blocked_import_omitted_go_sibling` (the runner writes a handwritten `helper.go` that imports `strings`), both `KIND=transpile_fail` on `CODE=GALA-E0023`. Add a `materialize_fixtures` case for the handwritten one, and a line in `docs/gala-translation.md`'s probe-group table so `audit.sh` check 2 sees no orphan.
2. `.agents/skills/gala-from-go/SKILL.md`: rewrite the mixed-package bullet to state the refusal, delete the "resolves against the imports of any file" clause, and rewrite the "A sibling file's import hides a missing import from the transpiler" trap as a refusal rather than a hidden build failure. Fix the closing `Never` bullet that repeats the claim.
3. `.agents/skills/gala-from-go/references/constructs.md`: give the two import rows the Status and Code the pinned compiler produces, and replace the generated-import-block Check with one that re-derives the refusal.
4. `docs/gala-translation.md`: carry the per-file rule on the roster's `named-imports` row, and add a short "Imports are checked per file" subsection under Workarounds naming both probes.
5. `server/GALA.md`: correct the stale `GALA-E0025` quote to the `GALA-E0023` diagnostic the probes pin, and name the probes.
6. Run `run.sh`, `audit.sh`, `go run ./inventory -check`, the inventory unit tests, `npm run format:markdown`, `npm run test:markdown-rules` and root `npm run fmt:check`.
7. Confirm no upstream issue is touched and `server/third_party/gala/` is unchanged.

## Prettier caution

`docs/gala-translation.md`'s tables are padded to a very wide column. Any cell edit must be made through `npm run format:markdown` so the padding is regenerated rather than hand-fitted.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reproduction confirmed on the pinned toolchains: the transpile is refused with `GALA-E0023` in every shape, whether the importing sibling is a `.gala` file or a handwritten `.go` file, and whether or not any file in the package imports the package. There is no second branch on 0.84.1.

`GALA-E0025` did not fire in any shape tested, including the two-file repro its own page documents. The page still states the per-file rule, so it is cited as the authority for the rule rather than for the emitted code.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

Two documents under `.agents/skills/gala-from-go/` claimed a GALA failure mode that `gala` 0.84.1 does not have: that a qualified name resolves against the imports of *any* file in the package, so a `.gala` file that omits an import its sibling has transpiles cleanly and then fails `go build`. On 0.84.1 the transpile is refused instead. Both claims are gone, the correct rule is stated in all five GALA documents, and the refusal is pinned by the corpus so it cannot drift back.

### The rule, as re-derived on 0.84.1 / go1.26.7

Import resolution is per file, and there is no second branch. A qualified name the file does not import is refused at transpile time with `GALA-E0023: undefined: <pkg>`, and the diagnostic is identical across all three sibling layouts: a `.gala` sibling that imports it, a handwritten `.go` sibling that imports it, and no sibling at all. The sibling's language is not what decides it, and neither is whether any file in the package imports the package.

An unqualified GALA-runtime name is the same rule from the other side, and its hint is the more useful diagnostic: `Future(2)` without importing `martianoff/gala/concurrent` is refused with `GALA-E0023: undefined: Future` and `Future is declared in the GALA package "martianoff/gala/concurrent", which this file does not import`, which names the declaring package and gives both spellings of the fix.

`GALA-E0025` did not fire in any shape measured, including the two-file repro its own page documents. Its page is still the clearest statement of the rule and says outright that sibling files' imports do not propagate, so both `server/GALA.md` and `docs/gala-translation.md` now cite it as the authority for the rule rather than for the code the pinned compiler emits. `server/GALA.md` previously quoted it as an observed diagnostic, which was stale; per the user's decision that quote is corrected to `GALA-E0023`.

### Corpus probes (AC #3)

Three new probes, all `KIND=transpile_fail` on `CODE=GALA-E0023`, so both the false claim and its "no file in the package imports it" branch now fail if a future release restores the old behaviour:

- `blocked_import_omitted_gala_sibling` — `main.gala` qualifies `strings`, `sibling.gala` imports it. The qualified Go-package shape.
- `blocked_import_omitted_go_sibling` — same `main.gala`, with `run.sh` writing a handwritten `helper.go` that imports `strings`. The same shape with a Go sibling, which is what shows the language is irrelevant.
- `blocked_runtime_name_omitted_import` — `main.gala` calls bare `Future(2)`, `sibling.gala` dot-imports `concurrent`. The unqualified shape, whose `ERR` marker is the naming-package half of the hint because that is what separates it from the generic spelling on a Go qualifier.

`run.sh` grew one `materialize_fixtures` case for the handwritten fixture, matching the existing convention that a Go fixture reproducing a mixed-package case is written by the runner and gitignored rather than committed.

### Document edits

- `.agents/skills/gala-from-go/SKILL.md` — the mixed-package bullet now states the per-file refusal and that no sibling layout changes it; the section lead-in no longer claims all three rules fail away from the transpile. The trap "A sibling file's import hides a missing import from the transpiler" became "A sibling file's import is not your import", which keeps the useful guidance and drops the build-failure story. The closing `Never` bullet, which repeated the false premise, now says the check is per file and the omission is refused rather than carried. A grep sweep for "imports of any/every file" and for the "transpiles cleanly then fails `go build`" phrasing finds nothing left in the skill.
- `.agents/skills/gala-from-go/references/constructs.md` — both import rows carry `Code` `GALA-E0023`, and both Checks transpile and read the diagnostic rather than reading the generated import block. The second row was renamed from "a qualified name no file in the package imports" to "an unqualified GALA-runtime name the file does not import", because 0.84.1 makes the sibling's behaviour a non-discriminator and the unqualified shape is the one with its own hint. The prose under the table now says the omission is refused at the transpile rather than left for the Go compiler.
- `docs/gala-translation.md` — the roster's `named-imports` row carries the per-file rule, `Works; GALA-E0023`, and links all three probes; a new "Imports are checked per file" subsection under Workarounds names them and cites both explain pages; the probes are in the Blocked group row so `audit.sh` check 2 sees no orphan.
- `server/GALA.md` — the bullet is retitled "Imports are checked per file, and a missing import is a refusal rather than a lookup", quotes the `GALA-E0023` diagnostics, and links the probes.

## Verification

`run.sh` 80 passed / 0 failed. `audit.sh` all seven checks pass, including check 7 over all five documents. `go run ./inventory -check` reports `doc matches (168 files)`. The inventory unit tests pass with `-count=1`. `npm run format:markdown`, `npm run format:markdown:check`, `npm run test:markdown-rules` (30 tests) and root `npm run fmt:check` all pass. No expectation was re-baselined and `server/third_party/gala/` has zero changed files.

The probes were also checked in the direction that matters: temporarily adding the missing import to `blocked_import_omitted_gala_sibling` makes it fail with `transpile unexpectedly succeeded`, which is the world the old documents described, so the expectation genuinely discriminates rather than passing vacuously.

## Out of scope, recorded as TASK-0340

The per-file check covers value positions only. A bare GALA type name in a **type position** with no import — `struct Holder(F Future[int])`, `var holder Future[int]`, a parameter type — transpiles cleanly and emits the bare `Future[int]`, which fails `go build`. `gala explain GALA-E0023` documents the gap in its own text: it covers identifiers in value position and checks only the package qualifier in type position. That is exactly the failure shape the corrected documents say cannot happen, so it is filed as TASK-0340 with the three reproductions, rather than folded in here.

## Note for the reviewer

`npm run lint:markdown` reports one pre-existing sentences-per-line error at `docs/environments.md:95`, on a file this change does not touch. It is not one of this task's gates and was left alone.
<!-- SECTION:FINAL_SUMMARY:END -->
