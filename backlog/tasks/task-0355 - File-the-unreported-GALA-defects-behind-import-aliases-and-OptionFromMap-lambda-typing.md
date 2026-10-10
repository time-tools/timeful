---
id: TASK-0355
title: >-
  File the unreported GALA defects behind import aliases and OptionFromMap
  lambda typing
status: Done
assignee: []
created_date: '2026-10-10 14:54'
updated_date: '2026-10-10 15:02'
labels:
  - gala
  - server
dependencies: []
references:
  - 'https://github.com/martianoff/gala/issues/749'
  - 'https://github.com/martianoff/gala/issues/750'
  - server/GALA.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
modified_files:
  - server/GALA.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-from-go/references/gaps.md
priority: high
type: chore
ordinal: 359300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After the TASK-0353/TASK-0354 runtime passes, two compiler defects are recorded in the ledger and skill references but have no upstream report on `martianoff/gala`:

1. A differently named import alias loses a Go call's `(T, error)` signature, so `std.GoTry` is not emitted and the generated Go does not build. Real occurrence: `pgstore "timeful/server/postgres"` in the bot command twins; the workaround was to drop the alias.
2. `go_interop.OptionFromMap(...).ForEach((v) => ...)` leaks the transpiler's own `T` into the generated Go when the map's value type is a GALA struct in another package and the binding's type comes from a handwritten sibling. Real occurrence: `server/discord_bot/init.gala`'s `commandMap`; the workaround is an annotated lambda.

The pass also re-checks the open finding `len` on a slice whose type is a handwritten `.go` sibling in another package (#613 closed for same-package declarations only). If the imported case now lowers on the pinned compiler, the finding row and the #613 report-index note are corrected rather than filed.

File each reproducing candidate as its own upstream issue per `.agents/skills/gala-from-go/references/gaps.md`'s template, and update `server/GALA.md`'s report index and the affected finding/prose entries with the issue URLs. A candidate that does not reproduce is recorded and not filed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Each candidate is re-derived on the pinned compiler (0.87.1, flake rev cd2fdcb5) with a minimal self-contained repro, and the emitted-Go contrast (absence vs presence of `std.GoTry`; leaked `T`) plus the `go build` error is recorded
- [x] #2 Each candidate is checked against the upstream tracker before filing, and an existing issue is commented instead of duplicated
- [x] #3 Filed issues follow the gaps.md report template, lead with the generated artifact, and assert absence as well as presence
- [x] #4 The #613 imported-declaration residual is re-checked on the pinned rev; if fixed, the finding row and report-index note are corrected instead of filing
- [x] #5 server/GALA.md's report index and the affected finding/prose entries name the filed issue URLs
- [x] #6 Changed Markdown is formatted with `npm run format:markdown`; no runtime file changes
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Provenance (2026-10-10)

`nix develop` gala `0.87.1` at `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; flake rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c` matches `server/GALA_COMPILER`; extraction marker `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`; Go `go1.26.7 linux/amd64`. All repros transpile with `--stdlib-module go.gala.fyi/stdlib` from the package directory, into scratch modules under `/tmp/opencode/gala-reports/`.

## Filed upstream

- [#749](https://github.com/martianoff/gala/issues/749) — A differently named import alias loses a Go call's `(T, error)` signature, so `std.GoTry` is not emitted and the generated Go does not build.
- [#750](https://github.com/martianoff/gala/issues/750) — `go_interop.OptionFromMap(...).ForEach` leaks the transpiler's own `T` into the generated Go when the map's value type is a GALA struct in another package.

Drafts per the `gaps.md` template are at `/tmp/opencode/gala-reports/drafts/issue-alias.md` and `issue-optionfrommap.md`. Tracker searched with `gh search issues` for `GoTry`, `OnFailure`, `import alias`, `aliased import`, `signature`, `OptionFromMap`, `ForEach`, `leak`, and `type parameter`; no duplicate existed (the nearest entries were #471 and #618, cross-referenced from #750).

## #749 reproduction (`/tmp/opencode/gala-reports/alias`)

Module `aliasrepro`; `store/store.go` is `func Load() (string, error)`; `aliased.gala` imports `st "aliasrepro/store"` and calls `st.Load().OnFailure((err error) => {}).Get()`; `unaliased.gala` is identical with `"aliasrepro/store"` and `store.Load()`.

- Aliased emitted Go: `var s = std.NewImmutable(st.Load().OnFailure(func(err error) {\n\t}).Get())` — no `std.GoTry`.
- Unaliased emitted Go: `var s = std.NewImmutable(std.GoTry(store.Load()).OnFailure(func(err error) {\n\t}).Get())`.
- `go mod tidy && go build ./...` fails with `aliased.gala:8: multiple-value st.Load() (value of type (string, error)) in single-value context`; with `aliased.go` moved aside the build exits 0. Both transpiles exit 0 with no diagnostic.

## #750 reproduction (`/tmp/opencode/gala-reports/optionfrommap2`)

Module `optionrepro2`; `sub/sub.gala` declares `struct Command(var Name string)`; `helper.go` returns `map[string]sub.Command`; `leak.gala` is `var commandMap = newMap()` plus `go_interop.OptionFromMap(commandMap, "x").ForEach((command) => {})`.

- Leak emitted Go: `go_interop.OptionFromMap(commandMap, "x").ForEach(func(command T) {\n\t})`; `go build` fails with `leak.gala:10: undefined: T`.
- `annotated.gala` with `(command sub.Command) => {}` and the `sub` import emits `func(command sub.Command)` and builds.
- Boundary on the same compiler, all with a package-level binding from the same-package sibling: `map[string]int` emits `func(v int)`; `map[string]sub.GoEntry` (Go struct, imported package) emits `func(entry sub.GoEntry)`; `map[string]sub.Command` (GALA struct, imported package) leaks `T`; a written `map[string]sub.Command` parameter emits `func(command sub.Command)`.

## #613 re-check — fixed, so no report

The `len` finding's residual no longer reproduces on the pinned rev. Module `sizemix` (`bag/bag.go` with `Bag{Items []string}`, `New()`, `Repo.List() []string`, `Repo.Entries() []Entry`, `Doc.Text string`): `CountParam(b *bag.Bag)` (written, imported) emits `len(b.Items)`; `CountLocal` (`var r = bag.NewRepo()`) emits `len(r.List())`; `CountMethodParam` emits `len(r.List())`; `CountEntries` emits `len(r.Entries())`; `TextSize` emits `utf8.RuneCountInString(d.Text)`; `TextBytes` emits `len(d.Text)`; the aliased-import spellings likewise emit `len(...)`. Module `leninferred` (same-package `List() ([]string, error)`) emits `len(logs)` for `var logs, err = List()`. All builds exit 0. The finding row was therefore deleted and #613 is recorded as fixed on the pinned rev instead of a follow-up comment.

## Ledger and references

- `server/GALA.md`: #749/#750 rows in the upstream report index; #613 row says fixed on the pinned rev; `len` finding row deleted; runtime idiom pass prose links #749/#750.
- `.agents/skills/gala-from-go/references/constructs.md`: inferred-receiver row deleted, receiver prose rewritten, #749 link on the alias row and prose, #750 link on the comma-ok row.
- `.agents/skills/gala-from-go/references/gaps.md`: `.Size()` section and What-Not-To-File bullet record the fix.
- `.agents/skills/gala-from-go/SKILL.md`: the sibling-struct trap no longer claims `.Size()`/`.ByteSize()` on a sibling's field fails.

## Checks

`npm run format:markdown` (only `server/GALA.md` changed) and `npm run format:markdown:check` pass. The `.agents/**` files are excluded by the root formatter and follow the existing sentence-per-line style. No runtime file, test, or generated twin changed.

A concurrent GALA-loop iteration (postgres) created `server/postgres/repository.gala`, `repository_methods.go`, and `repository_types.go` and modified `server/postgres/repository.go` mid-session; those files were left untouched, as was the pre-existing `backlog/backlog.md` modification.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Filed two upstream defects on `martianoff/gala` and corrected the ledger for a third finding that is fixed on the pinned compiler.

## Filed, 2026-10-10

- [#749](https://github.com/martianoff/gala/issues/749) — a differently named import alias loses a Go call's `(T, error)` signature, so `std.GoTry` is not emitted and the generated Go does not build; the package-name qualifier emits `std.GoTry` and builds. Real occurrence: `pgstore "timeful/server/postgres"` in the bot command twins.
- [#750](https://github.com/martianoff/gala/issues/750) — `go_interop.OptionFromMap(...).ForEach` leaks the compiler's own `T` into the generated Go when the map's value type is a GALA struct declared in another package and the binding's type comes from a handwritten sibling; the annotated lambda emits the real type and builds. Real occurrence: `discord_bot/init.gala`'s `commandMap`.

Both were re-derived minimally on the pinned compiler (GALA 0.87.1, flake rev `cd2fdcb5`, extraction `0.87.1 7b1dd208...`) with scratch modules under `/tmp/opencode/gala-reports/`, checked against the tracker before filing, and written to `gaps.md`'s report template: emitted-Go contrast first, absence assertions, minimal two-file repro, compiler and environment.

## Not filed: #613 is fixed on the pinned rev

The open finding `len` on a receiver whose type comes from a handwritten `.go` sibling no longer reproduces: `.Size()`/`.ByteSize()` lower for written and inferred receivers, for same-package and imported handwritten `.go` declarations (`len(b.Items)`, `len(r.List())`, `len(r.Entries())`, `utf8.RuneCountInString(d.Text)`, `len(d.Text)`, and `var logs, err = List()` then `len(logs)`). The finding row is deleted, the report-index row for [#613](https://github.com/martianoff/gala/issues/613) records it as fixed on the pinned rev, and the stale inferred/imported-receiver claims in `constructs.md`, `gaps.md`, and `SKILL.md` are corrected instead of filing a duplicate.

## Documents

`server/GALA.md` gains the #749/#750 report-index rows, the corrected #613 row, the deleted `len` finding, and links in the runtime idiom pass prose. `constructs.md` links both issues and drops the inferred-receiver row; `gaps.md`'s `.Size()` section and What-Not-To-File bullet now record the fix; `SKILL.md`'s sibling-struct trap no longer claims the lowering fails.

No runtime code, test, or generated twin changed. `npm run format:markdown` reformatted `server/GALA.md` and `format:markdown:check` passes; the `.agents/**` files are excluded by that command and follow the existing sentence-per-line style by hand. Unit and e2e checks are documentation-only exemptions, Swagger and the code graph are untouched, and no root `scripts/` or `prettier/` file changed.

A concurrent GALA-loop iteration created `server/postgres/repository.gala`, `repository_methods.go`, and `repository_types.go` and modified `repository.go` during this pass; those files were left untouched.
<!-- SECTION:FINAL_SUMMARY:END -->
