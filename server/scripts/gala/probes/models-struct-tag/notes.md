# models-struct-tag

`models/location.go`, `models/event.go`, and the nested model structs carry JSON tags that define their wire shape.
GALA has no tag syntax; the fields must be renamed through `Codec[T]`, which changes the wire shape.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-models-struct-tag.go
```

## Observed

```text
error: extraneous input '`json:"country_code"`' expecting ')'
  --> main.gala:5:40
   |
 5 | struct Location(var CountryCode string `json:"country_code"`)
   |                                        ^^^^^^^^^^^^^^^^^^^^^
   |
```

## Expected Go shape

```go
type Location struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	City        string  `json:"city"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	State       string  `json:"state"`
}
```

Every field name in the JSON wire format changes if the struct moves into GALA.

## Classification

Boundary gap (`references/gaps.md`): GALA expresses the struct, but not with the Go-shaped wire format.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) carried this triage and closed upstream on 2026-10-04; the construct still fails on the pinned compiler.
No specific follow-up report was found.

## Compiler

- `gala version`: `GALA version 0.85.0`
- flake rev: `a888e824ff53adb653bbd48dcc83f67788eeced4`
- extraction: `0.85.0 7a42d3c6ef6edde23cbbc9d44ed6aade21fc563cce0cc9d41bceaecff070184c`
- date: 2026-10-05
