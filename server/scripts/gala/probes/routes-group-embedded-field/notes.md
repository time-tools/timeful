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
