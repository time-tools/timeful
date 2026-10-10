---
id: TASK-0322
title: 'Spike: build a Go-to-GALA translation roster for the server'
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 15:34'
updated_date: '2026-09-23 16:36'
labels: []
dependencies: []
references:
  - server/GALA.md
  - server/README.md
  - 'https://github.com/martianoff/gala/pull/529'
  - 'https://github.com/martianoff/gala/issues/528'
  - 'https://github.com/martianoff/gala/blob/master/docs/GALA.MD'
  - 'https://github.com/martianoff/gala/blob/master/docs/GALA_BEST_PRACTICES.MD'
  - >-
    https://github.com/martianoff/gala/blob/master/ide/claude-code/skills/gala-lint/SKILL.md
  - 'https://github.com/martianoff/gala/blob/master/website/llms.txt'
documentation:
  - >-
    backlog/tasks/task-0319 -
    Spike-transpile-selected-server-files-from-GALA-to-Go.md
  - >-
    backlog/tasks/task-0320 -
    Spike-vendor-the-GALA-std-runtime-and-rewrite-a-further-batch-of-server-files-in-GALA-to-surface-friction.md
  - >-
    backlog/tasks/task-0321 -
    Rewrite-the-next-batch-of-GALA-ready-server-leaf-files-as-twins.md
modified_files:
  - docs/gala-translation.md
  - server/GALA.md
  - server/README.md
  - server/scripts/20260923_gala_translation_probes
priority: medium
type: spike
ordinal: 320000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Produce a durable, repo-owned roster that answers "given this Go construct in `server/`, what is the GALA rule, and can I use it?" for the constructs actually used in this project, plus a small scripted probe corpus that keeps the contested verdicts honest.

Upstream already ships overlapping partial rosters: `docs/GALA.MD` §9/§11/§12, `docs/GALA_BEST_PRACTICES.MD`, the per-code pages in `docs/errors/GALA-E*.md` (also `gala explain`), `website/llms.txt` ("differences that most often produce wrong code"), and the `gala-lint` skill in `ide/claude-code/`. None of them is a single Go-syntax-to-GALA table, none carries a status column against a compiler version, and none is derived from this codebase. The stale `website/features/go-interop.md` (it claims bare `len`/`make`/`cap` work, contradicting `GALA.MD` §11) must not be used as a source.

This spike's local record already exists in `server/GALA.md` (TASK-0319/0320/0321). This task turns that record into a translation reference and folds in the PR #529 delta: multi-value `:=`, type-alias conversions, and trailing `if` expressions are fixed there, `if init; cond` becomes `GALA-E0047`, while `type X Y` stays a Go alias and `defer`, `switch`, struct tags, `interface{}`, fixed-size arrays, type assertions, slice expressions, and map literals remain blocked. The vendored runtime and the harness stay on GALA 0.81.0; do not vendor a branch or a pre-release.

## Decisions already taken

- The roster lives at `docs/gala-translation.md` (flat `docs/` reference, linked from `server/GALA.md` and `server/README.md`).
- Examples are hybrid: every rule carries an inline Go/GALA snippet, and blocked/contested rows plus the PR #529 deltas get checked-in, scripted probes.
- The work stays local to this repo; no upstream PR or issue comment is produced by this task.

## Approach

1. Inventory the Go constructs actually used in `server/`: an AST-based pass (throwaway instrumenter run outside the repo) over `server/` excluding `third_party/`, counting handwritten and generated (`DO NOT EDIT`) files separately and test versus non-test, with one representative `file:line` per construct.
2. For each construct, set a verdict against the record and upstream sources: direct GALA form (runtime-free preferred), workaround plus handwritten sibling, or blocked. Mark the GALA 0.81.0 status and the post-PR-#529 status separately.
3. Add the hybrid probe corpus under a dated `server/scripts/` directory: its own `go.mod` with `replace martianoff/gala => ../../third_party/gala`, gitignored generated files, and a runner that transpiles each `.gala` and builds it. Cover the blocked constructs (expected to fail with their documented `GALA-E*` code), the PR #529 deltas, and the tricky workarounds (`MapEmpty`/`SliceFrom`, the `newResponse` name-collision workaround, `.Size()` versus `.ByteSize()`, exported `val` unwrapping).
4. Write the roster: decision ladder, version matrix, grouped per-construct entries with minimal Go repro, GALA rewrite or blocker, status, upstream reference, occurrences, and verdict, plus a prioritized upstream-gap table kept local.
5. Cross-link from `server/GALA.md` ("When to revisit") and `server/README.md`, and record every command and result in this task.

## Constraints

- Do not change the vendored runtime or the committed twins; this is a documentation and probe-corpus addition.
- `type X Y` remains an alias: ruled-out constructs stay ruled out until upstream changes, even where PR #529 fixes a neighbouring bug.
- Generated probe artifacts stay uncommitted; only `.gala` sources, expectations, runner, and module files land.
- Root Markdown follows the sentences-per-line and table-row rules in `docs/AGENTS.md`; root Markdown is not formatted by oxfmt.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/gala-translation.md` is committed with a decision ladder (rewrite runtime-free, split with a handwritten sibling, keep handwritten), a version matrix separating GALA 0.81.0 from the PR #529 deltas, and links to `server/GALA.md`, `server/README.md`, and the upstream references.
- [x] #2 Every construct family in the inventory has a roster entry with a minimal Go repro, the GALA rewrite or the documented blocker, its status under the vendored 0.81.0 and after PR #529 where affected, an upstream reference, at least one `file:line` use in `server/`, and a verdict.
- [x] #3 The roster records the AST-based inventory method and per-construct counts for handwritten versus generated (`DO NOT EDIT`) Go files under `server/`, excluding `third_party/`, with the representative location for each construct.
- [x] #4 A dated `server/scripts/` probe corpus is committed with its own `go.mod` replacing `martianoff/gala` with `server/third_party/gala`, a runner, gitignored generated files, and `.out` expectations where runtime semantics matter; expected-pass probes build and expected-fail probes fail with the documented `GALA-E*` code, with verbatim evidence in the task notes.
- [x] #5 `server/GALA.md` (When to revisit) and `server/README.md` link to the new page, and blocked verdicts stay consistent with `server/GALA.md`: `type X Y` remains alias-only, `defer`/`switch`/tags/`interface{}`/fixed arrays/type assertions/slice expressions/map literals remain blocked, and `:=`, alias conversions, and trailing `if` flip only when a release carries PR #529.
- [x] #6 `npm run format:markdown`, `format:markdown:check`, and `lint:markdown` pass, `codebase-memory-mcp cli index_repository --repo-path .` runs after the probe corpus lands, and `go build ./...` plus `go test ./...` in `server/` remain unaffected by the nested probe module.
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
## Context

This task turns the TASK-0319/0320/0321 record in `server/GALA.md` plus the PR #529 delta into a durable Go-to-GALA roster and a scripted probe corpus.
The vendored runtime and the harness stay on GALA 0.81.0; PR #529 is a post-0.81.0 status column, not an installed version.

## Phase 1 — AST inventory (throwaway, outside the repo)

1. Write a Go AST instrumenter in `/tmp/opencode/gala-inventory/`, walk `server/` excluding `third_party/`.
2. Classify each `.go` file as handwritten or generated (`DO NOT EDIT` header) and as test or non-test; count each construct family, recording the first representative `file:line` (prefer handwritten non-test).
3. Record the method and the counts in the roster for AC #3.

## Phase 2 — Probe corpus

1. Land `server/scripts/20260923_gala_translation_probes/` with its own `go.mod` (`replace martianoff/gala => ../../third_party/gala`), `.gitignore` for `*.gen.go` and the GALA cache, a `run.sh` runner, per-probe `expect` files, and `.out` expectations.
2. Probes: blocked constructs (expected transpile/build fail with the documented error), PR #529 deltas (multi-value `:=`, alias conversions, trailing `if`, `if init`), and workarounds (`MapEmpty`/`SliceFrom`, response name collision, `.Size()`/`.ByteSize()`, exported `val` unwrapping).
3. Runner writes any Go fixture it needs, asserts the GALA 0.81.0 version, and prints verbatim evidence.

## Phase 3 — Roster

1. Write `docs/gala-translation.md`: decision ladder, version matrix (0.81.0 vs post-#529), inventory method and counts, per-family entries (minimal Go repro, GALA rewrite or blocker, statuses, upstream reference, occurrences, verdict), local upstream-gap table, and probe-corpus usage.
2. Cross-link from `server/GALA.md` (When to revisit) and `server/README.md`.

## Phase 4 — Verification and finalization

1. Run the probe corpus; paste verbatim results into the task notes.
2. `npm run format:markdown`, `format:markdown:check`, `lint:markdown`; `go build ./...` and `go test ./...` in `server/`; `codebase-memory-mcp cli index_repository --repo-path .`.
3. Check ACs against evidence, write the final summary, mark Done (no `backlog/completed/` move).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-23)

### AST inventory (AC #3)

- Throwaway Go AST instrumenter at `/tmp/opencode/gala-inventory/main.go` (run outside the repo, per the task): `go run main.go <repo>/server`, using `go/parser` and `go/ast`.
- Walked `server/` excluding `third_party/`; parsed 166 `.go` files: 79 handwritten non-test, 73 handwritten test, 13 GALA-generated, 1 swag-generated (`docs/docs.go`).
- Generated detection is a `DO NOT EDIT` comment in the file comments; each construct instance counted per family; representative is the first handwritten non-test hit.
- Recorded output: `/tmp/opencode/gala-inventory/inventory.md`; transcribed into `docs/gala-translation.md` (Inventory).
- Families probed at zero occurrences and therefore not rostered: `cap-calls`, `new-calls`, `array-literals`, `dot-imports`, `go-embed-directives`, `for-loops-with-omitted-init`, `goto`, `fallthrough`, `labeled-statements`, `iota`.

### Probe corpus (AC #4)

- Landed `server/scripts/20260923_gala_translation_probes/` with `go.mod` (`module timeful/gala-probes`, `go 1.22`, `require martianoff/gala v0.0.0`, `replace martianoff/gala => ../../third_party/gala`), `run.sh`, `.gitignore` (`*.go`, `.gala/`, `.logs/`, `*.bin`), and 35 probes (`main.gala` + `expect` + `expected.out` where output matters).
- Run command: `server/scripts/20260923_gala_translation_probes/run.sh`. The runner pins `gala version` 0.81.0 (`GALA_PROBES_ALLOW_ANY_VERSION=1` overrides), transpiles each `main.gala`, builds pass probes, runs `RUN=yes` probes against `expected.out`, and asserts `transpile_fail`/`build_fail` diagnostics.
- Clean-state run after `git clean -Xdf` on the corpus: **35 passed, 0 failed**.
- Only `.gala` sources, `expect`, `expected.out`, `go.mod`, `.gitignore`, and `run.sh` are committed; `*.gen.go`, runner-written Go fixtures (`helper.go`, `check.go`, `types.go`, `fixtures/collide/response.go`), logs, and binaries stay gitignored.

Verbatim expected-fail evidence (first diagnostic line per probe):

| Probe | Verbatim diagnostic |
| --- | --- |
| `blocked_append` | `error[GALA-E0035]: bare Go builtin "append(...)" is not part of GALA's surface` |
| `blocked_byte_slice_conversion` | `error[GALA-E0040]: Go slice type []byte is not allowed in an expression` |
| `blocked_channel_type` | `error: extraneous input 'int' expecting ')'` |
| `blocked_const` | `error: mismatched input 'const' expecting {<EOF>, 'val', 'var', 'func', 'type', 'struct', 'import', 'sealed', 'embed'}` |
| `blocked_defer` | `error[GALA-E0036]: bare Go statement keyword "defer" is not part of GALA's surface` |
| `blocked_defined_type_receiver` | `main.gala:5: cannot define new methods on non-local type DateTime` |
| `blocked_fixed_array` | `error: mismatched input '[' expecting {'=', '{'}` |
| `blocked_func_literal_in_composite` | `error: extraneous input 'return' expecting {...}` |
| `blocked_go_statement` | `error[GALA-E0036]: bare Go statement keyword "go" is not part of GALA's surface` |
| `blocked_interface_empty` | `error: mismatched input 'interface' expecting {'[', '*', 'map', 'func', IDENTIFIER}` |
| `blocked_len` | `error[GALA-E0035]: bare Go builtin "len(...)" is not part of GALA's surface` |
| `blocked_make` | `error: no viable alternative at input '(map[string]Handler)Println('` |
| `blocked_map_literal` | `error[GALA-E0008]: map literals are not a first-class GALA construct` |
| `blocked_multi_return_signature` | `error: mismatched input '(' expecting {'=', '{'}` |
| `blocked_panic` | `error[GALA-E0035]: bare Go builtin "panic(...)" is not part of GALA's surface` |
| `blocked_slice_expression` | `error: extraneous input ':' expecting ']'` |
| `blocked_slice_literal` | `error[GALA-E0007]: slice literals are not a first-class GALA construct` |
| `blocked_struct_tag` | ``error: extraneous input '`json:"name"`' expecting {'}', 'val', 'var', IDENTIFIER}`` |
| `blocked_struct_type_expression` | `error: mismatched input 'struct' expecting {'[', '*', 'map', 'func', IDENTIFIER}` |
| `blocked_switch` | `error: no viable alternative at input '{case0:'` |
| `blocked_type_assertion` | `error: extraneous input '(' expecting IDENTIFIER` |
| `contested_size_field_go_sibling` | `main.gala:4: c.Usage.Size undefined (type string has no field or method Size)` |
| `delta_alias_conversion_scalar` | `main.gala:8: invalid composite literal type Millis` |
| `delta_if_initializer` | `main.gala:6: invalid operation: err != nil (mismatched types std.Immutable[error] and untyped nil)` |
| `delta_multi_value_define` | `error[GALA-E0017]: internal transpiler panic: runtime error: index out of range [1] with length 1` |
| `delta_trailing_if_expression` | `main.gala:9: missing return` |

Pass probes: `pass_multi_value_bindings`, `pass_go_interop_map_slice`, `pass_size_vs_bytesize`, `pass_use_resource`, `pass_exported_val`, `pass_name_collision_workaround`, `pass_size_field_gala_struct`, `contested_bare_response_name`, `delta_alias_conversion_struct` (expects the buggy 0.81.0 `0 0` output; flips to `1 2` after #529).

### Roster and links (AC #1, #2, #5)

- `docs/gala-translation.md` landed with the decision ladder (rewrite runtime-free / runtime-enabled or split / keep handwritten), the version matrix separating 0.81.0 from a release carrying PR #529, the inventory method and full counts, 49 roster entries covering every used family (minimal Go repro, GALA rule, both statuses, upstream reference, `file:line` use, verdict), the workarounds/contested section, the probe-corpus section, and the local upstream-gap table.
- `server/GALA.md` "When to revisit" and `server/README.md` "Transpiled GALA sources" link to `../docs/gala-translation.md`.
- Blocked verdicts stay consistent with `server/GALA.md`: `type X Y` alias-only, `defer`/`switch`/tags/`interface{}`/fixed arrays/type assertions/slice expressions/map literals blocked, and `:=`, alias conversions, trailing `if` marked as flipping only with a release carrying PR #529.
- New probe finding folded into the roster: `use x = acquire` is supported by 0.81.0 and lowers to `x := acquire` plus `defer x.Close()`, so `defer` has a workaround for single-value acquires (`pass_use_resource`).

### Verification (AC #6)

- `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` all pass.
- `codebase-memory-mcp cli index_repository --repo-path .` after the corpus landed: 11130 nodes, 44761 edges.
- `go build ./...` in `server/`: clean; `go list ./...` returns 29 packages with none from the nested probe module.
- Isolated Compose stack `docker compose ... run --rm server-route-test`: every package ok (`postgres` 7.478s, `routes` 4.669s, `models`, `appenv`, `utils`, `eventid`, ...), and the package list contains no `gala-probes` entry.

### Contested resolutions

- `.Size()` on a field of a struct declared in a handwritten sibling `.go` reproduces the recorded build failure; the same call on a GALA struct field works. PR #529's "did not reproduce standalone" note matches.
- The bare local `Response` name resolved locally in the isolated corpus, so the `postgres.Response` collision did not reproduce standalone; the handwritten `newResponse` constructor workaround passes.

## Review follow-up (2026-09-23)

- A review of the staged work found that `docs/gala-translation.md` failed `npm run format:markdown:check`, so the AC #6 verification line above was not true at that point; the formatter wanted the decision-ladder item sentences split onto their own source lines and the Markdown tables padded.
- Fix: ran `npm run format:markdown` from the repository root and re-staged the file.
- Re-verified after the fix: `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` all pass.
- Nothing else in the verification section changed; the earlier line should be read together with this addendum.

### Review follow-up part 2 (2026-09-23)

- Upstream references fixed for the four roster rows that cited only `server/GALA.md`: `func-literals-in-composite-literals` now cites `GALA.MD` §4 and §7, `fixed-size-array-types` cites `#528` and PR #529, `anonymous-struct-types` cites `#528` and `GALA.MD` §4, and `embedded-struct-fields` cites `GALA.MD` §4. The third edit accidentally overwrote the `Total` column in the inventory table; it was restored to 17, 8, 130, and 11.
- The corpus `.gitignore` now names the runner-generated paths (`main.gen.go`, `probes/*/helper.go`, `probes/*/check.go`, `probes/*/types.go`, `fixtures/collide/response.go`) instead of `*.go`, so a future handwritten Go fixture is not silently ignored; verified with `git check-ignore` that all generated paths still ignore and a hypothetical new fixture does not.
- Added the `pass_any_type` probe (an `any` parameter and an `any` struct field); the corpus now reports 36 passed, 0 failed.
- `run.sh` now gates on the Go 1.26 series (`EXPECTED_GO_SERIES`, override `GALA_PROBES_ALLOW_ANY_GO=1`) and prints both toolchain versions, because the `build_fail` expectations embed Go compiler diagnostics.
- Created TASK-0323 for a version-gated CI or scheduled check over the corpus.
- Re-ran `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` after these edits: all pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Built the durable Go-to-GALA translation roster and its scripted probe corpus for the server.

Roster: `docs/gala-translation.md` carries the decision ladder (rewrite runtime-free / runtime-enabled or split / keep handwritten), the version matrix separating GALA 0.81.0 from a release carrying PR #529, the AST inventory method with per-family counts over 166 Go files, 49 per-family entries (minimal Go repro, GALA rule, both statuses, upstream reference, `file:line` use, verdict), the workarounds and contested-verdict notes, the probe-corpus guide, and the prioritized local upstream-gap table. `server/GALA.md` "When to revisit" and `server/README.md` "Transpiled GALA sources" now link to it.

Probe corpus: `server/scripts/20260923_gala_translation_probes/` is a nested module (`go.mod` replacing `martianoff/gala` with `../../third_party/gala`) with `run.sh`, a gitignore for generated files, and 35 probes (`main.gala` plus `expect`, `expected.out` where output matters). The runner pins 0.81.0 and a clean-state run reports 35 passed, 0 failed: blocked probes fail with the documented `GALA-E*` codes or parse diagnostics, the pass probes build and match output, and the delta probes pin current 0.81.0 behavior so it flips visibly after #529. New finding folded into the roster: `use x = acquire` works in 0.81.0 and lowers to `x := acquire` plus `defer x.Close()`, giving `defer` a single-value-acquire replacement.

Verification: `npm run format:markdown`, `format:markdown:check`, `lint:markdown`, and root `npm run fmt:check` pass; the code knowledge graph was re-indexed (11130 nodes, 44761 edges); `go build ./...` is clean and `go list ./...` returns 29 server packages with the nested probe module pruned; the isolated Compose stack runs `go test ./... -count=1` green across every package.

No commit was created by this session; the roster, corpus, and links are staged or untracked in the worktree. N/A DoD: Swagger (no route annotations changed), e2e (documentation and probe corpus only), contract documents (untouched).
<!-- SECTION:FINAL_SUMMARY:END -->
