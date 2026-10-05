# GALA in the server

This is the single GALA ledger for the server: which twins exist, which compiler they were built with, what still blocks the rest of the cursor, and how to continue.
It does not restate construct substitutions or gap classification; those live in the project-independent translation reference at [`../.agents/skills/gala-from-go/SKILL.md`](../.agents/skills/gala-from-go/SKILL.md).
The loop that consumes both is [`../.agents/skills/gala-loop/SKILL.md`](../.agents/skills/gala-loop/SKILL.md), and verification is [`scripts/gala/verify.sh`](scripts/gala/verify.sh).

## Contract

Both halves of a twin are committed: the `.gala` source and the Go file the transpiler generates from it.
The Go build never invokes GALA, so a generated file without its source is unreadable to the next maintainer and a source without its generated file is not in the build at all.
Never hand-edit a generated file; it carries a `DO NOT EDIT` header and the next transpile discards the edit without a trace.
Transpile from the package directory so the emitted `//line` directives stay relative: `cd server/<pkg> && gala transpile -i <name>.gala -o <name>.go`.
The transpiler writes its analysis cache under `server/.gala/`, which is gitignored.

Every committed twin is runtime-free: its generated Go names no `martianoff/gala` package, so the repository vendors no GALA runtime and `server/go.mod` has no `martianoff/gala` requirement.
Write Go packages directly, bind raw Go values with `var`, and use `.ByteSize()`/`.Size()` in place of `len`.

GALA constructs that lower through the runtime stay out of a twin.
They are `val` and `:=` bindings, `Option`/`Try`/`Either`, sealed types, the immutable collections, and any GALA `struct` declaration, which emits `Copy`, `Equal`, `Unapply`, and struct metadata through `std`.
A value the twin needs from one of them belongs in a handwritten Go sibling in the same package, which the transpiler resolves without an import: a map literal, a slice expression, a `(T, bool)` type assertion, or an interface method with several results.
[GO_INTEROP.MD Part 3](https://github.com/martianoff/gala/blob/master/docs/GO_INTEROP.MD#part-3-one-package-two-languages) is the reference, and `middleware/auth_session.go`, `postgres/dailylogs_types.go`, and `discord_bot/interop.go` are the siblings that rule produced here.
A candidate whose Go-facing API cannot survive the split stays handwritten.

Read the emitted Go before committing a new twin and confirm it names no runtime package; `verify.sh` fails a twin whose generated Go imports one.

## Compiler

The translation target is the `gala` compiler the dev shell puts on `PATH`, and every committed twin regenerates byte-identically under it.
The compiler is the flake-locked commit; `gala version` reports `0.85.0`, and the current rev `a888e824` adds [#712](https://github.com/martianoff/gala/pull/712), which makes Go's `encoding/json`, YAML, and `fmt` see a `std.Immutable`'s value rather than `{}` (closes [#687](https://github.com/martianoff/gala/issues/687)), on top of the rev that carried [#699](https://github.com/martianoff/gala/issues/699), [#705](https://github.com/martianoff/gala/issues/705), [#706](https://github.com/martianoff/gala/issues/706), [#707](https://github.com/martianoff/gala/issues/707), [#708](https://github.com/martianoff/gala/issues/708), [#709](https://github.com/martianoff/gala/issues/709), and [#710](https://github.com/martianoff/gala/issues/710) on top of the rev that carried the post-0.85.0 [#691](https://github.com/martianoff/gala/issues/691), [#692](https://github.com/martianoff/gala/issues/692), [#695](https://github.com/martianoff/gala/issues/695), and [#702](https://github.com/martianoff/gala/issues/702) fixes and [GO_INTEROP.MD](https://github.com/martianoff/gala/blob/master/docs/GO_INTEROP.MD).

| Item           | Value                                                                                                         |
| -------------- | ------------------------------------------------------------------------------------------------------------- |
| `gala version` | `GALA version 0.85.0`                                                                                         |
| flake rev      | `a888e824ff53adb653bbd48dcc83f67788eeced4` (`jq -r '.nodes.gala.locked.rev' flake.lock`)                      |
| extraction     | `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c` in [`GALA_COMPILER`](GALA_COMPILER) |
| pinned tools   | `nix develop` provides `gala` and the Go toolchain from `flake.lock`                                          |

Record the provenance before any iteration:

```sh
gala version
which gala
jq -r '.nodes.gala.locked.rev' flake.lock
cat server/GALA_COMPILER
```

The GALA runtime is not vendored.
Every twin is runtime-free, so generated Go imports only the server's own packages and ordinary Go dependencies; `server/go.mod` has no `martianoff/gala` require or replace, and `server/Dockerfile` copies no runtime tree.
[#698](https://github.com/martianoff/gala/issues/698) still asks upstream to publish the transpiled stdlib as a release asset, but this repository no longer consumes one, and a runtime-enabled twin would have to reopen that packaging question.

When the flake rev and `GALA_COMPILER` disagree, stop translating and run a bump iteration first.
The version string alone cannot identify the compiler, because several commits share one version string, so the compiler is proved by the fingerprint in its `.stdlib-extracted` marker.

1. Enter `nix develop` so `gala` is the flake-locked compiler, and record `gala version`, `which gala`, and `jq -r '.nodes.gala.locked.rev' flake.lock`.
2. Force the pinned compiler to write its own extraction: delete `~/.gala/stdlib/v<version>/.stdlib-extracted`, then run one transpile from a package directory, for example `cd server/eventid && gala transpile -i eventid.gala -o /tmp/eventid.go`.
   The transpiler resolves the stdlib on every run and rewrites the version directory when the marker is missing or names a different embedded snapshot, so the extraction at that path afterwards belongs to the pinned compiler.
3. Read `~/.gala/stdlib/v<version>/.stdlib-extracted`; it is `<version> <snapshot fingerprint>`.
4. Write `GALA_COMPILER` with the compiler version, the flake rev, the extraction marker contents, and the date.
5. Run `scripts/gala/verify.sh --write`, review the per-twin diff, then run `go build ./...` in `server/` and the canonical backend test sequence in [`README.md`](README.md).

A rev mismatch is resolved by a bump iteration and never by translating across it, because a twin generated by one compiler is not reproducible under another.

## Twin registry

Seventeen `.gala` sources across thirteen packages carry a committed twin.
Every row regenerates with `cd server/<dir> && gala transpile -i <name>.gala -o <name>.go`; a handwritten sibling has no command because it is not generated.

| Package                   | GALA source                                   | Generated Go                                | Handwritten sibling                                                                                                                              | Style        |
| ------------------------- | --------------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ |
| `eventid`                 | `eventid/eventid.gala`                        | `eventid/eventid.go`                        | —                                                                                                                                                | runtime-free |
| `observability`           | `observability/redact.gala`                   | `observability/redact.go`                   | —                                                                                                                                                | runtime-free |
| `logger`                  | `logger/logger.gala`                          | `logger/logger.go`                          | —                                                                                                                                                | runtime-free |
| `appenv`                  | `appenv/appenv.gala`                          | `appenv/appenv.go`                          | `appenv/appenv_port.go` (`ResolvePort`)                                                                                                          | runtime-free |
| `utils`                   | `utils/array_utils.gala`                      | `utils/array_utils.go`                      | `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`)                                                          | runtime-free |
| `utils`                   | `utils/request_utils.gala`                    | `utils/request_utils.go`                    | —                                                                                                                                                | runtime-free |
| `services`                | `services/services.gala`                      | `services/services.go`                      | —                                                                                                                                                | runtime-free |
| `services/providerconfig` | `services/providerconfig/providerconfig.gala` | `services/providerconfig/providerconfig.go` | `services/providerconfig/doc.go` (package comment)                                                                                               | runtime-free |
| `routes`                  | `routes/guest_response_ownership.gala`        | `routes/guest_response_ownership.go`        | —                                                                                                                                                | runtime-free |
| `routes`                  | `routes/users.gala`                           | `routes/users.go`                           | —                                                                                                                                                | runtime-free |
| `discord_bot/commands`    | `discord_bot/commands/help.gala`              | `discord_bot/commands/help.go`              | —                                                                                                                                                | runtime-free |
| `discord_bot/commands`    | `discord_bot/commands/num_users.gala`         | `discord_bot/commands/num_users.go`         | —                                                                                                                                                | runtime-free |
| `discord_bot/commands`    | `discord_bot/commands/active_users.gala`      | `discord_bot/commands/active_users.go`      | `discord_bot/commands/active_users_extra.go` (slice, map, and count helpers)                                                                     | runtime-free |
| `discord_bot`             | `discord_bot/init.gala`                       | `discord_bot/init.go`                       | `discord_bot/interop.go` (`newCommandMap`, `argsFrom`)                                                                                           | runtime-free |
| `slackbot/commands`       | `slackbot/commands/num_users.gala`            | `slackbot/commands/num_users.go`            | `slackbot/commands/utils.go` (`newResponse`)                                                                                                     | runtime-free |
| `middleware`              | `middleware/auth.gala`                        | `middleware/auth.go`                        | `middleware/doc.go` (package comment), `middleware/auth_session.go` (`sessionIdentityID`)                                                        | runtime-free |
| `postgres`                | `postgres/dailylogs.gala`                     | `postgres/dailylogs.go`                     | `postgres/dailylogs_methods.go` (the four `*Repository` daily-log methods), `postgres/dailylogs_types.go` (`DailyUserLog`, `DailyUserLogMember`) | runtime-free |

`scripts/gala/verify.sh` checks that every one of them regenerates byte-identically, stays free of the GALA runtime, is `gofmt`-clean, and keeps `go build ./...` green.
Adding a twin means adding its registry row here and passing `verify.sh`.

## Open findings

These keep the cursor's candidates handwritten or split, and each is classified against [`references/gaps.md`](../.agents/skills/gala-from-go/references/gaps.md).
A finding gains a committed repro under `scripts/gala/probes/<slug>/` when the loop first reaches a candidate it blocks, and the probe is removed when the finding closes.

| Finding                                                                       | Class        | Workaround                                                                                                                          | Upstream                                                                                        | Probe |
| ----------------------------------------------------------------------------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ----- |
| Go-style multi-value return signature                                         | boundary gap | split the member into a handwritten sibling, or return `Tuple[T, error]` to GALA-internal callers                                   | none filed                                                                                      | —     |
| `len` on a slice whose type is a handwritten `.go` sibling in another package | workaround   | a handwritten sibling helper takes the count; `go_interop.SliceFrom(x, 0).Size()` is the reference substitute but names the runtime | [#613](https://github.com/martianoff/gala/issues/613) closed for same-package declarations only | —     |
| Struct tags that define a wire shape                                          | boundary gap | keep the tagged struct handwritten; `Codec` changes the wire shape                                                                  | [#528](https://github.com/martianoff/gala/issues/528)                                           | —     |
| Anonymous struct types                                                        | boundary gap | declare a named struct, which changes the type identity                                                                             | [#528](https://github.com/martianoff/gala/issues/528)                                           | —     |
| `struct{}` as a type expression                                               | boundary gap | declare a named empty struct, which changes the element type identity                                                               | [#528](https://github.com/martianoff/gala/issues/528)                                           | —     |
| A channel type that crosses the package boundary                              | boundary gap | a handwritten Go sibling owns the channel; GALA cannot express `chan` or `select`                                                   | none filed                                                                                      | —     |
| Embedded fields                                                               | language gap | flatten by hand and lose promotion                                                                                                  | none filed                                                                                      | —     |
| Fixed-size array types such as `[16]byte`                                     | language gap | a slice or a struct, which changes the semantics                                                                                    | [#528](https://github.com/martianoff/gala/issues/528)                                           | —     |
| `select`                                                                      | language gap | none; the repository's uses pair it with `chan` and `make`                                                                          | none filed                                                                                      | —     |
| In-place `recover`                                                            | language gap | `Try` captures a panic as a value but cannot resume in place                                                                        | none filed                                                                                      | —     |

Declaration comments and `swag` annotations are no longer a finding: the pinned compiler emits them, so `routes/users.gala` keeps its `@Router` block and `swag init --parseDependency` regenerates `server/docs` byte-identically (see the report index for [#619](https://github.com/martianoff/gala/issues/619)).
A comment inside a function body is still dropped, so the `InitUsers` route-ordering note lives only in `routes/users.gala`.
The post-release [#695](https://github.com/martianoff/gala/issues/695) makes a Go method's several results one `Try`/`Tuple` value like a Go function's, but a GALA function still cannot declare a Go-style multi-value signature (a transpile parse error), so the Go-style multi-value return signature row stays.
Fixed-size arrays and anonymous or empty struct types remain the reason `models/uuid.go`, `models/event.go`, and `models/set.go` stay handwritten.
A workaround that names a runtime construct (`go_interop`, `Try`) is no longer available in a committed twin and belongs in a handwritten Go sibling.

## Upstream report index

Reports this repository filed, and their state on the pinned compiler rev.
A report closed after the flake lock is fixed upstream but not yet in the pinned compiler, so its finding stays open until the next sync.

| #                                                      | Title                                                    | Upstream state         | On the pinned rev                                                                                 |
| ------------------------------------------------------ | -------------------------------------------------------- | ---------------------- | ------------------------------------------------------------------------------------------------- |
| [#528](https://github.com/martianoff/gala/issues/528)  | Triage language limitations                              | open                   | open                                                                                              |
| [PR #529](https://github.com/martianoff/gala/pull/529) | Multi-value define and alias fixes                       | merged in 0.84.1       | included                                                                                          |
| [#611](https://github.com/martianoff/gala/issues/611)  | `fallthrough` in a `match` arm                           | closed                 | fixed (closure predates the lock)                                                                 |
| [#612](https://github.com/martianoff/gala/issues/612)  | `match` on `var` fields emits an unwrap                  | closed                 | fixed (closure predates the lock)                                                                 |
| [#613](https://github.com/martianoff/gala/issues/613)  | `.Size()` on a Go-declared receiver                      | closed                 | fixed for same-package declarations; an imported handwritten `.go` declaration is still unlowered |
| [#614](https://github.com/martianoff/gala/issues/614)  | Method call on a `:=`/`val` binding                      | closed                 | fixed (closure predates the lock)                                                                 |
| [#615](https://github.com/martianoff/gala/issues/615)  | False `GALA-E0044` for a sibling method                  | closed                 | fixed (closure predates the lock)                                                                 |
| [#616](https://github.com/martianoff/gala/issues/616)  | Bare name in a declared-type position                    | closed                 | fixed (closure predates the lock)                                                                 |
| [#617](https://github.com/martianoff/gala/issues/617)  | `func F() = <expr>` loses its result type                | closed                 | fixed (verified here)                                                                             |
| [#618](https://github.com/martianoff/gala/issues/618)  | `resource.Using` over a Go sibling type                  | closed                 | fixed (closure predates the lock)                                                                 |
| [#619](https://github.com/martianoff/gala/issues/619)  | Comments dropped from generated Go                       | closed                 | fixed (verified here)                                                                             |
| [#620](https://github.com/martianoff/gala/issues/620)  | `var (a, b) = <tuple>` panics                            | closed                 | fixed (closure predates the lock)                                                                 |
| [#621](https://github.com/martianoff/gala/issues/621)  | No newtype for non-struct types                          | closed                 | fixed (verified here)                                                                             |
| [#648](https://github.com/martianoff/gala/issues/648)  | Type-position import name is not checked                 | closed                 | fixed (verified here)                                                                             |
| [#678](https://github.com/martianoff/gala/issues/678)  | `gala-local` refused the stdlib `test` package           | closed                 | fixed (verified here)                                                                             |
| [PR #680](https://github.com/martianoff/gala/pull/680) | Local bootstrap gives batch files their package siblings | merged before the lock | included                                                                                          |
| [#691](https://github.com/martianoff/gala/issues/691)  | Lowercase sealed variants count for exhaustiveness       | closed                 | fixed (verified here)                                                                             |
| [#692](https://github.com/martianoff/gala/issues/692)  | Strict JSON decoder; YAML escapes                        | closed                 | fixed (verified here)                                                                             |
| [#695](https://github.com/martianoff/gala/issues/695)  | Go method multi-results lifted to `Try`/`Tuple`          | closed                 | fixed (verified here)                                                                             |
| [#698](https://github.com/martianoff/gala/issues/698)  | Publish the transpiled stdlib as a release asset         | open                   | n/a (packaging request)                                                                           |

The loop re-checks a report's finding on the pinned compiler before relying on it, because "closed" and "in the pinned rev" are different claims.

## Cursor

The cursor is ordered, and the loop takes the next entry and re-checks its constructs on the pinned compiler before translating.
A fixed blocker moves the candidate forward; a persistent blocker with no workaround becomes a Backlog task and a probe, and the candidate stays.

1. `slackbot/commands/active_users.go`, `slackbot/commands/utils.go` — the same `active_users` split as `discord_bot/commands`; `.Size()` takes sibling count helpers and the chart/response literals take a sibling constructor in `slackbot/commands/utils.go`.
2. `models/datetime.go`, `models/uuid.go`, `models/set.go`, `models/location.go`, `models/event.go` — defined-type methods, tags, fixed-size arrays, and `struct{}`.
3. `errs/errors.go` — tags and `interface{}`.
4. `routes/respondent_identity.go`, `routes/group.go`.
5. `postgres/` — the `Repository` declarations-only split; the sibling-method defect is closed before the lock.
6. `main.go`.
7. `observability/provider.go`, `observability/readiness.go`, `observability/transport.go`.
8. `services/auth`, `services/calendar`, `services/contacts`, `services/listmonk`, `services/microsoftgraph`.
9. `mockprovider`, `services/gcloud/tasks.go`.

## Stop conditions

Stop and report when any condition in [`Stop And Report`](../.agents/skills/gala-from-go/SKILL.md#stop-and-report) holds, and obey its `Never` list.
The project additions are:
no hand-edited generated twin; no transpile from the repository root; no reshaped Go-facing signature or wire format; no relaxed test; and no translation across a compiler rev mismatch, because a bump is its own re-baseline iteration.
