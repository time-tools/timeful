# models-codec-wire

`models/location.go`'s JSON keys are a codec's job under GALA 0.95.0: `GALA-E0085` refuses the field tags, and its hint names `Codec[T]` and a handwritten `.go` file.
This probe records the codec substitute for the runtime `encoding/json` wire shape and the reason it does not retire a swagger-visible type.
It complements [`../models-struct-tag/`](../models-struct-tag/), which records the tag rejection and the synthesized-member boundary.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile --stdlib-module go.gala.fyi/stdlib -i main.gala -o /tmp/probe-models-codec-wire.go
```

## Observed

The source declares the tagged `models.Location` shape with plain `var` fields and no tags, plus `Codec[Location](SnakeCase())` inside `MarshalJSON` and `UnmarshalJSON`.
The generated Go carries the same seven fields without tags, the exported `StructMeta_Location` type, and the synthesized `Copy`, `Equal`, and `Unapply` members:

```go
type Location struct {
	CountryCode string
	CountryName string
	City        string
	Postal      string
	Latitude    float64
	Longitude   float64
	State       string
}
```

`MarshalJSON` emits `Success(go_interop.ToBytes(value))` or `Failure(err)` from the codec's `Try[string]`, and `UnmarshalJSON` assigns the decoded value through `*location = value` and returns the codec's error on failure.
A scratch Go test marshaled the generated type and a Go struct with the original tags side by side.
The bytes were identical for a populated location, the zero value, the float boundaries `1e21`, `1e-7`, `math.MaxFloat64`, and `math.SmallestNonzeroFloat64`, HTML and unicode escaping, and both types rejected `math.NaN()` with an error.
Decoding also matched for exact keys, unknown keys, duplicate keys, missing keys, and `null`; the one difference is that `encoding/json` matches a key case-insensitively against the tag and the codec matches it exactly, so `Country_Code` populates the tagged struct but not the codec one.

`swag` reads the `json` tag directly, so the tagless declaration would document `countryCode` instead of `country_code` (and the same camelCase change for the other keys); that is why `models/location.go` and the other swagger-visible models keep their handwritten declarations.
The GALA struct also adds the exported `StructMeta_Location` and the synthesized members the `models-struct-tag` probe records.

## Classification

The construct is the `answered` struct-tag row in [`constructs.md`](../../../../.agents/skills/gala-from-go/references/constructs.md) (`GALA-E0085`).
The codec substitute is the documented path for a wire shape whose only reader is `encoding/json`; a tag a Go reflection consumer reads is the handwritten-file path, not a gap.

## Compiler

- `gala version`: `GALA version 0.95.0`
- flake rev: `173f7d778f103a05553564d841e0daba0126cb16`
- extraction: `0.95.0 7b68112a768d6d4776444c8ab561ce629c219d86748f93f0ca1f485daf7d1648`
- date: 2026-10-31
