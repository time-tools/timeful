---
id: TASK-0206
title: Supersede rewritten ADRs instead of reformulating them in place
status: Done
assignee: []
created_date: '2026-09-11 20:45'
updated_date: '2026-09-11 20:46'
labels: []
dependencies: []
modified_files:
  - docs/design/architecture/adr/ADR-011.md
  - docs/design/architecture/adr/ADR-012.md
  - docs/design/architecture/adr/ADR-015.md
  - docs/design/architecture/adr/ADR-016.md
  - docs/design/architecture/adr/ADR-018.md
  - docs/design/architecture/adr/ADR-019.md
  - docs/design/architecture/adr/ADR-020.md
  - docs/design/README.md
  - server/docs/postgres-data-boundaries.md
priority: medium
type: docs
ordinal: 245000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Revise the TASK-0205 ADR cleanup so that history is preserved: restore the radically rewritten ADR bodies and let successor ADRs carry the current PostgreSQL-only decisions.

Scope:
- Restore ADR-011, ADR-012, ADR-015, and ADR-016 to their original bodies.
- Mark ADR-012 superseded by a new ADR-019, ADR-015 superseded by a new ADR-020, and ADR-016 superseded by ADR-018; keep the ADR-011 supersession chain (ADR-011 to ADR-016 to ADR-018).
- Expand ADR-018 to absorb the surviving single-authority and data-ownership decisions from ADR-011 and ADR-016, and to reference ADR-019 for account resolution.
- Add ADR-019 (account resolution through platform_identities.external_user_id) and ADR-020 (authenticated AES-256-GCM provider-credential envelope).
- Update the ADR index and the durable-decision links in server/docs/postgres-data-boundaries.md.
- Keep the TASK-0205 minimal ADR-010 cleanup and the ADR-013/ADR-014/ADR-017 deprecation edits.

Constraints:
- Superseded and deprecated records keep their historical MongoDB and legacy wording by decision.
- ADR-008/ADR-009 legacy wording refers to pre-existing frontend guest semantics and is out of scope.
- Documentation-only change per BACKLOG_WORKFLOW.md; no unit or e2e tests required.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ADR-012, ADR-015, and ADR-016 are restored to their original bodies, declare status superseded with superseded_by ADR-019, ADR-020, and ADR-018 respectively, and carry dated notes
- [x] #2 ADR-011 keeps its supersession by ADR-016 so the chain reaches ADR-018
- [x] #3 ADR-018 records the single-authoritative-store rule, the data ownership boundary, and its supersession of ADR-011 and ADR-016, and references ADR-019 and ADR-020
- [x] #4 ADR-019 and ADR-020 exist with valid MADR frontmatter, H1s matching their ids and titles, and index rows in docs/design/README.md
- [x] #5 docs/design/README.md rows match the restored and new titles, and server/docs/postgres-data-boundaries.md lists ADR-018, ADR-019, and ADR-020 as durable decisions
- [x] #6 npm run format:markdown, npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format, and npm run fmt:check pass, and graphify update . refreshes the graph
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-11)

Documentation-only change; no runtime code, tests, or configuration touched.

- Captured the in-place rewritten text from TASK-0205, then restored ADR-011/ADR-012/ADR-015/ADR-016 with `git checkout HEAD --`.
- ADR-012, ADR-015, and ADR-016: status superseded, `superseded_by` ADR-019/ADR-020/ADR-018, `updated_date: 2026-09-11`, and a dated note stating the successor restates the decision for the PostgreSQL-only system. Bodies unchanged.
- ADR-018: absorbed the single-authority principles and the data ownership boundary (calendar/credentials/sub-calendars/preferences, OTP, daily logs PostgreSQL-owned; friend requests retired), declared supersession of ADR-011 and ADR-016, referenced ADR-019 for account resolution and ADR-020 for credentials, and added the friend-request consequence.
- ADR-019 and ADR-020 created from the captured rewritten text with supersession statements.
- docs/design/README.md restored the original ADR-011/012/015/016 titles and added ADR-019/ADR-020; server/docs/postgres-data-boundaries.md now lists ADR-018, ADR-019, and ADR-020.
- TASK-0205 received a comment recording the revised approach.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Restored the radically rewritten ADR bodies and moved the current PostgreSQL-only decisions into explicit successor records, keeping supersession history intact.

Delivered:
- ADR-011, ADR-012, ADR-015, and ADR-016 are restored to their original bodies.
- ADR-012, ADR-015, and ADR-016 declare status superseded with superseded_by ADR-019, ADR-020, and ADR-018 and carry dated notes; ADR-011 keeps its supersession chain to ADR-016.
- ADR-018 records the single-authoritative-store rule, the data ownership boundary, and its supersession of ADR-011 and ADR-016, and references ADR-019 and ADR-020.
- New ADR-019 "Resolve Accounts By External User Identifier" and ADR-020 "Encrypt Provider Credentials At Rest With Authenticated AES-256-GCM" carry the current decisions with valid MADR frontmatter and H1s.
- docs/design/README.md rows match the restored and new titles; server/docs/postgres-data-boundaries.md lists ADR-018, ADR-019, and ADR-020 as durable decisions.
- The TASK-0205 ADR-010 stale-sentence cleanup and the ADR-013/ADR-014/ADR-017 deprecation edits stand.

Known residual (owner-approved): superseded and deprecated records ADR-011 through ADR-017 keep their historical MongoDB and legacy wording.

Evidence: rg shows MongoDB/legacy wording only in superseded/deprecated ADR-011 through ADR-017 and ADR-008/ADR-009; npm run format:markdown, format:markdown:check, lint:markdown, test:markdown-format (30 tests), and fmt:check pass; graphify update . rebuilt the graph (5513 nodes, 9738 edges, 413 communities). Documentation-only change, so unit and e2e tests are exempt per BACKLOG_WORKFLOW.md and the task DoD. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
