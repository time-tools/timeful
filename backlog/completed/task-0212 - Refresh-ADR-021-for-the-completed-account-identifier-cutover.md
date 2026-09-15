---
id: TASK-0212
title: Refresh ADR-021 for the completed account-identifier cutover
status: Done
assignee:
  - opencode
created_date: '2026-09-12 21:25'
updated_date: '2026-09-12 21:26'
labels: []
dependencies: []
modified_files:
  - docs/design/architecture/adr/ADR-021.md
priority: medium
type: docs
ordinal: 213000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
ADR-021 is accepted but still reads as a pre-cutover plan. Its Context says the account identity "currently" has two representations, its Decision Outcome prescribes a guarded backfill-and-drop migration, and its Consequences present one-time cutover costs as live risks. TASK-0210 deleted the cutover migration chain when it merged the consolidation into the single baseline schema, so those execution statements no longer match repository policy. The decision itself is unchanged and remains correct: platform_identities.id is the sole account identifier.

Scope: revise docs/design/architecture/adr/ADR-021.md in place (no supersession) so its context is explicitly historical and its execution statements describe the completed cutover and the baseline recreate policy. Keep one historical sentence explaining the MongoDB-era, ObjectID-shaped origin of the retired 24-character hexadecimal identifier; keep current-state prose store-neutral.

Constraints:
- Keep id, title, components, status accepted, created_date, the satisfies FR-121 traceability, the supersession of ADR-019, and the ADR-018 link.
- Do not change the decision and do not touch other ADR records, the ADR index, .graphifyignore, or linked documents.
- Documentation-only change per BACKLOG_WORKFLOW.md; unit and e2e tests are exempt.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Context describes the two-representation state and the MongoDB-era, ObjectID-shaped rationale in past tense as history, not as a current condition
- [x] #2 Decision Outcome keeps the sole native UUIDv7 identifier, canonical lowercase hyphenated wire form, all-zero sentinel, native uuid account references, tombstone without foreign key, and no-compatibility-lookup statements, and replaces the backfill-and-drop prescription with the completed cutover and the single-baseline recreate policy
- [x] #3 Consequences separate durable benefits from completed one-time cutover costs, with no pre-cutover work implied as pending
- [x] #4 Frontmatter keeps created_date, status, and traceability; updated_date is 2026-09-13; no other tracked file changes
- [x] #5 Root npm run fmt:check, npm run format:markdown:check, npm run lint:markdown, and npm run test:markdown-format pass, and graphify update . refreshes the graph
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm the current-state contract in `server/docs/postgres-data-boundaries.md` and the single-baseline policy from TASK-0210.
2. Revise `docs/design/architecture/adr/ADR-021.md` in place: historical Context, durable Decision Outcome plus completed-cutover and baseline-recreate sentences, past-tense one-time consequences, `updated_date` 2026-09-13.
3. Keep the decision, status, traceability, supersession of ADR-019, and ADR-018 link unchanged; touch no other ADR, index, or linked document.
4. Verify with root `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format`, `npm run fmt:check`, and `graphify update .`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-13)

Documentation-only change limited to `docs/design/architecture/adr/ADR-021.md`; no supersession and no other ADR, index, or linked-document edits.

- Frontmatter: `updated_date` 2026-09-12 to 2026-09-13; `id`, `title`, `components`, `status: accepted`, `created_date`, and `addresses.satisfies: FR-121` unchanged.
- Context: reframed as history ("At the time of this decision...") with past-tense verbs, kept the legacy column inventory, compressed the origin into one MongoDB-era/ObjectID-shaped rationale sentence, and moved the hard-cutover allowance to past tense.
- Decision Outcome: kept every durable statement (sole native UUIDv7 identifier, canonical lowercase hyphenated RFC 9562 wire form, all-zero sentinel, native uuid references, tombstone without foreign key, no compatibility lookup); replaced the present-tense guarded backfill-and-drop prescription with the executed cutover and the TASK-0210 single-baseline recreate policy; removed the forced re-sign-in clause from the no-compatibility-lookup sentence; changed the sentinel contrast to "retired legacy 24-character hexadecimal sentinel" so the only MongoDB/ObjectID wording is the Context rationale.
- Consequences: the three durable benefits are unchanged; the three "Bad" items now record completed one-time costs in past tense.

Evidence:
- AC1: Context lines 21-26 use past tense and no longer say "currently"; `grep -in "currently"` and the MongoDB/ObjectID check return one historical Context sentence.
- AC2: Decision Outcome retains the identifier, wire-form, sentinel, native-uuid, tombstone, and no-compatibility-lookup statements; the backfill-and-drop sentence is past tense and the baseline recreate policy sentence follows it.
- AC3: Consequences separate three durable "Good" items from three completed one-time "Bad" items; no "must complete"/"must be rewritten"/pending cutover wording remains.
- AC4: `git diff` shows only `docs/design/architecture/adr/ADR-021.md`; `updated_date: "2026-09-13"`, `status: accepted`, `created_date: "2026-09-12"`, and `satisfies: FR-121` are intact.
- AC5: `npm run format:markdown:check`, `npm run lint:markdown`, `npm run test:markdown-format` (30 tests), and `npm run fmt:check` all pass; `graphify update .` reports no code-graph topology changes (expected for a docs-only edit; semantic re-index is deferred to the `/graphify --update` flow per TASK-0209).

Unit and e2e tests are exempt: documentation-only change per BACKLOG_WORKFLOW.md and the task Definition of Done. `backlog/backlog.md` and task-0178 were already dirty before this task.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Refreshed ADR-021 in place for the completed account-identifier cutover without superseding it: the decision is unchanged, so the record now reads as history plus the steady-state contract instead of a pending migration plan.

Changes (docs/design/architecture/adr/ADR-021.md only):
- Context: "At the time of this decision..." past-tense framing, legacy column inventory kept as history, one sentence recording the MongoDB-era, ObjectID-shaped origin of the retired 24-character hexadecimal identifier, hard-cutover allowance moved to past tense.
- Decision Outcome: all durable statements retained (sole native UUIDv7 identifier, canonical lowercase hyphenated RFC 9562 wire form, all-zero sentinel, native uuid account references, tombstone without foreign key, no compatibility lookup); the present-tense guarded backfill-and-drop prescription replaced by the executed cutover and the single-baseline recreate policy from TASK-0210; sentinel contrast reworded to "retired legacy 24-character hexadecimal sentinel"; forced re-sign-in removed from the no-compatibility-lookup sentence.
- Consequences: three durable benefits unchanged; three "Bad" items now record completed one-time costs in past tense.
- Frontmatter: updated_date 2026-09-13; status, created_date, FR-121 traceability, the ADR-019 supersession, and the ADR-018 link unchanged.

Verification: root npm run format:markdown:check, npm run lint:markdown, npm run test:markdown-format (30 tests), and npm run fmt:check all pass; git diff shows only ADR-021.md; graphify update . reports no code-graph topology changes (docs-only; semantic re-index deferred to /graphify --update). Unit and e2e tests exempt per BACKLOG_WORKFLOW.md. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
