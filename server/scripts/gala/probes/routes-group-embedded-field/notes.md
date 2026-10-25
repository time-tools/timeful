# routes-group-embedded-field

`routes/group.go`'s `eventInput` embeds `models.Event` so the request body sees the event's fields at the top level next to the group-only `attendees` key.
GALA has no embedded-field syntax; the documented substitute is a named field plus explicit delegation, which loses promotion and changes the request wire shape.

## Invocation

From this directory, with generated Go written to a scratch path:

```sh
gala transpile -i main.gala -o /tmp/probe-routes-group-embedded-field.go
```

## Observed

```text
error: mismatched input '}' expecting {'[', '*', 'map', 'func', IDENTIFIER}
  --> main.gala:11:1
   |
11 | }
   | ^
   |
```

The parser reads `Event` as a field name and then expects a type, so the bare embed spelling dies on the next token.
The `embed Event` spelling was checked on the same compiler and is not in the grammar either (`extraneous input 'embed'`).

## Expected Go shape

```go
type eventInput struct {
	models.Event
	Attendees []string `json:"attendees"`
}
```

`event_routes.go` binds request JSON into this type in three handlers, so both the promotion and the outer tag define the wire format.

## Classification

Answered row, boundary cost (`references/gaps.md`): the substitute exists (a named field plus delegation) but loses the promotion the wire format depends on, so the member stays handwritten.

## Upstream

[#528](https://github.com/martianoff/gala/issues/528) carried this triage and closed upstream on 2026-10-04; `constructs.md` lists embedded fields as an answered row and `gaps.md` says not to file it.
No specific follow-up report was found.

## Compiler

- `gala version`: `GALA version 0.87.1`
- flake rev: `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`
- extraction: `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`
- date: 2026-10-10

## Re-verified (2026-10-18)

- `gala version`: `GALA version 0.90.0`
- flake rev: `014c2f23e45a2cd631b73cab38c0bdcb5e773347`
- extraction: `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`
- Result: the same parse error for the bare embed spelling; the expected-token list now includes the new `default` keyword.

## Re-verified (2026-10-21)

- `gala version`: `GALA version 0.92.0`
- flake rev: `4bd2d4e89cb8699e94b1f209356dd679db8d54ac`
- extraction: `0.92.0 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same parse error for the bare embed spelling; the expected-token list still names the `default` keyword.

## Re-verified (2026-10-24)

- `gala version`: `GALA version 0.92.1`
- flake rev: `911cc05932486626e684931aac9f1707939a91e6`
- extraction: `0.92.1 2c6059cade6f8331c75c084d870dbf4f9af574030a1e211d539c98b456f294a3`
- Result: the same parse error for the bare embed spelling; the expected-token list still names the `default` keyword.

## Re-verified (2026-10-25)

- `gala version`: `GALA version 0.93.0`
- flake rev: `8ba02b2b18eca7ca3bc410f5e467b112366bdf6c`
- extraction: `0.93.0 6368c09381c6f9d8d72f1551101ca48207cb2e1c34122627ca928261ad05b13c`
- Result: the same parse error for the bare embed spelling; the expected-token list still names the `default` keyword.
