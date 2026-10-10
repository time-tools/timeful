---
id: TASK-0334
title: Add an emitted-Go expectation kind to the GALA probe corpus
status: Done
assignee: []
created_date: '2026-09-29 19:46'
updated_date: '2026-09-29 22:03'
labels: []
dependencies: []
modified_files:
  - server/scripts/20260923_gala_translation_probes/run.sh
  - server/scripts/20260923_gala_translation_probes/audit.sh
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_multi_value_define_wraps/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_multi_value_define_wraps/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_multi_value_var_plain/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_multi_value_var_plain/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_generic_struct_instance/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_generic_struct_instance/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_plain_struct_no_instance/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_plain_struct_no_instance/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_field_wrapped_immutable/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_field_wrapped_immutable/expect
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_var_field_plain/main.gala
  - >-
    server/scripts/20260923_gala_translation_probes/probes/emit_var_field_plain/expect
  - docs/gala-translation.md
priority: high
type: task
ordinal: 339000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Give the GALA probe corpus an expectation kind that asserts on the text of the generated Go, so a claim about emitted shape can be checked rather than asserted in prose.

## Why

The corpus has three expectation kinds, and none of them can observe the shape of the emitted Go. `KIND=pass` with `RUN=yes` asserts that the program builds and prints the expected output, which cannot distinguish a plain Go binding from a wrapped one, because a wrapper the program never observes is invisible to an output assertion. `KIND=transpile_fail` and `KIND=build_fail` assert that something is rejected, and a rejection is not the claim: a construct can be accepted with the wrong shape.

The consequence is already paid for. A wrong claim about the multi-value `:=` receive survived in three documents and a 54-probe corpus, because the probe that covers it pins the program's recovered value and not the emitted shape. The probe's note now says so, which records the gap without closing it.

## What an emitted-text assertion buys

It closes claims of three kinds, each of which is currently unchecked:

- A binding that wraps versus one that does not, checked by asserting a wrapper marker is present in one spelling and absent in the other. The two spellings are the same program, so the absence assertion is the one that carries the claim.
- A declaration that gains synthesized members versus one that does not, checked across a generic and a non-generic declaration of the same shape.
- A field that is wrapped versus one that is plain, checked by asserting the wrapper's type name in the generated field list.

## Design constraints to resolve before implementing

The runner reads at most the first value of each `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` key, so a single-valued new key gives one assertion per probe. A decision is needed on whether that is enough, or whether multiple assertions per probe are wanted and the reader has to change. One probe per assertion is the cheaper option and keeps the reader unchanged.

An absence assertion is the interesting half, and it is the half that interacts with the audit. The audit requires a gap's blocking probe not to be `KIND=pass`, on the reasoning that a blocker has to fail for the construct it names. An absence assertion fits that reasoning, because a construct is blocked exactly when the thing it needs is not emitted; a presence assertion does not fit it. Whatever kind is chosen, the audit's check has to classify it correctly rather than inherit the pass rule.

Assertions must target semantic markers rather than layout, because the generated Go is formatted and carries source-map directives that name the probe's own path. A marker like a wrapper constructor call or a synthesized interface name is stable across formatting; an indentation or a line break is not.

These assertions are coupled to the compiler release more tightly than any existing kind, because they pin codegen rather than behaviour. The `build_fail` expectations are already coupled to the Go toolchain, and the roster documents re-baselining for that; a new kind needs the same treatment stated up front, and the corpus page's re-baseline guidance has to cover it. A probe that fails on a release bump for a reason nobody recorded is worse than no probe.

## What this does not cover

The mixed-package defect, where a method call is emitted on a field whose type the transpiler does not know, is already covered: the generated Go does not compile, so `KIND=build_fail` with the compiler's diagnostic is the right kind for it. An emitted-text assertion adds nothing there and should not be used for it.

The upstream transpiler defects found in TASK-0332 belong to the upstream tracker and are not in scope.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `run.sh` supports an expectation that asserts on the text of the generated Go, with both a presence assertion and an absence assertion, and a probe using it passes when the generated text matches and fails with a readable diff when it does not
- [x] #2 `audit.sh` accepts the new expectation kind rather than rejecting it as an unknown kind
- [x] #3 At least one probe pins that a multi-value `:=` receive wraps its bindings and one pins that the `var` spelling does not, so the two spellings can no longer be documented as equivalent
- [x] #4 At least one probe pins that a generic struct declaration gains an `Instance` interface and an `Is<T>()` method while a non-generic declaration does not, so the qualified claim is checked in both directions
- [x] #5 The roster's notes state that these assertions are coupled to the compiler release by construction, that a compiler bump is expected to fail them, and what the re-baseline procedure is
- [x] #6 The runner's documentation and the corpus page state which of the existing probe kinds assert on program behaviour, which assert on a build failure, and which assert on emitted text, so a future author knows which kind a new claim needs
- [x] #7 `audit.sh` and `run.sh` both pass, and the full probe count is reported
- [x] #8 Any probe that is added rather than extended uses a name that does not collide with an existing probe, and the roster's probe links resolve to it
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation plan

### The kind and its keys

`KIND=emit` asserts on the text of `probes/<name>/main.gen.go` after a successful transpile; it does not build or run the program, because the claim is about codegen rather than behaviour.

Two new expect keys carry the assertions:

- `CONTAINS=<marker>`: the marker must appear in the emitted Go.
- `ABSENT=<marker>`: the marker must not appear in the emitted Go.

Both keys are repeatable and the polarity lives in the key name, so a marker that contains a colon or an `=` needs no escaping and the expect file reads as an assertion list. This is the "multiple assertions per probe" branch of the design decision in the description: AC #4 needs one probe to pin both the `Instance` interface and the `Is<T>()` method, which a single-valued key cannot express without a layout-coupled multi-line marker. The reader change is a second `sed`-based helper that returns *all* values of a key instead of the first; the existing single-valued `expect_value` helper is untouched, so `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` behave exactly as before.

### Failure report

`CONTAINS` failures print the numbered generated Go, truncated at 60 lines, because there is no anchor line to centre on.
`ABSENT` failures print `grep -nF -C 2` of the offending marker, which is the diff-like output.
Every failing assertion in a probe is reported, not just the first.
The full numbered generated Go is always written to `.logs/<name>.emit.txt`.

### Audit classification

`audit.sh` check 1 accepts `KIND=emit` and requires at least one non-empty `CONTAINS` or `ABSENT` marker, and rejects an empty marker line.

`audit.sh` check 3 keeps the existing "not `KIND=pass`" rule and adds one case rather than inheriting it: a `KIND=emit` probe with a `CONTAINS` marker is not a blocker either, because a presence assertion pins a shape that is there rather than a shape that is missing. A `KIND=emit` probe whose markers are all `ABSENT` is a valid blocker, which is the reasoning in the description: a construct is blocked exactly when the thing it needs is not emitted.

### Probes

Four new `emit_` probes, none of which collide with an existing name:

- `emit_multi_value_define_wraps`: `n, err := strconv.Atoi("42")` wraps each result and unwraps every read. Presence markers: `std.NewImmutable`, `n.Get()`, `err.Get()`.
- `emit_multi_value_var_plain`: the same program in the `var` spelling emits none of them. Absence markers, the same three strings, so the absence half is what carries the claim.
- `emit_generic_struct_instance`: `struct Pair[T any]` gains `type PairInstance interface` and `IsPair()`. Presence markers: `type PairInstance interface`, `IsPair() bool`, `func (_ Pair[T]) IsPair() bool`.
- `emit_plain_struct_no_instance`: `struct Coord` of the same shape gains neither. Absence markers: `type CoordInstance interface`, `IsCoord()`.

Every marker is a semantic name a formatter cannot move: a wrapper constructor call, a synthesized interface and method name, or a method call on a binding. No marker is whitespace, indentation, or a line break, and none matches the `//line main.gala:N` directives.

### Documentation

- `run.sh` header comment: the four kinds and which of them assert on behaviour, on a build failure, and on emitted text.
- `docs/gala-translation.md` [Probe corpus]: the same taxonomy, and the compiler-release coupling plus the re-baseline procedure.
- `docs/gala-translation.md` [When to revisit]: the `emit_*` probes join the set that is expected to move on a compiler bump.
- `docs/gala-translation.md` [Multi-value bindings]: replace "Neither probe pins the emitted shape" with the two new `emit_` probes.
- Construct roster Probe column: the `multi-value-define`, `multi-value-var`, `methods`, and `struct-declarations` rows gain the `emit_` probes that pin their emitted-shape claims.
- Probe group table: a new "Emitted shape" row.
- Each new probe's `NOTE=` records the compiler-release coupling and the re-baseline step, so the note travels with the probe.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation notes

### What was built

`KIND=emit` in `run.sh`: transpile, then assert on the text of `probes/<name>/main.gen.go`.
The assertions are two repeatable expect keys whose polarity lives in the key name, `CONTAINS=<marker>` and `ABSENT=<marker>`, so a marker needs no escaping and the expect file reads as an assertion list.
The reader change is one added helper, `expect_values`, which returns every value of a key; the existing single-valued `expect_value` is untouched, so `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` behave exactly as before.

The design decision the description left open is resolved in favour of multiple assertions per probe, because AC #4 needs one probe to pin both the `Instance` interface and the `Is<T>()` method and a single-valued key could only do that with a multi-line marker, which the description rules out as layout-coupled.

### Six probes, not four

The description lists three claim classes the new kind buys, and the acceptance criteria only name the first two, so the third was delivered as well rather than left prose-only: the `struct-declarations` roster row and the `Mechanical rewrites` catalog both assert that a non-`var` field is wrapped and a `var` field is not, and nothing checked it.

| Probe | Kind of assertion | Claim pinned |
| ----- | ----------------- | ------------ |
| `emit_multi_value_define_wraps` | 3 `CONTAINS` | `n, err := f()` wraps each result in `std.Immutable` and unwraps every read with `.Get()` |
| `emit_multi_value_var_plain` | 3 `ABSENT`, the same strings over the same program | the `var` spelling emits no wrapper and no unwrapping |
| `emit_generic_struct_instance` | 3 `CONTAINS` | a generic declaration gains `type PairInstance interface` and `IsPair()` |
| `emit_plain_struct_no_instance` | 2 `ABSENT` | a non-generic declaration of the same shape gains neither |
| `emit_field_wrapped_immutable` | 2 `CONTAINS` | a non-`var` field is emitted as `std.Immutable[string]` |
| `emit_var_field_plain` | 2 `ABSENT` | a `var` field stays a plain Go field |

Every marker is a semantic name a formatter cannot move: a wrapper constructor call, a wrapper type name, a synthesized interface or method name, or a method call on a binding.
No marker is whitespace, indentation, or a line break, and none can match the `//line main.gala:N` directives that carry the probe's own path.

### A latent audit bug the new kind exposed

`audit.sh` check 3 built its blocking-probe list by matching backticked tokens against the `blocked_` and `contested_` name prefixes, so no probe under any other prefix could ever be named as a gap blocker.
The new classification in check 3 would have been dead code without fixing that.
The parser now reads the Genuine gaps table's `Blocking probe` cell by its column header, the way checks 4 and 6 already find their tables, so a new probe prefix is pickable without editing the script and a probe named in the workaround or upstream cell cannot be mistaken for the blocker.

### Blocker classification, verified in both directions

The description was explicit that the audit must classify the new kind rather than inherit the pass rule.
`KIND=pass` asserts behaviour and a `KIND=emit` probe carrying a `CONTAINS` marker pins a shape that is there rather than a shape that is missing, so neither can back a gap claim.
A `KIND=emit` probe whose markers are all `ABSENT` can, because a construct is blocked exactly when the thing it needs is not emitted.

Observed, by temporarily pointing GAP-1's blocking probe at each probe in turn:

```
FAIL GAP-1: probe emit_multi_value_define_wraps is KIND=emit with a presence marker, not a blocker
audit failed: 1 problem(s)
```

and for the absence-marker probe in the same slot, `audit passed`.

### Readable failure output, observed

A broken `CONTAINS`:

```
FAIL emit_multi_value_define_wraps: 1 of 2 emitted-text assertion(s) did not hold
  no match for CONTAINS=thisMarkerIsNotEmitted
  the generated Go, numbered; a full copy is at .logs/emit_multi_value_define_wraps.emit.txt
       1	// Code generated by GALA transpiler. DO NOT EDIT.
       ...
      14			n              = std.NewImmutable(_tmp_1)
```

A broken `ABSENT`, which shows the offending lines with two lines of context:

```
FAIL emit_generic_struct_instance: 1 of 2 emitted-text assertion(s) did not hold
  matched ABSENT=NewImmutable, which the emitted Go must not contain
  39-func main() {
  40-//line main.gala:6
  41:	p := std.NewImmutable(Pair[int]{First: 1, Second: 2})
  42-//line main.gala:7
  43-	fmt.Println(p.Get().First)
```

Every failing assertion in a probe is reported, not just the first, and the runner exits 1.

### Rejection cases, observed

Three throwaway probes confirmed the guard rails rather than assuming them:

```
FAIL tmp_check_a: KIND=emit without a CONTAINS or ABSENT marker
FAIL tmp_check_b: empty CONTAINS or ABSENT marker
FAIL tmp_check_c: unknown KIND 'emitt'
```

`audit.sh` reported the same three conditions, so the runner and the audit agree on what a valid `expect` file is.

### Documentation

- `run.sh` header: the four kinds and a "which kind a new claim needs" list, the marker rules, and the compiler-release coupling with a pointer to the re-baseline procedure.
- `docs/gala-translation.md` [Probe corpus]: a four-row kind taxonomy table, the marker rules, the failure-output contract, the compiler-release coupling, the re-baseline procedure, and the audit's new blocker classification.
- `docs/gala-translation.md` [When to revisit]: the six `emit_*` probes join the set expected to move, with the re-baseline step.
- `docs/gala-translation.md` [Multi-value bindings]: "Neither probe pins the emitted shape" is replaced by the two new probes and the reason the absence half carries the claim.
- `docs/gala-translation.md` [Compiler history]: the `:=` bullet now names the two probes, and records that a wrong claim about it once survived in three places.
- `docs/gala-translation.md` [Sources and scope]: the corpus is described as covering emitted-shape verdicts too.
- Construct roster: `multi-value-define`, `multi-value-var`, `methods`, and `struct-declarations` rows gain the `emit_` probes that pin their emitted-shape claims.
- `Mechanical rewrites` and `Replaced by analog or workaround`: the `var` field rule gains the two field probes.
- Probe group table: a new `Emitted shape` row.

Each new probe's `NOTE=` records the pairing, which half carries the claim, the compiler-release coupling, and the re-baseline pointer, so the note travels with the probe.

### Checks

- `run.sh`: `go go1.26.7; gala 0.84.1; 60 probes` then `60 passed, 0 failed`, up from 54.
- `audit.sh`: `audit passed` with all six checks, `probe directories (60)`, `no orphan probes, all roster links resolve`, and the gap, class, and catalog checks unchanged.
- `go run ./inventory -check`: `inventory check: doc matches (166 files)`, so the new probe directories are still skipped by the inventory.
- `go test ./inventory/...`: `ok timeful/gala-probes/inventory`.
- `npm run format:markdown` then `format:markdown:check`: clean.
- `npm run lint:markdown`: the one error is pre-existing in `docs/environments.md:95` and is present on the unmodified tree too.
- `npx vitest run eslint/markdown/ prettier/markdown/`: 4 files, 60 tests, all passed.
- `npm run fmt:check`: clean.
- `bash -n` on both scripts: clean.
- `codebase-memory-mcp_index_repository --repo-path .`: re-indexed, 11361 nodes and 45788 edges.

Swagger regeneration does not apply: no route annotation or handler changed.
`docs/environments.md` and `PLUGIN_API_README.md` are untouched because no contract moved; the only contract this change alters is the probe corpus's own `expect` file format, which is documented in the runner header and in the corpus page.

### Definition of Done judgement for the test and Swagger items

DoD 2 and 3 were checked on the observation that the diff touches no frontend, no server runtime code, and no build or deployment configuration: the changed paths are the nested probe Go module under `server/scripts/20260923_gala_translation_probes/`, `docs/gala-translation.md`, and the Backlog records. The unit test that covers this directory, `go test ./inventory/...`, was run and passes; the e2e suite is Playwright browser coverage of the frontend and has no case that exercises the probe corpus, so re-running it could not observe this change. DoD 5 (Swagger) does not apply because no route annotation or handler changed, and is left unchecked rather than checked.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Summary

The GALA probe corpus can now assert on the shape of the emitted Go, so a claim about codegen is checked instead of asserted in prose.

### The new kind

`KIND=emit` transpiles a probe and then asserts on the text of `main.gen.go`; it neither builds nor runs the program, because the claim is about codegen rather than behaviour.
Two repeatable expect keys carry the assertions, `CONTAINS=<marker>` and `ABSENT=<marker>`, with the polarity in the key name so a marker needs no escaping.
The reader change is one added helper, `expect_values`; the existing single-valued `expect_value` is untouched, so `KIND`, `CODE`, `ERR`, `RUN`, and `NOTE` behave exactly as before.

The design decision the task left open is resolved in favour of multiple assertions per probe: acceptance criterion 4 needs one probe to pin both the `Instance` interface and the `Is<T>()` method, and a single-valued key could only do that with a multi-line marker, which the task itself rules out as layout-coupled.

A failure prints the numbered generated Go, plus the offending lines with two lines of context for a broken `ABSENT`; every failing assertion in a probe is reported, not just the first, and the full copy lands in `.logs/<name>.emit.txt`.

### Six probes

`emit_multi_value_define_wraps` and `emit_multi_value_var_plain` are the same program in the two spellings, asserting presence and absence of the same three strings; the absence half is what carries the claim, because a transpiler that started wrapping both spellings would pass every output probe in the corpus and fail only these two.
`emit_generic_struct_instance` and `emit_plain_struct_no_instance` check the synthesized-members claim in both directions, so the qualified wording in the roster can no longer drift to the unconditional.
`emit_field_wrapped_immutable` and `emit_var_field_plain` close the third claim class the task description named but the acceptance criteria did not: the `var` field rule that the `Mechanical rewrites` catalog depends on was prose-only until now.

Every marker is a semantic name a formatter cannot move, and none can match the `//line main.gala:N` directives.

### Audit

`audit.sh` accepts the kind, requires at least one non-empty marker, and rejects an empty one.
Its blocker check classifies kinds rather than inheriting the pass rule: a `KIND=emit` probe with a `CONTAINS` marker cannot back a gap claim, while one whose markers are all `ABSENT` can.
Both directions were observed rather than assumed.

Fixing that exposed a latent bug: the gap-row parser built its blocking-probe list from the `blocked_` and `contested_` name prefixes, so no other prefix could ever be named as a blocker and the new classification would have been dead code. The parser now reads the `Blocking probe` cell by its column header, the way checks 4 and 6 already find their tables.

### Documentation

The runner header and the corpus page both state which kind asserts on behaviour, on a build failure, and on emitted text, and both state that these assertions are coupled to the compiler release by construction, that a bump is expected to fail them, and how to re-baseline.
Each probe's `NOTE=` repeats that, so the coupling travels with the probe.
The `Multi-value bindings` section no longer says the shape is unpinned, and the `Compiler history` bullet now records that a wrong claim about `:=` once survived in three places.

### Verification

- `run.sh`: `60 probes`, `60 passed, 0 failed`, up from 54.
- `audit.sh`: `audit passed`, all six checks.
- `go run ./inventory -check`: `doc matches (166 files)`.
- `go test ./inventory/...`: pass.
- `npm run format:markdown:check`, `npm run fmt:check`, and the markdown lint and format vitest suites: clean, the one lint error being pre-existing in `docs/environments.md`.
- Unit and e2e suites: the diff is confined to the nested probe Go module, two bash scripts, and Markdown, so neither suite has changed inputs; the corpus's own Go tests were run and pass.
- Swagger regeneration does not apply; no route annotation or handler changed.
- Code knowledge graph re-indexed.
<!-- SECTION:FINAL_SUMMARY:END -->
