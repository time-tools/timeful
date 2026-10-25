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

## Re-verified (2026-10-18)

- `gala version`: `GALA version 0.90.0`
- flake rev: `014c2f23e45a2cd631b73cab38c0bdcb5e773347`
- extraction: `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`
- Result: the same two parse errors; the expected-token lists now include the new `default` keyword.

## Re-verified (2026-10-21)

- `gala version`: `GALA version 0.92.0`
- flake rev: `4bd2d4e89cb8699e94b1f209356dd679db8d54ac`
- extraction: `0.92.0 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same two parse errors; the expected-token lists still name the `default` keyword.

## Re-verified (2026-10-24)

- `gala version`: `GALA version 0.92.1`
- flake rev: `911cc05932486626e684931aac9f1707939a91e6`
- extraction: `0.92.1 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same two parse errors; the expected-token lists still name the `default` keyword.

## Re-verified (2026-10-25)

- `gala version`: `GALA version 0.93.0`
- flake rev: `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`
- extraction: `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`
- Result: the same two parse errors; the expected-token lists still name the `default` keyword.
