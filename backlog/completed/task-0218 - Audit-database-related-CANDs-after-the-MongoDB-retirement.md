---
id: TASK-0218
title: Audit database-related CANDs after the MongoDB retirement
status: Done
assignee: []
created_date: '2026-09-13 14:25'
updated_date: '2026-09-13 14:26'
labels: []
dependencies: []
priority: medium
type: docs
ordinal: 223000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Resolve the database-related records in the backlog requirement candidate inventory now that PostgreSQL is the only supported store.

Scope: review CAND-168, CAND-169, CAND-170, CAND-177, CAND-191, CAND-212, and CAND-213 against the PostgreSQL-only architecture. CAND-212 and CAND-213 are the only MongoDB-worded candidates and are deprecated by the store retirement. CAND-212's durable behavior is reformulated as a new canonical quality requirement for staging and production PostgreSQL access. CAND-170 becomes completed migration evidence.

Constraints: no CAND record is deleted, no controlled verdict values change, and the new quality requirement follows the quality requirement template and stays independently verifiable without its source candidate.

Acceptance criteria:
- CAND-212 records deprecation with the MongoDB retirement and points to the reformulated quality requirement.
- CAND-213 records deprecation with the MongoDB retirement and states why no reformulation applies.
- CAND-170 records the completed one-off migration as retained evidence.
- The new quality requirement exists with one ISO/IEC 25010 classification pair and all six scenario elements.
- The new quality requirement is indexed in docs/requirements/README.md.
- CAND-168, CAND-169, CAND-177, and CAND-191 remain unchanged because they are already PostgreSQL-worded or store-neutral.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CAND-212 records deprecation with the MongoDB retirement and points to the reformulated quality requirement.
- [x] #2 CAND-213 records deprecation with the MongoDB retirement and states why no reformulation applies.
- [x] #3 CAND-170 records the completed one-off migration as retained evidence.
- [x] #4 The new quality requirement exists with one ISO/IEC 25010 classification pair and all six scenario elements.
- [x] #5 The new quality requirement is indexed in docs/requirements/README.md.
- [x] #6 CAND-168, CAND-169, CAND-177, and CAND-191 remain unchanged because they are already PostgreSQL-worded or store-neutral.
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
1. Update CAND-212 disposition: deprecated with the MongoDB retirement and reformulated as QR-014.
2. Update CAND-213 disposition: deprecated with the MongoDB retirement and not reformulated.
3. Update CAND-170 disposition: completed migration evidence retained.
4. Create docs/requirements/quality/qr/QR-014.md from the quality template: authenticated staging/production PostgreSQL access, security/confidentiality, infrastructure.
5. Add the QR-014 row to the quality table in docs/requirements/README.md.
6. Format changed Markdown with npm run format:markdown and run npm run lint:markdown.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Audited the database-related candidates and resolved the two MongoDB-worded records.

Changes:
- CAND-212 disposition now deprecates the MongoDB access behavior with the 2026-09-11 retirement and links the reformulated QR-014.
- CAND-213 disposition now deprecates the development MongoDB behavior and records why it was not reformulated (development and test PostgreSQL access requires configured credentials).
- CAND-170 disposition now marks the one-off report as completed migration evidence.
- Added docs/requirements/quality/qr/QR-014.md: authenticated staging/production PostgreSQL access with distinct administrative, migration, application, and backup roles (security/confidentiality, infrastructure), with all six quality-scenario elements.
- Indexed QR-014 in the quality table in docs/requirements/README.md.
- CAND-168, CAND-169, CAND-177, and CAND-191 remain unchanged as required.

Validation:
- npm run format:markdown: no files changed (all edits already compliant).
- npx prettier --check docs/requirements/quality/qr/QR-014.md: passes.
- npm run lint:markdown: passes.
- eslint on the new QR-014: passes.
- Documentation-only change, so unit and e2e tests are exempt.
<!-- SECTION:FINAL_SUMMARY:END -->
