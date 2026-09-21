# Backlog Migration Inventory

## Authority

[`backlog-fr-inventory.md`](backlog-fr-inventory.md) is the consolidated index for the non-normative review inventory.
Its [`backlog-fr-inventory-candidates/`](backlog-fr-inventory-candidates/) directory contains one canonical candidate record per selected completed entry in [`backlog/backlog.md`](../../../backlog/backlog.md) or section from a retired temporary requirements migration source.
The records preserve source wording, provenance, and review assessments so migration decisions can be audited.
They are not product requirements records.

Accepted `FR-*` and `QR-*` records in the parent requirements hierarchy are the canonical normative requirements. `CAND-*` identifiers are temporary review and traceability identifiers only; they never become permanent requirement IDs.
A resolved candidate remains in its inventory file with its final disposition and, where applicable, permanent `FR-*` or `QR-*` ID.

## Scope And Sources

The inventory reviews selected completed source material and wording retained verbatim from retired temporary requirements migration sources, indexed in [`backlog-fr-inventory.md`](backlog-fr-inventory.md).
It does not assert that every source item is a requirement, and it does not create requirements from plans, implementation details, one-off migrations, defect reports, or unresolved product choices.

The candidate files are durable review artifacts, not disposable staging files.
The consolidated index groups candidates by product area for navigation.

A candidate file may represent one coherent checked source-item group, including its indented child text.
Its filename is the permanent temporary review identifier, for example `CAND-001.md`; the identifier never becomes a permanent requirement ID.

## Candidate Schema

### Machine-Readable Triage Metadata

Every candidate file begins with YAML front matter that normalizes its review outcome for inventory queries:

```yaml
---
id: CAND-001
verdict: proposed-requirement
requirement_type: FR
related_requirements: []
confidence: inferred
---
```

`verdict` is exactly one of:

- `proposed-requirement` for a candidate that may be promoted to a canonical
  requirement. `requirement_type` is then required and is either `FR` or `QR`.
- `covered` for source material already covered by one or more canonical
  requirements. `related_requirements` lists those permanent IDs.
- `excluded` for an implementation detail, ADR or decision, bug or
  investigation, or source material that does not assert durable requirement
  behavior.
- `needs-decision` for source material whose product intent or verification
  boundary remains unresolved.

`related_requirements` is always an array of permanent `FR-*` or `QR-*` IDs; use `[]` when none is applicable. `confidence` is exactly one of `confirmed`, `inferred`, or `needs-product-decision`.

The metadata is a normalized index of the human-readable review fields below: `candidate FR` and `candidate QR` map to `proposed-requirement`; `existing requirement` and applicable `duplicate or refinement` entries map to `covered`; `ADR or decision`, `implementation detail`, `bug or investigation`, and non-requirement duplicates map to `excluded`; and unresolved entries map to `needs-decision`.
The metadata and review fields must be updated together.

Every candidate section uses this exact field sequence:

```md
## CAND-NNN: Optional concise title

#### Source

> Original source wording, quoted verbatim.

[Source lines N-N](relative-source-file.md#anchor) (when the source is retained)

#### Candidate behavior

One observable, independently testable outcome, or an explicit statement that
no requirement behavior is asserted.

#### Applicability

Actor: ...
Location: ...
Event kind: ...
Interaction mode: ...
Viewport: ...
State: ...
Exclusions: ...

#### Classification

One controlled classification value.

#### Existing Requirements and Confidence

Relevant accepted or proposed FR/QR records and one controlled confidence value.

#### Disposition

The review outcome, including a permanent FR/QR ID when resolved into one.

#### Open Questions

Required when product scope, intent, terminology, or a verification boundary is unresolved; otherwise `None.` or omitted only where the source batch's established schema omits it.
```

Field rules:

- **Source** is a raw block quote of the selected source wording.
  Preserve it
  verbatim, including spelling, punctuation, URLs, and checked bullets.
  Do not
  add glossary links inside raw Source quotes.
  When the source file is
  retained, its source reference must link to the corresponding stable source
  range or heading anchor.
  For retired source files, the verbatim quote in the
  candidate is the durable provenance record.
- **Candidate behavior** describes only an observable, independently testable
  outcome supported by the source.
  If no durable outcome is established, say
  so rather than inventing one.
- **Applicability** records actor, location, event kind, interaction mode,
  viewport, state, and exclusions.
  Use `unspecified`, `unconfirmed`, `none`,
  or `not applicable` where source evidence does not establish a value; do not
  generalize beyond the source.
- **Existing Requirements and Confidence** identifies overlap with existing
  accepted or proposed requirements, or `None`, and states exactly one
  confidence value.
- **Disposition** records whether the candidate maps to an existing permanent
  requirement, remains a candidate, is excluded, is retained as provenance, or
  needs a later decision.
  A permanent ID is recorded here only after the
  canonical record resolves the behavior.
- **Open Questions** captures unresolved scope, product intent, terminology,
  measurement, or verification questions.
  Do not use questions to imply a
  requirement that the source does not establish.

## Controlled Values

`Classification` is exactly one of:

- `candidate FR`
- `candidate QR`
- `existing requirement`
- `ADR or decision`
- `implementation detail`
- `bug or investigation`
- `duplicate or refinement`
- `needs product decision`

`Confidence` is exactly one of:

- `confirmed`
- `inferred`
- `needs product decision`

## Terminology

Candidate behavior and review-authored prose follow the controlled terminology guide in [`../terminology/README.md`](../../terminology/README.md).
The inventory itself remains non-normative, and uncertain terms belong in Open Questions or product review rather than becoming implied definitions.
Raw Source quotes are provenance, not authored prose, and must remain free of added glossary links.

## Review And Disposition

Review candidates against the accepted requirements first.
Map behavior already covered by an accepted `FR-*` or `QR-*` record to that record and retain the candidate as regression or provenance evidence.
Promote a new candidate only after product intent, applicability, boundaries, and verifiability are clear; create the canonical atomic FR or measurable QR under the parent requirements rules, not in this inventory.
Preserve decisions, defects, implementation details, and exclusions as non-normative history when they explain why no requirement was created.

### Resolved Existing-Requirement Example

## CAND-098

#### Source

> - [x] Given on the edit availability page, when no timeslot is marked as available/if needed, then the Save button should be disabled

[Source lines 411-411](../../../backlog/backlog.md#L411-L411)

#### Candidate behavior

No new requirement behavior asserted; accepted FR-004 already requires a non-empty Availability Response before saving.

#### Applicability

Actor: availability editor; Location: edit availability page; Event kind: timed; Interaction mode: editing; Viewport: unspecified; State: no Available or If needed slot; Exclusions: add availability not explicit.

#### Classification

existing requirement

#### Existing Requirements and Confidence

Overlap: accepted FR-004 directly covers the save precondition.
Confidence: confirmed.

#### Disposition

Map to FR-004; retain the disabled-button presentation only if separately needed.

### Unresolved Needs-Product-Decision Example

## CAND-054

#### Source

> - [x] add flag to enable privacy policy

[Source lines 339-339](../../../backlog/backlog.md#L339-L339)

#### Candidate behavior

No new requirement behavior asserted; this names a configuration mechanism without an observable policy outcome.

#### Applicability

Actor: maintainer.
Location: deployment configuration.
Event kind: none.
Interaction mode: configuration.
Viewport: any.
State: privacy policy feature flag.
Exclusions: policy content.

#### Classification

needs product decision

#### Existing Requirements and Confidence

None.
Confidence: needs product decision.

#### Disposition

Do not migrate pending product and compliance scope.

#### Open Questions

When enabled, where must the privacy policy appear and who controls the flag?
