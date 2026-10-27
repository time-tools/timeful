# models-struct-tag

`models/location.go`, `models/event.go`, and the nested model structs carry JSON tags that define their wire shape.
The earlier probe tested only the shorthand form with a backtick tag and concluded that GALA had no tag syntax.
That conclusion was wrong: the block form accepts a Go tag as an interpreted string and copies it verbatim into the generated Go.
This probe records the working spelling, the emitted Go, the two rejected spellings, and the synthesized-member boundary that remains.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o /tmp/probe-models-struct-tag.go
```

## Working spelling

`main.gala` carries the `models.Location` shape with `var` fields and interpreted-string tags:

```gala
type Location struct {
    var CountryCode string "json:\"country_code\""
    ...
}
```

The emitted declaration keeps the Go field layout and tags exactly:

```go
type Location struct {
	CountryCode string  "json:\"country_code\""
	...
}
```

A scratch Go test marshaled the generated struct and a handwritten struct with the same fields and tags.
Both produced the same bytes, starting `{"country_code":"US","country_name":"United States",...}`, and `reflect.StructTag.Get("json")` returned `country_code`.
The generated file built in a scratch module requiring `go.gala.fyi/stdlib v0.93.0`, and `go test` passed.

The block form also emits `Copy`, `Equal`, `Unapply`, and `StructMeta_Location`, and imports `go.gala.fyi/stdlib/std`.
That is the remaining blocker for an exported DTO under the repository's surface rule; it is not a tag gap.

## Rejected spellings

The shorthand form has no tag slot:

```text
error: extraneous input '"json:\"country_code\""' expecting ')'
  --> short-quoted.gala:3:40
```

A raw-string tag is not in the grammar, even in the block form:

```text
error: extraneous input '`json:"country_code"`' expecting {'}', 'val', 'var', 'default', IDENTIFIER}
  --> block-backtick.gala:4:28
```

The shorthand form with a backtick tag fails the same way, so the only working tag spelling is the interpreted string with escaped quotes.
Every field must be `var` to stay a plain Go field; a non-`var` field becomes `std.Immutable[T]`.

## Classification

The tag itself is a `workaround` row in [`constructs.md`](../../../../.agents/skills/gala-from-go/references/constructs.md): the block form preserves the Go tag and the wire shape, while the shorthand form and a raw-string tag are parse errors.
A `Codec[T]` remains the substitute when the wire shape may move, not when a Go tool reads the tag.
The synthesized members are a boundary gap, and the opt-out report is conditional on the consuming repository's API rule; no tag gap is filed.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) closed 2026-10-04.
The block-form tag capability is a documentation defect: [#885](https://github.com/martianoff/gala/issues/885) (filed 2026-10-27).
The tag itself is not refiled; #528's answer treated tags as absent because the shorthand form was the only one tested.

## Compiler

- `gala version`: `GALA version 0.93.0`
- flake rev: `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`
- extraction: `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`
- date: 2026-10-27

## History

The shorthand parse error was re-verified on each earlier pin; the block-form spelling was first verified on 2026-10-27.

- `0.85.0` (flake rev `a888e824`, 2026-10-05): the shorthand backtick tag was a parse error.
- `0.90.0` (flake rev `014c2f23`, 2026-10-18): the same parse error; the expected-token list added the `default` keyword.
- `0.92.0` (flake rev `4bd2d4e8`, 2026-10-21): the same parse error.
- `0.92.1` (flake rev `911cc059`, 2026-10-24): the same parse error.
- `0.93.0` (flake rev `8ba02b2b`, 2026-10-25): the same shorthand parse error.
- `0.93.0` (flake rev `8ba02b2b`, 2026-10-27): the block-form interpreted-string tag emits the exact Go tag and passes the JSON comparison; the shorthand and raw-string spellings still fail.
