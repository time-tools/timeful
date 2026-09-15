---
id: TASK-0202
title: >-
  Prune the deployed-MongoDB decommission runbook section and refresh the stale
  MongoDB test fixture
status: Done
assignee: []
created_date: '2026-09-11 15:55'
updated_date: '2026-09-11 15:56'
labels: []
dependencies: []
references:
  - docs/postgres-operations.md
  - prettier/markdown/sentences-per-line.test.mjs
  - docs/requirements/quality/qr/QR-011.md
  - DEPLOYMENT.md
  - docs/environments.md
modified_files:
  - docs/postgres-operations.md
  - prettier/markdown/sentences-per-line.test.mjs
priority: low
type: chore
ordinal: 241000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The staging and production MongoDB services have both been decommissioned, so the live PostgreSQL operations runbook no longer needs to retain the legacy decommission procedure. Also, the sentences-per-line formatter fixture still embeds an obsolete QR-011 response-measure sentence that names MongoDB, which the current QR-011 record no longer contains.

Outcome:
- docs/postgres-operations.md scopes the runbook to backup/restore and retention cleanup only, with no MongoDB wording.
- The prettier/markdown/sentences-per-line.test.mjs QR-011 fixture matches the current QR-011 response measure and still verifies that multi-sentence table rows are not split.

Constraints:
- Keep history: do not edit ADRs, deprecated requirements or their index rows, CAND-212/CAND-213 verbatim source quotes, applied server/migrations comments, or backlog archive records.
- No runtime, configuration, or generated-artifact changes; test fixture update must not change formatter behavior.
- Do not commit unless explicitly asked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/postgres-operations.md contains no MongoDB mention and documents only backup/restore and retention cleanup.
- [x] #2 prettier/markdown/sentences-per-line.test.mjs contains no obsolete MongoDB sentence and its QR-011 fixture matches docs/requirements/quality/qr/QR-011.md, with the table-row assertions still passing.
- [x] #3 npm run test:markdown-format, npm run fmt:check, npm run format:markdown:check, and npm run lint:markdown pass from the repo root.
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
1. Rewrite docs/postgres-operations.md intro to scope only backup/restore and retention cleanup, and delete the Deployed MongoDB Decommission section; confirm no inbound anchor links reference it.
2. Update prettier/markdown/sentences-per-line.test.mjs: drop the obsolete MongoDB sentence from the QR-011 response-measure fixture row and retarget the integrity assertion to the row terminator.
3. Verify: rg -i mongo over the two files, npm run test:markdown-format, npm run fmt:check, npm run format:markdown:check, npm run lint:markdown, graphify update .
4. Record evidence, finalize per BACKLOG_WORKFLOW.md.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation (2026-09-11)

- `docs/postgres-operations.md`: rewrote the intro to "This runbook documents the ongoing operational procedures: backup and restore, and retention cleanup." and deleted the `## Deployed MongoDB Decommission` section; the file now ends with the Retention Cleanup section. Checked inbound links first: `DEPLOYMENT.md:155` and `docs/environments.md:328` link the document, not the removed heading.
- `prettier/markdown/sentences-per-line.test.mjs`: removed the obsolete MongoDB sentence from the QR-011 response-measure fixture row so it matches `docs/requirements/quality/qr/QR-011.md:40`, and replaced the old substring assertion with `expect(once).toMatch(/including source revocation\. +\|/)` so the row-integrity check still verifies the cell and row terminator stay on one physical line.
- First assertion attempt used the literal `"including source revocation. |"` and failed because the formatter pads table cells with spaces before the trailing pipe; the regex accounts for the padding without allowing a newline.

## Verification

- `rg -i mongo docs/postgres-operations.md prettier/markdown/sentences-per-line.test.mjs` exits 1 (no matches).
- `npm run test:markdown-format`: 30 passed.
- `npm run fmt:check`: all matched files formatted.
- `npm run format:markdown` and `npm run format:markdown:check`: pass.
- `npm run lint:markdown`: pass.
- `graphify update .`: rebuilt graph.json, graph.html, and GRAPH_REPORT.md.
- Historical records untouched: ADRs, FR-122/QR-012 and their index rows, CAND-212/CAND-213, applied migrations, backlog archive. No commit.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Pruned the deployed-MongoDB decommission guidance from the live PostgreSQL operations runbook and refreshed the stale MongoDB sentence in the sentences-per-line formatter fixture. No runtime or configuration changes; historical records were left untouched.

Delivered:
- docs/postgres-operations.md now scopes itself to backup/restore and retention cleanup and ends after the Retention Cleanup section; the 12-line Deployed MongoDB Decommission section is removed. No inbound anchor links referenced it (DEPLOYMENT.md and docs/environments.md link the document only).
- prettier/markdown/sentences-per-line.test.mjs no longer carries the obsolete "MongoDB transfer behavior is outside this requirement's scope." sentence; its QR-011 fixture now matches docs/requirements/quality/qr/QR-011.md, and the second integrity assertion targets the response-measure row terminator with /including source revocation\. +\|/.

Evidence:
- `rg -i mongo docs/postgres-operations.md prettier/markdown/sentences-per-line.test.mjs` returns no matches.
- `npm run test:markdown-format` passes (30/30 tests).
- `npm run fmt:check`, `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown` pass.
- `graphify update .` rebuilt the graph (5495 nodes, 9713 edges, 406 communities).
- ADRs, deprecated requirements and index rows, CAND-212/CAND-213, applied server/migrations comments, and backlog history are unchanged. No commit was made.
<!-- SECTION:FINAL_SUMMARY:END -->
