---
id: TASK-0332
title: Author a project-independent gala-from-go skill for translating Go to GALA
status: Done
assignee: []
created_date: '2026-09-29 18:49'
updated_date: '2026-09-29 19:01'
labels: []
dependencies: []
modified_files:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
  - .agents/skills/gala-rewrite/SKILL.md
priority: medium
type: docs
ordinal: 337000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Author a GALA-from-Go translation skill that is independent of the timeful project and staged for upstreaming to the GALA plugin.

The translation knowledge needed to rewrite Go in GALA currently lives only in timeful-specific documents, so it cannot be reused outside this repository and cannot be upstreamed. An agent doing the same rewrite in another project would re-derive every construct verdict from scratch.

The skill is authored here first so it can be exercised against real files, then copied upstream.

## Shape

Three files under `.agents/skills/gala-from-go/`:

- `SKILL.md`: version check, three-verdict triage, output-style choice, the mechanical pass, the verify loop, the blocked path, the traps, and the never list.
- `references/constructs.md`: one row per Go construct with minimal Go, GALA spelling, status, diagnostic code, and the pinning check. Grouped by declarations, functions, control flow, builtins, containers, errors, and concurrency.
- `references/gaps.md`: the blocked-construct list, the gap-class vocabulary, and the gap report template.

## Independence

The skill must contain no timeful path, filename, task ID, or verdict. Its authoritative sources are the upstream `docs/GALA.MD` sections, `gala explain GALA-Exxxx`, `gala doc <pkg>`, and the upstream `gala-lint` skill for style.

`.agents/skills/gala-rewrite/SKILL.md` keeps the timeful-specific procedure and gains a one-line pointer to the new skill. The two overlap on replacement spellings but index from opposite ends: one by repository procedure, the other by Go syntax.

## Constraints

- Rows are version-agnostic: each names a diagnostic code or a check to run, and the agent verifies a row before building on it. The document is not pinned to one compiler release.
- Every construct row is derived by transpiling it against the compiler installed on `PATH`, not copied from any existing roster.
- A row that contradicts an existing project document is flagged rather than silently reconciled.
- `scripts/markdown.mjs` excludes `.agents/**`, so the Markdown formatting scripts do not cover these files; the `docs/AGENTS.md` sentence-per-line and one-row-per-line house style still applies by choice.

## Upstream follow-up

Copying the skill into the GALA checkout, adding it to the plugin and marketplace descriptions, and opening a pull request to the upstream repository are separate later work.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `SKILL.md` states the compiler-version check, the three triage verdicts, the output-style choice, the mechanical translation pass, the verification loop, the stop-and-report path, the traps, and an explicit never-do list
- [x] #2 `references/constructs.md` has one row per Go construct with minimal Go, GALA spelling, status, diagnostic code where one exists, and the check that pins the row, grouped by declarations, functions, control flow, builtins, containers, errors, and concurrency
- [x] #3 `references/gaps.md` lists the blocked constructs, defines the gap-class vocabulary, and provides a gap report template addressed to the upstream triage issue
- [x] #4 Every construct row's status is derived by transpiling that construct against the compiler on `PATH` in this session, not copied from an existing project document
- [x] #5 No timeful path, filename, task ID, roster verdict, or probe name appears anywhere in the three files
- [x] #6 No skill row is pinned to one compiler release; each names a diagnostic code or a check the reader can run
- [x] #7 Any row contradicting `docs/gala-translation.md` is recorded as a finding rather than reconciled silently, and each such finding is reported in the task notes
- [x] #8 `.agents/skills/gala-rewrite/SKILL.md` gains a one-line pointer to the new skill and keeps its project-specific procedure
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
1. Re-verify the load-bearing claims on the compiler actually installed, rather than trusting either the existing roster or the prior session's notes. Every construct row is then derived by transpiling a minimal repro and, for a shape claim, reading the emitted Go and then building and running it.
2. Author `SKILL.md` with the version check, the three triage verdicts, the output-style rules, the ordered mechanical pass, the verify loop, the stop-and-report conditions, the traps, and the never list.
3. Author `references/constructs.md` as one row per construct, grouped by declarations and types, functions and lambdas, control flow, builtins and expressions, containers, errors and failure, and concurrency and resources. Each row carries minimal Go, the GALA spelling, a status, a diagnostic code, and a check the reader can run.
4. Author `references/gaps.md` with the gap-class vocabulary, the blocked list re-derived on the compiler in hand, the defect findings that are issues rather than gaps, and a report template.
5. Add the one-line pointer to the existing rewrite skill.
6. Sweep all three files for project-specific terms and for release pins, then run the repository's checks.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Verification method

Each row was derived by writing a minimal `.gala` repro under `/tmp/opencode/gala-verify/`, transpiling it, and then, for any row making a shape or acceptance claim, reading the emitted Go and building and running it in a scratch module that replaces the GALA module with the vendored runtime. Roughly 90 scratch programs were transpiled; nothing in the repository was written by these checks. Compiler `gala` 0.84.1, Go go1.26.7.

The distinction that shaped the work: a shape claim cannot be settled by a test run, because a wrapper the program never observes is invisible to the program's own assertions.

## Finding 1: the project roster's `:=` claim is wrong about the emitted shape

`docs/gala-translation.md` states in three places that `a, b := f()` "lowers like `var a, b = f()`" and lists `multi-value-define` as a direct form. On 0.84.1 the two spellings emit different code. `var n, err = strconv.Atoi("42")` emits a plain Go binding and plain reads. `n, err := strconv.Atoi("42")` emits a `var` block binding each result to `std.NewImmutable(...)` and emits `.Get()` at every read.

Upstream `docs/GALA.MD` §2 states this directly ("Variables declared this way are immutable in GALA"), so the roster is the document out of step, not the compiler. The roster's `delta_multi_value_define` probe asserts the program prints the right value and passes, because the wrapper is invisible to that assertion; the probe pins acceptance, not shape.

Why it matters: the roster's mechanical-rewrite catalog tells a rewriter to move an `if` initializer's binding onto the preceding line as `err := f()`. That is correct inside GALA but is the wrong shape for a value a Go sibling reads. The skill states the `var` rule instead.

## Finding 2: the prior handoff's `Try` claim is wrong

`backlog/handoffs/handoff-2026-09-29T18-32-34Z.md` records that "`Try(...)` around an error-only Go call never reports failure; `FromError(...)` is the correct wrapper". Measured on 0.84.1, `Try` reports failure correctly in both directions, for a removable temp file and for a call that fails:

```
Try on a real removable file reports failure: false
FromError on a real removable file reports failure: false
Try on a failing call reports failure: true
FromError on a failing call reports failure: true
```

The two lower differently: `Try(f())` wraps the call in a closure that panics on the error and catches it, while `FromError(f())` calls the runtime helper directly. Both are correct, so `FromError` is the more direct spelling rather than a correctness fix. The skill says exactly that.

A first run appeared to confirm the handoff. The cause was a bug in the measurement, not the compiler: the probe removed the same file twice, so the second removal failed legitimately. Recorded because it is the exact failure mode the skill warns about when it tells the reader to re-derive a row rather than trust it.

## Finding 3: `gala doc` cannot describe the interop helper package

`go_interop` is hand-written Go rather than GALA source, so `gala doc go_interop` fails with "package not found in any search path", while `gala doc std`, `resource`, `collection_immutable`, and `concurrent` all succeed. A skill that pointed an agent at `gala doc` for the sanctioned helpers would send it to a package that is not there, so the skill tells the reader to use the transpiler's own diagnostic hint instead. Every `go_interop` helper named in the skill was verified by transpiling a program that uses it: `SliceOf`, `SliceEmpty`, `SliceWithSize`, `SliceWithCapacity`, `SliceCap`, `SliceAppend`, `SliceAppendAll`, `SliceCopy`, `MapEmpty`, `MapPut`, `MapLen`, `MapDelete`, `New`, `NewMutex`, `ToBytes`.

## Finding 4: the roster's `Instance` interface claim is not reproducible

The roster states twice that a GALA struct grows "`Copy`, `Equal`, `Unapply`, and an `Instance` interface". The emitted Go on 0.84.1 contains `Copy`, `Equal`, and `Unapply` and no `Instance`, and `p.Is[P]()` transpiles and then fails to build. Upstream `docs/GALA.MD` §4 documents only `Copy` and `Equal` and says to test a variant with `match`. The skill states what the emitted code actually contains.

## Transpiler defects found, for upstream issue reports

None of these is a gap; each is a construct GALA should handle where the handling is wrong, so each is an issue rather than a request. All three are recorded in `references/gaps.md` under defects observed while pinning rows, with the repro shape needed to file them.

1. **`fallthrough` inside a `match` arm yields `GALA-E0017` rather than a diagnostic.** A bare `fallthrough` statement is correctly rejected as a forbidden statement keyword, so the rejection path is inconsistent: the word passes the parser in arm position, reaches codegen, and emits Go that does not parse.
2. **A `match` on a struct whose fields are declared `var` emits a `.Get()` on each field read and then fails to build**, because a `var` field is a plain Go field. The identical struct with the fields declared without `var` builds and runs, so the transpiler unwraps the already-unwrapped field.
3. **`.Size()` and `.ByteSize()` on a field of a struct declared in a handwritten Go sibling emit a method call on a plain Go type and fail to build.** The same calls on a GALA-declared struct lower correctly, so the trigger is a mixed package. This is the one defect that bears on a triage decision, because it is an argument for the split verdict over rewrite-whole when a package will have a `.go` sibling.

Defect 3 independently re-derives the roster's contested `contested_size_field_go_sibling` probe, which reproduced exactly in a scratch module with a handwritten `types.go` beside a transpiled `twin.gala`. That finding held up; it is carried into the skill as both a trap and a defect entry.

## Two claims this session got wrong first, caught by the check the row prescribes

Both are the argument for the version-agnostic design, since in each case the draft was written from the roster or from memory and the row's own check refuted it.

1. A first draft asserted that a `for` post statement has no `++` spelling. Transpiling `i++` in a post slot succeeds and emits Go `i++`. The rule is about mutability, not the operator: `++` on a `val` or a `:=` binding is rejected, while the `:=` in a `for` init slot is mutable.
2. A first draft named `concurrent.Future.WithTimeout` as a free function. `gala doc concurrent` shows `WithTimeout` as a method on `Future[T]`, alongside `AwaitFor(timeout)`. The row now names the method.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Authored a project-independent `gala-from-go` skill at `.agents/skills/gala-from-go/`, staged for upstreaming to the GALA plugin, and pointed the existing `gala-rewrite` skill at it. Tracked as TASK-0332.

## What was built

Three files, each owning one question:

- `SKILL.md` — the version check, three triage verdicts (rewrite whole, split with a handwritten sibling, keep handwritten), the two output-style rules, the ordered mechanical pass, the verify loop, five stop-and-report conditions, fifteen traps, and a nine-item never list.
- `references/constructs.md` — 48 rows across seven groups (declarations and types, functions and lambdas, control flow, builtins and expressions, containers, errors and failure, concurrency and resources). Each row carries minimal Go, the GALA spelling, a status, a diagnostic code where one fires, and a check the reader can run.
- `references/gaps.md` — the gap-class vocabulary, a blocked list of 20 constructs re-derived on the compiler in hand, three defect findings that are issues rather than gaps, and a report template addressed to the upstream triage issue.

## Independence

A grep sweep across all three files for project-specific terms and for release pins returns nothing. The skill names no repository path, filename, task, roster verdict, or probe. It references only upstream sources: the GALA specification, `gala explain`, `gala doc`, and the sibling lint skill. `scripts/markdown.mjs` excludes `.agents/**` from the formatting pipeline, so the `docs/AGENTS.md` sentence-per-line and one-row-per-line house style was applied by choice; a structural check confirms every table row is closed on one physical line and every table has a consistent cell count.

## Method

Roughly 90 minimal `.gala` programs were transpiled with the compiler on `PATH`. Every row's status comes from that transpile, not from an existing document. Shape claims were settled by reading the emitted Go; acceptance claims were settled by building and running in a scratch module replacing the GALA module with the vendored runtime. Where a row made no difference between the two readings, the row says so rather than implying a verification that did not happen.

## Findings

Four contradictions with existing documents, recorded rather than reconciled, with the measurement for each in the task notes:

1. The project roster's claim that `a, b := f()` lowers like `var a, b = f()` is wrong about the emitted shape. The `:=` form binds through `std.NewImmutable` and emits `.Get()` at every read; the `var` form emits a plain Go binding. Upstream's own specification says `:=` bindings are immutable, so the roster is the document out of step. The roster's probe does not catch this because it asserts the program's output rather than the emitted shape.
2. The prior session's handoff claim that `Try` around an error-only Go call never reports failure is wrong. Measured in both directions, `Try` and `FromError` both report correctly; they differ only in how direct the lowering is.
3. `gala doc` cannot describe the interop helper package, because that package's source is handwritten Go. Every helper named in the skill was instead verified by transpiling a program that calls it.
4. The roster's claim that a GALA struct gains an `Instance` interface is not reproducible; the emitted struct has `Copy`, `Equal`, and `Unapply` and no `Instance`, and the corresponding call transpiles and then fails to build.

Three transpiler defects, each a construct GALA should handle where the handling is wrong, recorded in `references/gaps.md` with the repro shape needed to file them: `fallthrough` inside a `match` arm produces an internal transpiler error instead of a diagnostic; a `match` on a struct with `var` fields emits an unwrap those fields do not have and then fails to build; and `.Size()`/`.ByteSize()` on a field of a Go-sibling-declared struct emits a method call on a plain Go type and fails to build. The third independently reproduced an existing roster probe, so that finding held up.

Two claims the first drafts got wrong, both caught by the check the row itself prescribes, which is the argument for the version-agnostic design: a `for` post statement does accept `++` and the real rule is about mutability, and `WithTimeout` is a method on `Future` rather than a free function.

## Checks

`npm run format:markdown:check`, `npm run fmt:check`, `npm run test:markdown-rules`, and `npm run test:markdown-format` all pass. Swagger regeneration and the code-graph refresh are not applicable: this change touches no annotations, no runtime code, and no generated artifact. Backend and e2e tests are not required for a documentation-only change and were not run.

## Not done here

The `:=` correction against the roster, and the three defect reports to the upstream tracker, are follow-up work outside this task's scope and are recorded in the task notes. Copying the skill into the upstream checkout, adding it to the plugin and marketplace descriptions, and opening a pull request remain separate later work.
<!-- SECTION:FINAL_SUMMARY:END -->
