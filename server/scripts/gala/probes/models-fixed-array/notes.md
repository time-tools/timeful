# models-fixed-array

`models/uuid.go` mints and formats UUIDs through a fixed-size `[16]byte` (`NewUUID`, `formatUUID`).
GALA has no fixed-size array type; `Array[T]` or a struct substitute changes the semantics.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-models-fixed-array.go
```

## Observed

```text
error: no viable alternative at input '[16'
  --> main.gala:6:13
   |
 6 | 	var value [16]byte
   | 	           ^^
   |

error: extraneous input ']' expecting {'{', '}', '(', '[', ';', '+', '-', '^', '*', '&', 'map', 'true', 'false', 'nil', '!', '<-', 'val', 'var', 'bind', 'also', 'use', 'func', 'type', 'if', 'for', 'return', 'import', INTERPOLATED_STRING, FORMAT_STRING, IDENTIFIER, INT_LIT, FLOAT_LIT, STRING, CHAR_LIT, RAW_STRING}
  --> main.gala:6:15
   |
 6 | 	var value [16]byte
   | 	             ^
   |
```

## Expected Go shape

```go
func formatUUID(value [16]byte) string { ... }
```

The array is both a local and a function parameter in `models/uuid.go`; a slice or struct substitute changes the function's Go-facing signature.

## Classification

Language gap (`references/gaps.md`): no GALA syntax expresses a fixed-size array.

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

## Status (2026-10-05)

`models/uuid.go` landed as `models/uuid.gala` plus `models/uuid_extra.go`; the fixed-size `[16]byte` helpers `NewUUID` and `formatUUID` moved into the sibling unchanged, so the construct still keeps any `.gala` file from declaring those members but no longer keeps the whole file handwritten.
This probe stays because the gap is still open on the pinned compiler.

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
