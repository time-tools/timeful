# GALA in the server

This document records the findings from rewriting server files in GALA and transpiling them back to Go.
It is the durable record for spikes TASK-0319, TASK-0320, and TASK-0321.
The upstream language references are [GALA.MD](https://github.com/martianoff/gala/blob/master/docs/GALA.MD) and [GALA Best Practices](https://github.com/martianoff/gala/blob/master/docs/GALA_BEST_PRACTICES.MD).

## Current usage

Thirteen packages carry a GALA source and a committed, generated Go twin (fifteen `.gala` sources in total).

| Package                   | GALA source                                   | Generated Go                                | Handwritten sibling                                                                     |
| ------------------------- | --------------------------------------------- | ------------------------------------------- | --------------------------------------------------------------------------------------- |
| `eventid`                 | `eventid/eventid.gala`                        | `eventid/eventid.go`                        | —                                                                                       |
| `observability`           | `observability/redact.gala`                   | `observability/redact.go`                   | —                                                                                       |
| `logger`                  | `logger/logger.gala`                          | `logger/logger.go`                          | —                                                                                       |
| `appenv`                  | `appenv/appenv.gala`                          | `appenv/appenv.go`                          | `appenv/appenv_port.go` (`ResolvePort`)                                                 |
| `utils`                   | `utils/array_utils.gala`                      | `utils/array_utils.go` (`Find`)             | `utils/array_utils_extra.go` (`ArrayToSet`, `ElementWithIndex`, `FindAddedRemovedKept`) |
| `utils`                   | `utils/request_utils.gala`                    | `utils/request_utils.go`                    | —                                                                                       |
| `services`                | `services/services.gala`                      | `services/services.go`                      | —                                                                                       |
| `services/providerconfig` | `services/providerconfig/providerconfig.gala` | `services/providerconfig/providerconfig.go` | `services/providerconfig/doc.go` (package comment)                                      |
| `routes`                  | `routes/guest_response_ownership.gala`        | `routes/guest_response_ownership.go`        | —                                                                                       |
| `discord_bot/commands`    | `discord_bot/commands/help.gala`              | `discord_bot/commands/help.go`              | —                                                                                       |
| `discord_bot/commands`    | `discord_bot/commands/num_users.gala`         | `discord_bot/commands/num_users.go`         | —                                                                                       |
| `discord_bot`             | `discord_bot/init.gala`                       | `discord_bot/init.go`                       | —                                                                                       |
| `slackbot/commands`       | `slackbot/commands/num_users.gala`            | `slackbot/commands/num_users.go`            | `slackbot/commands/utils.go` (`newResponse`)                                            |
| `middleware`              | `middleware/auth.gala`                        | `middleware/auth.go`                        | `middleware/doc.go` (package comment)                                                   |
| `postgres`                | `postgres/dailylogs.gala`                     | `postgres/dailylogs.go`                     | `postgres/dailylogs_methods.go` (the four `*Repository` daily-log methods)              |

Regenerate a package from its own directory so the embedded `//line` directives stay relative:

```sh
cd server/eventid && gala transpile -i eventid.gala -o eventid.go
cd server/observability && gala transpile -i redact.gala -o redact.go
cd server/logger && gala transpile -i logger.gala -o logger.go
cd server/appenv && gala transpile -i appenv.gala -o appenv.go
cd server/utils && gala transpile -i array_utils.gala -o array_utils.go
cd server/utils && gala transpile -i request_utils.gala -o request_utils.go
cd server/services && gala transpile -i services.gala -o services.go
cd server/services/providerconfig && gala transpile -i providerconfig.gala -o providerconfig.go
cd server/routes && gala transpile -i guest_response_ownership.gala -o guest_response_ownership.go
cd server/discord_bot/commands && gala transpile -i help.gala -o help.go
cd server/discord_bot/commands && gala transpile -i num_users.gala -o num_users.go
cd server/discord_bot && gala transpile -i init.gala -o init.go
cd server/slackbot/commands && gala transpile -i num_users.gala -o num_users.go
```

A handwritten sibling has no regeneration command because it is not generated.
GALA 0.84.1 produced the committed output, and regeneration with the same version is deterministic.
Re-running the commands above must produce no diff.
Do not hand-edit a generated file, and never commit one without its `.gala` source.
The Go build never invokes GALA, which is why both artifacts are committed.
The transpiler emits no comments, so package comments live in handwritten `eventid/doc.go` and `services/providerconfig/doc.go`, and the comment on an exported declaration is lost when its file is transpiled (`services/providerconfig`).
The transpiler writes an analysis cache under `server/.gala/`, which is gitignored.

## Ground rules

Two output styles are now committable, and the choice depends on whether the file needs the GALA runtime.

### Runtime-free output

Use the runtime-free style for leaf twins whose exported Go API must stay exactly Go-shaped:

- Import Go packages directly, for example `import "regexp"`.
- Bind raw Go values with `var`, not `val`.
- Avoid the GALA standard library (`Option`, `Try`, immutable collections).
- Use `.ByteSize()` or `.Size()` instead of `len`, and `go_interop` wrappers instead of `make` or `append`.

This is what `eventid`, `observability`, `logger`, `request_utils`, `services`, `guest_response_ownership`, `providerconfig`, the generated half of `utils`, and the Discord and Slack command files use.
Its generated Go imports only the packages the source names, which keeps the twin readable next to its Go callers.

### Runtime-enabled output

Use the runtime-enabled style when the file benefits from GALA-native features and its Go callers tolerate the emitted shapes:

- `val`, `Option`, `Try`, `Either`, sealed types, and `collection_immutable` compile because `server/go.mod` resolves the vendored runtime.
- Generated code imports `martianoff/gala/std`, `martianoff/gala/go_interop`, or another vendored subpackage, every one of which resolves through the single local replace.
- Every `val` emits a `std.Immutable[...]` wrapper, so an exported package-level `val` changes the Go type a caller sees, and Go code must call `.Get()` to unwrap it.
- Prefer `var` when a value crosses into Go or a handwritten sibling, and keep `val` for local, GALA-internal bindings.
- `discord_bot/init.gala` is the first committed file that needs the runtime: it uses `go_interop.MapEmpty` and `go_interop.SliceFrom` because map literals, `make`, and slice expressions do not parse.
- A bare type or value name that an imported package exports resolves to that package even when a handwritten sibling `.go` declares the same name locally; `slackbot/commands/num_users.gala` calls the handwritten `newResponse` helper instead of naming the local `Response` type, which otherwise transpiles to `postgres.Response`.

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
- Multi-value Go calls received with `var a, b = f()` and reassigned with `a, b = f()`; blank results work, as in `var bodyBytes, _ = json.Marshal(body)`.
- Blank parameters such as `_ string` in `shouldExposeGuestSignUpResponsePayload`, and package-level `var` composite literals of Go-declared sibling types, including named unexported functions assigned to function-typed fields in `discord_bot/commands`.
- Map index assignment on a `go_interop.MapEmpty` map, as in `commandMap[command.Name] = command`.
- `go_interop` helpers `MapEmpty`, `MapPut`, `SliceFrom`, and `SliceAppend` replace `make`, map literals, slice expressions, and `append` in runtime-enabled files such as `discord_bot/init.gala`.
- Deterministic, gofmt-formatted output with a `DO NOT EDIT` header and `//line` directives back to the source.

## What does not work

- A `swag` annotation on a handler.
  The transpiler emits no comments at all, so an annotation written in the `.gala` source is absent from the generated Go and `swag init` drops the endpoint from the OpenAPI document.
  A handwritten sibling cannot supply it either, because an annotation has to sit immediately above its declaration.
  A package comment _is_ recoverable from a `doc.go` sibling; a declaration comment and an annotation are not, because the only spelling is a bodiless re-declaration, which is a Go error.
  `routes/users.go` is the one annotated file the translation roster reports as rewritable, and it stays handwritten for this reason.
  49 `@Router` directives live across 12 files here, so every route handler is exposed to this and the file-level verdict does not show it.
  Filed upstream as [#619](https://github.com/martianoff/gala/issues/619).
- Calling a method that a handwritten sibling declares.
  `GALA-E0044`, and it is loud: `Repository has no method withTransaction`.
  The transpiler resolves a method call against the declarations in the same `.gala` file and reads no Go sibling.
  The refusal only fires when the _type_ is declared in the transpiled file: a method on a type the transpiler does not know passes through and Go resolves it, so the failure is a declarations-in-GALA split rather than any mixed package.
  `Repository` is declared in `repository.go`, which stays handwritten, and its 106 methods are spread across twelve files, so twelve of the fourteen rung-2 `postgres` candidates are blocked by this alone.
  The only viable shape is a declarations-only split, which is what `dailylogs` now is.
- `.Size()` on a receiver whose type was inferred from a Go call.
  The transpiler types a receiver from the `.gala` file's own text, so a receiver written out resolves even across packages while one inferred from a Go-declared function does not, even in the same package.
  This is what blocks both `active_users.go` files, which are rung 1 in the roster.
  It is not the absence of a substitute: `go_interop.SliceFrom(logs, 0).Size()` reaches the same length without naming the type, and `SliceFrom` is `s[from:]`, so it is a view and allocates nothing.
  A multi-value binding still cannot be annotated, which is why the roster classes this as a workaround rather than a boundary gap, and a `string` receiver still has no zero-cost spelling because `go_interop` ships `MapLen` and `SliceCap` but no length helper for either kind.
  The lowering defect is filed upstream as [#613](https://github.com/martianoff/gala/issues/613).
- A method call on a `:=` or `val` binding.
  `usersRouter := router.Group("/users")` followed by `usersRouter.GET(...)` emits `usersRouter.Get().GET(...)`, a call that does not exist in Go.
  `val` emits the same thing, so `var` is the only safe spelling of the three rather than a style preference.
  The transpile is clean and the failure appears only at `go build`, and only for a pointer-receiver method: a value-receiver method on the unwrapped copy builds and runs, so the same source is either a build failure or a quiet copy.
  The `var` spelling is correct and emits no wrapper at all.
- Go-style multi-value returns.
  A declaration like `func f() (string, error)` is a parse error, and GALA returns tuples instead.
  This blocks drop-in rewrites of Go helpers such as `appenv.ResolvePort`, `utils.CORSOrigins`, and `utils.GetListmonkOtpFromAddress`.
  The workaround is the split: keep multi-value members in a handwritten sibling and transpile the rest, which `appenv` and `utils` now do.
  This is boundary gap `GAP-2` in the translation roster.
- Receiving a multi-value Go call with `:=`.
  This used to abort with an internal transpiler panic; PR [#529](https://github.com/martianoff/gala/pull/529) fixed it and it transpiles on the pinned release.
  It is a val-style binding rather than a neutral one, so every twin still takes the `var` spelling: `:=` wraps each result and the transpiler inserts `.Get()` at every read, while `var a, b = f()` emits a plain Go binding.

  ```text
  // n, err := strconv.Atoi("42")
  var (
  	_tmp_1, _tmp_2 = strconv.Atoi("42")
  	n              = std.NewImmutable(_tmp_1)
  	err            = std.NewImmutable(_tmp_2)
  )
  ```

- `if err := f(); err != nil`.
  The initializer slot is rejected at transpile time as `GALA-E0047`; use `var err = f()` on the previous line:

  ```text
  error[GALA-E0047]: `if` takes no initializer statement
    --> probe_ifinit.gala:6:1
    |
  6 | 	if err := strconv.Atoi(text); err != nil {
    | 	^^ wrap the call in `Try(...)` and `match` on `Success(v)` / `F…
    |
  ```

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
  This is language gap `GAP-4`; 0.84.1 rejects the declaration as `GALA-E0048`, because an alias takes no methods of its own, while alias conversions themselves now work.

- Backtick struct tags.
  Every tagged field is a parse error, so the wire shape cannot be reproduced and `models/location.go` is blocked.
  This is language gap `GAP-1` in the translation roster, and 0.84.1 leaves it unchanged:

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

  This is boundary gap `GAP-5` in the translation roster.

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

  This is language gap `GAP-8` in the translation roster.

- `switch` statements.
  No `switch` parses, single-case or multi-case, which blocks `models/uuid.go`, `observability/record.go`, `services/calendar/types.go`, and `routes/respondent_identity.go`; a plain `switch` fails at the first case label:

  ```text
  error: no viable alternative at input '{case0:'
    --> probe_switch.gala:5:8
    |
  5 | 	case 0:
    | 	      ^
    |
  ```

  The `models/uuid.go` probe hit the same rule on its multi-case labels (`error: no viable alternative at input '{case4,'`), and ordinary bitwise indexing and `strings.Builder` calls around it do parse.
  The replacement is `match` or an `if`/`else` chain.

- `defer`.
  GALA has no `defer`; the vendored runtime uses `resource.Using`, and everything else must close explicitly, which blocks most of `postgres`, `main.go`, `observability/{provider,readiness,transport}`, `routes/group.go`, `services/{auth,calendar,contacts,listmonk,microsoftgraph}`, and `slackbot/commands/utils.go`:

  ```text
  error[GALA-E0036]: bare Go statement keyword "defer" is not part of GALA's surface
    --> probe_defer_plain.gala:6:2
    |
  6 | 	defer Reset()
    | 	^^^^^ GALA has no `defer`
    |
  ```

- `len()`.
  The call is rejected with a hint, and `.Size()` is rune-based while `.ByteSize()` preserves Go byte semantics; `.Size()` on a field access such as `command.Usage.Size()` does not lower and emits invalid Go (`command.Usage.Size undefined (type string has no field or method Size)`), so the twins compare strings with `!= ""`:

  ```text
  error[GALA-E0035]: bare Go builtin "len(...)" is not part of GALA's surface
    --> probe_len.gala:4:9
    |
  4 | 	return len(s)
    | 	       ^^^ use `.Size()`
    |
    = hint: use `.Size()` (logical size — characters for strings) or `.ByteSize()` (raw bytes) instead of `len(...)`
  ```

- Type assertions.
  `v.(T)` does not parse, which blocks `middleware/auth.go` and its `session.Get("userId").(string)`:

  ```text
  error: extraneous input '(' expecting IDENTIFIER
    --> probe_c.gala:4:13
    |
  4 | 	s, ok := v.(string)
     | 	           ^
     |
  ```

  A type-pattern `match` (`case s: string`) lowers to `std.As[T]` and replaces the comma-ok form, so this row is an analog rather than a gap (`pass_type_match`).

- Slice expressions.
  `args[1:]` does not parse; `discord_bot/init.gala` uses `go_interop.SliceFrom(args, 1)`:

  ```text
  error: extraneous input ':' expecting ']'
    --> init.gala:83:15
     |
  83 | 	args = args[1:]
     | 	             ^
     |
  ```

- Map literals.
  `map[string]T{...}` is not a first-class construct, which blocks `mockprovider`, `services/gcloud/tasks.go`, and both `active_users.go` files; `discord_bot/init.gala` uses `go_interop.MapEmpty` and map index assignment instead:

  ```text
  error[GALA-E0008]: map literals are not a first-class GALA construct
    --> probe_e.gala:3:13
    |
  3 | var empty = map[string]Handler{}
    |             ^^^ use collection_immutable.HashMap or collection_mutable.HashM…
    |
  ```

- Inline function literals inside composite literals.
  `Command{Execute: func(...) {...}}` does not parse, so the Discord and Slack command files declare a named unexported function and reference it from the literal:

  ```text
  error: extraneous input 'return' expecting {'{', '}', '(', '[', '+', '-', '^', '*', '&', 'map', 'true', 'false', 'nil', '!', '<-', 'func', 'if', INTERPOLATED_STRING, FORMAT_STRING, IDENTIFIER, INT_LIT, FLOAT_LIT, STRING, CHAR_LIT, RAW_STRING}
    --> probe_a.gala:5:31
    |
  5 | 	Run: func(s string) string { return s },
    | 	                             ^^^^^^
    |
  ```

- GALA struct declarations are not drop-in twins.
  Every GALA struct gains exported `Copy`, `Equal`, and `Unapply` methods, a generic struct additionally gains an exported `<Type>Instance` interface and an `Is<Type>()` method, and fields are wrapped in `std.Immutable[T]` unless declared with `var`.
  `var` fields emit plain Go types, but the synthesized members still grow the exported API, so an existing Go struct stays handwritten.
- `panic`, `append`, and `make` as bare builtins.
  Use `go_builtins.Panic`, `go_interop.SliceAppend`, and `go_interop.MapEmpty`/`MapPut` instead.
- `[]byte(s)` conversions.
  Use `go_interop.ToBytes`; `string(bytes)` works directly.
- Documentation comments and every other comment.
  Generated Go contains no comments, so package comments live in handwritten `eventid/doc.go` and `services/providerconfig/doc.go`, and comments attached to exported declarations are lost when the file is transpiled.
  Both comment syntaxes parse and neither reaches the emitted Go, so this is an emit gap rather than a grammar gap, which [`emit_comments_dropped`](scripts/20260923_gala_translation_probes/probes/emit_comments_dropped/) pins by asserting both comment bodies are absent from `main.gen.go`.
- A bare name in a declared-type position.
  When an imported package exports a name a sibling `.gala` file also declares, the return type, parameter type, and `var` annotation resolve to the import while a constructor position resolves locally, so one file can end up using two different types under one name.
  Only a declaration in the same `.gala` file resolves correctly, which is why `slackbot/commands/utils.go` carries a handwritten `newResponse`.
  Filed upstream as [#616](https://github.com/martianoff/gala/issues/616).
- A resource combinator over a Go-declared resource type.
  `resource.Using(res, body)` infers both type arguments when the resource type is declared in a `.gala` file, and upstream's own example relies on that; with a handwritten Go sibling it binds the body parameter to `any`, and a partial type-argument list emits the transpiler's own parameter name as a Go type.
  Filed upstream as [#618](https://github.com/martianoff/gala/issues/618).
- Two forms that lose a result type.
  `func F() = <expr>` never infers one, so the emitted Go function is void and its body returns a value; `var (a, b) = f()` panics the transpiler with `GALA-E0017` where `val` transpiles.
  Filed upstream as [#617](https://github.com/martianoff/gala/issues/617) and [#620](https://github.com/martianoff/gala/issues/620).
- Byte-level string slicing.
  `value[:n]` does not parse, so `truncate` uses a rune-accumulation loop that preserves the original byte semantics.
- Matching a string-typed constant such as `case Development =>`, which binds a new name instead of comparing; `string(Development)` is not an extractor either.
  `appenv` therefore uses `if`/`else` with `string(...)` comparisons.
- Imports are checked per file, and a missing import is a refusal rather than a lookup.
  Each file must import what it uses; a qualified name the file does not import is refused at transpile time with `error[GALA-E0023]: undefined: <pkg>`, whether or not another file in the package imports that package, and whether that other file is a `.gala` file or a handwritten `.go` one.
  An unqualified GALA-runtime name is the same rule from the other side, and its hint names the package that declares it: a concurrency probe calling `Future(2)` without importing `martianoff/gala/concurrent` reads `error[GALA-E0023]: undefined: Future` with `Future is declared in the GALA package "martianoff/gala/concurrent", which this file does not import`.
  `gala explain GALA-E0025` states the same per-file rule and says outright that sibling files' imports do not propagate, but `GALA-E0025` did not fire in any shape measured on 0.84.1, so it is the authority for the rule rather than for the code the pinned compiler emits.
  The corpus pins all of it in [`blocked_import_omitted_gala_sibling`](scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_gala_sibling/), [`blocked_import_omitted_go_sibling`](scripts/20260923_gala_translation_probes/probes/blocked_import_omitted_go_sibling/), and [`blocked_runtime_name_omitted_import`](scripts/20260923_gala_translation_probes/probes/blocked_runtime_name_omitted_import/).
  The construct roster's `named-imports` row records the import syntax itself as a direct form, and this per-file check is the limit that row does not carry.
- Name collisions with imported packages.
  A bare type name that an imported package exports resolves to that package even when a sibling `.go` declares it locally; `slackbot/commands/num_users.gala` calling `Response{...}` emitted `pgstore.Response{...}` and failed to compile, while the same pattern works for `Command` because no import exports that name.
  The workaround is a handwritten constructor (`newResponse` in `slackbot/commands/utils.go`).
- An exported package-level `val` becomes `var X = std.NewImmutable(...)`, so Go callers see a `std.Immutable[T]` wrapper rather than the underlying value.

## Runtime probes

Four probes in `/tmp` closed the TASK-0319 "Not yet exercised" list against the vendored runtime.
Each probe was transpiled with GALA 0.84.1 and compiled in a scratch module whose `go.mod` replaces `martianoff/gala` with `server/third_party/gala`.

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

- Provenance: the gala CLI v0.84.1 extraction at `~/.gala/stdlib/v0.84.1`, which is the ready-to-compile snapshot built from `martianoff/gala` 0.84.1.
  The Apache-2.0 `LICENSE` was copied from the upstream source checkout.
  Neither the extraction directory nor the upstream checkout shipped a license file next to the snapshot.
- Layout: all 21 packages (`collection_immutable`, `collection_mutable`, `concurrent`, `crypto`, `fs`, `go_builtins`, `go_interop`, `io`, `json`, `lazy`, `path`, `regex`, `resource`, `std`, `stream`, `strings`, `subprocess`, `test`, `time_utils`, `validation`, `yaml`) live as subdirectories of the module.
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
  The 0.84.1 snapshot reports the same five findings, two of them in `std` and three in `json` plus `yaml`.
- Update procedure: install the matching gala CLI version, re-copy `~/.gala/stdlib/<version>` over the vendored tree, delete the extraction marker and per-package module files, rewrite the root `go.mod`, restore the `LICENSE`, run `go build ./...` inside the vendored tree, run the server build and tests, then regenerate every twin and compare hashes.

## Verification

- The isolated Compose stack ran `go test ./... -count=1` with every package green, including `routes`, `postgres`, and the wire-shape-sensitive `models` package.
- `go build ./...` and `go vet` pass on the server module for every touched package, and `gofmt -l` reports no generated file.
- The Compose image build path rebuilt `server` from the changed `Dockerfile` and produced the server binary.
- `appenv`, `utils`, `routes`, `discord_bot`, `slackbot`, and `providerconfig` tests pass unchanged against the committed twins.
- Regenerating each twin twice produced byte-identical output; the sha256 of the committed files matched both runs:

  | Committed file                              | sha256                                                             |
  | ------------------------------------------- | ------------------------------------------------------------------ |
  | `eventid/eventid.go`                        | `dfa1bff953a8dc528eddc45f6a900010f9cf5fffa00a9d1c3c77ad13a468b988` |
  | `observability/redact.go`                   | `b16f0b85b5e9c713810151450df5e6fa75b43a49fce9dea9a2dcea7b0b76274d` |
  | `logger/logger.go`                          | `ff196cb720fd218da1b61af6fb6ae4f52964d59ddb740a9d779da804de4a5a58` |
  | `appenv/appenv.go`                          | `6a4dfd20650a517932caf39ae08d4e93d105de6dc9ed0ea5637c083572f95762` |
  | `utils/array_utils.go`                      | `5b303a4077bdf47e79f1525b4e15f177db69bd2aead3fbcce6023eb5fe869276` |
  | `utils/request_utils.go`                    | `b636a83486e1e33019846cb2d30c094c7ebcd52747dcd4c0c1b03e36dd11f7f3` |
  | `services/services.go`                      | `0c65dab9869fca16f43063ed8605e37801e8d0e8d4d19946a083b778eeac3ce2` |
  | `services/providerconfig/providerconfig.go` | `e77c2513833030c645b1e502fb76346fa76fed7a61b3daf49a246992aee59d64` |
  | `routes/guest_response_ownership.go`        | `5ef803dfe4f3a13f72d8726974b338134ec9a09f768d2f60920c49bcc7c6e450` |
  | `discord_bot/commands/help.go`              | `45afe9bae3ebd48520d7e1b3f34fdba754b83c35e7d002e3ae57a742f2d0a637` |
  | `discord_bot/commands/num_users.go`         | `d699d3116829b604fd048f8f9f5d6082a31e122d0f6ec0b7e9fd60b58ccd830e` |
  | `discord_bot/init.go`                       | `87dcafb9086c5764ec97d1d1125778291a79c667824f074c995f558e38bfd6cb` |
  | `slackbot/commands/num_users.go`            | `6ddeea45ec41555cd07751c52125cc3de794231b8599e5586ca8f2868aed00f9` |
  | `middleware/auth.go`                        | `3b1cb9a2a86d94ae8906a4c2950963e3a367b90a9786eadf4c97c63753a69743` |
  | `postgres/dailylogs.go`                     | `ae3fe90adc34a683ed2bf6fa91f960cce2b23b3a6ae17c064b2e524e2e15c078` |

- The probe module compiled and ran against the vendored runtime, and its `go run` output confirmed the sealed, monad, collection, and concurrency shapes described above.

## When to revisit

Adoption beyond mostly interop-shaped leaf packages is still limited by the upstream gaps the roster classifies; in rough priority order:

- `GAP-11`, now filed upstream as [#619](https://github.com/martianoff/gala/issues/619): a comment-preserving emit, or an interop escape hatch letting a generated file carry a Go-facing comment.
  A `swag` annotation is the case that bites here, because it is what deletes a documented endpoint from the OpenAPI document rather than only a godoc comment.
- Upstream reports filed on 2026-09-30 from minimal reproductions, none of which is a substitute problem:
  [#614](https://github.com/martianoff/gala/issues/614) for a method-call receiver on a `:=` or `val` binding, [#615](https://github.com/martianoff/gala/issues/615) for the false-positive `GALA-E0044`, [#616](https://github.com/martianoff/gala/issues/616) for a bare name in a declared-type position, [#617](https://github.com/martianoff/gala/issues/617) and [#620](https://github.com/martianoff/gala/issues/620) for the two forms that lose a result type, and [#618](https://github.com/martianoff/gala/issues/618) for a resource combinator over a Go-declared type.
  The two feature requests the maintainer invited in #528 are [#619](https://github.com/martianoff/gala/issues/619) for comments in generated Go and [#621](https://github.com/martianoff/gala/issues/621) for a newtype over a non-struct type.
- `GAP-2`: Go-style multi-value return signatures, or an interop escape hatch for them.
  The `:=` receive panic that used to be part of this ask was fixed by PR [#529](https://github.com/martianoff/gala/pull/529); what remains is the signature itself.
  This is the top ask; it alone would unblock `appenv.ResolvePort`, `routes/respondent_identity.go`, every marshaler in `models`, and most of `postgres`.
- `GAP-1`: preservation of struct tags, so JSON-shaped structs like `models/location.go` and `errs/errors.go` can be rewritten.
- `GAP-4`: defined types for non-struct types, or an explicit newtype declaration, so named scalars with methods (`DateTime`, `UUID`) can move.
- `GAP-3`, `GAP-5`, `GAP-6`, `GAP-7`, `GAP-8`, `GAP-9`, and `GAP-10`: anonymous and empty struct types, channel types at the boundary, embedded fields, fixed-size arrays, `select`, and in-place `recover`.
- Beyond the gaps, a `switch` lowering to `match` and a `defer` replacement that accepts `(T, error)` acquires would remove the manual rewrites for provider dispatch and most of `postgres`.
- Documentation comments in generated Go, which stay in handwritten `doc.go` files.

Until then, use GALA where the exported Go API stays Go-shaped, the file needs no struct tags, named scalar receivers, multi-value returns, `switch`, or `defer`, and either the runtime-free style applies or the runtime-enabled shape is acceptable at the boundary.

The per-construct roster with the inventory counts, the 0.84.1 status column, the compiler history for the PR #529 fixes, the classified genuine gaps, and the scripted probe corpus is in [`../docs/gala-translation.md`](../docs/gala-translation.md).
The ordered procedure for adding another twin, from candidate selection to registration, is in [`../docs/gala-rewrite-playbook.md`](../docs/gala-rewrite-playbook.md).
