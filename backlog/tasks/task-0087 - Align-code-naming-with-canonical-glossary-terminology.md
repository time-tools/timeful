---
id: TASK-0087
title: Align code naming with canonical glossary terminology
status: To Do
assignee: []
created_date: '2026-08-27 11:22'
updated_date: '2026-09-15 12:23'
labels:
  - terminology
  - refactoring
dependencies: []
documentation:
  - docs/terminology/glossary.md
  - docs/terminology/README.md
  - frontend/AGENTS.md
priority: medium
type: chore
ordinal: 95000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The canonical controlled terminology now lives in docs/terminology/glossary.md, but code was written before that vocabulary existed, so variables, functions, types, composables, components, file names, and directories may use older ad-hoc words instead of the canonical terms.

Update code-level naming across the repository so names are derived from the established glossary terminology. The full inventory of divergent names is not known yet; the first phase of work is a repo-wide comparison of code naming against the glossary, followed by targeted renames. If the discovered scope is too large for one reviewable change, split deliverable slices into follow-up tasks and ask before expanding silently.

Constraints: renames are code-identity changes only and must not alter external contracts such as API paths, JSON payload keys, persisted database fields, or browser-plugin message shapes; those spellings are governed by their own documentation. Where a glossary term's canonical prose differs only by word casing (for example `Show all hours` as a UI label), use judgment: UI-label spellings stay as labels, while type, function, and file identifiers use the language's standard casing derived from the same words.

Read docs/terminology/README.md before starting; it defines the canonicalization rules used here.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every identifier, file, directory, symbol, and type name whose concept corresponds to a controlled term in docs/terminology/glossary.md uses a spelling derived from that term's canonical form, following each language's identifier casing conventions (e.g., PascalCase for Go/Vue/TS exported types and file names, camelCase for TS variables and functions)
- [ ] #2 Undocumented ad-hoc synonyms for controlled terms in code are replaced with the glossary-derived name, or the term family is added to the glossary through its documented process when code legitimately needs a distinct term
- [ ] #3 Renames preserve observable behavior and public contracts: HTTP routes, JSON field names, persistence keys, query params, browser-plugin window.postMessage payload shapes (see PLUGIN_API_README.md), and other wire-format spellings are unchanged
- [ ] #4 Compatibility shims or aliases introduced only for transport boundaries live at explicit boundary layers per frontend AGENTS.md rules, not scattered through views, composables, or services
- [ ] #5 Repository-required checks pass for every touched area (frontend lint, typecheck, build, unit tests; server build/tests when server files changed) and any skipped check is reported
- [ ] #6 The final summary lists the terminology-driven renames applied, notable deliberate non-renames with reasons, and any glossary gaps discovered
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-15 12:23
---
Coordination (TASK-0228, 2026-09-15): TASK-0228 renamed `server/eventsource` to `server/eventid` and the `accounts_test.go` helper `accountObjectID` to `accountUUID`, because both names carried retired Mongo event-source, dispatch, and ObjectID vocabulary. Treat `eventid.Canonical` and `accountUUID` as the current names in this task's naming inventory rather than re-litigating or reverting the cleanup. The glossary has no controlled term for the public event identifier yet, so TASK-0228 did not add one; a future glossary-derived rename may supersede these names.
---
<!-- COMMENTS:END -->
