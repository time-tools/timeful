# Architecture Decision Records

Read `../README.md` for the design-record catalog and the relationship between ADRs, requirements, specifications, and Backlog tasks.

## ADR Format

Each ADR has a permanent, zero-padded identifier and an ID-only filename:

```text
adr/ADR-001.md
```

IDs are assigned sequentially across the repository, not per component.

This format is adapted from [MADR 4](https://adr.github.io/madr/).
See its [minimal template](https://github.com/adr/madr/blob/4.0.0/template/adr-template-minimal.md) and [full template](https://github.com/adr/madr/blob/4.0.0/template/adr-template.md).

Start new records from [`adr/TEMPLATE.md`](adr/TEMPLATE.md).

## Metadata

Every ADR starts with YAML frontmatter containing:

```yaml
---
id: ADR-001
title: Frontend Boundary Models and Canonical Internal Shapes
components:
  - frontend
status: accepted
created_date: 2026-05-05
updated_date: 2026-05-05
---
```

- `id` is the permanent ADR identifier and must match the filename.
- `title` is the ADR title.
- The H1 must be `# <id>: <title>` using the frontmatter `id` and `title`.
- `components` lists the responsible components.
  Allowed values are `frontend`,
  `backend`, and `infrastructure`.
- `status` is one of `proposed`, `accepted`, `deprecated`, `superseded`, or
  `rejected`.
  A superseded ADR also declares `superseded_by: ADR-###`.
- Inactive statuses are `deprecated`, `superseded`, and `rejected`.
  An inactive ADR carries a status banner immediately after the H1, appears in the inactive table in [`../README.md`](../README.md), and stays out of the repository knowledge graph through `adr/.graphifyignore`.
- `created_date` is the immutable date the decision record was created.
- `updated_date` is the date of its most recent material revision.
  Do not
  update it for formatting or link-only changes.

Dates use ISO 8601 calendar-date form: `YYYY-MM-DD`.

### Addressed Requirements

An ADR may address functional requirements (FRs) and quality requirements (QRs).
Record those relationships in optional `addresses` metadata:

```yaml
addresses:
  satisfies:
    - FR-018
  enables:
    - FR-063
  constrains:
    - QR-002
    - QR-003
```

- `satisfies` identifies requirements directly fulfilled by the decision.
- `enables` identifies requirements that the decision makes feasible or
  sustainable without directly fulfilling them.
- `constrains` identifies requirements for which the decision limits valid
  implementations.

Omit empty relation categories and omit `addresses` when an ADR addresses no requirements.
A requirement may occur in more than one category when each relationship is materially true.
Keep incidental contextual references in the body rather than adding a generic traceability relation.

FRs and QRs remain self-contained specifications and must not cite ADRs as normative dependencies.
Requirement provenance is recorded separately from the requirement statement.

### Untracked Quality Attributes

Addressed QRs provide the quality attributes they specify.
When an ADR has a material consequence for a quality attribute without an addressed QR, record it in optional `affected_untracked_quality_attributes` metadata:

```yaml
affected_untracked_quality_attributes:
  - characteristic: maintainability
    subcharacteristic: modularity
```

Use ISO/IEC 25010:2023 characteristic and subcharacteristic names.
Omit this field when every affected quality attribute is represented by an addressed QR.
Explain the impact in the ADR's Consequences section, not in the frontmatter.

## Sections

New ADRs use these sections:

```md
## Context

## Considered Options

## Decision Outcome

## Consequences
```

`Decision Outcome` records the selected option and why it was chosen.
`Consequences` records the resulting benefits, tradeoffs, constraints, risks, and follow-on work.
Existing ADRs are not required to reconstruct undocumented considered options or decision drivers; add those only when supported by evidence in a later revision.
