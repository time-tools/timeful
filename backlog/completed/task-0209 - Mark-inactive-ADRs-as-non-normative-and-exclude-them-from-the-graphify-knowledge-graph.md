---
id: TASK-0209
title: >-
  Mark inactive ADRs as non-normative and exclude them from the graphify
  knowledge graph
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-12 16:15'
updated_date: '2026-09-12 16:17'
labels:
  - docs
  - architecture
  - graphify
dependencies: []
references:
  - docs/design/README.md
  - docs/design/architecture/README.md
  - docs/design/architecture/adr/
  - .graphifyignore
  - docs/graphify.md
priority: medium
type: docs
ordinal: 253000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Inactive ADRs (status `deprecated`, `superseded`, or `rejected`) are retained for decision history, but their status is currently invisible in titles and the design index, so agents reading or querying them can mistake them for active decisions. Make inactive records unmistakably non-normative in the docs and remove them from graphify results.

Confirmed approach:
- Split the ADR index in `docs/design/README.md` into separate Active and Inactive tables that show status (and the superseding ADR where applicable), and state that inactive records must not be followed.
- Give each inactive ADR a standardized status banner immediately after the H1. Keep the existing `# <id>: <title>` H1 format unchanged; do not add status text to titles.
- Document the status-sync convention in `docs/design/architecture/README.md` so future supersessions/deprecations update both the index and the graphify ignore list.
- Add `docs/design/architecture/adr/.graphifyignore` listing the inactive ADR filenames. Graphify 0.9.48 honors nested ignore files anchored at their directory.

Current inactive records: ADR-003 (deprecated), ADR-011 through ADR-017 (ADR-011, 012, 015, 016 superseded; ADR-013, 014, 017 deprecated), and ADR-019 (superseded). Active records stay unchanged apart from being listed in the Active table.

Durable reasoning: ignoring inactive ADRs is intentional even though active ADRs 018, 020, and 021 mention them in supersession prose; that history lives in the docs, not in graph queries.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/design/README.md` lists all ADRs in separate Active and Inactive tables with status shown, the superseding ADR identified for superseded records, and an explicit statement that inactive records must not be followed
- [x] #2 Every `deprecated` or `superseded` ADR carries a standardized banner immediately after its H1 naming its status, date, and (when superseded) the replacement ADR, while the H1 stays exactly `# <id>: <title>`
- [x] #3 `docs/design/architecture/README.md` states that changing an ADR to an inactive status requires updating the design index and the ADR-directory graphify ignore list
- [x] #4 `docs/design/architecture/adr/.graphifyignore` excludes all inactive ADR filenames
- [x] #5 After a graphify refresh, the graph no longer contains `source_file` nodes from inactive ADRs and active ADR files still appear
- [x] #6 Root Markdown formatting and lint checks pass (`npm run format:markdown:check`, `npm run lint:markdown`), and the ADR index statuses are consistent with ADR frontmatter
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
1. Update docs/design/README.md: add a normative rule sentence (inactive records are retained for history and must not be followed), then replace the single ADR table with an Active table (statuses proposed/accepted; columns ID, Scope, Status, Title) and an Inactive table (deprecated/superseded/rejected; columns ID, Scope, Status, Superseded by, Title). Preserve the existing unstaged heading rename and one-line-per-row formatting.
2. In each of the nine inactive ADRs (003, 011-017, 019), replace the current first status sentence after the H1 with a standardized banner: "**Inactive. Superseded on <date> by [ADR-0YY](ADR-0YY.md). Retained for decision history and must not be followed as a current decision.**" or the deprecated variant. Keep substantive follow-up sentences and keep the H1 unchanged.
3. Extend the status bullet in docs/design/architecture/README.md so an inactive status change requires updating the design index and docs/design/architecture/adr/.graphifyignore.
4. Add docs/design/architecture/adr/.graphifyignore listing the nine inactive filenames; graphify 0.9.48 honors nested ignore files anchored at the containing directory.
5. Run `graphify update .` to prune the newly excluded sources; verify no inactive ADR source_file nodes remain in graphify-out/graph.json and active ADRs remain; inspect tracked semantic cache churn.
6. Run npm run format:markdown:check and npm run lint:markdown from the repo root; reconcile any formatting and confirm index statuses match ADR frontmatter.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the approved split-index + standardized-banner approach; no H1 titles were changed. The graphify ignore list lives in a nested `docs/design/architecture/adr/.graphifyignore` so it travels with the records, and the architecture authoring guide now keeps future status changes in sync with the index and ignore list.

`graphify update .` classified the nine files as newly excluded (still on disk) rather than deleted, and the CLI pruned them correctly: 55 nodes removed, 0 inactive-ADR nodes remain, 99 active-ADR nodes remain. The tracked semantic cache was untouched.

Verified index/frontmatter/ignore consistency with a throwaway script over all 21 ADR files: 0 status mismatches, no duplicate listings, ignore set identical to the inactive set, and every inactive ADR has a banner.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Inactive ADRs are now unmistakably non-normative in the docs and absent from the graphify knowledge graph.

What changed:
- `docs/design/README.md`: replaced the single ADR table with separate Active (12 records) and Inactive (9 records) tables showing status and, for superseded records, the superseding ADR, plus explicit sentences that active records are normative and inactive records must not be followed.
- Nine inactive ADRs (003, 011-017, 019) now carry a standardized blockquote banner immediately after the H1: status, date, superseding link where applicable, and "Retained for decision history and must not be followed as a current decision." H1s remain exactly `# <id>: <title>`; substantive history sentences were preserved.
- `docs/design/architecture/README.md`: the `status` metadata guidance now defines inactive statuses and requires the banner, the inactive index table, and the `adr/.graphifyignore` entry when an ADR becomes inactive.
- Added `docs/design/architecture/adr/.graphifyignore` listing the nine inactive filenames; graphify 0.9.48 honors nested ignore files anchored at their directory.

Verification:
- `graphify update .` pruned 55 nodes from the 9 newly-ignored files (graph now 5571 nodes, 10293 links). A scripted check confirms 0 remaining nodes with inactive-ADR `source_file` values and 99 nodes from the 12 active ADR files.
- Scripted consistency check: index statuses match all 21 ADR frontmatter statuses; no record is listed twice; the ignore list equals the inactive set exactly.
- `npm run format:markdown` (only `docs/design/README.md` needed formatting), `npm run format:markdown:check`, `npm run lint:markdown`, and `git diff --check` all pass. Documentation-only change: unit and e2e tests are exempt per the Definition of Done.
- No tracked semantic-cache churn (`graphify-out/cache/semantic/` unchanged).

Notes:
- The CLI `graphify update .` re-extracts code only; it reports that changed docs still need the assistant-driven `/graphify --update` flow. The two edited READMEs will be semantically re-indexed on the next such run, but this does not affect the inactive-ADR pruning or repo-tracked artifacts.
<!-- SECTION:FINAL_SUMMARY:END -->
