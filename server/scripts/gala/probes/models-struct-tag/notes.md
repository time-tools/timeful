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

## Re-verified (2026-10-18)

- `gala version`: `GALA version 0.90.0`
- flake rev: `014c2f23e45a2cd631b73cab38c0bdcb5e773347`
- extraction: `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`
- Result: the same parse error; the expected-token list now includes the new `default` keyword where it lists declaration starts.

## Re-verified (2026-10-21)

- `gala version`: `GALA version 0.92.0`
- flake rev: `4bd2d4e89cb8699e94b1f209356dd679db8d54ac`
- extraction: `0.92.0 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same parse error; the expected-token list still names the `default` keyword where it lists declaration starts.

## Re-verified (2026-10-24)

- `gala version`: `GALA version 0.92.1`
- flake rev: `911cc05932486626e684931aac9f1707939a91e6`
- extraction: `0.92.1 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same parse error; the expected-token list still names the `default` keyword where it lists declaration starts.
