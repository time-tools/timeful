# GALA in the server

This document records the findings from rewriting server files in GALA and transpiling them back to Go.
It is the durable record for spike TASK-0319.
The upstream language references are [GALA.MD](https://github.com/martianoff/gala/blob/master/docs/GALA.MD) and [GALA Best Practices](https://github.com/martianoff/gala/blob/master/docs/GALA_BEST_PRACTICES.MD).

## Current usage

Two packages carry a GALA source and a committed, generated Go twin.

| Package | GALA source | Generated Go |
| --- | --- | --- |
| `eventid` | `eventid/eventid.gala` | `eventid/eventid.go` |
| `observability` | `observability/redact.gala` | `observability/redact.go` |

Regenerate a package from its own directory so the embedded `//line` directives stay relative:

```sh
cd server/eventid && gala transpile -i eventid.gala -o eventid.go
cd server/observability && gala transpile -i redact.gala -o redact.go
```

GALA 0.81.0 produced the committed output, and regeneration with the same version is deterministic.
Re-running the commands above must produce no diff.
Do not hand-edit a generated file, and never commit one without its `.gala` source.
The Go build never invokes GALA, which is why both artifacts are committed.
`eventid/doc.go` carries the package comment that the transpiler does not emit.
The transpiler writes an analysis cache under `server/.gala/`, which is gitignored.

## Ground rules for runtime-free output

The generated Go must compile in the existing module without new dependencies.
This code therefore follows GALA's Go-interop guidance rather than full idiomatic GALA:

- Import Go packages directly, for example `import "regexp"`.
- Bind raw Go values with `var`, not `val`.
  Every `val` emits `std.NewImmutable(...)` and imports `martianoff/gala/std`, which `server/go.mod` cannot resolve.
- Avoid the GALA standard library (`Option`, `Try`, immutable collections) for the same reason.
- Use `.ByteSize()` or `.Size()` instead of `len`, and `go_interop` wrappers instead of `make` or `append`.

## What works well

- Expression-bodied functions, for example `func Canonical(id string) bool = canonicalShortID.MatchString(id)`.
- Direct Go standard library calls, including multi-value calls received with `var`, such as `var port, err = strconv.Atoi(value)`.
- `match` over strings, which compiles to a plain Go `if`/`else` chain with no runtime imports.
- `for ... range` loops, `break`, char literals such as `' '` and `'\t'`, raw-string regexes, and `string(rune)` conversions.
- Package-level `var` declarations with inferred types.
- Mutable interop parameters, for example `func truncate(var value string, limit int) string`.
- Deterministic, gofmt-formatted output with a `DO NOT EDIT` header and `//line` directives back to the source.

## What does not work

- Go-style multi-value returns.
  A declaration like `func f() (string, error)` is a parse error, and GALA returns tuples instead.
  This blocks drop-in rewrites of Go helpers such as `appenv.ResolvePort`, `utils.CORSOrigins`, and `utils.GetListmonkOtpFromAddress`.
- `const`.
  A package constant must be a `var`, or a `val` that wraps the value in `std.Immutable` and drags in the runtime.
- Any `val` at all in runtime-free output.
  The `std` runtime ships as GALA source under `~/.gala/stdlib/<version>` with machine-local `replace` directives, so generated code that imports it cannot be committed without vendoring the runtime or publishing it as a normal Go module.
- Documentation comments.
  Generated Go contains no comments, so the package comment for `eventid` lives in a handwritten `doc.go`.
- Byte-level string slicing.
  `value[:n]` does not parse, so `truncate` uses a rune-accumulation loop that preserves the original byte semantics.
- Matching a string-typed constant such as `case Development =>`, which binds a new name instead of comparing; `string(Development)` is not an extractor either.

## Not yet exercised

Structs and methods, sealed types, generics, collections, error monads, and concurrency have not been tried in this codebase.
Each needs either the GALA runtime or a runtime-free design, so evaluate them together with the runtime-dependency question before relying on them.

## Verification

- The isolated Compose stack runs `go test ./... -count=1` with every package green.
- `go build ./...` and `go vet ./eventid/ ./observability/` pass.
- A differential test compared the transpiled `truncate` and `Redact` against the original Go implementations on 200,000 random strings (ASCII, whitespace, multi-byte runes, and invalid UTF-8) and 20,000 redaction input combinations, with identical outputs.
- Regenerating both files twice produced byte-identical output, confirmed by sha256.

## When to revisit

Adoption beyond interop-shaped leaf packages is blocked until at least one of these changes lands upstream:

- Go-style multi-value return signatures, or an interop escape hatch for them.
- A `std` runtime that the server can depend on through `go.mod`, or an officially supported runtime-free mode.
- Preservation of documentation comments in generated Go.
- Byte-level string operations without `go_interop`.

Until then, use GALA only where the exported Go API stays Go-shaped, the functions return at most one value, and the file needs no GALA runtime imports.
