---
id: TASK-0321
title: Rewrite the next batch of GALA-ready server leaf files as twins
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 10:32'
updated_date: '2026-09-23 10:55'
labels: []
dependencies: []
references:
  - server/GALA.md
  - >-
    backlog/tasks/task-0320 -
    Spike-vendor-the-GALA-std-runtime-and-rewrite-a-further-batch-of-server-files-in-GALA-to-surface-friction.md
modified_files:
  - server/utils/request_utils.gala
  - server/utils/request_utils.go
  - server/services/services.gala
  - server/services/services.go
  - server/routes/guest_response_ownership.gala
  - server/routes/guest_response_ownership.go
  - server/services/providerconfig/providerconfig.gala
  - server/services/providerconfig/providerconfig.go
  - server/services/providerconfig/doc.go
  - server/discord_bot/commands/help.gala
  - server/discord_bot/commands/help.go
  - server/discord_bot/commands/num_users.gala
  - server/discord_bot/commands/num_users.go
  - server/discord_bot/init.gala
  - server/discord_bot/init.go
  - server/slackbot/commands/num_users.gala
  - server/slackbot/commands/num_users.go
  - server/slackbot/commands/utils.go
  - server/GALA.md
  - server/README.md
priority: medium
type: enhancement
ordinal: 319000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0320 vendored the GALA 0.81.0 runtime and proved the twin/split pattern, but the remaining server files were never triaged against the then-recorded blockers. A follow-up batch should convert the leaf files whose exported Go API can stay unchanged, because those carry no behavioral risk and are the cheapest evidence for where GALA is actually adoptable. The same batch should record the parser and codegen blockers that this triage surfaced (switch, defer, multi-value `:=`, if-init emission, len, type assertions, slice expressions, map literals, inline function literals in composite literals) in server/GALA.md, since they constrain every future rewrite and are the upstream asks.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All eight verified leaf files are committed as a .gala source plus its generated .go twin: utils/request_utils, services/services, routes/guest_response_ownership, services/providerconfig, discord_bot/commands/help, discord_bot/commands/num_users, discord_bot/init, and slackbot/commands/num_users; each keeps its exported Go API shape and behavior identical and its package comment is preserved (a handwritten doc.go where the transpiler cannot emit one).
- [x] #2 Every committed twin regenerates byte-identically from its committed .gala source with the command recorded in server/GALA.md, and go build ./... plus go vet on the touched packages pass.
- [x] #3 Existing tests pass unchanged: the affected package tests (utils, providerconfig, discord_bot, slackbot, routes, services) and the isolated Compose stack go test ./... -count=1.
- [x] #4 server/GALA.md records the new twins, their regeneration commands, and the newly discovered frictions with verbatim transpiler errors: switch statements, defer, multi-value `:=` receive, `if err :=` emission, len, type assertions, slice expressions, map literals, and inline function literals in composite literals; Verification hashes and When to revisit are refreshed.
- [x] #5 server/README.md Transpiled GALA sources lists every new twin with its regeneration command.
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

The candidate triage for this batch is already done and verified in a scratch copy of `server/` at `/tmp/opencode/twins` (transpiled with GALA 0.81.0, then `go build`, `go vet`, and package tests green).
The eight twins and the exact workarounds are known, so the remaining work is landing them, documenting the findings, and finalizing.

Constraints: exported Go API shape and behavior stay unchanged; no struct/interface/const declarations that would grow or alter the API; blocked files (tags, `defer`, `switch`, fixed arrays, named scalar receivers, map literals, type assertions) are out of scope.

## Phase 1 — Land the eight twins

1. `utils/request_utils.gala`: rewrite `decoded, err :=` as `var decoded, err =`; keep `logger.StdErr.Panicln`.
2. `services/services.gala`: same multi-value `var` rewrite; split the two `http.NewRequest` branches into `var created, _ = ...; req = created`.
3. `routes/guest_response_ownership.gala`: `const` becomes `var`; blank parameter and `models.SignUpResponse` field method calls carry over.
4. `services/providerconfig/providerconfig.gala`: both const blocks become `var`; the scoped `if value := ...; value != ""` becomes `var value = ...` plus a plain `if`; add `services/providerconfig/doc.go` to carry the package comment the transpiler drops.
5. `discord_bot/commands/help.gala` plus `num_users.gala`: package vars use named unexported exec functions because inline function literals in composite literals do not parse; `len(command.Usage) > 0` becomes `command.Usage != ""`; multi-value `:=` receives become `var` with distinct error names (multi-value reassignment does not parse).
6. `discord_bot/init.gala`: `make(map)` via `go_interop.MapEmpty`, `args[1:]` via `go_interop.SliceFrom`, multi-value `:=` becomes `var`, map index assignment stays.
7. `slackbot/commands/num_users.gala`: same named-func and `var` pattern as its Discord sibling.
8. Transpile each file from its own package directory so `//line` directives stay relative, and replace the original `.go` with the generated twin; the `.gala` source is the only hand-edited member.

## Phase 2 — Verification

1. Regenerate each twin twice and compare sha256, recording the committed hash.
2. `go build ./...` and `go vet` on the touched packages.
3. Package tests unchanged: `utils`, `services/providerconfig`, `discord_bot`, `slackbot`, `routes`, `services`.
4. Isolated Compose stack `go test ./... -count=1` per `server/README.md`.

## Phase 3 — Documentation

1. `server/GALA.md`: extend the Current usage table and regeneration block with the eight twins; add the new frictions with verbatim transpiler errors (`switch`, `defer` GALA-E0036, multi-value `:=` GALA-E0017, `if err :=` Immutable emission, `len`, type assertions, `args[1:]`, map literals GALA-E0008, inline function literals in composite literals); refresh Verification and When to revisit.
2. `server/README.md`: add every new twin and handwritten sibling to the Transpiled GALA sources table.

## Phase 4 — Finalization

1. `npm run format:markdown` / `format:markdown:check` / `lint:markdown`.
2. `codebase-memory-mcp cli index_repository --repo-path .`.
3. Mark DoD items, record fail/pass evidence, check acceptance criteria, write the final summary, mark Done.
4. Commit only task-relevant files; leave the pre-existing `backlog/backlog.md` UX notes unstaged.

## Risks

- The original `providerconfig.go` comments (exported consts and functions) are dropped by the transpiler; only the package comment survives via `doc.go`.
- `discord_bot/init.gala` now imports `martianoff/gala/go_interop`, so the file is runtime-enabled rather than runtime-free; the vendored module makes this compile.
- `go.mod` keeps the deliberately unused `require martianoff/gala v0.0.0`; do not run bare `go mod tidy`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-23 session)

### Phase 1 — eight twins landed

- Full twins committed with their `.gala` sources: `utils/request_utils`, `services/services`, `routes/guest_response_ownership`, `services/providerconfig` (+ handwritten `doc.go` for the package comment), `discord_bot/commands/help`, `discord_bot/commands/num_users`, `discord_bot/init` (runtime-enabled: `go_interop.MapEmpty`/`SliceFrom`), and `slackbot/commands/num_users`.
- Workarounds used: multi-value `:=` becomes `var a, b =`; inline function literals in `Command{...}` become named unexported functions; `len(x) > 0` becomes `x != ""`; `const` becomes `var`; `make`/slicing become `go_interop` helpers.
- Unplanned finding: `slackbot/commands/num_users.gala` calling `Response{...}` resolved the bare name to the imported `postgres.Response` and generated non-compiling Go. Fixed with a handwritten `newResponse` constructor in `slackbot/commands/utils.go` and recorded in GALA.md as a name-collision friction. No API change.
- `services/providerconfig` exported consts become vars (same names); declaration comments are dropped by the transpiler, so the package comment moved to `doc.go`.

### Phase 2 — verification evidence

- All eight twins regenerated twice in place; sha256 identical both runs:
  - `utils/request_utils.go` `b636a83486e1e33019846cb2d30c094c7ebcd52747dcd4c0c1b03e36dd11f7f3`
  - `services/services.go` `0c65dab9869fca16f43063ed8605e37801e8d0e8d4d19946a083b778eeac3ce2`
  - `routes/guest_response_ownership.go` `5ef803dfe4f3a13f72d8726974b338134ec9a09f768d2f60920c49bcc7c6e450`
  - `services/providerconfig/providerconfig.go` `e77c2513833030c645b1e502fb76346fa76fed7a61b3daf49a246992aee59d64`
  - `discord_bot/commands/help.go` `45afe9bae3ebd48520d7e1b3f34fdba754b83c35e7d002e3ae57a742f2d0a637`
  - `discord_bot/commands/num_users.go` `d699d3116829b604fd048f8f9f5d6082a31e122d0f6ec0b7e9fd60b58ccd830e`
  - `discord_bot/init.go` `87dcafb9086c5764ec97d1d1125778291a79c667824f074c995f558e38bfd6cb`
  - `slackbot/commands/num_users.go` `6ddeea45ec41555cd07751c52125cc3de794231b8599e5586ca8f2868aed00f9`
- `go build ./...` clean, `go vet` clean on all touched packages, `gofmt -l` reports nothing.
- Exported API shape unchanged in every twin; the only deliberate shape change is providerconfig's exported consts becoming vars (same names, string values), matching the `appenv` precedent.
- Affected package tests pass on the host, and the isolated Compose stack ran `go test ./... -count=1` green across every package (`routes` 3.1s, `postgres` 7.9s, `models`, `observability`, `accounts`, `utils`, `providerconfig`, `discord_bot`, ...).

### Phase 3 — documentation

- `server/GALA.md`: 13-row Current usage table and regeneration block, two-mode ground rules updated, new works/does-not-work entries with verbatim errors (switch, defer GALA-E0036, multi-value `:=` GALA-E0017, `if err :=` Immutable emission, `len`, type assertions, slice expressions, map literals GALA-E0008, inline function literals, name collision), 13-row sha256 Verification table, refreshed When to revisit.
- `server/README.md`: 13 generated files with regeneration commands and the four handwritten siblings.
- `npm run format:markdown`, `format:markdown:check`, `lint:markdown`, and root `npm run fmt:check` all pass.

### Phase 4 — hygiene

- `codebase-memory-mcp cli index_repository --repo-path .` refreshed: 11091 nodes, 44702 edges.
- Swagger N/A (no route annotations changed); e2e N/A (backend-only, no user-visible or API change); contract documents untouched.
- `backlog/backlog.md` carries pre-existing unrelated UX notes and stays unstaged; the commit includes only task files.

### Friction findings recorded for upstream

- `switch` is wholly unsupported (not only multi-case).
- `defer` is GALA-E0036 with runtime `resource.Using` as the replacement.
- Multi-value `:=` receive is GALA-E0017; `var` and plain reassignment work.
- `if err := f(); err != nil` emits `std.Immutable[error]`.
- `len` is GALA-E0035; `.Size()` is rune-based and field-level `.Size()` does not lower.
- Type assertions, slice expressions, map literals, and inline function literals in composite literals do not parse.
- An imported package symbol shadows a sibling handwritten type of the same name.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Extended the GALA twin set from five files to thirteen, with every exported Go API shape and behavior unchanged.

Twins: `utils/request_utils`, `services/services`, `routes/guest_response_ownership`, `services/providerconfig` (consts to vars plus a handwritten `doc.go` for the package comment), `discord_bot/commands/{help,num_users}` (named exec functions and string-emptiness checks), `discord_bot/init` (runtime-enabled with `go_interop.MapEmpty`/`SliceFrom`), and `slackbot/commands/num_users` (handwritten `newResponse` constructor to avoid the `postgres.Response` name collision).

Verification: all eight new twins regenerate byte-identically (sha256 table in `server/GALA.md`); `go build ./...`, `go vet` on touched packages, and `gofmt -l` are clean; affected package tests and the isolated Compose stack `go test ./... -count=1` are green; markdown format/check/lint and root `npm run fmt:check` pass; the code knowledge graph was re-indexed.

Documentation: `server/GALA.md` records all thirteen twins and regeneration commands, the new frictions with verbatim transpiler errors (`switch`, `defer` GALA-E0036, multi-value `:=` GALA-E0017, `if err :=` emission, `len`, type assertions, slice expressions, map literals, inline function literals, imported-name collisions), and refreshed Verification and When to revisit; `server/README.md` lists every regeneration command and handwritten sibling.

N/A DoD: Swagger (no route annotations changed), e2e (backend-only, no behavior or API change), contract documents (untouched).
<!-- SECTION:FINAL_SUMMARY:END -->
