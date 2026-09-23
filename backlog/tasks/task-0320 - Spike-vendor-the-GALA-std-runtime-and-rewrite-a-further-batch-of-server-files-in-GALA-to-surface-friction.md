---
id: TASK-0320
title: >-
  Spike: vendor the GALA std runtime and rewrite a further batch of server files
  in GALA to surface friction
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 09:42'
updated_date: '2026-09-23 10:16'
labels: []
dependencies: []
modified_files:
  - server/third_party/gala
  - server/go.mod
  - server/Dockerfile
  - server/GALA.md
  - server/README.md
  - server/logger/logger.gala
  - server/logger/logger.go
  - server/appenv/appenv.gala
  - server/appenv/appenv.go
  - server/appenv/appenv_port.go
  - server/utils/array_utils.gala
  - server/utils/array_utils.go
  - server/utils/array_utils_extra.go
priority: medium
type: spike
ordinal: 318000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up to TASK-0319, whose durable record is server/GALA.md. That spike transpiled two leaf packages from GALA to Go and identified blockers; the top blocker was the martianoff/gala std runtime, which generated code imports for val, monads, and collections but server/go.mod could not resolve. The runtime exists as ready-to-compile Go in the installed stdlib snapshot (the tree the gala CLI extracts at ~/.gala/stdlib/v0.81.0, containing every stdlib package as .gen.go plus handwritten Go). The sibling checkout at ../gala/gala is upstream source whose .gen.go files only exist as bazel outputs and track master, so it is the wrong vendor source for version parity.

Outcome: the runtime is vendored into the repo so runtime-using transpiled output compiles in the server module, a further batch of server files is rewritten in GALA to exercise the feature areas TASK-0319 left untested, and the findings are recorded in server/GALA.md.

Scope:
- Vendor the v0.81.0 stdlib snapshot as one flattened Go module martianoff/gala under server/third_party/gala/, wired into server/go.mod with a single local replace. Flattening means one root go.mod and no nested per-package go.mod, so the import paths generated code emits (martianoff/gala/std, martianoff/gala/go_interop, martianoff/gala/go_builtins, ...) resolve as subdirectories.
- Rewrite batch, simple to hard: server/models/location.go (struct decl plus JSON tags), server/models/set.go (generic type decl over a Go map), server/logger/logger.go (interface-typed params, package pointer vars), server/errs/errors.go (struct with methods implementing error, interface fields, const block), server/appenv/appenv.go (named string type, typed consts, match on consts, multi-value ResolvePort), server/utils/array_utils.go (generic functions and generic struct, go_interop make/append, three-value FindAddedRemovedKept), server/models/datetime.go (named type with value and pointer methods, json.Marshaler), server/models/uuid.go (multi-value ParseUUID, arrays, bit ops, strings.Builder, switch multi-case, panic).
- Each file lands one of three outcomes: full twin committed beside the original package with the exported Go API identical; split where multi-value-return members stay in a handwritten sibling .go and the rest transpiles (validating the mixed .gala/.go package pattern); or blocked, keeping the original Go and recording the exact transpiler error verbatim as friction evidence.
- Scratch probes in /tmp (findings only, nothing committed) closing the Not yet exercised list from server/GALA.md against the now-vendored runtime: sealed types, val with Option/Try/Either including Try wrapping a Go multi-return and the std.Immutable effect on exported signatures, collection_immutable Array/HashMap, and concurrency via go_interop.Spawn or Future.
- Record everything in server/GALA.md following docs/AGENTS.md authoring rules (one sentence per source line, one table row per line).

Out of scope:
- Changing any exported Go API or behavior; existing Go tests must compile and pass unchanged.
- Regenerating or restyling the existing eventid and observability twins; they stay runtime-free.
- CI integration for regeneration, upstream GALA changes, or vendoring anything besides the v0.81.0 stdlib snapshot.

Known risks to resolve during execution and record either way: the docs say type X Y transpiles to a Go alias (type X = Y), which would make methods on named types like UUID or DateTime illegal Go receivers and block datetime, uuid, and set; backtick JSON struct tags may not parse, blocking errs and location; a stdlib package may pull an external module import, in which case fix the flattened go.mod requires or drop the package and document why.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The v0.81.0 GALA stdlib snapshot is vendored under server/third_party/gala/ as one flattened Go module named martianoff/gala (per-package go.mod/go.sum removed, upstream LICENSE kept) and go build ./... passes over the tree
- [x] #2 server/go.mod resolves the runtime with require martianoff/gala v0.0.0 plus replace martianoff/gala => ./third_party/gala and the server still builds in the Compose Docker path unchanged
- [x] #3 Every file in the rewrite batch is attempted and lands one of three recorded outcomes: a committed .gala plus generated twin, a split with multi-value members kept in a handwritten sibling .go, or blocked with the exact transpiler error recorded verbatim in server/GALA.md
- [x] #4 Scratch probes for sealed types, val with Option/Try/Either, collections, and concurrency each record the transpile result, whether the generated Go compiles in the module, and the API shape Go callers would see
- [x] #5 server/GALA.md is updated: Current usage table with new packages and split members, two-mode ground rules covering runtime-free and runtime-enabled output, works/does-not-work extended with batch results, Not yet exercised replaced by probe results, a Vendoring section with provenance and update procedure, and refreshed Verification and When to revisit sections
- [x] #6 Existing package tests (appenv, utils, models including wire_shape) pass unchanged against any committed twins; each twin regenerates byte-identically (sha256); go build ./... and go vet on touched packages pass; the isolated Compose stack runs go test ./... -count=1 green
- [x] #7 server/README.md Transpiled GALA sources section lists every committed twin with its regeneration command
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

TASK-0319 landed two GALA twins (eventid, observability/redact) and recorded findings in server/GALA.md.
Its top blocker was the martianoff/gala std runtime: val, monads, and collections emit imports that server/go.mod could not resolve.
The runtime exists as ready-to-compile Go — ~/.gala/stdlib/v0.81.0/ is the complete transpiled snapshot (every stdlib package as .gen.go plus the handwritten Go, about 1.8 MB, version-matched to GALA 0.81.0).
The sibling checkout ../gala/gala is upstream source whose .gen.go files only exist as bazel outputs and track master rather than 0.81.0, so the vendor source is the v0.81.0 snapshot.

Vendoring that snapshot lifts the runtime blocker and lets the rewrite batch probe the Not yet exercised list (structs and methods, sealed types, generics, collections, error monads, concurrency) against a resolvable runtime.

## Phase 0 — Task setup

1. This task carries the plan; mark it In Progress and assign it before implementation, per the Executing Tasks section of BACKLOG_WORKFLOW.md.

## Phase 1 — Vendor the runtime (server/third_party/gala/)

1. Copy ~/.gala/stdlib/v0.81.0/* into server/third_party/gala/ — 20 packages: collection_immutable, collection_mutable, concurrent, crypto, fs, go_builtins, go_interop, io, json, lazy, path, regex, resource, std, stream, strings, subprocess, test, time_utils, validation, yaml.
2. Flatten to one Go module martianoff/gala: write a single root go.mod (module martianoff/gala; go 1.22) and delete each package's own go.mod and go.sum.
Nested modules would carve the packages out of the parent, and the snapshot's per-package replaces (martianoff/gala/std => ../std and similar) only work inside its multi-module layout.
3. Keep the upstream LICENSE (Apache-2.0) at server/third_party/gala/LICENSE.
4. Import-path check: generated code emits martianoff/gala/std, martianoff/gala/go_interop, martianoff/gala/go_builtins, martianoff/gala/collection_immutable and similar paths, which resolve as subdirectories of the flattened module.
5. server/go.mod: add require martianoff/gala v0.0.0 and replace martianoff/gala => ./third_party/gala.
A local replace needs no go.sum entry.
6. Docker needs no change: the Compose build context and the server-route-test mount are ./server, so third_party/ comes along, and go mod download skips replaced modules.
7. Run go build ./... and go vet ./... over third_party/gala to confirm the flattened tree compiles.
Watch the test and subprocess packages for stray external imports (the likely module-dependency candidates); add the require to the flattened go.mod if trivial, otherwise drop the package and record why in GALA.md.
8. Pure plumbing proof: nothing imports the runtime yet, so the server test suite must still pass at this point.
9. Provenance is recorded in GALA.md in Phase 4: source tree ~/.gala/stdlib/v0.81.0 (extracted by the gala CLI, built from martianoff/gala 0.81.0), the flattening steps, and the update procedure (re-copy, re-flatten, rebuild, re-run tests, regenerate twins).

## Phase 2 — Rewrite batch (simple to hard)

Per-file procedure: write <name>.gala beside the original, transpile from the package directory (cd server/<pkg> && gala transpile -i <name>.gala -o <name>.go) so the //line directives stay relative, diff the exported API against the original, then run the package tests.

Every file lands one of three outcomes:

- full twin — commit the .gala source and its generated .go (replacing the original file) with the exported Go API unchanged in shape; the .gala source is the only hand-edited member.
- split — multi-value-return members stay in a handwritten sibling .go (for example datetime_json.go) and the rest transpiles; both files are committed.
This validates the mixed .gala/.go package pattern, which is the practical adoption path.
- blocked — keep the original Go file untouched and record the exact transpiler error verbatim in server/GALA.md as friction evidence.

| File | Probes | Expected friction |
| --- | --- | --- |
| models/location.go | struct decl + JSON tags | do backtick tags survive transpile? |
| models/set.go | type Set[T comparable] map[T]struct{} | generic type decls; alias-vs-defined-type emit |
| logger/logger.go | io.Writer params, package *log.Logger vars | interface-typed params |
| errs/errors.go | struct + method + error impl, interface{} fields, const block | tags again, pointer receivers, const to var |
| appenv/appenv.go | named string type, typed consts, match on consts | known const-matching friction; ResolvePort (string, error) forces a split |
| utils/array_utils.go | Find[T any], generic struct, make/append via go_interop | generics; FindAddedRemovedKept three-value return forces a split |
| models/datetime.go | named type + value/pointer methods, json.Marshaler | multi-value MarshalJSON () ([]byte, error) forces a split |
| models/uuid.go | kitchen sink: ParseUUID (UUID, bool), [16]byte, bit ops, strings.Builder, switch multi-case, panic | everything above plus byte and range grammar gaps |

Ordering matters: location and set prove a mixed generated/handwritten models package works before the datetime and uuid splits.

Per-file detail:

1. models/location.go (13 lines) — Location struct with seven backtick-tagged fields and float64 members.
If tags do not survive transpile the wire shape changes, so the file is blocked and that becomes the headline JSON-tag finding.
2. models/set.go (3 lines) — type Set[T comparable] map[T]struct{}, used across routes and services.
Tiny probe of generic type declarations and of what the transpiler emits for type X Y (Go alias versus Go defined type).
3. logger/logger.go (17 lines) — Init(io.Writer) building two *log.Logger package vars via io.MultiWriter and log.New flags.
Probes interface-typed parameters and pointer-typed package vars.
4. errs/errors.go (44 lines) — GoogleAPIError struct with JSON tags and interface{} fields, its Error() string method (pointer receiver implementing error), a json.Marshal call inside it, and a const block of exported error-code strings that must become vars.
5. appenv/appenv.go (78 lines) — Environment named string type, four typed consts, Parse/Current/IsProductionLike/ShouldUseReleaseMode/Port as match-based helpers, and ResolvePort (env Environment, override string) (string, error).
Expect consts to vars, the string-const match friction TASK-0319 already recorded, and ResolvePort forcing a split into a handwritten sibling .go.
6. utils/array_utils.go (75 lines) — Find[T any], ArrayToSet[T comparable] returning models.Set[T], ElementWithIndex[T any] generic struct, and FindAddedRemovedKept returning three []ElementWithIndex[T] values.
Probes generics, generic structs, cross-package generic use, make/append through go_interop (now resolvable), and the three-value return that forces the split.
7. models/datetime.go (41 lines) — DateTime int64 named type with Time() and IsZero() value methods, MarshalJSON/UnmarshalJSON pointer methods implementing json.Marshaler and json.Unmarshaler, and the NewDateTimeFromTime constructor.
MarshalJSON () ([]byte, error) forces the split: marshalers stay in a handwritten datetime_json.go while the type and scalar helpers transpile; wire_shape_test.go guards the output.
8. models/uuid.go (133 lines) — UUID string named type with String/IsZero/MarshalJSON/MarshalText/UnmarshalJSON/UnmarshalText, ParseUUID (UUID, bool), NewUUID over a [16]byte with crypto/rand.Read multi-receive and bit ops (value[6] & 0x0f and similar), strings.Builder with Grow/WriteByte, a switch on the loop index with multi-value cases (8, 13, 18, 23), and panic.
Expect the multi-value and byte/range grammar gaps to force a split or a block; uuid_test.go and wire_shape_test.go guard the output.

## Phase 3 — Scratch probes in /tmp/opencode (findings only, nothing committed)

Close the Not yet exercised list from server/GALA.md against the now-vendored runtime.
For each probe record: the transpile result verbatim (including any GALA-E codes), whether the generated Go compiles in the server module against the vendored runtime (a scratch package wired with the same replace), and the API shape Go callers would see.

9. Sealed types — a small sealed type (for example Shape with Circle and Rectangle variants) exercising construction and exhaustive match.
Inspect the generated _variant discriminator, companion structs, Apply/Unapply methods, and the std.Equal dispatch inside the synthesized Copy/Equal.
10. val with Option/Try/Either — package-level and function-local val (which std.Immutable wrapper appears in generated signatures and whether exported API shapes survive), Some/None matching, Try(strconv.Atoi(...)) auto-wrapping a Go multi-return into Try[int], Either error paths, and FlatMap chaining.
11. Collections — collection_immutable Array and HashMap: construction (ArrayOf/HashMapOf), Map/Filter/FoldLeft, GetOption, and how .Size() lowers.
12. Concurrency — go_interop.Spawn and concurrent Future: what generated Go looks like (goroutines, channels, Sendable marker erasure) and whether the documented concurrency checks fire (for example GALA-E0037 on an unshareable capture).

## Phase 4 — Record findings in server/GALA.md

Follow the docs/AGENTS.md authoring rules: one sentence per source line, one table row per line.

- Current usage table: every new committed twin and split member (for example models/datetime.gala plus models/datetime.go plus models/datetime_json.go) with its regeneration command.
- Ground rules section: becomes two-mode.
Runtime-free style (var bindings, no std imports — what eventid and redact use today) stays preferred for leaf twins; runtime-enabled style (val, monads, collections, go_interop) is now committable because the module resolves.
Document when to choose which.
- What works well and What does not work: extend with the batch results verbatim.
The multi-value-return entry gains the split-file workaround; the runtime-dependency entry moves to resolved-by-vendoring (keeping the historical note); new frictions are added with verbatim transpiler errors — JSON struct tags, alias-versus-defined-type emit, interface{} fields, const, panic and byte indexing, switch multi-case.
- Not yet exercised: replaced by the probe results 9 to 12, or narrowed to whatever remains genuinely untried.
- New Vendoring section: provenance (source tree, upstream version 0.81.0, Apache-2.0), the flattened layout and why (per-package go.mod removal; a single replace), and the update procedure (re-copy, re-flatten, rebuild, re-run tests, regenerate twins).
- Verification section: the Phase 5 evidence.
- When to revisit: multi-value Go returns remain the top upstream ask; refresh the remaining items against what the batch resolved.

## Phase 5 — Verification and finalization

1. Every committed twin regenerates byte-identically: run each regeneration command twice and compare sha256.
2. go build ./... and go vet on the touched packages pass (server root plus any split packages).
3. Existing package tests pass unchanged against the twins: appenv_test.go, utils_test.go, datetime_test.go, uuid_test.go, wire_shape_test.go — the wire-shape-sensitive files.
Differential testing is added only where test coverage is thin; the uuid and datetime wire shapes are already covered.
4. Full backend go test ./... -count=1 in the isolated Compose test stack (the canonical commands in server/README.md).
5. DoD hygiene: npm run format:markdown for the GALA.md change and codebase-memory-mcp cli index_repository --repo-path . for the code change.
Swagger regeneration, scripts/prettier fmt:check, and contract documents are N/A unless something unexpectedly touches them (no route, window.postMessage, or environment-contract changes are in scope).
6. Update the server/README.md Transpiled GALA sources section to list every committed twin and split member with its regeneration command.
7. Keep task notes current through execution and record fail/pass evidence per outcome in the final summary, then mark Done.
Moving Done tasks to backlog/completed/ stays a separate periodic step per BACKLOG_WORKFLOW.md.

## Risks / open items to resolve during execution and record either way

- Alias versus defined type: GALA.MD says type X Y transpiles to a Go alias (type X = Y).
If so, type UUID string with func (id UUID) String() is illegal Go (non-local receiver), and datetime, uuid, and set are blocked outright — that becomes a headline finding rather than a twin.
- Backtick JSON struct tags may not parse or may not survive onto the generated struct; errs and location depend on them, and the models wire shapes depend on them throughout.
- A stdlib package (test and subprocess are the candidates) may pull an external module import into the flattened tree; handle by fixing the flattened go.mod requires, or drop the package and document why.
- Mixed .gala/.go package mechanics: if the transpiler refuses to emit beside handwritten .go members, or go build rejects the combination, splits degrade to blocked and the adoption-path finding changes accordingly.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Progress (2026-09-23 session, status In Progress)

### Completed

Phase 1 — vendoring (AC #1, #2 largely done):
- `server/third_party/gala/` is one flattened Go module `martianoff/gala` (20 packages, per-package go.mod/go.sum deleted, root go.mod declares `module martianoff/gala` + `go 1.22`, Apache-2.0 LICENSE copied from `../gala/gala/LICENSE`).
- `go build ./...` is clean inside the vendored tree; `go vet ./...` there reports 5 pre-existing upstream findings (ReadRune signature hints and unreachable code in std/json/yaml) — upstream noise, not introduced.
- The vendored tree is already committed as `059f7b62 feat(server): vendor third_party/gala` (author Danila Danko; this session did not create that commit).
- `server/go.mod` has `require martianoff/gala v0.0.0` (direct block) + `replace martianoff/gala => ./third_party/gala`. No committed code imports the runtime yet, so `go mod tidy` would drop the require; there is no CI tidy check, and the require is kept deliberately for AC #2.
- `server/Dockerfile`: added `COPY third_party/gala/go.mod third_party/gala/` before `RUN go mod download` in the `testdeps` stage. This is required because a local replace must resolve inside that dependency layer; verified with `docker compose --env-file .env.development -f compose.yaml -f compose.development.yaml build server` (image built, binary produced).
- Compose config itself is unchanged; the test stack mounts `./server`, so `third_party/` is present.

Phase 2 — rewrite batch (AC #3 done):
- Full twin: `logger/logger.gala` → generated `logger/logger.go` (runtime-free).
- Splits (both compile, tests pass):
  - `appenv/appenv.gala` → generated `appenv/appenv.go` + handwritten `appenv/appenv_port.go` (`ResolvePort`, multi-value).
  - `utils/array_utils.gala` → generated `utils/array_utils.go` (`Find`) + handwritten `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`; blocked by missing `struct{}` type expression, GALA struct API growth, and multi-value returns).
- Blocked, original Go untouched, verbatim transpiler errors recorded in `server/GALA.md`: `models/location.go` (backtick JSON tags), `models/set.go` (`map[T]struct{}`), `errs/errors.go` (`const` + tags + `interface{}`), `models/datetime.go` (alias emit + multi-value), `models/uuid.go` (`const`, alias, multi-value, `[16]byte`, Go `switch` multi-case).

Phase 3 — probes (AC #4 done, findings only in `/tmp`):
- Sealed types, `val`/`Option`/`Try`/`Either`, `collection_immutable` Array/HashMap, `go_interop.Spawn`/`concurrent.Future`: all transpile clean, compile in a scratch module replacing `martianoff/gala` with `server/third_party/gala`, and run.
- Extra evidence recorded: `GALA-E0037` on reassignable var capture in `Future`; `GALA-E0025` when a sibling `.gala` file lacks its own import; `GALA-E0044` for `HashMap.GetOption` (correct method is `Get`); exported package-level `val` becomes `std.Immutable[T]`.
- Scratch evidence kept at `/tmp/opencode/gala-probes/phase3` and `/tmp/opencode/blocked-probes`.

Phase 4 — docs (AC #5, #7 done):
- `server/GALA.md` rewritten (two-mode ground rules, extended works/does-not-work with verbatim errors, Runtime probes table + details, Vendoring section with provenance/layout/update procedure, Verification, refreshed When to revisit).
- `server/README.md` Transpiled GALA sources section lists all five twins with regeneration commands plus the two handwritten siblings.
- `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` all pass.

Phase 5 — verification already run (AC #6 evidence):
- Isolated Compose stack `go test ./... -count=1`: all packages green (`routes`, `postgres`, `models`, `appenv`, `utils`, ...).
- `go build ./...` and full `go vet ./...` pass on the server module.
- All five committed twins regenerate byte-identically; two runs each, sha256 recorded in the GALA.md Verification table.

### Remaining work

1. DoD #6 refresh is not yet run: `codebase-memory-mcp cli index_repository --repo-path .` (binary resolves from the nix devShell).
2. Commit state: everything except `third_party/gala` is currently staged, not committed — `server/Dockerfile`, `server/go.mod`, `server/GALA.md`, `server/README.md`, `server/logger/{logger.gala,logger.go}`, `server/appenv/{appenv.gala,appenv.go,appenv_port.go}`, `server/utils/{array_utils.gala,array_utils.go,array_utils_extra.go}`, `backlog/backlog.md`, and this task file. `third_party/gala` is in commit `059f7b62`. This session made no commit.
3. Task finalization (Phase 5.7): verify/check off the seven acceptance criteria against the evidence above, complete remaining DoD items, write the final summary, and mark the task Done. Do not move it to `backlog/completed/` (separate periodic step).
4. Swagger and `scripts/`/`prettier/` fmt:check are N/A (no route or script changes); contract documents were not touched.

### Environment notes for the next session

- GALA CLI v0.81.0 (`~/.nix-profile/bin/gala`); snapshot at `~/.gala/stdlib/v0.81.0`.
- Regeneration commands and expected hashes are in `server/GALA.md` and `server/README.md`.
- `server/.gala/` is the gitignored transpiler analysis cache.
- Diff review caveat: `server/go.mod` intentionally carries an unused `require martianoff/gala v0.0.0`; do not run bare `go mod tidy` or the require disappears.

## Continuation session (2026-09-23, session 2)

Phase 5 evidence re-run against the committed state (`5a6ff936`):
- All five twins regenerated in place; worktree stayed clean, so output is byte-identical to the committed files.
- sha256 of all five generated files matches the table in `server/GALA.md`.
- `go build ./...` and `go vet ./logger/ ./appenv/ ./utils/` pass on the server module.
- Vendored tree: `go build ./...` clean; `go vet ./...` reports exactly the five documented upstream findings.
- Isolated Compose stack `go test ./... -count=1` green across every package.
- `npm run format:markdown:check`, `npm run lint:markdown`, and `npm run fmt:check` pass.
- DoD #6 refresh ran: `codebase-memory-mcp cli index_repository --repo-path .` (11086 nodes, 44695 edges).

Session 1 remaining work is complete; task marked Done (not moved to `backlog/completed/`, which stays a separate periodic step).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Vendored the GALA 0.81.0 std runtime and extended GALA rewrite coverage, with all findings recorded in `server/GALA.md`.

Runtime: `server/third_party/gala/` is one flattened Go module `martianoff/gala` (20 packages, per-package go.mod/go.sum removed, Apache-2.0 LICENSE kept); `server/go.mod` carries `require martianoff/gala v0.0.0` plus `replace martianoff/gala => ./third_party/gala`; the Dockerfile copies `third_party/gala/go.mod` before `go mod download`; the Compose build and route-test paths stay unchanged.

Batch outcomes: `logger` is a full runtime-free twin; `appenv` (`ResolvePort`) and `utils` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`) are splits with handwritten sibling .go files; `models/location.go`, `models/set.go`, `errs/errors.go`, `models/datetime.go`, and `models/uuid.go` are blocked, with the exact transpiler errors recorded verbatim (JSON struct tags, `map[T]struct{}`, `const` + `interface{}`, alias-versus-defined-type emit, multi-value returns, fixed-size arrays, Go switch multi-case).

Probes: sealed types, `val` with Option/Try/Either, `collection_immutable` Array/HashMap, and `go_interop.Spawn`/`concurrent.Future` all transpile, compile against the vendored runtime, and run; `GALA.md` records the emitted Go API shapes and the `GALA-E0037`/`GALA-E0025`/`GALA-E0044` findings.

Verification: all five twins regenerate byte-identically (sha256 table in `GALA.md`); `go build ./...` and `go vet` on touched packages pass; the isolated Compose stack runs `go test ./... -count=1` green; markdown format/lint checks pass; the code knowledge graph was re-indexed.

N/A DoD: e2e not required (backend-only, exported API unchanged); no Swagger annotations changed; no `scripts/` or `prettier/` edits; no contract documents affected.
<!-- SECTION:FINAL_SUMMARY:END -->
