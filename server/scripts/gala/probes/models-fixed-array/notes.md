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
