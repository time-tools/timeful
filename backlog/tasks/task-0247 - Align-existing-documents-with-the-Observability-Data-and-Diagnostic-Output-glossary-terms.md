---
id: TASK-0247
title: Align existing documents with the observability glossary terms
status: Done
assignee:
  - OpenCode
created_date: '2026-09-16 14:20'
updated_date: '2026-09-16 15:26'
labels: []
dependencies:
  - TASK-0248
references:
  - docs/terminology/glossary.md
  - docs/terminology/README.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/requirements/quality/qr/QR-019.md
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/quality/qr/QR-004.md
  - docs/requirements/quality/qr/QR-009.md
  - docs/requirements/quality/qr/QR-010.md
  - docs/requirements/quality/qr/QR-015.md
  - docs/requirements/quality/qr/QR-016.md
  - docs/requirements/quality/qr/QR-018.md
  - docs/requirements/README.md
modified_files:
  - docs/requirements/quality/qr/QR-004.md
  - docs/requirements/quality/qr/QR-009.md
  - docs/requirements/quality/qr/QR-010.md
  - docs/requirements/quality/qr/QR-015.md
  - docs/requirements/quality/qr/QR-016.md
  - docs/requirements/quality/qr/QR-017.md
  - docs/requirements/quality/qr/QR-018.md
  - docs/requirements/quality/qr/QR-019.md
  - docs/design/architecture/adr/ADR-022.md
  - docs/requirements/README.md
priority: medium
type: docs
ordinal: 248000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-0243 added Observability Data and Diagnostic Output, and TASK-0248 expanded the Observability glossary with Signal, Structured Log Records, Service Health Status, Retention Window, Stream, and Environment-Scoped Observability Credential. Align pre-existing references to all Observability glossary terms in QR-004, QR-009, QR-010, QR-015, QR-016, QR-017, QR-018, QR-019, the ADR-022 body, and the QR-018 row in docs/requirements/README.md with the canonical term forms, first-occurrence links, and bold repeats required by docs/terminology/README.md, without changing normative meaning. Replace enumerations that only restate a term's definition, and keep clauses that add facts beyond the definitions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 QR-004, QR-009, QR-010, QR-015, QR-016, QR-017, QR-018, and QR-019 use the canonical Observability glossary term forms wherever they refer to those concepts, with first-occurrence glossary links and bold repeats per docs/terminology/README.md, and redundant signal enumerations replaced by the terms
- [x] #2 ADR-022 uses the canonical forms in its body where it refers to those concepts, and its updated_date is updated only if the revision is material beyond links per docs/design/architecture/README.md
- [x] #3 The QR-018 row in docs/requirements/README.md matches its frontmatter title after capitalization alignment
- [x] #4 No requirement changes normative meaning
- [x] #5 npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass from the repository root
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. QR-004: canonicalize the "diagnostic output" references to Diagnostic Output in Environment, Artifact, Response, and Response Measure with first-occurrence glossary links; leave generic diagnostic-event wording in Source and Stimulus unchanged.
2. QR-009: canonicalize and link Diagnostic Output in Response Measure.
3. QR-010: replace "structured log records" and "diagnostic records" with Structured Log Records and Service Health Status in Artifact, Response, and Response Measure with first-occurrence links.
4. QR-015: replace the signal enumeration in Stimulus with Observability Data and link representative data in Response Measure; link Retention Window in Response.
5. QR-016: link Environment-Scoped Observability Credential in Response and Response Measure; leave the endpoint and listener wording unchanged.
6. QR-017: replace the redundant "for logs, metrics, and traces" enumeration in Environment with canonical Observability Data.
7. QR-018: capitalize the frontmatter title to "Delete Observability Data after its Retention Window"; canonicalize and link Observability Data, Retention Window, and Stream in Stimulus, Artifact, Response, and Response Measure; update the README row to match the frontmatter title with glossary links.
8. QR-019: canonicalize and link Environment-Scoped Observability Credential and Stream throughout; keep organization generic.
9. ADR-022: canonicalize Diagnostic Output, Structured Log Records, Service Health Status, Observability Data, Signal, Stream, Environment-Scoped Observability Credential, and Retention Window where the text refers to them, with first-occurrence links and bold repeats, and keep updated_date unchanged because the revision is casing and link only.
10. Run npm run format:markdown, then format:markdown:check, lint:markdown, test:markdown-format, and fmt:check from the repository root.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: OpenCode review
created: 2026-09-16 15:26
---
Post-review corrections: replaced the shortened credential repeats with bold canonical forms in ADR-022 (Decision Outcome paragraph) and QR-019 (Response Measure); restored the plural Environment-Scoped Observability Credentials in QR-016 so its original normative plurality is preserved; reworded QR-018's Environment cell to "a finite 14-day period"; and tightened the Structured Log Records, Service Health Status, and Retention Window glossary definitions. CAND-216's dangling Show all hours link is repointed under TASK-0248. Root checks re-run and passing.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Aligned the observability requirements, ADR-022, and the requirements index with the expanded Observability glossary.

- QR-004: Environment, Artifact, Response, and Response Measure use Diagnostic Output with first-occurrence links; generic diagnostic-event wording in Source and Stimulus is unchanged.
- QR-009: Response Measure links Diagnostic Output.
- QR-010: Artifact, Response, and Response Measure use Structured Log Records and Service Health Status in place of "structured log records", "diagnostic records", and "health state".
- QR-015: Stimulus and Response Measure use Observability Data; Response links Retention Window.
- QR-016: Response and Response Measure reference Environment-Scoped Observability Credential; endpoint and listener wording is unchanged.
- QR-017: the redundant signal enumeration is dropped; Observability Data export is enabled.
- QR-018: frontmatter title capitalized to "Delete Observability Data after its Retention Window"; Source, Stimulus, Environment, Artifact, Response, and Response Measure use Observability Data, Retention Window, and Stream with links and bold repeats; the README row matches the frontmatter title with both term links.
- QR-019: Environment-Scoped Observability Credential and Stream are used with links and bold repeats, including the README row term link; organization stays generic.
- ADR-022: replaces pre-existing phrasings with Diagnostic Output, Structured Log Records, Service Health Status, Observability Data, Signal, Retention Window, Stream, and Environment-Scoped Observability Credential; definitional enumerations are removed while informational clauses remain (current absence of log search, metrics, and traces; "the first delivered Signal is server logs"; OTLP transport; local logs-and-traces inspection); updated_date stays 2026-09-16 because the revision is casing and link only.
- Verification: a script confirms every occurrence of the eight terms in the changed files is linked, bold, or exempt (heading/frontmatter); npm run format:markdown changed only README table alignment; format:markdown:check, lint:markdown, test:markdown-format (2 files, 30 tests), and fmt:check all pass from the repository root. Graphify resync was skipped because this repository keeps Graphify cache updates in separate chore(graphify) commits.
<!-- SECTION:FINAL_SUMMARY:END -->
