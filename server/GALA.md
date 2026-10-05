# GALA in the server

This is the single GALA ledger for the server: which twins exist, which compiler and runtime they were built with, what still blocks the rest of the cursor, and how to continue.
It does not restate construct substitutions or gap classification; those live in the project-independent translation reference at [`../.agents/skills/gala-from-go/SKILL.md`](../.agents/skills/gala-from-go/SKILL.md).
The loop that consumes both is [`../.agents/skills/gala-loop/SKILL.md`](../.agents/skills/gala-loop/SKILL.md), and verification is [`scripts/gala/verify.sh`](scripts/gala/verify.sh).

## Contract

Both halves of a twin are committed: the `.gala` source and the Go file the transpiler generates from it.
The Go build never invokes GALA, so a generated file without its source is unreadable to the next maintainer and a source without its generated file is not in the build at all.
Never hand-edit a generated file; it carries a `DO NOT EDIT` header and the next transpile discards the edit without a trace.
Transpile from the package directory so the emitted `//line` directives stay relative: `cd server/<pkg> && gala transpile -i <name>.gala -o <name>.go`.
The transpiler writes its analysis cache under `server/.gala/`, which is gitignored.

Two output styles are committable, and the choice is a decision rather than a style preference.

- Runtime-free output suits a leaf twin whose exported Go API must stay Go-shaped: import Go packages directly, bind raw Go values with `var`, avoid the GALA standard library, and use `.ByteSize()`/`.Size()` plus the `go_interop` helpers in place of `len`, `make`, `append`, and `[]byte`.
  The generated Go imports only what the source names, which keeps the twin readable beside its Go callers.
- Runtime-enabled output suits a file that benefits from GALA-native features and whose Go callers tolerate the emitted shapes: `val`, `Option`/`Try`/`Either`, sealed types, and immutable collections compile through the vendored runtime, and the generated Go imports it.
  Every `val` or `:=` binding becomes a `std.Immutable` wrapper that a Go caller has to unwrap, so keep `var` for anything that crosses into Go or a handwritten sibling.

Read the emitted Go before committing a new twin, because a wrapper the program never observes still changes what a Go caller sees.

## Compiler and runtime

The translation target is the `gala` compiler the dev shell puts on `PATH`, and the vendored runtime under `third_party/gala/` is always re-vendored to match it before translating.
The compiler is the flake-locked commit; `gala version` reports `0.85.0`, and the current rev `d38fadb` adds the post-release fixes [#691](https://github.com/martianoff/gala/issues/691), [#692](https://github.com/martianoff/gala/issues/692), [#695](https://github.com/martianoff/gala/issues/695), and [#702](https://github.com/martianoff/gala/issues/702) on top of the release that carried the #648 and #621 fixes.

| Item             | Value                                                                                                                      |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `gala version`   | `GALA version 0.85.0`                                                                                                      |
| flake rev        | `d38fadb844f0849c6c6ac8a47e8f91e025eace6b` (`jq -r '.nodes.gala.locked.rev' flake.lock`)                                   |
| vendored runtime | [`third_party/gala/`](third_party/gala/), provenance in [`third_party/gala/VENDORED_FROM`](third_party/gala/VENDORED_FROM) |
| pinned tools     | `nix develop` provides `gala` and the Go toolchain from `flake.lock`                                                       |

Record the provenance before any iteration:

```sh
gala version
which gala
jq -r '.nodes.gala.locked.rev' flake.lock
cat server/third_party/gala/VENDORED_FROM
```

The vendored runtime is one flattened Go module named `martianoff/gala`.
`server/go.mod` requires it at `v0.0.0` and replaces it with the local directory, so generated imports such as `martianoff/gala/std` and `martianoff/gala/go_interop` resolve with no network fetch and no `go.sum` entry.
`server/Dockerfile` copies `third_party/gala/go.mod` before `go mod download`, because the local replace has to resolve inside the dependency layer.
The Apache-2.0 `LICENSE` is copied from the upstream source checkout; the CLI extraction does not ship one.
The committed copy is a stopgap: [#698](https://github.com/martianoff/gala/issues/698) asks upstream to publish the transpiled tree as a per-release, checksum-pinned asset, which would replace the vendoring step with a download.

When the flake rev and `VENDORED_FROM` disagree, stop translating and run a sync iteration first.
The version string alone cannot identify the compiler, because several commits share one version string, so the extraction is proved by the fingerprint in its `.stdlib-extracted` marker.

1. Enter `nix develop` so `gala` is the flake-locked compiler, and record `gala version`, `which gala`, and `jq -r '.nodes.gala.locked.rev' flake.lock`.
2. Force the pinned compiler to write its own extraction: delete `~/.gala/stdlib/v<version>/.stdlib-extracted`, then run one transpile from a package directory, for example `cd server/eventid && gala transpile -i eventid.gala -o /tmp/eventid.go`.
   The transpiler resolves the stdlib on every run and rewrites the version directory when the marker is missing or names a different embedded snapshot, so the extraction at that path afterwards belongs to the pinned compiler.
3. Read `~/.gala/stdlib/v<version>/.stdlib-extracted`; it is `<version> <snapshot fingerprint>` and is recorded in `VENDORED_FROM` as the extraction provenance.
4. Copy `~/.gala/stdlib/v<version>/` over `third_party/gala/`.
5. Delete the per-package `go.mod` and `go.sum` files and the `.stdlib-extracted` marker that the extraction ships.
6. Keep the `LICENSE`, and rewrite the root `go.mod` as one module: `module martianoff/gala` with the snapshot's `go` directive.
7. Run `go build ./...` inside `third_party/gala/`.
8. Write `VENDORED_FROM` with the compiler version, the flake rev, the source path, the extraction marker contents, and the date.
9. Run `scripts/gala/verify.sh --write`, review the per-twin diff, then run `go build ./...` in `server/` and the canonical backend test sequence in [`README.md`](README.md).

A rev mismatch is resolved by a sync iteration and never by translating across it, because the compiler and the vendored runtime are pinned together and a twin generated by one against the other is not reproducible.

## Twin registry

Fifteen `.gala` sources across thirteen packages carry a committed twin.
Every row regenerates with `cd server/<dir> && gala transpile -i <name>.gala -o <name>.go`; a handwritten sibling has no command because it is not generated.

| Package                   | GALA source                                   | Generated Go                                | Handwritten sibling                                                                     | Style           |
| ------------------------- | --------------------------------------------- | ------------------------------------------- | --------------------------------------------------------------------------------------- | --------------- |
| `eventid`                 | `eventid/eventid.gala`                        | `eventid/eventid.go`                        | —                                                                                       | runtime-free    |
| `observability`           | `observability/redact.gala`                   | `observability/redact.go`                   | —                                                                                       | runtime-free    |
| `logger`                  | `logger/logger.gala`                          | `logger/logger.go`                          | —                                                                                       | runtime-free    |
| `appenv`                  | `appenv/appenv.gala`                          | `appenv/appenv.go`                          | `appenv/appenv_port.go` (`ResolvePort`)                                                 | runtime-free    |
| `utils`                   | `utils/array_utils.gala`                      | `utils/array_utils.go`                      | `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`) | runtime-free    |
| `utils`                   | `utils/request_utils.gala`                    | `utils/request_utils.go`                    | —                                                                                       | runtime-free    |
| `services`                | `services/services.gala`                      | `services/services.go`                      | —                                                                                       | runtime-free    |
| `services/providerconfig` | `services/providerconfig/providerconfig.gala` | `services/providerconfig/providerconfig.go` | `services/providerconfig/doc.go` (package comment)                                      | runtime-free    |
| `routes`                  | `routes/guest_response_ownership.gala`        | `routes/guest_response_ownership.go`        | —                                                                                       | runtime-free    |
| `discord_bot/commands`    | `discord_bot/commands/help.gala`              | `discord_bot/commands/help.go`              | —                                                                                       | runtime-free    |
| `discord_bot/commands`    | `discord_bot/commands/num_users.gala`         | `discord_bot/commands/num_users.go`         | —                                                                                       | runtime-free    |
| `discord_bot`             | `discord_bot/init.gala`                       | `discord_bot/init.go`                       | —                                                                                       | runtime-enabled |
| `slackbot/commands`       | `slackbot/commands/num_users.gala`            | `slackbot/commands/num_users.go`            | `slackbot/commands/utils.go` (`newResponse`)                                            | runtime-free    |
| `middleware`              | `middleware/auth.gala`                        | `middleware/auth.go`                        | `middleware/doc.go` (package comment)                                                   | runtime-enabled |
| `postgres`                | `postgres/dailylogs.gala`                     | `postgres/dailylogs.go`                     | `postgres/dailylogs_methods.go` (the four `*Repository` daily-log methods)              | runtime-enabled |

`scripts/gala/verify.sh` checks that every one of them regenerates byte-identically, is `gofmt`-clean, and keeps `go build ./...` green.
Adding a twin means adding its registry row here and passing `verify.sh`.

## Open findings

These keep the cursor's candidates handwritten or split, and each is classified against [`references/gaps.md`](../.agents/skills/gala-from-go/references/gaps.md).
A finding gains a committed repro under `scripts/gala/probes/<slug>/` when the loop first reaches a candidate it blocks, and the probe is removed when the finding closes.

| Finding                                          | Class        | Workaround                                                                                        | Upstream                                              | Probe |
| ------------------------------------------------ | ------------ | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------- | ----- |
| Go-style multi-value return signature            | boundary gap | split the member into a handwritten sibling, or return `Tuple[T, error]` to GALA-internal callers | none filed                                            | —     |
| Struct tags that define a wire shape             | boundary gap | keep the tagged struct handwritten; `Codec` changes the wire shape                                | [#528](https://github.com/martianoff/gala/issues/528) | —     |
| Anonymous struct types                           | boundary gap | declare a named struct, which changes the type identity                                           | [#528](https://github.com/martianoff/gala/issues/528) | —     |
| `struct{}` as a type expression                  | boundary gap | declare a named empty struct, which changes the element type identity                             | [#528](https://github.com/martianoff/gala/issues/528) | —     |
| A channel type that crosses the package boundary | boundary gap | `concurrent.Future`/`go_interop` only when no channel type crosses                                | none filed                                            | —     |
| Embedded fields                                  | language gap | flatten by hand and lose promotion                                                                | none filed                                            | —     |
| Fixed-size array types such as `[16]byte`        | language gap | a slice or a struct, which changes the semantics                                                  | [#528](https://github.com/martianoff/gala/issues/528) | —     |
| `select`                                         | language gap | none; the repository's uses pair it with `chan` and `make`                                        | none filed                                            | —     |
| In-place `recover`                               | language gap | `Try` captures a panic as a value but cannot resume in place                                      | none filed                                            | —     |

Declaration comments and `swag` annotations are no longer a finding: the pinned compiler emits them, so `routes/users.go` is unblocked (see the report index for [#619](https://github.com/martianoff/gala/issues/619)).
The post-release [#695](https://github.com/martianoff/gala/issues/695) makes a Go method's several results one `Try`/`Tuple` value like a Go function's, but a GALA function still cannot declare a Go-style multi-value signature (a transpile parse error), so the Go-style multi-value return signature row stays.
Fixed-size arrays and anonymous or empty struct types remain the reason `models/uuid.go`, `models/event.go`, and `models/set.go` stay handwritten.

## Upstream report index

Reports this repository filed, and their state on the pinned compiler rev.
A report closed after the flake lock is fixed upstream but not yet in the pinned compiler, so its finding stays open until the next sync.

| #                                                      | Title                                                    | Upstream state         | On the pinned rev                 |
| ------------------------------------------------------ | -------------------------------------------------------- | ---------------------- | --------------------------------- |
| [#528](https://github.com/martianoff/gala/issues/528)  | Triage language limitations                              | open                   | open                              |
| [PR #529](https://github.com/martianoff/gala/pull/529) | Multi-value define and alias fixes                       | merged in 0.84.1       | included                          |
| [#611](https://github.com/martianoff/gala/issues/611)  | `fallthrough` in a `match` arm                           | closed                 | fixed (closure predates the lock) |
| [#612](https://github.com/martianoff/gala/issues/612)  | `match` on `var` fields emits an unwrap                  | closed                 | fixed (closure predates the lock) |
| [#613](https://github.com/martianoff/gala/issues/613)  | `.Size()` on a Go-declared receiver                      | closed                 | fixed (closure predates the lock) |
| [#614](https://github.com/martianoff/gala/issues/614)  | Method call on a `:=`/`val` binding                      | closed                 | fixed (closure predates the lock) |
| [#615](https://github.com/martianoff/gala/issues/615)  | False `GALA-E0044` for a sibling method                  | closed                 | fixed (closure predates the lock) |
| [#616](https://github.com/martianoff/gala/issues/616)  | Bare name in a declared-type position                    | closed                 | fixed (closure predates the lock) |
| [#617](https://github.com/martianoff/gala/issues/617)  | `func F() = <expr>` loses its result type                | closed                 | fixed (verified here)             |
| [#618](https://github.com/martianoff/gala/issues/618)  | `resource.Using` over a Go sibling type                  | closed                 | fixed (closure predates the lock) |
| [#619](https://github.com/martianoff/gala/issues/619)  | Comments dropped from generated Go                       | closed                 | fixed (verified here)             |
| [#620](https://github.com/martianoff/gala/issues/620)  | `var (a, b) = <tuple>` panics                            | closed                 | fixed (closure predates the lock) |
| [#621](https://github.com/martianoff/gala/issues/621)  | No newtype for non-struct types                          | closed                 | fixed (verified here)             |
| [#648](https://github.com/martianoff/gala/issues/648)  | Type-position import name is not checked                 | closed                 | fixed (verified here)             |
| [#678](https://github.com/martianoff/gala/issues/678)  | `gala-local` refused the stdlib `test` package           | closed                 | fixed (verified here)             |
| [PR #680](https://github.com/martianoff/gala/pull/680) | Local bootstrap gives batch files their package siblings | merged before the lock | included                          |
| [#691](https://github.com/martianoff/gala/issues/691)  | Lowercase sealed variants count for exhaustiveness       | closed                 | fixed (verified here)             |
| [#692](https://github.com/martianoff/gala/issues/692)  | Strict JSON decoder; YAML escapes                        | closed                 | fixed (verified here)             |
| [#695](https://github.com/martianoff/gala/issues/695)  | Go method multi-results lifted to `Try`/`Tuple`          | closed                 | fixed (verified here)             |
| [#698](https://github.com/martianoff/gala/issues/698)  | Publish the transpiled stdlib as a release asset         | open                   | n/a (packaging request)           |

The loop re-checks a report's finding on the pinned compiler before relying on it, because "closed" and "in the pinned rev" are different claims.

## Cursor

The cursor is ordered, and the loop takes the next entry and re-checks its constructs on the pinned compiler before translating.
A fixed blocker moves the candidate forward; a persistent blocker with no workaround becomes a Backlog task and a probe, and the candidate stays.

1. `routes/users.go` — unblocked by the comment emit; verify the `@Router` block survives and `swag init` still emits the endpoints.
2. `discord_bot/commands/active_users.go`, `slackbot/commands/active_users.go`, `slackbot/commands/utils.go` — previously blocked by the inferred-receiver `.Size()` and bare-name type-position defects, both closed before the lock.
3. `models/datetime.go`, `models/uuid.go`, `models/set.go`, `models/location.go`, `models/event.go` — defined-type methods, tags, fixed-size arrays, and `struct{}`.
4. `errs/errors.go` — tags and `interface{}`.
5. `routes/respondent_identity.go`, `routes/group.go`.
6. `postgres/` — the `Repository` declarations-only split; the sibling-method defect is closed before the lock.
7. `main.go`.
8. `observability/provider.go`, `observability/readiness.go`, `observability/transport.go`.
9. `services/auth`, `services/calendar`, `services/contacts`, `services/listmonk`, `services/microsoftgraph`.
10. `mockprovider`, `services/gcloud/tasks.go`.

## Stop conditions

Stop and report when any condition in [`Stop And Report`](../.agents/skills/gala-from-go/SKILL.md#stop-and-report) holds, and obey its `Never` list.
The project additions are:
no hand-edited generated twin; no transpile from the repository root; no reshaped Go-facing signature or wire format; no relaxed test; and no translation across a compiler or runtime rev mismatch, because the re-vendor is its own iteration.
