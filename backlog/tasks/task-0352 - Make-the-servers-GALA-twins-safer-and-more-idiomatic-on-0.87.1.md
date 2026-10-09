---
id: TASK-0352
title: Make the server's GALA twins safer and more idiomatic on 0.87.1
status: Done
assignee: []
created_date: '2026-10-09 09:35'
updated_date: '2026-10-09 09:49'
labels:
  - gala
  - server
dependencies: []
references:
  - server/GALA.md
  - server/scripts/gala/verify.sh
  - .claude/skills/gala-from-go/SKILL.md
  - .claude/skills/gala-from-go/references/constructs.md
modified_files:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/GALA.md
  - server/appenv/appenv.gala
  - server/appenv/appenv.go
  - server/appenv/appenv_port.go
  - server/appenv/appenv_environment.go
  - server/appenv/appenv_type_test.go
  - server/discord_bot/commands/active_users.gala
  - server/discord_bot/commands/active_users.go
  - server/discord_bot/commands/help.gala
  - server/discord_bot/commands/help.go
  - server/discord_bot/commands/num_users.gala
  - server/discord_bot/commands/num_users.go
  - server/discord_bot/init.gala
  - server/discord_bot/init.go
  - server/routes/guest_response_ownership.gala
  - server/routes/guest_response_ownership.go
  - server/services/providerconfig/providerconfig.gala
  - server/services/providerconfig/providerconfig.go
  - server/services/services.gala
  - server/services/services.go
  - server/slackbot/commands/active_users.gala
  - server/slackbot/commands/active_users.go
  - server/slackbot/commands/num_users.gala
  - server/slackbot/commands/num_users.go
  - server/utils/request_utils.gala
  - server/utils/request_utils.go
priority: medium
ordinal: 355005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The committed GALA twins in `server/` were written as Go spelled in GALA syntax, because earlier compilers forced runtime-free twins to avoid most GALA idioms. GALA 0.86/0.87 changed what is available: `match` with literal, value, and guard patterns, `if` expressions, `s"..."`/`f"..."` interpolation, expression-bodied functions, and `opaque type` with a same-package `Hash`/`Compare` sibling all lower to plain Go with no runtime import on the pinned 0.87.1 compiler.

The runtime itself is still not consumable: `gala transpile` emits `martianoff/gala/...` imports, and `replace martianoff/gala => go.gala.fyi/stdlib v0.87.1` is rejected by Go ("used for two different module paths"), because only `gala export` can rewrite stdlib imports. Twins therefore stay runtime-free, and `val`, `Option`/`Try`, collections, structs, and sealed types stay out.

Investigation also found two safety problems in the existing translation:
- `type Environment string` in `appenv/appenv.gala` emits `type Environment = string`, a Go type alias, so the twin silently dropped the distinct `appenv.Environment` type the handwritten Go had. GALA's `type X Y` is an alias by design; `opaque type` is the newtype.
- The `gala-from-go` construct reference recommends combining match patterns with `|`. GALA has no pattern alternatives: `case 1 | 2` is an expression pattern that emits `obj == 1|2` (bitwise OR, compares with 3) and silently miscompiles, while `case "a" | "b"` fails `go build`.

Rewrite the twins with the runtime-free idioms where that improves readability or narrows mutation, restore the distinct `Environment` type, and correct the stale or dangerous rows in the translation reference.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `appenv.Environment` is a distinct Go defined type again (not an alias), with only the `Hash`/`Compare` methods added by the documented `opaque type` suppression path, and the generated Go stays runtime-free.
- [x] #2 Twins whose bodies are value dispatch, if/else chains, or `fmt.Sprintf` formatting use the runtime-free GALA idioms (`match` with literal/value/guard patterns, `if` expressions, interpolation, expression-bodied functions) where that preserves behavior.
- [x] #3 Exported Go-facing signatures, wire formats, and observable output are unchanged apart from the `Environment` restoration; no generated file is hand-edited and no test is relaxed.
- [x] #4 `server/scripts/gala/verify.sh` passes, `go build ./...` and `go vet` on touched packages pass, and the canonical backend test sequence in `server/README.md` passes.
- [x] #5 The `gala-from-go` construct reference no longer claims `type X Y` is a defined type or recommends `|` pattern alternatives, records the silent `case 1 | 2` miscompile as a trap, and records which match/if/interpolation forms lower runtime-free on the pinned compiler.
- [x] #6 `server/GALA.md` records the runtime-module finding (the `replace` path is rejected), the `Environment` restoration, and any finding changes.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm what 0.86/0.87 lifted on the pinned compiler: try the `go.gala.fyi/stdlib` module as the runtime, and probe which idioms lower runtime-free.
2. Bug fix (Bug Fix Protocol): add a reflect-based regression test for `appenv.Environment`, watch it fail, then switch to `opaque type` with a `Hash`/`Compare` sibling.
3. Rewrite the other twins with runtime-free idioms that preserve behavior: `match`, `if` expressions, interpolation, and expression bodies.
4. Regenerate with `verify.sh --write`, read the generated diff, then run vet, the build, and the canonical Compose backend tests.
5. Correct the `gala-from-go` construct rows and traps, and update the `server/GALA.md` ledger.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Provenance
`nix develop` gala `GALA version 0.87.1` at `/nix/store/sy03nzcg76p7429jkrkgkpgj73fhzvrm-gala-0.87.1/bin/gala`; flake rev `669c958cfb1f0ef2fe69ecb8e199b33c88456664` matches `server/GALA_COMPILER`, so no bump was needed.

## Runtime module (scratch probe)
- `gala transpile` emits `martianoff/gala/...` imports.
- `replace martianoff/gala => go.gala.fyi/stdlib v0.87.1` fails `go mod tidy` with `go.gala.fyi/stdlib@v0.87.1 used for two different module paths`.
- `--stdlib-module` exists only on `gala export`.
- Twins stay runtime-free.

## Runtime-free idioms (scratch probes on 0.87.1)
- Plain Go: value-used `match` with literal, stable-identifier, guard, or `_` patterns, and `if` expressions. Both lower to an immediately invoked closure.
- Plain Go: `s"..."`/`f"..."` lower to `fmt.Sprintf`, with `%d` for ints and `%v` for strings.
- Plain Go: expression bodies, and a Go result list whose body is a Go call.
- Runtime: a typed pattern (`std.As`) and a Tuple/`Success` body for a result list (`std.Try`).
- Stable identifiers compare (`obj == Prod`), as GALA.MD §Stable Identifiers documents, so constructs row 92 was stale.

## Defects found
1. `type Environment string` emitted `type Environment = string` (GALA `type X Y` is an alias by design). This was present since c99c0764.
2. `case 1 | 2` emits `obj == 1|2`. A run of `small(1), small(2), small(3)` printed `other other one-or-two`. `case "a" | "b"` fails `go build`. No upstream issue was found, and GALA documents no alternative pattern. The skill's `fallthrough` row recommended `|`.

## Bug Fix Protocol evidence (Environment alias)
- Before: `go test ./appenv/ -run TestEnvironmentIsDefinedType` failed with `Environment resolves to .string, want the defined type timeful/server/appenv.Environment`.
- After: the same command passes, and the generated `appenv.go` declares `type Environment string`. The old generated `appenv.go` was moved aside to scratch before the transpile, per the opaque-type suppression requirement.

## Verification
- `server/scripts/gala/verify.sh --write` regenerated 11 twins. A second `verify.sh` printed `OK (21 twins)`: runtime-free, gofmt-clean, byte-identical, and `go build ./...` passes.
- `go vet` is clean on appenv, discord_bot, slackbot, routes, services, providerconfig, and utils.
- The canonical Compose backend sequence is green in every package (`routes` 2.142s, `postgres` 5.731s, `appenv` 0.011s). The existing `.env.test` was kept, and the volumes were created one per command.
- The generated diff reviewed by hand shows only verb changes (`%s`→`%v` on string fields), the removal of the `bodyBuffer` intermediate in `CallApi`, closure-lowered matches and ifs, and `//line` renumbering.
- No swag annotations changed. `codebase-memory-mcp cli index_repository --repo-path .` was refreshed, and `npm run format:markdown` was run.

## Upstream reports (filed 2026-10-09 at the user's request)
- [#739](https://github.com/martianoff/gala/issues/739): `case 1 | 2` is lowered as bitwise OR and silently matches only 3 (defect).
- [#740](https://github.com/martianoff/gala/issues/740): follow-up to #698, asking `gala transpile` to emit `go.gala.fyi/stdlib` imports (`--stdlib-module`). Evidence: the `replace` is refused, and `gala export` in `server/` fails with `gala.mod not found`.

Both are linked from the open findings table and the report index in `server/GALA.md`.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The twins now use the runtime-free GALA idioms available on 0.87.1, `appenv.Environment` is a distinct type again, and the translation reference no longer recommends a construct that silently miscompiles.

- **Bug fix:** `appenv.Environment` had been a `string` alias since its first translation, because GALA's `type X Y` is an alias. It is now an `opaque type`, with `Hash`/`Compare` in `appenv/appenv_environment.go` and a regression test in `appenv/appenv_type_test.go`. The test failed before the change and passes after it.
- **Idioms:**
  - `appenv` uses `match` with stable identifiers and guards.
  - `providerconfig` and `discord_bot` use `if` expressions.
  - Guest ownership uses a boolean expression body.
  - The bot commands and `services` use `s"..."` interpolation.
  - `CallApi` no longer builds a typed-nil buffer intermediate.
- **Docs:** the `gala-from-go` rows for defined types, stable identifiers, `fallthrough`, several values in one `case`, guards, type assertions, and interpolation are corrected. Two traps were added: the alias, and `|` as bitwise OR.
- **Ledger:** `server/GALA.md` records why `go.gala.fyi/stdlib` cannot back a transpiled twin (the `replace` is refused), the idiom list, the `appenv` sibling, and two new findings.

Verification: `verify.sh` OK (21 twins), vet clean, and the canonical Compose backend tests green.
<!-- SECTION:FINAL_SUMMARY:END -->

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
