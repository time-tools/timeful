---
id: CAND-170
verdict: excluded
related_requirements: []
confidence: confirmed
---

# CAND-170: Postgres Migration Link Report

### Source

> - [x] Migrate to postgres and report new links:
>   - <http://127.0.0.1:4173/e/BFfB4> -> <http://127.0.0.1:4173/e/M7FVZFYP>
>   - <http://127.0.0.1:4173/e/6df78> -> <http://127.0.0.1:4173/e/C9ZC3WZS>

[Source lines 538-540](../../../../backlog/backlog.md#L538-L540)

### Candidate behavior

No new requirement behavior asserted; this is a one-off migration verification report with concrete local links.

### Applicability

- Actor: migration operator
- Location: migration execution
- Event kind: unspecified
- Interaction mode: not applicable
- Viewport: not applicable
- State: one-off completed migration
- Exclusions: ongoing event-link behavior

### Classification

implementation detail

### Existing Requirements and Confidence

Existing requirements: None identified.

Confidence: confirmed

### Disposition

Retained as completed migration evidence; the one-off report records links migrated during the MongoDB-to-PostgreSQL transition, which completed on 2026-09-11, and does not create an FR or QR.
