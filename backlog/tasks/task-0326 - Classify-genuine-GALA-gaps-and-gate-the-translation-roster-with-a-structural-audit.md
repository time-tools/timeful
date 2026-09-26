---
id: TASK-0326
title: >-
  Classify genuine GALA gaps and gate the translation roster with a structural
  audit
status: Done
assignee:
  - '@opencode'
created_date: '2026-09-23 17:48'
updated_date: '2026-09-23 18:01'
labels: []
dependencies: []
references:
  - docs/gala-translation.md
  - server/GALA.md
  - server/scripts/20260923_gala_translation_probes/run.sh
  - server/scripts/20260923_gala_translation_probes/inventory/main.go
  - server/scripts/20260923_gala_translation_probes/audit.sh
documentation:
  - >-
    backlog/tasks/task-0322 -
    Spike-build-a-Go-to-GALA-translation-roster-for-the-server.md
  - >-
    backlog/tasks/task-0323 -
    Add-a-version-gated-check-for-the-GALA-translation-probe-corpus.md
  - >-
    backlog/tasks/task-0324 -
    Make-every-GALA-equivalent-verdict-in-the-translation-roster-deterministic.md
modified_files:
  - docs/gala-translation.md
  - server/GALA.md
  - server/scripts/20260923_gala_translation_probes/audit.sh
  - server/scripts/20260923_gala_translation_probes/inventory/main.go
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_anonymous_struct
  - >-
    server/scripts/20260923_gala_translation_probes/probes/blocked_embedded_field
  - server/scripts/20260923_gala_translation_probes/probes/blocked_select
  - server/scripts/20260923_gala_translation_probes/probes/blocked_recover
  - server/scripts/20260923_gala_translation_probes/probes/pass_type_match
priority: medium
type: task
ordinal: 324000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Outcome

Make genuine GALA gaps a defined, classified, and audited class in `docs/gala-translation.md` instead of an implicit subset of the `Keep handwritten` verdicts, and give the probe corpus a read-only consistency gate plus a committed inventory tool with a `--check` drift mode.

Today the roster's `Verdict` column mixes expressibility, file strategy, and runtime style, the "Upstream gaps, prioritized" table mixes workaroundable constructs with true gaps (for example `switch`, `const`, `interface{}`, map/slice literals), several gap claims have no probe (`anonymous-struct-types`, `embedded-struct-fields`, `select-statements`, `recover-calls`), priority rests on counts produced by a throwaway `/tmp` instrumenter, and gap facts are restated by hand in `server/GALA.md`.

## Decisions already taken

- Taxonomy is two genuine-gap kinds: `language gap` (no GALA syntax or semantics) and `boundary gap` (expressible, but no Go-shaped drop-in); the non-gap classes are `analog`, `workaround`, and `none` (direct form).
- A genuine-gap claim requires at least one failing, non-contested probe.
- Delivery is documentation plus a read-only `audit.sh`; TASK-0323 stays the version-gated corpus run.
- Occurrence counts come from a committed stdlib-only inventory tool with `--check`, walking the server Go module and skipping nested modules (`third_party/` and the probe corpus).

## Constraints

- Keep the work local: no upstream issue or PR is produced.
- Do not change the vendored runtime, the committed twins, or server runtime code.
- Corpus commits stay limited to `.gala` sources, `expect`, `expected.out`, `go.mod`, `.gitignore`, `run.sh`, and the new tooling; generated artifacts stay gitignored.
- Root Markdown follows the sentences-per-line and table-row rules in `docs/AGENTS.md`.
- GALA 0.81.0 and the Go 1.26 series stay the pinned toolchains.

## Evidence required

- Corpus run output, `audit.sh` failure-then-pass output for a seeded inconsistency, `inventory --check` failure-then-pass output for seeded drift, and the formatting checks, all recorded in the task notes.

## Acceptance criteria

1. `docs/gala-translation.md` defines the gap classes and the genuine-gap evidence rule in a dedicated section, and every roster table carries a `Gap class` cell consistent with its rule and verdict.
2. The rebuilt gaps table lists only `language`/`boundary` rows, each with a stable `GAP-n` ID, kind, construct families, occurrence counts, blocking probe(s), partial workaround, upstream reference, status, and revisit trigger; workaroundable constructs move to a separate replacement table.
3. Unproven gap claims gain blocking probes (`anonymous-struct-types`, `embedded-struct-fields`, `select-statements`, `recover-calls`, and the empty `struct{}` type expression when promoted to a roster family), and the corpus runner reports all probes passing.
4. `audit.sh` validates probe directory integrity, orphan probes, doc probe links, gap-row blockers, class-to-GAP consistency, and `GAP-n` cross-references in `server/GALA.md`, and is demonstrated failing on a seeded inconsistency before passing again.
5. `inventory/main.go` is committed with a `--check` mode that compares the doc's inventory counts, class totals, and representatives against a fresh walk, and it is demonstrated failing on seeded drift before passing again.
6. `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` pass; `go build ./...` in `server/` is unaffected; and the code knowledge graph is re-indexed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/gala-translation.md` defines the gap classes and the genuine-gap evidence rule in a dedicated section, and every roster table carries a `Gap class` cell consistent with its rule and verdict.
- [x] #2 The rebuilt gaps table lists only `language`/`boundary` rows, each with a stable `GAP-n` ID, kind, construct families, occurrence counts, blocking probe(s), partial workaround, upstream reference, status, and revisit trigger; workaroundable constructs move to a separate replacement table.
- [x] #3 Unproven gap claims gain blocking probes (`anonymous-struct-types`, `embedded-struct-fields`, `select-statements`, `recover-calls`, and the empty `struct{}` type expression when promoted to a roster family), and the corpus runner reports all probes passing.
- [x] #4 `audit.sh` validates probe directory integrity, orphan probes, doc probe links, gap-row blockers, class-to-GAP consistency, and `GAP-n` cross-references in `server/GALA.md`, and is demonstrated failing on a seeded inconsistency before passing again.
- [x] #5 `inventory/main.go` is committed with a `--check` mode that compares the doc's inventory counts, class totals, and representatives against a fresh walk, and it is demonstrated failing on seeded drift before passing again.
- [x] #6 `npm run format:markdown`, `npm run format:markdown:check`, `npm run lint:markdown`, and root `npm run fmt:check` pass; `go build ./...` in `server/` is unaffected; and the code knowledge graph is re-indexed.
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
## Implementation plan (2026-09-23)

Scope: `docs/gala-translation.md`, `server/GALA.md`, and `server/scripts/20260923_gala_translation_probes/` (new probes, `audit.sh`, `inventory/main.go`). No server runtime changes.

### Phase 1 — Probe evidence

1. Add blocking probes for the unproven gap claims: `blocked_anonymous_struct`, `blocked_embedded_field`, `blocked_select`, `blocked_recover`, `blocked_empty_struct_type`, and `blocked_const_array_length` if the const compile-time residue is claimed.
2. Test whether a `match` on an `any` value can replace a Go type assertion; if it can, `type-assertions` is an analog, otherwise a genuine gap.
3. Run the corpus and record diagnostics.

### Phase 2 — Classification and gap ledger in `docs/gala-translation.md`

1. Add a `Gap classification` section defining `none`/`analog`/`workaround`/`language`/`boundary` and the genuine-gap evidence rule.
2. Add a `Gap class` column to the five roster tables and assign classes per the rule-first decision in the plan.
3. Rebuild the upstream-gap table as `GAP-1..n` rows (kind, families, occurrences, blocking probe, partial workaround, upstream reference, status, revisit trigger) and add a separate replacement table for analog/workaround constructs.
4. Refresh the inventory with the committed tool and update counts, representatives, and probe-corpus lists.

### Phase 3 — Tooling

1. Commit `inventory/main.go` (stdlib only, nested module) with `--server`, `--doc`, and `--check`; the walk skips nested modules.
2. Add `audit.sh` validating probe integrity, orphan probes, doc links, gap-row blockers, class-to-GAP consistency, and `server/GALA.md` `GAP-n` references.

### Phase 4 — `server/GALA.md` and verification

1. Cite `GAP-n` IDs in "What does not work" and "When to revisit" instead of restating the list.
2. Evidence: corpus run, audit fail-then-pass on a seeded inconsistency, inventory `--check` fail-then-pass on seeded drift, markdown checks, root `fmt:check`, `go build ./...`, re-index.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation notes (2026-09-23)

### Classification and gap ledger
- Added a `Gap classification` section defining `—`/`analog`/`workaround`/`language gap`/`boundary gap` and the genuine-gap evidence rule (blocking, non-contested probe required).
- Added a `Gap class` column to all five roster tables (50 family rows).
- Replaced the old priority table with `Genuine gaps`: 10 `GAP-n` rows ordered by occurrence count (GAP-1 struct-tags 387, GAP-2 multi-return-signature 181, GAP-3 anonymous-struct-types 130, GAP-4 defined-non-struct-types + methods-on-defined-non-struct 32, GAP-5 empty-struct-type 24, GAP-6 channel-types 20, GAP-7 embedded-struct-fields 11, GAP-8 fixed-size-array-types 8, GAP-9 select-statements 7, GAP-10 recover-calls 4), each with kind, families, occurrences, blocking probe, partial workaround, upstream reference, status, and revisit trigger.
- Added `Replaced by analog or workaround` for the 22 non-gap non-direct rows, moving `switch`, `defer`, `const`, `interface{}`, literals, and `make` out of the gap list.
- Promoted the empty `struct{}` type expression to an inventory family (`empty-struct-type`, 19/5/0/0/24, representative `models/set.go:3`) and a roster row with the existing `blocked_struct_type_expression` probe.
- Corrected `type-assertions`: a type-pattern `match` lowers to `std.As[T]` and replaces comma-ok, so the row is `analog` with a new `pass_type_match` probe, not a handwritten blocker.

### Probes
- New: `blocked_anonymous_struct` (`extraneous input 'struct' expecting`), `blocked_embedded_field` (GALA-E0034 `parameter "Base" has no declared type`), `blocked_select` (`no viable alternative at input '{case<-ch:'`), `blocked_recover` (GALA-E0035), `pass_type_match` (RUN=yes, `x`/`other`).
- Corpus is 54 probes; final run `54 passed, 0 failed`.

### Tooling
- `inventory/main.go` (committed, stdlib-only, nested module) walks the server Go module skipping nested modules, reproduces the Inventory table, and `-check` compares the summary, per-family class counts, totals, and representatives against the doc.
- `audit.sh` (read-only, no toolchain) checks probe integrity, orphan probes and unresolved links, gap-row blockers, class-to-`GAP-n` consistency, and `GAP-n` cross-references with `server/GALA.md`.
- The audit caught a delimiter-collapse bug during seed testing (empty probe cell read as `0`), fixed with a unit-separator delimiter.

### Evidence
- Corpus: `server/scripts/20260923_gala_translation_probes/run.sh` -> `54 passed, 0 failed`.
- Audit seeded drift (GAP-7 probe swapped to `pass_any_type`): `FAIL GAP-7: no blocking probe named`, exit 1; reverted: `audit passed`, exit 0.
- Inventory seeded drift (`struct-tags` total 387 -> 386): `FAIL struct-tags total: doc says 386, computed 387`, exit 1; reverted: `inventory check: doc matches (166 files)`, exit 0.
- `npm run format:markdown`, `format:markdown:check`, `lint:markdown`, root `npm run fmt:check`: pass.
- `go build ./...` in `server/`: clean; `go list ./...` returns 29 packages with the nested probe module excluded.
- `codebase-memory-mcp cli index_repository --repo-path .`: indexed (11194 nodes, 45065 edges).
- No commit created; changes are in the worktree.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Made genuine GALA gaps a defined, classified, and audited class instead of an implicit subset of the roster's handwritten verdicts.

## What changed

`docs/gala-translation.md` now has a `Gap classification` section defining `—`, `analog`, `workaround`, `language gap`, and `boundary gap`, plus the evidence rule that a genuine gap needs a failing, non-contested probe. All five roster tables carry a `Gap class` column. The old mixed priority table is replaced by `Genuine gaps`: 10 stable `GAP-1`..`GAP-10` rows ordered by occurrence count with kind, construct families, occurrences, blocking probe, partial workaround, upstream reference, status, and revisit trigger, and by a `Replaced by analog or workaround` table for the 22 analog/workaround rows, so `switch`, `defer`, `const`, `interface{}`, literals, and `make` no longer masquerade as gaps. The empty `struct{}` type expression was promoted to an inventory family (`empty-struct-type`, 24) and a roster row. `type-assertions` was corrected to an analog after `pass_type_match` proved that a type-pattern `match` lowers to `std.As[T]`.

`server/GALA.md` cites `GAP-1`..`GAP-10` in "What does not work" and "When to revisit" instead of restating the gap list.

## Tooling and probes

- Four new blocking probes (`blocked_anonymous_struct`, `blocked_embedded_field`, `blocked_select`, `blocked_recover`) and one pass probe (`pass_type_match`); the corpus is 54 probes.
- `inventory/main.go` (committed, stdlib-only) re-derives the Inventory table and `-check` compares summary totals, per-family class counts, totals, and representatives, exiting non-zero on drift.
- `audit.sh` (read-only, no toolchain) checks probe integrity, orphan probes, unresolved roster links, gap-row blockers, class-to-`GAP-n` consistency, and `GAP-n` cross-references with `server/GALA.md`.

## Verification

- `run.sh`: 54 passed, 0 failed.
- Audit seeded drift (`GAP-7` probe swapped to a passing probe) failed with `FAIL GAP-7: no blocking probe named` and passed after revert; the seed exposed and fixed a delimiter-collapse bug in the script.
- Inventory seeded drift (`struct-tags` 387 to 386) failed with `FAIL struct-tags total` and passed after revert.
- `format:markdown`, `format:markdown:check`, `lint:markdown`, root `fmt:check`: pass; `go build ./...` clean with the nested module excluded; code graph re-indexed (11194 nodes, 45065 edges).

## Risks and follow-ups

Gap kinds are judgment calls where a partial substitute exists (`empty-struct-type`, `anonymous-struct-types`), recorded in the partial-workaround column so later evidence can reclassify them. TASK-0323 remains the version-gated corpus run; the new toolchain-free checks can be wired into that same CI path. No commit was created.
<!-- SECTION:FINAL_SUMMARY:END -->
