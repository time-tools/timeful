# models-defined-scalar-methods

`models/datetime.go` and `models/uuid.go` declare methods on defined scalar types (`type DateTime int64`, `type UUID string`).
GALA refuses that declaration with `GALA-E0048` and names `opaque type` as the substitute.
By default the substitute transpiles to Go that imports `martianoff/gala/std` for the synthesized `Hash` and `Compare` methods.
A same-package `.go` sibling declaring `Hash` and `Compare` suppresses both synthesized methods, so the generated Go names no GALA package and can enter a runtime-free twin; the cost is two exported methods the original Go API did not carry.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-models-defined-scalar-methods.go
```

To reproduce the suppressed run, place the sibling listing below next to `main.gala` as `suppress.go` and repeat the command; remove it to restore the default output.

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

## Suppression (runtime-free path)

A same-package `.go` sibling declaring `Hash() uint32` and `Compare(DateTime) int` suppresses the synthesized methods, verified on the pinned compiler on 2026-10-05.
The sibling used for the check:

```go
package probe

func (d DateTime) Hash() uint32 {
	return uint32(d) ^ uint32(d>>32)
}

func (d DateTime) Compare(other DateTime) int {
	if d < other {
		return -1
	}
	if d > other {
		return 1
	}
	return 0
}
```

With the sibling present, the same invocation emits:

```go
package probe

import "time"

//line main.gala:5
type DateTime int64

//line main.gala:7
func (d DateTime) Time() time.Time {
	return time.UnixMilli(int64(d))
}

//line main.gala:9
func (d DateTime) IsZero() bool {
	return d == 0
}
```

The generated Go names no GALA package, and the generated file plus the sibling build and vet clean in a scratch module (`go build ./...`, `go vet ./...`).
The same suppression holds for `opaque type UUID string`.
Upstream documents the behavior in PR #665 (the fix for #621): the synthesized methods are skipped when GALA or a same-package `.go` file already declares them.

## Expected Go shape

The type keeps the Go identity and the wire shape and names no GALA package once the sibling suppresses the synthesized methods:

```go
type DateTime int64

func (d DateTime) Time() time.Time { ... }
func (d DateTime) IsZero() bool { ... }
func (d DateTime) MarshalJSON() ([]byte, error) { ... }
func (d *DateTime) UnmarshalJSON(data []byte) error { ... }
```

## Classification

`GALA-E0048` names `opaque type`, so the construct is a documented answer under `references/gaps.md` and is listed there under "What Not To File".
The runtime-free contract does not make the substitute unavailable: the sibling suppression above keeps the generated Go free of the GALA runtime.
The cost is the two exported methods the sibling adds, which the original Go API did not carry; whether to pay that cost is a repository decision, so `models/datetime.go` and `models/uuid.go` stay handwritten until it is accepted.
No upstream comment or new report is needed, because upstream already documents the suppression in PR #665.

## Compiler

- `gala version`: `GALA version 0.85.0`
- flake rev: `a888e824ff53adb653bbd48dcc83f67788eeced4`
- extraction: `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`
- date: 2026-10-05

## Status (2026-10-05)

The repository accepted the two-added-methods cost for `models/datetime.go`, which landed as `models/datetime.gala` plus `models/datetime_extra.go` (`Hash`/`Compare` and the multi-value `MarshalJSON`/pointer-receiver `UnmarshalJSON`).
One transpile-order requirement showed up while landing it: the suppression only takes effect when the superseded handwritten file is absent at analysis time.
Transpiling `datetime.gala` while the old `datetime.go` still declared `DateTime` emitted `martianoff/gala/std`-backed `Hash`/`Compare`, while moving the old file aside first produced the runtime-free output above.
`models/uuid.go` remains handwritten pending its fixed-size `[16]byte` split, so this probe stays.
