# GALA in the server

This document records the findings from rewriting server files in GALA and transpiling them back to Go.
It is the durable record for spike TASK-0319 and its follow-up TASK-0320.
The upstream language references are [GALA.MD](https://github.com/martianoff/gala/blob/master/docs/GALA.MD) and [GALA Best Practices](https://github.com/martianoff/gala/blob/master/docs/GALA_BEST_PRACTICES.MD).

## Current usage

Five packages carry a GALA source and a committed, generated Go twin.

| Package         | GALA source                 | Generated Go                    | Handwritten sibling                                                                     |
| --------------- | --------------------------- | ------------------------------- | --------------------------------------------------------------------------------------- |
| `eventid`       | `eventid/eventid.gala`      | `eventid/eventid.go`            | —                                                                                       |
| `observability` | `observability/redact.gala` | `observability/redact.go`       | —                                                                                       |
| `logger`        | `logger/logger.gala`        | `logger/logger.go`              | —                                                                                       |
| `appenv`        | `appenv/appenv.gala`        | `appenv/appenv.go`              | `appenv/appenv_port.go` (`ResolvePort`)                                                 |
| `utils`         | `utils/array_utils.gala`    | `utils/array_utils.go` (`Find`) | `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`) |

Regenerate a package from its own directory so the embedded `//line` directives stay relative:

```sh
cd server/eventid && gala transpile -i eventid.gala -o eventid.go
cd server/observability && gala transpile -i redact.gala -o redact.go
cd server/logger && gala transpile -i logger.gala -o logger.go
cd server/appenv && gala transpile -i appenv.gala -o appenv.go
cd server/utils && gala transpile -i array_utils.gala -o array_utils.go
```

A handwritten sibling has no regeneration command because it is not generated.
GALA 0.81.0 produced the committed output, and regeneration with the same version is deterministic.
Re-running the commands above must produce no diff.
Do not hand-edit a generated file, and never commit one without its `.gala` source.
The Go build never invokes GALA, which is why both artifacts are committed.
`eventid/doc.go` carries the package comment that the transpiler does not emit.
The transpiler writes an analysis cache under `server/.gala/`, which is gitignored.

## Ground rules

Two output styles are now committable, and the choice depends on whether the file needs the GALA runtime.

### Runtime-free output

Use the runtime-free style for leaf twins whose exported Go API must stay exactly Go-shaped:

- Import Go packages directly, for example `import "regexp"`.
- Bind raw Go values with `var`, not `val`.
- Avoid the GALA standard library (`Option`, `Try`, immutable collections).
- Use `.ByteSize()` or `.Size()` instead of `len`, and `go_interop` wrappers instead of `make` or `append`.

This is what `eventid`, `observability`, `logger`, and the generated half of `utils` use.
Its generated Go imports only the packages the source names, which keeps the twin readable next to its Go callers.

### Runtime-enabled output

Use the runtime-enabled style when the file benefits from GALA-native features and its Go callers tolerate the emitted shapes:

- `val`, `Option`, `Try`, `Either`, sealed types, and `collection_immutable` compile because `server/go.mod` resolves the vendored runtime.
- Generated code imports `martianoff/gala/std`, `martianoff/gala/go_interop`, or another vendored subpackage, every one of which resolves through the single local replace.
- Every `val` emits a `std.Immutable[...]` wrapper, so an exported package-level `val` changes the Go type a caller sees, and Go code must call `.Get()` to unwrap it.
- Prefer `var` when a value crosses into Go or a handwritten sibling, and keep `val` for local, GALA-internal bindings.

## What works well

- Expression-bodied functions, for example `func Canonical(id string) bool = canonicalShortID.MatchString(id)`.
- Direct Go standard library calls, including multi-value calls received with `var`, such as `var port, err = strconv.Atoi(value)`.
- `match` over strings, which compiles to a plain Go `if`/`else` chain with no runtime imports.
- `for ... range` loops, `break`, char literals such as `' '` and `'\t'`, raw-string regexes, and `string(rune)` conversions.
- Package-level `var` declarations with inferred types, with explicit types, and with no initializer at all, such as `var StdOut *log.Logger` in `logger.gala`.
- Interface-typed parameters and package-level pointer variables, which `logger` now relies on.
- Bitwise-or flag expressions such as `log.Ldate|log.Ltime|log.Lmsgprefix|log.Llongfile`.
- `if`/`else` chains over normalized strings, including `string(Environment)` conversions, for `appenv`'s parse and port helpers.
- Generic functions with function parameters, such as `Find[T any](arr []T, equals func(T) bool) int`.
- The split pattern: a generated `.go` file and a handwritten sibling coexist in one package, and the package tests pass unchanged (`appenv`, `utils`).
- Sealed types: the transpiler emits the parent struct with a `_variant` discriminator, companion structs, `Apply`/`Unapply`, `is*` helpers, `Copy`, `Equal` via `std.Equal`, and a `String` method; an exhaustive `match` compiles to a closure over `Unapply`.
- `Try(strconv.Atoi(text))` auto-wraps a Go `(T, error)` call, converts a non-nil error into `Failure`, and supports `ToOption`, `Map`, `FlatMap`, `Fold`, and `GetOrElse`.
- `collection_immutable` `Array` and `HashMap` constructors with `Map`, `Filter`, `FoldLeft`, `GetOption`, `Get`, and `Size`.
- `go_interop.Spawn` and `concurrent.Future` compile and run, and the data-race check fires as documented.
- Deterministic, gofmt-formatted output with a `DO NOT EDIT` header and `//line` directives back to the source.

## What does not work

- Go-style multi-value returns.
  A declaration like `func f() (string, error)` is a parse error, and GALA returns tuples instead.
  This blocks drop-in rewrites of Go helpers such as `appenv.ResolvePort`, `utils.CORSOrigins`, and `utils.GetListmonkOtpFromAddress`.
  The workaround is the split: keep multi-value members in a handwritten sibling and transpile the rest, which `appenv` and `utils` now do.
- `type X Y` emits a Go alias.
  The transpiler turns `type DateTime int64` into `type DateTime = int64`, so any method on a named scalar type is illegal Go.
  A reduced `datetime.gala` probe (type, two methods, constructor) transpiles cleanly and then fails to compile:

  ```text
  # models
  datetime.gala:7: cannot define new methods on non-local type DateTime
  datetime.gala:9: cannot define new methods on non-local type DateTime
  datetime.gala:12: invalid composite literal type DateTime
  ```

  `models/datetime.go`, `models/uuid.go`, and `models/set.go` are blocked by this rule and stay handwritten.

- Backtick struct tags.
  Every tagged field is a parse error, so the wire shape cannot be reproduced and `models/location.go` is blocked:

  ```text
  error: extraneous input '`json:"country_code"`' expecting {'}', 'val', 'var', IDENTIFIER}
    --> location.gala:4:22
    |
  4 | 	CountryCode string  `json:"country_code"`
    | 	                    ^^^^^^^^^^^^^^^^^^^^^
    |
  ```

- `const`.
  A package constant must be a `var`, or a `val` that wraps the value in `std.Immutable` and drags in the runtime.

  ```text
  error: extraneous input 'const' expecting {<EOF>, 'val', 'var', 'func', 'type', 'struct', 'import', 'sealed', 'embed'}
    --> uuid.gala:10:1
     |
  10 | const zeroUUIDValue = "00000000-0000-0000-0000-000000000000"
     | ^^^^^
     |
  ```

- `interface{}` field or parameter types.
  `errs/errors.go` is blocked by `interface{}` fields on top of its tags:

  ```text
  error: mismatched input 'interface' expecting {'[', '*', 'map', 'func', IDENTIFIER}
    --> errors.gala:19:10
     |
  19 | 	Details interface{} `json:"details"`
     | 	        ^^^^^^^^^
     |
  ```

- `struct{}` as a type expression.
  `type Set[T comparable] map[T]struct{}` and its element literal do not parse, which blocks `models/set.go` and `utils.ArrayToSet`:

  ```text
  error: mismatched input 'struct' expecting {'[', '*', 'map', 'func', IDENTIFIER}
    --> set.gala:3:30
    |
  3 | type Set[T comparable] map[T]struct{}
    |                              ^^^^^^
    |
  ```

- Fixed-size arrays.
  There is no `[N]T` grammar, so `models/uuid.go`'s `[16]byte` cannot be expressed:

  ```text
  error: mismatched input '[' expecting {'=', '{'}
    --> uuid_arrays.gala:3:17
    |
  3 | func NewValue() [16]byte {
    |                 ^
    |
  ```

- Go `switch` and multi-case case labels.
  `models/uuid.go`'s `case 4, 6, 8, 10:` in a switch over the loop index does not parse; ordinary bitwise indexing and `strings.Builder` calls do parse:

  ```text
  error: no viable alternative at input '{case4,'
    --> uuid_switch.gala:9:9
    |
  9 | 		case 4, 6, 8, 10:
    | 		      ^
    |
  ```

- GALA struct declarations are not drop-in twins.
  Every GALA struct gains exported `Copy`, `Equal`, `Unapply`, and `Is<Type>` methods plus an exported `<Type>Instance` interface, and fields are wrapped in `std.Immutable[T]` unless declared with `var`.
  `var` fields emit plain Go types, but the synthesized members still grow the exported API, so an existing Go struct stays handwritten.
- `panic`, `append`, and `make` as bare builtins.
  Use `go_builtins.Panic`, `go_interop.SliceAppend`, and `go_interop.MapEmpty`/`MapPut` instead.
- `[]byte(s)` conversions.
  Use `go_interop.ToBytes`; `string(bytes)` works directly.
- Documentation comments.
  Generated Go contains no comments, so the package comment for `eventid` lives in a handwritten `doc.go`.
- Byte-level string slicing.
  `value[:n]` does not parse, so `truncate` uses a rune-accumulation loop that preserves the original byte semantics.
- Matching a string-typed constant such as `case Development =>`, which binds a new name instead of comparing; `string(Development)` is not an extractor either.
  `appenv` therefore uses `if`/`else` with `string(...)` comparisons.
- Imports do not propagate between sibling `.gala` files.
  Each file must import what it uses, which a concurrency probe hit as `error[GALA-E0025]: undefined: Future ... 'concurrent' is not imported in this file`.
- The transpiler does not know types declared in sibling `.go` files.
  A method on a Go-declared type transpiles and compiles, but constructing that type emits a composite literal, as the reduced datetime probe shows with `invalid composite literal type DateTime`.
- An exported package-level `val` becomes `var X = std.NewImmutable(...)`, so Go callers see a `std.Immutable[T]` wrapper rather than the underlying value.

## Runtime probes

Four probes in `/tmp` closed the TASK-0319 "Not yet exercised" list against the vendored runtime.
Each probe was transpiled with GALA 0.81.0 and compiled in a scratch module whose `go.mod` replaces `martianoff/gala` with `server/third_party/gala`.

| Probe            | Transpile | Compiles in module | Go API shape                                                                                                                                                           |
| ---------------- | --------- | ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Sealed type      | clean     | yes                | `type Shape struct` with `std.Immutable` fields and a `_variant uint8`; companions `Circle`, `Rectangle`, `Point` with `Apply`/`Unapply`; `Copy`, `Equal`, `String`    |
| `val` and monads | clean     | yes                | package and local `val` lower to `var x = std.NewImmutable(...)` with automatic `.Get()`; exported functions keep plain Go return types when the body unwraps          |
| Collections      | clean     | yes                | `Array[int]` and `HashMap[string, int]` parameters and returns; method calls lower to plain calls or package helpers such as `Array_FoldLeft`                          |
| Concurrency      | clean     | yes                | `Future[int]{}.Apply(func() int {...})` and `go_interop.Spawn(func() {...})`; `Await()` returns `std.Try[T]`; a reassignable var capture is rejected with `GALA-E0037` |

The four probes in detail:

1. Sealed types.
   `sealed type Shape { case Circle(Radius float64); case Rectangle(Width float64, Height float64); case Point }` produces a merged parent struct with one field per variant, wrapped in `std.Immutable`, plus a private `_variant uint8` and an iota block of `_Shape_*` constants.
   `Match` lowers to an immediately-invoked closure over the companion's `Unapply`, with `panic("unreachable")` on the impossible final branch.
   A runtime check returned `circle 2.00` for a `Circle` and exhaustive matches with `case _` compiled.
2. `val` with `Option`, `Try`, and `Either`.
   Package-level `val answer = 42` emits `var answer = NewImmutable(42)`, and `func Answer() int = answer` emits `return answer.Get()`, so unexported `val` does not leak the wrapper.
   An exported `val ExportedAnswer = 99` emits `var ExportedAnswer = std.NewImmutable(99)`, so its Go type is `std.Immutable[int]` and the exported API shape changes.
   `Try(strconv.Atoi(text))` emits a closure that panics on the Go error and is recovered by `Try.Apply`; `Describe("42")` returned `parsed 42` and `Describe("nope")` returned the `strconv.Atoi` error text as a `Failure`.
   `Either` with `Map` and `Fold` returned `value: 4` and `error: division by zero`, and `ParseAndDouble("21").GetOrElse(0)` returned `42`.
3. Collections.
   `Filter` and `Map` stay method calls, `FoldLeft` lowers to `Array_FoldLeft`, and `.Size()` stays a method call.
   A `val` holding a `HashMap` is wrapped, so `m.Get("a")` in GALA lowers to `m.Get().Get("a")` in Go.
   Runtime results: even count `2`, sum `10`, `GetOption(2)` `3`, `ArrayOf(1, 2, 3).Size()` `3`, lookup `got 1`, map-then-sum `100`.
   `HashMap` has no `GetOption`; `Get` is the lookup method, and using the wrong name produces `error[GALA-E0044]: HashMap has no method GetOption`.
4. Concurrency.
   `Future(42)` emits `Future[int]{}.Apply(func() int { return 42 })`, `Await()` returns `std.Try[int]`, and `GetOrElse` recovers the value.
   `go_interop.Spawn` emits a plain `go func` wrapper and is unchecked; a probe that read a captured var immediately after `Spawn` observed the zero value because the goroutine had not run yet.
   Capturing a reassignable `var` in a `Future` body is rejected before codegen:

   ```text
   error[GALA-E0037]: closure crossing a concurrency boundary captures reassignable var "count" — a data race, because the enclosing scope may reassign it while the goroutine runs
     --> bad_capture.gala:7:16
     |
   7 | 	return Future(count + 1)
     | 	              ^^^^^ snapshot it into an immutable `val` before the boundary
     |
   ```

Not yet exercised after these probes: `collection_mutable`, the `json`/`yaml` codecs, `stream`, `fs`, `io`, `subprocess`, `time_utils`, `validation`, `embed val`, `Try` over three-or-more-value Go returns, generic sealed types, and `Future` timeouts and cancellation.

## Vendoring the GALA runtime

The runtime is vendored at `server/third_party/gala/` as one flattened Go module named `martianoff/gala`.

- Provenance: the gala CLI v0.81.0 extraction at `~/.gala/stdlib/v0.81.0`, which is the ready-to-compile snapshot built from `martianoff/gala` 0.81.0.
  The Apache-2.0 `LICENSE` was copied from the upstream source checkout.
  Neither the extraction directory nor the upstream checkout shipped a license file next to the snapshot.
- Layout: all 20 packages (`collection_immutable`, `collection_mutable`, `concurrent`, `crypto`, `fs`, `go_builtins`, `go_interop`, `io`, `json`, `lazy`, `path`, `regex`, `resource`, `std`, `stream`, `strings`, `subprocess`, `test`, `time_utils`, `validation`, `yaml`) live as subdirectories of the module.
  Per-package `go.mod` and `go.sum` files are deleted, and a single root `go.mod` declares `module martianoff/gala` and `go 1.22`.
  The `.gala` sources stay beside their `.gen.go` outputs so each package's origin is visible.
- Why flattened: generated code imports `martianoff/gala/std`, `martianoff/gala/go_interop`, `martianoff/gala/go_builtins`, and similar paths.
  Nested modules would carve those packages out of the parent module, and the snapshot's per-package `replace` directives only work inside its original multi-module layout.
  One root module makes every import path resolve as a subdirectory.
- Wiring: `server/go.mod` has `require martianoff/gala v0.0.0` and `replace martianoff/gala => ./third_party/gala`.
  A local replace needs no `go.sum` entry.
- Docker: `server/Dockerfile` copies `third_party/gala/go.mod` before `RUN go mod download`, because the local replace must resolve inside the dependency layer even though the module has no external dependencies.
  The build context and the route-test mount are both `./server`, so the rest of `third_party/` arrives with `COPY . .`.
- Upstream vet noise: `go build ./...` over the vendored tree is clean, while `go vet ./...` reports five pre-existing findings in the snapshot (a `ReadRune` signature suggestion and unreachable-code warnings).
  They are upstream, not introduced here.
- Update procedure: install the matching gala CLI version, re-copy `~/.gala/stdlib/<version>` over the vendored tree, delete the extraction marker and per-package module files, rewrite the root `go.mod`, restore the `LICENSE`, run `go build ./...` inside the vendored tree, run the server build and tests, then regenerate every twin and compare hashes.

## Verification

- The isolated Compose stack ran `go test ./... -count=1` with every package green, including `routes`, `postgres`, and the wire-shape-sensitive `models` package.
- `go build ./...` and `go vet ./logger/ ./appenv/ ./utils/` pass on the server module.
- The Compose image build path rebuilt `server` from the changed `Dockerfile` and produced the server binary.
- `appenv`, `utils`, and `models` tests pass unchanged against the committed twins.
- Regenerating each of the five twins twice produced byte-identical output; the sha256 of the committed files matched both runs:

  | Committed file            | sha256                                                             |
  | ------------------------- | ------------------------------------------------------------------ |
  | `eventid/eventid.go`      | `dfa1bff953a8dc528eddc45f6a900010f9cf5fffa00a9d1c3c77ad13a468b988` |
  | `observability/redact.go` | `b16f0b85b5e9c713810151450df5e6fa75b43a49fce9dea9a2dcea7b0b76274d` |
  | `logger/logger.go`        | `ff196cb720fd218da1b61af6fb6ae4f52964d59ddb740a9d779da804de4a5a58` |
  | `appenv/appenv.go`        | `6a4dfd20650a517932caf39ae08d4e93d105de6dc9ed0ea5637c083572f95762` |
  | `utils/array_utils.go`    | `5b303a4077bdf47e79f1525b4e15f177db69bd2aead3fbcce6023eb5fe869276` |

- The probe module compiled and ran against the vendored runtime, and its `go run` output confirmed the sealed, monad, collection, and concurrency shapes described above.

## When to revisit

Adoption beyond mostly interop-shaped leaf packages is still blocked by upstream gaps, in rough priority order:

- Go-style multi-value return signatures, or an interop escape hatch for them.
  This is the top ask; it alone would unblock `appenv.ResolvePort`, `utils.FindAddedRemovedKept`, and every marshaler in `models`.
- Preservation of struct tags, so JSON-shaped structs like `models/location.go` and `errs/errors.go` can be rewritten.
- Defined types for non-struct types, or an explicit newtype declaration, so named scalars with methods (`DateTime`, `UUID`) can move.
- `const`, `interface{}`, fixed-size arrays, and Go `switch` with multi-case labels, all currently parse errors.
- Documentation comments in generated Go, which stay in handwritten `doc.go` files.

Until then, use GALA where the exported Go API stays Go-shaped, the file needs no struct tags, named scalar receivers, or multi-value returns, and either the runtime-free style applies or the runtime-enabled shape is acceptable at the boundary.
