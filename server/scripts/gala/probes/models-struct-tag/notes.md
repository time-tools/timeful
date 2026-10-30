# models-struct-tag

`models/location.go`, `models/event.go`, and the nested model structs carry JSON tags that define their wire shape.
GALA 0.95.0 rejects a Go struct tag on any field form with `GALA-E0085`, whose hint names `Codec[T]` for a wire shape GALA controls and a handwritten `.go` file for a type a Go reflection library must read through tags.
Before 0.95.0 the block form accepted the tag as an interpreted string and copied it verbatim into the generated Go; this probe records that removed capability, the current rejection, and the synthesized-member boundary that remains.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o /tmp/probe-models-struct-tag.go
```

## Working spelling on 0.93.0 and earlier (removed in 0.95.0)

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

The shorthand form with a backtick tag fails the same way, so on 0.93.0 the only working tag spelling was the interpreted string with escaped quotes.
Every field had to be `var` to stay a plain Go field; a non-`var` field becomes `std.Immutable[T]`.

## Current behavior (0.95.0)

The block-form spelling above is now refused, and every tagged field reports `GALA-E0085`; `main.gala` produces seven errors, one per field:

```text
error[GALA-E0085]: struct field CountryCode has a tag: GALA has no struct tags
  --> main.gala:4:28
   |
 4 |     var CountryCode string "json:\"country_code\""
   |                            ^^^^^^^^^^^^^^^^^^^^^^^ remove the tag
   |
   = hint: remove the tag; set JSON and YAML keys on Codec[T] with .Rename(field, key), .Omit and .OmitEmpty, and write a type that a Go library reads through its tags in a .go file of the package
```

`gala explain GALA-E0085` states that the diagnostic covers a block-form struct, a shorthand struct, and a sealed case, and both the quoted and the raw-string spelling.
The error is reported after the field's type, so it replaces the earlier parse errors for the shorthand and raw-string forms rather than leaving them reachable.

## Classification

The construct is a `documented answer` row in [`constructs.md`](../../../../.agents/skills/gala-from-go/references/constructs.md) (`answered`, `GALA-E0085`): GALA has no struct tags, and the hint names both substitutes — `Codec[T]` (`.Rename`, `.Omit`, `.OmitEmpty`) for a wire shape GALA controls, and a handwritten `.go` file for a type a Go reflection library reads through its tags.
No gap is filed: [#885](https://github.com/martianoff/gala/issues/885) closed completed by the rejection in PR #894, and the 0.95.0 release notes record it as a breaking change.
The synthesized members are a separate boundary gap, and the opt-out report is conditional on the consuming repository's API rule.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) closed 2026-10-04.
The block-form tag capability was filed as a documentation defect in [#885](https://github.com/martianoff/gala/issues/885) on 2026-10-27; upstream closed it completed on 2026-10-28 by rejecting tags as `GALA-E0085` in [PR #894](https://github.com/martianoff/gala/pull/894) instead of documenting the workaround.

## Compiler

- `gala version`: `GALA version 0.93.0`
- flake rev: `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`
- extraction: `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`
- date: 2026-10-27

## Re-verified (2026-10-29)

- `gala version`: `GALA version 0.95.0`
- flake rev: `173f7d778f103a05553564d841e0daba0126cb16`
- extraction: `0.95.0 7b68112a768d6d4776444c8ab561ce629c219d86748f93f0ca1f485daf7d1648`
- Result: the block-form interpreted-string tag now fails with seven `GALA-E0085` diagnostics, one per field, and no Go is emitted; the hint names `Codec[T]` and a handwritten `.go` file.

## History

The shorthand parse error was re-verified on each earlier pin; the block-form spelling was first verified on 2026-10-27 and removed upstream on 2026-10-28.

- `0.85.0` (flake rev `a888e824`, 2026-10-05): the shorthand backtick tag was a parse error.
- `0.90.0` (flake rev `014c2f23`, 2026-10-18): the same parse error; the expected-token list added the `default` keyword.
- `0.92.0` (flake rev `4bd2d4e8`, 2026-10-21): the same parse error.
- `0.92.1` (flake rev `911cc059`, 2026-10-24): the same parse error.
- `0.93.0` (flake rev `8ba02b2b`, 2026-10-25): the same shorthand parse error.
- `0.93.0` (flake rev `8ba02b2b`, 2026-10-27): the block-form interpreted-string tag emits the exact Go tag and passes the JSON comparison; the shorthand and raw-string spellings still fail.
- `0.95.0` (flake rev `173f7d77`, 2026-10-29): the block-form spelling is refused with `GALA-E0085`; the workaround is gone and a tagged type belongs in a handwritten `.go` file.
