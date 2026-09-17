---
id: CAND-216
title: Desktop Show-All-Hours Placement
verdict: proposed-requirement
requirement_type: FR
related_requirements:
  - FR-111
confidence: confirmed
---

# CAND-216: Desktop Show-All-Hours Placement

## Source

> Show all hours sits to the right of the event description on desktop, where the other options and action buttons such as cancel, save, and delete live.

Recorded from the session triage decision round of 2026-08-26; this record has no retained backlog source, so the quoted ruling is its provenance.
It corrects CAND-198, whose mobile premise was rejected by the owner.

## Candidate behavior

On desktop, the ["Collapse disabled times" Option](../../../terminology/glossary.md#collapse-disabled-times-option) sits to the right of the event description alongside the other options and action buttons such as cancel, save, and delete.

## Applicability

Actor: event visitor.
Location: desktop event page.
Event kind: timed.
Interaction mode: viewing.
Viewport: desktop.
State: event description visible.
Exclusions: mobile layouts; supersedes the CAND-198 mobile premise.

## Classification

candidate FR

## Existing Requirements and Confidence

Existing requirements: [FR-011](../../functional/fr/FR-011.md) defines expansion behavior and [FR-048](../../functional/fr/FR-048.md) covers another placement state; neither fixes this desktop position.
Confidence: confirmed.

## Disposition

Promoted to [FR-111](../../functional/fr/FR-111.md); this record corrects CAND-198, whose mobile premise was rejected.

## Open Questions

None.
