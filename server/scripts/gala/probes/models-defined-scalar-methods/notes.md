# models-defined-scalar-methods

`models/datetime.go` and `models/uuid.go` declare methods on defined scalar types (`type DateTime int64`, `type UUID string`).
GALA refuses that declaration with `GALA-E0048` and names `opaque type` as the substitute.
The substitute transpiles, but its generated Go imports `martianoff/gala/std` for the synthesized `Hash` and `Compare` methods, so it cannot enter a runtime-free twin.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-models-defined-scalar-methods.go
```

The alias form is refused before codegen.
Its minimal repro is `type DateTime int64` followed by `func (d DateTime) IsZero() bool = d == 0`, transpiled the same way.

## Observed

The `opaque type` form transpiles (exit 0) and the emitted file begins:

```go
package probe

import "martianoff/gala/std"
import "time"

type DateTime int64

func (s DateTime) Hash() uint32 {
	return std.HashInt(int64(s))
}
func (s DateTime) Compare(other DateTime) int {
	return std.CompareInt(int64(s), int64(other))
}
```

The alias form fails with:

```text
error[GALA-E0048]: cannot declare a method on "DateTime": it resolves to the built-in type int64
  --> main.gala:5:6
   |
 5 | func (d DateTime) IsZero() bool = d == 0
   |      ^ a type alias is the same type as its target, so it takes no…
   |
   = hint: a type alias is the same type as its target, so it takes no methods of its own — declare `opaque type DateTime int64` for a distinct type with methods, or write the method as a plain function
```

## Expected Go shape

The type keeps the Go identity and the wire shape and names no GALA package:

```go
type DateTime int64

func (d DateTime) Time() time.Time { ... }
func (d DateTime) IsZero() bool { ... }
func (d DateTime) MarshalJSON() ([]byte, error) { ... }
func (d *DateTime) UnmarshalJSON(data []byte) error { ... }
```

## Classification

Documented answer (`references/gaps.md`): `opaque type` is the named substitute, so this is not an upstream gap.
Under this repository's runtime-free twin contract the substitute is unavailable, so the file stays handwritten.

## Compiler

- `gala version`: `GALA version 0.85.0`
- flake rev: `a888e824ff53adb653bbd48dcc83f67788eeced4`
- extraction: `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`
- date: 2026-10-05
