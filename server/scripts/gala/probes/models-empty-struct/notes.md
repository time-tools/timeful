# models-empty-struct

`models/set.go` declares `type Set[T comparable] map[T]struct{}`.
GALA has no `struct{}` type expression; a named empty struct changes the map's element type identity.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-models-empty-struct.go
```

## Observed

```text
error: mismatched input 'struct' expecting {'[', '*', 'map', 'func', IDENTIFIER}
  --> main.gala:5:30
   |
 5 | type Set[T comparable] map[T]struct{}
   |                              ^^^^^^
   |

error: mismatched input '{' expecting IDENTIFIER
  --> main.gala:5:36
   |
 5 | type Set[T comparable] map[T]struct{}
   |                                    ^
   |
```

## Expected Go shape

```go
type Set[T comparable] map[T]struct{}
```

Set literals and membership checks are built on the zero-size element type, so any named empty struct is a different type.

## Classification

Boundary gap (`references/gaps.md`): GALA expresses the map, but not with the Go-shaped element type.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) carried this triage and closed upstream on 2026-10-04; the construct still fails on the pinned compiler.
No specific follow-up report was found.

## Compiler

- `gala version`: `GALA version 0.85.0`
- flake rev: `a888e824ff53adb653bbd48dcc83f67788eeced4`
- extraction: `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`
- date: 2026-10-05
