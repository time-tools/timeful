---
id: TASK-0324
title: Make every GALA-equivalent verdict in the translation roster deterministic
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 17:07'
updated_date: '2026-09-23 17:19'
labels: []
dependencies: []
references:
  - docs/gala-translation.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - server/GALA.md
  - 'https://github.com/martianoff/gala/pull/529'
documentation:
  - >-
    backlog/tasks/task-0322 -
    Spike-build-a-Go-to-GALA-translation-roster-for-the-server.md
  - >-
    backlog/tasks/task-0323 -
    Add-a-version-gated-check-for-the-GALA-translation-probe-corpus.md
modified_files:
  - docs/gala-translation.md
  - server/scripts/20260923_gala_translation_probes
priority: medium
type: task
ordinal: 322000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

`docs/gala-translation.md` is the repository's Go-to-GALA reference, but several verdicts name alternatives or conditional workarounds without a criterion for choosing, so a future transcriber has to guess which GALA equivalent to use.

The affected entries are:

- Conditional verdicts without a stated criterion: `switch-statements`, `defer`, `panic-calls`, `slice-expressions`, `const-declarations`, `struct-declarations`, `methods`, `generic-declarations`.
- Multiple equivalents offered with no selection rule: `if-initializers` (`var` versus `Try`), `len-calls` (`.Size()` versus `.ByteSize()`), `map-literals` and `delete-calls` (`go_interop` helpers versus `HashMap`), `slice-literals` and `append-calls` (`go_interop` helpers versus `Array`/collections), `channel-types` (handwritten versus `go_interop`/`concurrent`).
- Version-matrix rows recorded as "not probed": `for ; cond; post` with the init omitted, and alias to an instantiated generic struct.

## Constraints

- Keep the work local: no upstream issue or PR.
- Verdicts must stay consistent with `server/GALA.md` and with the 0.81.0 versus post-PR-#529 split in the roster.
- The corpus in `server/scripts/20260923_gala_translation_probes/` commits only `.gala` sources, `expect`, `expected.out`, `go.mod`, `.gitignore`, and `run.sh`; generated artifacts stay gitignored.
- Root Markdown follows the sentences-per-line and table-row rules in `docs/AGENTS.md`.
- TASK-0323 will run the corpus in CI, so probes added here are covered by that gate once it lands.

## Evidence required

Record the rule changes and the corpus run output in the task notes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every roster entry with a conditional or multi-option verdict states one concrete decision rule instead of leaving the choice to the reader.
- [x] #2 The conditional entries name the criterion that selects each branch.
- [x] #3 The multiple-equivalent entries state the runtime-free versus runtime-enabled selection rule inline.
- [x] #4 Each rule branch is pinned by a corpus probe (new or existing), and the roster entry names the probe that pins it.
- [x] #5 The two version-matrix rows recorded as 'not probed' are either probed or explicitly marked out of roster scope with a reason.
- [x] #6 The corpus runner reports all probes passing, and `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass.
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
## Implementation plan (2026-09-23)

Scope stays in `docs/gala-translation.md` and `server/scripts/20260923_gala_translation_probes/`; no runtime, twin, or server Go changes.

### Confirmed decisions to write into the roster (each inline in the entry)

1. `switch-statements`: rewrite to `match` only when every case is a literal pattern (`case "x"`, `case 0`, `case _`); an identifier `case` binds instead of comparing, and init/fallthrough stay handwritten.
2. `defer`: single-value acquire -> `use x = acquire`; named-resource release -> `resource.Using`/`Bracket` (mutexes: `resource.WithLock`); panic recovery or named-return mutation -> keep handwritten.
3. `panic-calls`: recoverable failure -> `Try`; unrecoverable invariant in statement position -> `go_builtins.Panic`; tail-position `= Panic(...)` does not compile in 0.81.0.
4. `slice-expressions`: always expressible with `go_interop.Slice(s, from, to)` for bounded ranges and `SliceFrom`/`SliceTo`/`SliceTake`/`SliceDrop` for one-sided ones; no handwritten sibling needed.
5. `const-declarations`: package-level `var` (or `val`) unless the name must stay a Go compile-time constant (exported `const` API, array length), which stays handwritten.
6. `struct-declarations` and `methods`: declare `var` fields for Go-shaped access; direct form for GALA-internal types, handwritten when the Go API must stay plain or the type needs tags/anonymous/embedded fields or a named-scalar receiver.
7. `generic-declarations`: direct form for `any`/`comparable` constraints; union constraints or blocked underlying shapes stay handwritten.
8. `if-initializers`: preceding `var` for a direct port, `Try` when failure should flow as a value; GALA-E0047 after #529.
9. `len-calls`: `.ByteSize()` preserves Go `len` byte semantics, `.Size()` only for intended character counts; collections use `.Size()`.
10. `map-literals`, `delete-calls`, `append-calls`, `slice-literals`, `make-calls`: the container type decides between `go_interop` helpers (Go `map`/`slice`) and collection methods (`HashMap.Put`/`Remove`, `Array.Append`, `ArrayOf`, `EmptyHashMap`).
11. `channel-types`: handwritten when a channel type appears in a signature/field/local; `concurrent.Future`/`go_interop` only when no channel type crosses the boundary.

### Corpus additions (13 probes, 36 -> 49)

- pass: `pass_match_switch_shape`, `pass_resource_using`, `pass_go_builtins_panic`, `pass_try_recover`, `pass_slice_ranges`, `pass_plain_var`, `pass_generic_constraints`, `pass_struct_methods` (plus a runner-written `check.go` fixture), `pass_hashmap_array`, `pass_concurrent_future`.
- build_fail: `blocked_panic_tail_position` (ERR `(no value) used as value`).
- deltas: `delta_for_omitted_init` (break-bounded loop prints `1` in 0.81.0, `0` after #529), `delta_generic_alias` (`IntPair(1, 2)` prints `0` in 0.81.0, `1` after #529).
- Extend `pass_go_interop_map_slice` with `MapDelete` and update its NOTE/expected.out.

### Documentation changes

- Add a `Probe` column to the five roster tables; every branch of a decision rule names its probe.
- Rewrite the rule/verdict cells for the entries above; keep the 0.81.0 vs #529 split intact.
- Correct the two version-matrix rows now probed: the generic-alias row currently claims Go that "does not compile"; 0.81.0 actually emits `IntPair{}` and silently returns zero values.
- Extend the workarounds section with the new rules and add the new probes to the corpus table and "When to revisit" flip list.

### Verification

- `server/scripts/20260923_gala_translation_probes/run.sh` reports 49 passed, 0 failed.
- `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, root `npm run fmt:check`.
- `go build ./...` in `server/` unaffected; re-index with `codebase-memory-mcp cli index_repository --repo-path .`.

Implementation refinement (2026-09-23): the switch rule also documents the manual `if`/`else` comparison chain for identifier cases (the committed `appenv` style) instead of leaving those cases unaddressed; literal patterns still use `match`, and `fallthrough` stays handwritten.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation notes (2026-09-23)

### Decision rules written into the roster

- Every conditional or multi-option verdict in `docs/gala-translation.md` is now an imperative decision rule with a probe per branch, and the five roster tables gained a `Probe` column.
- `switch`: `match` for literal patterns, manual `if`/`else` chain for identifier cases (`appenv` style), `fallthrough` stays handwritten.
- `defer`: `use` for single-value acquires; `resource.Using[*os.File, T]`/`Bracket` with explicit type arguments when the acquire already happened or the release is not `Close()`; panic recovery and named-return mutation stay handwritten.
- `panic`: `Try` for recoverable failures; `go_builtins.Panic` in statement position; tail position keeps the invariant as a statement with an early return.
- `slice-expressions`: always expressible with `Slice`/`SliceFrom`/`SliceTo`/`SliceTake`/`SliceDrop`, so no handwritten sibling is needed.
- `const`: package-level `var`/`val` unless a Go compile-time constant (exported `const` API, array length) is required.
- `struct-declarations`/`methods`: direct form for GALA-internal types with `var` fields for Go-shaped access; handwritten when the Go API must stay plain or the type needs tags/anonymous/embedded fields or a named-scalar receiver.
- `generic-declarations`: direct for `any`/`comparable`; union constraints or blocked underlying shapes stay handwritten.
- `if-initializers`: preceding `var` for a direct port, `Try` for value flow; GALA-E0047 after #529.
- `len-calls`: `.ByteSize()` preserves Go `len` byte semantics; `.Size()` only for intended character counts and for collections.
- `map`/`slice` literals, `delete`, `append`, `make`: the container type decides between `go_interop` helpers and collection methods.
- `channel-types`: handwritten when a channel type crosses a boundary; `concurrent.Future`/`go_interop` otherwise.

### Probe corpus (36 -> 49)

- New pass probes: `pass_match_switch_shape`, `pass_resource_using`, `pass_go_builtins_panic`, `pass_try_recover`, `pass_slice_ranges`, `pass_plain_var`, `pass_generic_constraints`, `pass_struct_methods` (runner-written `check.go` fixture), `pass_hashmap_array`, `pass_concurrent_future`.
- New build_fail probe: `blocked_panic_tail_position` (`(no value) used as value`).
- New delta probes: `delta_for_omitted_init` (`1` in 0.81.0, `0` after #529), `delta_generic_alias` (`0` -> `1`).
- Extended `pass_go_interop_map_slice` with `MapDelete` and `pass_slice_ranges` with `SliceCopy`; `run.sh` materializes the `pass_struct_methods` fixture.

### Findings that corrected the roster

- The go_builtins docstring suggests a tail-position `Panic` lowers to Go's builtin, but 0.81.0 emits void `panic(...)` in value position and `go build` fails; pinned by `blocked_panic_tail_position`.
- `resource.Using(f, (file) => ...)` cannot infer `R`; inference binds the body parameter to `any`, so explicit type arguments are required.
- The version matrix claimed an alias to an instantiated generic struct emits Go that does not compile; it actually emits `Alias{}` and silently returns zero values. Row corrected and pinned by `delta_generic_alias`.

### Verification (objective evidence)

- `server/scripts/20260923_gala_translation_probes/run.sh`: `go go1.26.7; gala 0.81.0; 49 probes` ... `49 passed, 0 failed`.
- `npm run format:markdown` reformatted `docs/gala-translation.md`; `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` all pass.
- `go build ./...` in `server/` is clean; the nested probe module stays out of the server module.
- `codebase-memory-mcp cli index_repository --repo-path .` -> `{"nodes":11138,"edges":44774,"status":"indexed"}`.
- No commit created; the doc, corpus changes, and the new probe directories are in the worktree.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Made every GALA-equivalent verdict in the translation roster deterministic and pinned each rule branch with a corpus probe.

## What changed

`docs/gala-translation.md`: all conditional or multi-option verdicts are now imperative decision rules with named criteria, and the five roster tables carry a `Probe` column. Rules added for `switch` (`match` for literal patterns, manual `if`/`else` for identifier cases), `defer` (`use`, then `resource.Using`/`Bracket` with explicit type arguments, handwritten for panic recovery/named-return mutation), `panic` (`Try` vs `go_builtins.Panic` statement position), slice expressions (always expressible via `go_interop`), `const` (package `var` vs Go compile-time constant), structs/methods (`var` fields, GALA-internal direct form), generics (`any`/`comparable`), `if` initializers (`var` vs `Try`), `len` (`.ByteSize()` vs `.Size()`), map/slice/delete/append/make (container type decides `go_interop` vs collections), and channels (handwritten when the type crosses a boundary). The workarounds section gained the corresponding rule sections, and the upstream-gap and "When to revisit" tables were aligned.

`server/scripts/20260923_gala_translation_probes/`: 13 new probes (10 pass, `blocked_panic_tail_position`, `delta_for_omitted_init`, `delta_generic_alias`), `MapDelete` and `SliceCopy` coverage added, and a runner-written `check.go` fixture for `pass_struct_methods`; the corpus is now 49 probes.

Two findings corrected the roster: tail-position `= Panic(...)` emits Go's void `panic(...)` in value position and fails to build, and `resource.Using` needs explicit type arguments because inference binds the body parameter to `any`. The version-matrix row for aliases to instantiated generic structs now says 0.81.0 emits `Alias{}` and silently returns zero values, matching `delta_generic_alias`.

## Verification

- `server/scripts/20260923_gala_translation_probes/run.sh`: 49 passed, 0 failed.
- `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, root `npm run fmt:check`: pass.
- `go build ./...` in `server/`: clean; the nested probe module stays separate.
- `codebase-memory-mcp cli index_repository --repo-path .`: indexed (11138 nodes, 44774 edges).

## Risks and follow-ups

Expectations pin GALA 0.81.0 and the Go 1.26 series; `delta_for_omitted_init` (`1` -> `0`) and `delta_generic_alias` (`0` -> `1`) join the visible flips once a release carries PR #529. TASK-0323's gated CI run will cover the new probes once it lands. No commit was created; the changes are in the worktree.
<!-- SECTION:FINAL_SUMMARY:END -->
