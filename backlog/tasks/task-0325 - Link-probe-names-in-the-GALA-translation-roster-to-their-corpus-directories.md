---
id: TASK-0325
title: Link probe names in the GALA translation roster to their corpus directories
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 17:27'
updated_date: '2026-09-23 17:29'
labels: []
dependencies: []
references:
  - docs/gala-translation.md
  - server/scripts/20260923_gala_translation_probes/run.sh
documentation:
  - >-
    backlog/tasks/task-0324 - Make every GALA-equivalent verdict in the
    translation roster deterministic.md
  - >-
    backlog/tasks/task-0322 -
    Spike-build-a-Go-to-GALA-translation-roster-for-the-server.md
priority: medium
type: docs
ordinal: 323000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

`docs/gala-translation.md` cites all 49 corpus probe names in table cells but every mention is plain inline code with no path, so a reader has to reconstruct the probe location from the `Probe corpus` section and a directory listing.

TASK-0324 added the `Probe` column and pinned every decision-rule branch, which makes the roster's probe names load-bearing; TASK-0323 will gate the corpus in CI, so the names are also becoming an operational handle.

## Outcome

Every probe name in the five roster `Probe` columns and in the `Probe corpus` table links to its directory under `server/scripts/20260923_gala_translation_probes/probes/<name>/`, and the doc states that convention.

## Scope decisions

- Link target is the probe directory, which holds both `main.gala` and `expect`; do not link `main.gala` directly.
- Probe mentions in the workarounds prose and in `When to revisit` stay plain inline code; only table cells gain links, keeping the prose readable.
- No corpus, runner, or server source changes.

## Constraints

- Root Markdown follows the sentences-per-line and table-row rules in `docs/AGENTS.md`; table rows stay on one physical line.
- All 49 cited probe names already resolve to committed directories; no new probes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every probe name in any table row (the version matrix `Probe` column, the five roster `Probe` columns, and the four `Probe corpus` rows) links to `../server/scripts/20260923_gala_translation_probes/probes/<name>/`.
- [x] #2 The `Construct roster` Probe-column note and the `Probe corpus` section state the link convention and target directory.
- [x] #3 Every linked path resolves to an existing probe directory, verified by a scripted check over the changed document.
- [x] #4 Probe mentions outside tables remain inline code with no links.
- [x] #5 `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [x] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Plan (2026-09-23)

1. Link every backticked probe name inside table rows to `../server/scripts/20260923_gala_translation_probes/probes/<name>/`: the version matrix, the five roster `Probe` columns, and the four `Probe corpus` table rows.
2. Leave probe mentions in `Workarounds and contested verdicts`, the intro bullets, and `When to revisit` as plain inline code.
3. Add a sentence to the `Construct roster` Probe-column note and the `Probe corpus` section stating the link convention and target directory.
4. Verify: extract every probe link from the changed document, assert each maps to an existing directory, assert no backticked probe name remains unlinked in any table row, then run `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Evidence (2026-09-23)

- A one-off transform linked 115 probe mentions inside table rows: the version matrix `Probe` column (7), the five roster `Probe` columns, and the four `Probe corpus` rows (49 unique names).
- Verification script output: `linked probe names: 49 (unique)`, `probe directories: 49`, `OK: all table probe names linked, all targets resolve, prose unchanged`.
  The script asserts every table-row mention is a link, every link target is an existing `probes/<name>/` directory, name and target match, and no non-table line contains a probe link.
- Convention sentences added at `docs/gala-translation.md:121` and `docs/gala-translation.md:283`.
- `npm run format:markdown` reformatted `docs/gala-translation.md` (table dividers reflowed for the longer Probe column); `npm run format:markdown:check` and `npm run lint:markdown` are clean after formatting.
- Docs-only change: unit/e2e exempt per BACKLOG_WORKFLOW.md; no swagger, runtime code, `scripts/`, `prettier/`, or contract documents touched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## What changed

`docs/gala-translation.md`: every probe name inside a table now links to its directory under `../server/scripts/20260923_gala_translation_probes/probes/<name>/`; prose mentions stay plain inline code.

- Version matrix `Probe` column: 7 links.
- Five roster `Probe` columns: per-branch links, including multi-probe cells.
- Four `Probe corpus` rows: all 49 probe names linked.
- Added the link-convention sentence to the `Construct roster` note and the `Probe corpus` section.

## Verification

- Link/coverage script: 49 unique linked names, 49 probe directories, all targets resolve, no unlinked table mention, no probe link in prose.
- `npm run format:markdown` reformatted the table dividers; `npm run format:markdown:check` and `npm run lint:markdown` pass.
- Docs-only change: no unit/e2e requirement, no swagger, no runtime code, no `scripts/` or `prettier/` change, no contract documents affected.

## Risks

None beyond link rot if a probe directory is renamed; the new convention sentence and TASK-0323's CI gate make such a rename visible.
<!-- SECTION:FINAL_SUMMARY:END -->
