---
id: TASK-0319
title: 'Spike: transpile selected server files from GALA to Go'
status: Done
assignee: []
created_date: '2026-09-23 08:48'
updated_date: '2026-09-23 09:00'
labels: []
dependencies: []
priority: medium
type: spike
ordinal: 317000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Evaluate using GALA (martianoff/gala, transpiles to Go) for parts of the Go server. Rewrite a small set of low-dependency server files in GALA, transpile them back to Go, and keep the existing Go tests green.

Scope:
- Candidate files: server/eventid/eventid.go, server/observability/redact.go; optionally server/appenv/appenv.go if the API can stay drop-in.
- Keep the package exported API and behavior identical so existing Go tests compile and pass unchanged.
- Commit the .gala sources and the generated .go output, or document a deterministic command that regenerates the .go.
- Do not add the GALA standard-library runtime as a dependency unless the spike documents exactly what that would require.

Out of scope:
- Migrating large files or PostgreSQL code.
- CI integration for regeneration.

Deliverable: the rewritten files plus a written findings section covering what GALA can and cannot express for this codebase (Go multi-value returns, package constants, val/Immutable runtime dependency, import resolution), and a recommendation on whether to continue.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At least two server files have equivalent GALA sources checked in and transpiled Go output that compiles in the existing server module
- [x] #2 Existing Go tests for the affected packages pass unchanged against the transpiled Go
- [x] #3 The full server test suite for the affected packages still builds with go build ./...
- [x] #4 Findings and a continue/stop recommendation are recorded in the task
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
1. Confirm GALA 0.81.0 idiom that transpiles to runtime-free Go for drop-in interop (var for raw Go values, match expressions, Go stdlib imports).
2. Rewrite server/eventid/eventid.go as eventid.gala, transpile to eventid.go, run package tests.
3. Rewrite server/observability/redact.go as redact.gala, transpile to redact.go, run package tests and go build ./... in server.
4. Clean up generated output, record findings (multi-return limitation, const/val runtime dependency, import resolution) and recommendation in the task notes.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Findings (2026-09-23)

### What was done
- `server/eventid/eventid.gala` -> `server/eventid/eventid.go` (generated); `server/eventid/doc.go` preserves the package comment the transpiler drops.
- `server/observability/redact.gala` -> `server/observability/redact.go` (generated).
- Regeneration: `cd server/<pkg> && gala transpile -i <name>.gala -o <name>.go` with GALA 0.81.0; two successive regenerations are byte-identical (sha256 checked).
- `server/.gitignore` ignores the transpiler's `server/.gala/` analysis cache; `server/README.md` documents the workflow.

### Verification
- `go build ./...` in `server/`: pass.
- `go vet ./eventid/ ./observability/`: pass.
- Isolated Compose stack `go test ./... -count=1`: every package ok (eventid, observability, postgres, routes, ...).
- Differential test of the transpiled `truncate`/`Redact` against the original implementations: 200k random strings (ASCII, spaces, tabs, newlines, multi-byte runes, invalid UTF-8) x random limits, plus 20k redaction input combinations; outputs identical.

### What GALA handles well here
- Free functions importing Go stdlib directly (`regexp`, `strings`, `unicode/utf8`) and expression-bodied functions.
- `match` on strings compiles to plain Go with no runtime imports.
- Package-level `var` with inferred types emits a plain Go var (no runtime import).
- `for ... range` over a string, `break`, char literals, and `string(rune)` conversions all work.

### Blockers for broader adoption
1. No Go-style multi-value returns: `func f() (string, error)` is a parse error (GALA.MD section 3); GALA returns `Tuple`. This blocks drop-in rewrites of `appenv.ResolvePort`, `utils.CORSOrigins`, `utils.GetListmonkOtpFromAddress`, `normalizeBaseUrl`, and most Go helpers.
2. No `const`. Package constants must be `var`, or `val`, and `val` wraps the value in `std.Immutable[T]` (e.g. `appenv.Development` becomes `std.Immutable[Environment]`), changing the exported Go API and breaking callers.
3. Any `val` (even function-local int/string/bool) emits `std.NewImmutable(...)` and imports `martianoff/gala/std`. The runtime is distributed as GALA source under `~/.gala/stdlib/<version>` with machine-local `replace` directives, so runtime-using output cannot be committed into `server/` without vendoring the runtime or publishing it as a normal Go module.
4. Documentation comments are not emitted into generated Go (no comments at all). The eventid package comment needs a handwritten `doc.go`; comments in `redact.gala` exist only in the GALA source.
5. Byte-level string slicing (`s[:n]`) is unsupported by the grammar; `truncate` was rewritten as a rune-accumulation loop with identical byte semantics.
6. Matching a string-typed constant (`case Development =>`) binds a new variable instead of comparing, and `string(Development)` is not an extractor, so string-typed enums lose idiomatic switch/match.
7. The transpiler writes an analysis cache at the module root (`server/.gala/cache`); it needs a gitignore entry (added).

### Recommendation
- The mechanism works for small, pure, interop-shaped leaf packages whose API has no multi-value returns and needs no GALA runtime; output is deterministic and tests stay green.
- Broad migration is not viable yet. The hard blockers are Go multi-value returns and the std runtime being distributed as machine-local GALA source. Continue only if upstream adds multi-return signatures (or an interop escape hatch) and a consumable runtime module; otherwise keep GALA out of the server build.
- Answer the runtime-dependency question (vendor vs. publish as a Go module) before expanding scope.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-09-23 09:00
---
Findings are now documented in `server/GALA.md` (linked from `server/README.md`): current usage and regeneration commands, runtime-free ground rules, what works, what does not, what is untested, verification evidence, and the upstream changes that would unblock broader adoption.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Rewrote two pure server files in GALA and committed their transpiled Go output: `server/eventid/eventid.gala` -> `eventid.go` (package comment preserved in `doc.go`) and `server/observability/redact.gala` -> `redact.go`. Added the `server/.gala/` cache to `server/.gitignore` and documented the regeneration commands in `server/README.md`.

Verification: `go build ./...` and `go vet` pass; the isolated Compose stack runs `go test ./... -count=1` green across all packages; differential tests match the original `truncate`/`Redact` on 200k random strings and 20k redaction combinations; regeneration is byte-identical.

Outcome: works for leaf packages with Go-shaped APIs and no runtime dependency. Broad adoption is blocked by (1) no Go multi-value `(T, error)` returns, (2) no `const`, and (3) the `std.Immutable` runtime shipping as machine-local GALA source rather than a go.mod-consumable module. Recommendation: stop at this spike unless those transpiler/runtime gaps are resolved upstream.

N/A DoD: no browser-visible change (e2e), no Swagger annotations, no `scripts/` or `prettier/` edits, no contract documents affected.
<!-- SECTION:FINAL_SUMMARY:END -->
