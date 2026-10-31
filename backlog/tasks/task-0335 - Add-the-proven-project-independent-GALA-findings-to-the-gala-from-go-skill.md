---
id: TASK-0335
title: Add the proven project-independent GALA findings to the gala-from-go skill
status: Done
assignee: []
created_date: '2026-09-29 22:24'
updated_date: '2026-09-29 22:39'
labels: []
dependencies: []
documentation:
  - docs/gala-translation.md
  - server/GALA.md
  - server/scripts/20260923_gala_translation_probes/
modified_files:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
priority: medium
type: docs
ordinal: 340000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The `gala-from-go` skill is authored here and staged for upstreaming to the GALA plugin, so whatever is not in it has to be re-derived by whoever uses it.

The repository's spike record and probe corpus have since proven a body of knowledge that is about GALA rather than about this repository, and none of it reached the skill. The gaps that matter most are the ones a rewriter cannot reach by reading rows: the rules that apply only to a package mixing a translated file with a handwritten sibling, the method for checking a claim about the shape of the emitted Go rather than about the program's behaviour, and the runtime behaviours that transpile, build, and run while being wrong. An agent upstreaming the skill today would re-derive all of it from scratch, and two of the costs are silent.

Two claims this repository's own documents state as fact did not reproduce on the pinned compiler. That is upstreamable knowledge in its own right, in a contested form rather than as a correction, because a claim that cannot be reproduced on the compiler in hand is a fact about the reproduction rather than about the language.

The stripping of the now-duplicated project documents is deliberately excluded: the roster and the spike record are machine-read and audit-gated, so removing content from them is a separate change with its own verification.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The mixed-package rules are recorded as one family rather than scattered: imports do not propagate between sibling files in a package so every file imports what it uses, and a bare name that both a translated file and an imported package declare is carried as contested with the handwritten-constructor workaround rather than as fact
- [x] #2 The `.Size()` and `.ByteSize()` rows state the emitted lowering per receiver kind, so a reader can tell a character count from a byte count from a collection element count by reading the generated Go rather than by guessing
- [x] #3 The verification guidance states that an absence assertion is the half that carries a claim about emitted shape, and that a marker must name a semantic thing rather than layout because the generated Go is formatted and carries source-map directives
- [x] #4 The verification guidance states that the generated file's import list is the report of what a translation cost, so whether a rewrite pulled in the runtime is a read of the emitted file rather than a judgement
- [x] #5 The traps cover the runtime behaviours that transpile, build, and run while being wrong: an unchecked spawn whose result is not synchronised, an immutable binding around a collection that double-unwraps at the call boundary, and a byte-boundary truncation whose substitute changes what the count means
- [x] #6 The triage section states what a sealed type costs at the Go boundary, alongside the existing shape rules, so a file that introduces one is judged on the same footing as one that introduces a wrapped binding
- [x] #7 The gap report guidance gains the upstream source index, the warning that one upstream interoperability page contradicts the specification and the compiler, and the rule that a construct's substitute can itself carry a cost that belongs in the report
- [x] #8 The defect guidance states that a report leads with the artifact the transpiler itself generated that states the correct shape, and that a presence check cannot back a claim that a construct is blocked while an absence check can
- [x] #9 The pinning guidance names the two interop packages that a documentation query cannot describe and why, so a reader is not sent to look for a helper list that is not there
- [x] #10 Every added row and claim is derived by transpiling on the compiler installed on PATH during this task rather than copied from a project document, and each new row names a diagnostic code or a check the reader can run
- [x] #11 No project path, filename, task identifier, probe name, or compiler release appears anywhere in the three skill files, and the skill names only upstream sources as its references
- [x] #12 A structural check confirms every table row in the three skill files is closed on one physical line and every table has a consistent cell count, because the repository's Markdown formatting scripts exclude this directory
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
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Derive every added claim on the compiler in hand, in a scratch tree outside the repository: a sibling-import case to separate the two undefined-name diagnostics, a sealed type's emitted member set, an immutable binding around a collection, a missing collection member, a byte-boundary truncation, a resource combinator with and without explicit type arguments, a runtime-free file's import list, and a spawn whose result is read without synchronisation (built and run against the runtime).
2. `SKILL.md`: fold the mixed-package family into the triage shape rules, add the sealed-type boundary cost, and extend Verify with the emitted-shape method (absence assertions, semantic markers, the import list as the cost report).
3. `references/constructs.md`: state the per-receiver lowering on the size rows, add the byte-boundary truncation row, name the two interop packages a documentation query cannot describe, and give the resource rows the reason explicit type arguments are required.
4. `references/gaps.md`: add the upstream source index with the contradicted page named, the substitute-carries-a-cost rule, and the defect-framing rule about leading with the artifact the transpiler generated.
5. Sweep the three files for project paths, probe names, task identifiers, and release pins; run the structural check on tables; record the two contested findings in the notes.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-29 22:39
---
## The mixed-package finding refutes the repository's record, in the opposite direction

`server/GALA.md` records "imports do not propagate between sibling `.gala` files", cites `GALA-E0025`, and treats the failure as a transpile-time diagnostic. Measured on 0.84.1, both halves are wrong.

- Two `.gala` files in one package, the first importing a package and the second using it qualified without importing it: **the second transpiles cleanly**. The generated file's import block is missing the package, and `go build` fails with `two.gala:5: undefined: concurrent` — attributed to the `.gala` path through the `//line` directive.
- The same file in a package where **no** file imports the package: refused at transpile time with `GALA-E0023: undefined: concurrent`. `GALA-E0023` is "Undefined symbol"; `GALA-E0025` is "Unresolved cross-package symbol" and appeared in no case measured.

So the rule is the inverse of the recorded one, and it is a worse trap: the transpiler resolves the name and the compiler does not, so a clean transpile is what makes the omission invisible. The skill states the measured rule and names both diagnostics.

## A contested claim, carried as contested

The bare-name collision between a local declaration and an imported package's export did **not** reproduce on 0.84.1 in a minimal mixed package: with a `.gala` file declaring `Response`, a handwritten sibling declaring `Response`, and an import that also exports `Response`, the local declaration wins in the emitted Go. The repository's own probe record already called this unreproducible standalone, so the skill carries it as contested and prescribes the workaround (never name the type; call a handwritten constructor in the sibling) rather than the outcome.

## The capture guard does not cover both concurrency boundaries

`concurrent.Future(out + 1)` capturing a reassignable `var` is refused with `GALA-E0037` and a hint to snapshot into a `val`. `go_interop.Spawn(() => { out = 42 })` capturing the same `var` in the same package is **accepted**: it transpiles, the enclosing function returns `0`, and `go run -race` reports a data race between the spawned write and the caller's read. The existing row documented the refusal without saying the guard is boundary-specific, which would lead a reader to assume the spawn form is covered too.
---

created: 2026-09-29 22:39
---
## The resource-combinator claim was half wrong, now measured both ways

The repository records that omitting explicit type arguments binds the body parameter to `any`. Measured: the parameter binds correctly (`func(x *os.File) any`), and it is the **result** type argument that defaults to `any` — which strips the return type off the enclosing function. The transpile is clean and `go build` then reports `too many return values` and `WithFile(f) (no value) used as value`. With both type arguments written out the enclosing function keeps its `string` result.

## Other claims derived rather than copied

- `.ByteSize()` on a string emits Go `len`; `.Size()` on a string emits `utf8.RuneCountInString`; `.Size()` on a Go slice emits `len`; on a GALA collection it stays a method call. The three previous rows gave one spelling for three lowerings.
- A `val`-held `HashMap` read emits `m.Get().Get(k)`.
- `HashMap.GetOption` is refused with `GALA-E0044`, and the hint enumerates the type's declared members, so the diagnostic is the member list.
- `value[:n]` is a parse error, so a byte-boundary truncation has no spelling and the rune-loop substitute changes the count.
- A sealed type emits a merged parent with `std.Immutable` fields, `_variant`, an iota constant per variant, per-variant `Apply` and `Unapply` returning `std.Option`, `is<Variant>()` helpers, `Copy`/`Equal`/`String`, and `panic("unreachable")` on the final match branch. The parent has no positionally constructible form, which is why the skill now treats it as a boundary cost beside `val` and `:=`.
- A runtime-free file using `regexp` and `.ByteSize()` emitted an import block containing only `regexp`, zero runtime imports — the evidence for the import-list-as-cost check.
- `gala doc` on an interop package fails with "not found in any search path" while `std` and `concurrent` succeed; those two packages are the only handwritten-Go ones in a runtime where every other package has GALA sources.

## Checks

`npm run format:markdown:check` clean, `npm run test:markdown-rules` 30 passed, `npm run fmt:check` clean, and `audit.sh` passes all six checks with the roster and the spike record untouched. A structural check over the three skill files reports 12 tables, 147 rows, 0 inconsistent cell counts; the one row a naive checker flags uses a pre-existing escaped pipe and is unchanged by this task. The independence sweep for project paths, filenames, task identifiers, probe names, and any version pin returns nothing, which also caught one pre-existing upstream path that read as project-relative and is now named as the specification.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Extended the project-independent `gala-from-go` skill with the GALA knowledge that had been proven in this repository but never reached it. Tracked as TASK-0335; all twelve acceptance criteria are satisfied.

## What was added

**`SKILL.md`** — a third triage shape rule for a `sealed type`, whose emitted parent struct a Go caller cannot construct positionally at all, so it belongs beside `val` and `:=` as a boundary change rather than a local spelling.
A new "The Mixed Package" section gathers the three rules that belong to a package rather than to a file, which is what a split verdict creates.
A new "Checking A Shape Claim" section states the three claim types and the check each needs, the rule that an absence assertion is the half that carries a shape claim, the rule that a marker must name a semantic thing because the output is formatted and carries source-map directives, and the import block as the transpiler's own account of what the translation cost.
Six new traps: the sealed type's cost, a `val`-held collection's double unwrap, a sibling's import hiding a missing one, a byte-boundary truncation whose substitute changes the count, a resource combinator whose result type argument defaults to `any` and strips the enclosing return type, and a capture guard that covers `concurrent.Future` but not `go_interop.Spawn`.
Two new never-items.

**`references/constructs.md`** — the three `len` rows split into four, each naming its receiver kind and the Go it actually emits, because `.Size()` lowers to a rune count on a string, to `len` on a Go slice, and to a method call on a GALA collection.
New rows for a string truncation, a type sum translated into a sealed type, and a small "Packages And Files" section for the import and sibling-declaration rules.
The resource row's check now says what to read rather than merely what to compare, and the pinning section names why the two interop packages are the ones a documentation query cannot describe.

**`references/gaps.md`** — a "Where To Look Before You Decide" section indexing the four upstream sources and naming the one page that claims bare `len`, `make`, and `cap` work, so a report is not written against it.
The rule that a substitute can carry its own cost and that the cost belongs in the report.
Two rules for making a defect report land: lead with the artifact the transpiler itself generated that states the correct shape, and notice that a presence check cannot back a claim that a construct is blocked while an absence check can, with the matching bullet in "What Not To File".

## Two of the repository's own claims were refuted, and one of them the other way

The spike record states that imports do not propagate between sibling `.gala` files and that the failure is a transpile-time diagnostic with a named code. Measured on the compiler in hand, the opposite is true and the trap is worse: a file that omits an import its sibling already has **transpiles cleanly**, because the transpiler resolves the name against the whole package while the generated file carries only its own source's imports, and the failure is a `go build` error blamed on the `.gala` path through a source-map directive. A package where nothing imports that package is refused at transpile time instead, with a different code. The skill states the measured rule and both diagnostics.

The bare-name collision between a local declaration and an imported package's export did not reproduce at all on this compiler, so the skill carries it as contested and prescribes the workaround rather than an outcome the reader might rely on.

A third repository claim was half wrong: omitting a resource combinator's type arguments does not bind the body parameter to `any` — it binds correctly — and what defaults to `any` is the result, which strips the return type off the enclosing function and fails at build time in a generated file.

A fourth is new rather than corrected: the capture guard refuses a reassignable-variable capture on one concurrency boundary and silently accepts it on the other, where the program returns before the spawned body runs and the race detector fires.

## Independence and checks

Every added row was derived by transpiling in this session, and the shape claims by reading the emitted Go; the two run-time claims were built and run against the runtime, one of them under the race detector.
A sweep for project paths, filenames, task identifiers, probe names, and any version pin across the three files returns nothing; it also caught a pre-existing upstream path that read as project-relative, now named as the specification.
A structural check reports 12 tables, 147 rows, and no inconsistent cell counts, which the repository's own Markdown pipeline does not cover for this directory.

`audit.sh` passes all six checks with the roster and the spike record untouched, `npm run format:markdown:check` and `npm run fmt:check` are clean, and the Markdown rules suite passes 30 tests.
Swagger regeneration and the code-graph refresh do not apply: no annotation, runtime code, or generated artifact changed.
Unit and e2e suites are not required for a documentation-only change and were not run.

## Not done here

Stripping the now-duplicated project documents is deliberately out of scope, as the task description records: the roster's decision ladder and construct roster are parsed by the inventory tool and the mechanical-rewrite catalog and the gap cross-references are gated by the audit, so removing content from them is a separate change with its own verification.
The two refuted claims in `server/GALA.md` are recorded in the task comments rather than corrected here, because that file is audit-gated and the correction deserves its own task.
Nothing is committed.
<!-- SECTION:FINAL_SUMMARY:END -->
