---
id: TASK-0341
title: Consolidate the GALA document stack into a ledger and a translate-verify loop
status: Done
assignee:
  - opencode
created_date: '2026-10-04 11:54'
updated_date: '2026-10-04 13:29'
labels:
  - gala
  - documentation
  - tooling
dependencies: []
references:
  - server/GALA.md
  - server/README.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/gaps.md
  - server/scripts/gala/verify.sh
  - server/scripts/gala/probes/type-position-import/
  - flake.nix
  - flake.lock
  - 'https://github.com/martianoff/gala/issues/648'
documentation:
  - server/third_party/gala/VENDORED_FROM
  - .agents/skills/gala-from-go/references/constructs.md
  - server/scripts/gala/probes/type-position-import/notes.md
modified_files:
  - server/GALA.md
  - server/README.md
  - server/scripts/gala/verify.sh
  - server/scripts/gala/.gitignore
  - server/scripts/gala/probes/type-position-import/main.gala
  - server/scripts/gala/probes/type-position-import/notes.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/gaps.md
  - server/third_party/gala/VENDORED_FROM
  - server/third_party/gala/go.mod
  - server/eventid/eventid.gala
  - server/eventid/eventid.go
  - server/eventid/doc.go
  - server/middleware/auth.gala
  - server/middleware/auth.go
  - server/middleware/doc.go
  - server/observability/redact.go
  - server/postgres/dailylogs.go
  - docs/gala-translation.md
  - docs/gala-rewrite-playbook.md
  - .agents/skills/gala-rewrite/SKILL.md
  - server/scripts/20260923_gala_translation_probes/
priority: medium
type: task
ordinal: 345005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Four local GALA artifacts overlap: `.agents/skills/gala-rewrite/`, `docs/gala-rewrite-playbook.md`, `docs/gala-translation.md`, and `server/GALA.md` (479 lines), while the only current, project-independent knowledge is `.agents/skills/gala-from-go/`.
The local stack duplicates the skill, carries stale inventory counts and verification hashes, and is loaded into the context of every translation agent.
The 85 MB probe corpus keeps an inventory tool, an audit script, and 80 probes alive to hold verdicts the skill already carries as version-agnostic checks.
The repository also has no loop for continuing translation: candidate selection, blocker classification, runtime provenance, and reporting are spread across the documents.

The agreed outcome is one translation reference, one ledger, one loop skill, and one verification script.
The translation target is the `gala` compiler on `PATH`, which the dev shell builds from source at the commit in `flake.lock` (`e2e28c318ff1eb606f7f607199b629d93b78ab1f` at planning time), and the vendored runtime under `server/third_party/gala/` is always re-vendored to match that compiler before translating.

Decisions taken in the 2026-10-04 planning session, all durable:
- The single ledger is `server/GALA.md`, rewritten.
- The probe corpus, `run.sh`, `audit.sh`, and the inventory tool are deleted; minimal repros are recreated per open blocker under `server/scripts/gala/probes/`.
- The loop is a new project skill at `.agents/skills/gala-loop/SKILL.md`; it loads `gala-from-go` for translation knowledge and reads the ledger for state.
- A blocker with no workaround becomes a Backlog task containing a filled `gaps.md` report; the loop does not file upstream issues itself.
- The vendored runtime is re-vendored whenever the compiler rev differs from the recorded provenance, and all committed twins are re-baselined after a re-vendor.
- This task supersedes TASK-0339 and its five subtasks, TASK-0323, TASK-0330.05, and TASK-0340; TASK-0330 is finalized Done with a comment that `.05` was superseded.

Live mismatch to resolve first: `~/.gala/stdlib/v0.84.1` was regenerated from master on 2026-10-02, while `server/third_party/gala/` was copied from the 0.84.1 release snapshot on 2026-09-29, so the first sync is expected to re-baseline all 15 twins.
The repository's own compiler history (`delta_*` records) is release-pinned and is dropped rather than carried into the version-agnostic skill.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `server/GALA.md` is the single GALA ledger: twin registry, compiler/runtime provenance and re-vendor procedure, open findings with classification/workaround/upstream link, upstream report index, and candidate cursor, without restating construct substitutions or gap classification
- [x] #2 `.agents/skills/gala-from-go/` is the only translation reference and contains no repository-relative link; the two links in `references/gaps.md` are inlined as facts
- [x] #3 `.agents/skills/gala-loop/SKILL.md` defines the provenance-and-sync step, one translation iteration, per-blocker probe and Backlog-task handling, the stop conditions, and the never list
- [x] #4 `server/scripts/gala/verify.sh` prints compiler and vendored-runtime provenance and fails on any twin that does not regenerate byte-identically, on a `gofmt -l` hit, or on `go build ./...`; a `--write` mode regenerates in place
- [x] #5 `docs/gala-translation.md`, `docs/gala-rewrite-playbook.md`, `.agents/skills/gala-rewrite/`, and `server/scripts/20260923_gala_translation_probes/` are deleted, and no tracked file references them
- [x] #6 `server/README.md`'s GALA section points at `server/GALA.md`, the loop skill, and `verify.sh`
- [x] #7 `server/third_party/gala/VENDORED_FROM` records the matching compiler version, flake rev, source path, and date, and after the sync every committed twin regenerates byte-identically and `go build ./...` passes
- [x] #8 The type-position import hole (#648, formerly TASK-0340) is re-checked on the synced compiler; a persistent hole has a minimal repro under `server/scripts/gala/probes/` and a Backlog report task, and a fixed hole is recorded as fixed
- [x] #9 Superseded tasks are archived with their durable facts carried into this task and its ledger: TASK-0339 and its five subtasks, TASK-0323, TASK-0330.05, TASK-0340; TASK-0330 is finalized Done with a supersession comment for `.05`
- [x] #10 Verification evidence is recorded: the dangling-reference sweep, `verify.sh` output, `npm run format:markdown:check`, `go build ./...`, and the codebase-memory index refresh
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
Agreed plan from the 2026-10-04 planning session.
The executing session must confirm it against the then-current tree, may revise it, and records any deviation in the task notes.

## Phase 0 - preserve facts before deletion

Read these fully and carry their durable facts into the new ledger or skill before deleting anything:
- `server/GALA.md`: the runtime-free vs runtime-enabled output styles, the vendoring provenance and update procedure, the twin registry and handwritten siblings, the open upstream issues, and the `.gala/` analysis-cache note.
- `docs/gala-translation.md`: "Genuine gaps", "Replaced by analog or workaround", and "Workarounds and contested verdicts" are already in `.agents/skills/gala-from-go/references/`; verify that before dropping them. Do not carry inventory counts, occurrence totals, verification hashes, compiler history, or probe names.
- `docs/gala-rewrite-playbook.md` and `.agents/skills/gala-rewrite/SKILL.md`: the ordered procedure is already in the skill's mechanical pass and verify sections; carry nothing else.
- `server/scripts/20260923_gala_translation_probes/`: seed the ledger cursor from the blocked candidate list, then treat the corpus as disposable.

Known blocked candidates from the old roster, to be re-checked on the current compiler: `models/{datetime,uuid,set,location,event}.go`, `errs/errors.go`, `routes/{users,respondent_identity,group}.go`, `postgres/` (the `Repository` sibling-method constraint), `main.go`, `observability/{provider,readiness,transport}.go`, `services/{auth,calendar,contacts,listmonk,microsoftgraph}`, `mockprovider`, `services/gcloud/tasks.go`, `discord_bot/commands/active_users.go`, `slackbot/commands/{active_users,utils}.go`.
`routes/users.go` stays handwritten because a `swag` annotation cannot survive the transpiler's comment-less emit.

## Phase 1 - rewrite `server/GALA.md` as the ledger

Target roughly 120-180 lines, in this order:
1. Contract: both halves of a twin are committed; never hand-edit a generated file; transpile from the package directory; `server/.gala/` is gitignored.
2. Compiler and runtime: `gala version`, `which gala`, the flake rev from `jq -r '.nodes.gala.locked.rev' flake.lock`, and the vendored provenance from `server/third_party/gala/VENDORED_FROM`; the re-vendor procedure; the rule that a rev mismatch is resolved by a sync iteration and never by translating across it.
3. Twin registry: one row per twin with source, generated, sibling, and style, plus a single command template `cd server/<pkg> && gala transpile -i <name>.gala -o <name>.go` instead of a per-row command column.
4. Open findings only: construct, classification, workaround or none, upstream link, probe path. Current open upstream items: `#528`, `#621`, `#619`, `#648`.
5. Upstream report index: number, title, state, last-checked compiler rev.
6. Cursor: the ordered candidates from Phase 0.
7. Stop conditions: point at the skill's stop-and-report list and never list.

Do not restate construct rows, gap classification, or the report template; those live in `gala-from-go`.

## Phase 2 - edit `gala-from-go`

- Inline the two repository-relative links in `.agents/skills/gala-from-go/references/gaps.md` (the probe-corpus references under "Gaps Worth Filing" and "Constructs That Were Never Gaps") as facts without paths.
- Sweep all three files for `server/`, `docs/`, `backlog/`, probe names, and release pins; none may remain.
- Keep rows version-agnostic; do not add compiler history.

## Phase 3 - add the loop skill

Create `.agents/skills/gala-loop/SKILL.md` with frontmatter name `gala-loop` and a description for translating server Go files to GALA or continuing the loop.
- Preconditions: `nix develop`; record `gala version`, `which gala`, and the flake rev; read `server/third_party/gala/VENDORED_FROM`; confirm the twins are committed and the tree is clean.
- Step zero, sync: when the rev differs from the vendored provenance, run the ledger's re-vendor procedure, then `verify.sh --write`, review the per-twin diff, `go build ./...`, and the canonical backend test sequence in `server/README.md`; record the new provenance and any shape changes. The sync is its own iteration.
- Iteration: pick the next cursor candidate; read the ledger findings for its package; load `gala-from-go` and follow its triage, mechanical pass, and verify; register the twin in the ledger and `server/README.md`; run `verify.sh` and the backend test sequence.
- Workaround blockers: record the substitution in the ledger and continue.
- No-workaround blockers: classify with `references/gaps.md`; add `server/scripts/gala/probes/<slug>/` with `main.gala` and notes (invocation, observed diagnostic, expected Go shape, compiler rev); search the upstream tracker; create a Backlog task containing the filled report template. The loop does not file upstream itself.
- Stop conditions and never list: restate the discipline, not the construct rows (no hand-edited twins, no root transpile, no signature or wire reshaping, no relaxed tests).

## Phase 4 - add `verify.sh` and the probe scheme

`server/scripts/gala/verify.sh`:
- Print `gala version`, `which gala`, the flake rev, and `VENDORED_FROM`.
- Find `server/**/*.gala` excluding `third_party/`; for each run `(cd <pkg> && gala transpile -i <name>.gala -o <tmp>)` and `cmp` against the committed `.go`.
- Run `gofmt -l` over generated twins and `go build ./...` in `server/`.
- Default mode is drift detection; `--write` regenerates in place; exit non-zero on any mismatch.
- Add `server/scripts/gala/.gitignore` for generated probe outputs (`*.gen.go`) and any transpiler cache.
- Probe convention: one committed directory per open blocker with `main.gala` plus `notes.md`, removed when the blocker closes.

## Phase 5 - delete the old stack and update pointers

- Move `server/scripts/20260923_gala_translation_probes/` aside under `/tmp/opencode/` before removing it, per the rewrite-safety rule, then `git rm` the tracked files.
- `git rm -r docs/gala-translation.md docs/gala-rewrite-playbook.md .agents/skills/gala-rewrite/`.
- Update `server/README.md`'s GALA section to point at `server/GALA.md`, `.agents/skills/gala-loop/SKILL.md`, and `server/scripts/gala/verify.sh`; drop the twin table and the per-row commands.
- Sweep with ripgrep for `gala-translation`, `gala-rewrite-playbook`, `gala-rewrite`, and `20260923_gala_translation_probes`; fix or record every hit, leaving completed Backlog tasks as historical records.

## Phase 6 - sync runtime and re-baseline

- Enter the dev shell; record `gala version`, `which gala`, and `jq -r '.nodes.gala.locked.rev' flake.lock`.
- Re-vendor per the ledger: copy `~/.gala/stdlib/v<version>` over `server/third_party/gala/`; delete per-package `go.mod` and `go.sum`; rewrite the root `go.mod` (`module martianoff/gala`, go directive from the snapshot); keep `LICENSE`; write `VENDORED_FROM` (version, rev, source path, date); `go build ./...` inside the vendored tree.
- `verify.sh --write`, then review the diff per twin: an expected codegen change is kept and recorded, a regression is a finding and stops the sync rather than being papered over.
- `go build ./...` in `server/` and the canonical backend test sequence.

## Phase 7 - first loop iteration (#648)

- Reproduce the type-position import hole from TASK-0340 on the synced compiler and record whether it persists.
- If it persists: probe at `server/scripts/gala/probes/type-position-import/`, a ledger finding, and a Backlog report task with the filled template.
- If it is fixed: drop it from the open list, record the fix in the ledger, and comment on upstream `#648`.

## Phase 8 - bookkeeping and verification

- Archive with carried facts: TASK-0339 and 0339.01-.05, TASK-0323, TASK-0330.05, TASK-0340; finalize TASK-0330 as Done with a comment that `.05` was superseded.
- Run the dangling-reference sweep, `npm run format:markdown:check`, `go build ./...`, and `codebase-memory-mcp cli index_repository --repo-path .`.
- Record evidence in notes, write the final summary, mark Done.

## Constraints

- `docs/AGENTS.md` applies to `server/GALA.md` and `server/README.md`: one sentence per line and one table row per line.
- `.agents/**` is outside the Markdown pipeline; apply the same style by hand.
- No hand-edited generated twin; no reshaped Go-facing signature; no relaxed test.
- Root Markdown uses the Prettier sentences-per-line pipeline, not oxfmt.
- The re-vendor is a separate iteration from a translation; do not mix them.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Execution recon (2026-10-04)

Confirmed against the then-current tree:
- `gala version` = `GALA version 0.84.1`; `which gala` = `/nix/store/k7lxw6inxv6zydnylkh0xwbc1llcskmm-gala-0.84.1/bin/gala`; flake rev = `e2e28c318ff1eb606f7f607199b629d93b78ab1f` (locked 2026-10-02T12:21:09Z).
- `server/third_party/gala/VENDORED_FROM` does not exist; the current vendored tree carries only the old root `go.mod` comment naming the 0.84.1 release extraction. The extraction at `~/.gala/stdlib/v0.84.1` was regenerated by the pinned compiler and differs from the vendored tree in all 60 `.gala`/`.gen.go` files plus the extraction layout, so a sync is required.
- The only open upstream issue is #528 ("Triage language limitations"). Every filed report #611-#621 and #648 is closed; #611-#620 closed 2026-10-02 before the flake lock, #648 on 2026-10-03 and #621 on 2026-10-04 after it.

Deviations from the agreed plan, all within its confirmed scope and recorded here:

1. #619 is fixed in the pinned compiler. `gala transpile` now emits declaration comments, package comments, and `swag`-shaped annotations; only comments inside function bodies are dropped. The plan's premise that declaration comments cannot survive is stale, so `routes/users.go` is no longer blocked by annotation loss and joins the cursor, and the ledger records #619 as fixed on this rev.
2. #621 (defined-type newtype) and #648 (type-position import hole) are closed upstream but their fixes are not in the pinned rev; both were re-verified as still present on the pinned compiler. The #648 repro indeed persists, so AC #8's persistent branch applies.
3. The measured twin re-baseline is 4 of 15, not all 15: comment output changes `eventid`, `middleware/auth`, and `observability/redact`; new `StructMeta_*` codec codegen changes `postgres/dailylogs`.
4. Two tracked statements in the middleware sources claim the transpiler emits no comments (`middleware/auth.gala`, `middleware/doc.go`); they are stale and get a minimal correction. The skill's comment claims in `SKILL.md` step 5 and `gaps.md` "Gaps Worth Filing" are corrected to the verified version-agnostic behavior while inlining the two repository-relative probe links required by AC #2.
5. `verify.sh` excludes `server/third_party/` and `server/scripts/gala/probes/` from the twin walk, so committed probe repros are not mistaken for twins.
6. Only the #648 probe is created by this task; the findings table lists a probe where one exists and the loop skill creates one per blocker when the cursor reaches it.

## Carried facts from the superseded tasks (AC #9)

- TASK-0339 (typed-data roster) and its five subtasks never landed an artifact. Its D1-D3 are replaced by this task's decisions: the ledger is `server/GALA.md` with no generated roster page; the twin inventory lives in the ledger registry rather than in `server/README.md`, which now points at it; the probe corpus, the typed verdict table, the emitter, and the audit split are dropped. The one goal that survives is skill independence, completed here by inlining the probe-corpus and `go_interop` links and sweeping repository paths out of `.agents/skills/gala-from-go/`. The pin-admissibility method was not extracted into a `references/probes.md`; the loop skill's probe convention and the skill's existing shape-claim and trap sections carry it.
- TASK-0323 (a CI or scheduled version-gated corpus run) was never implemented. The corpus is deleted; the version gate is now `verify.sh`'s provenance print plus its flake-rev mismatch failure against `VENDORED_FROM`. CI integration stays deliberately out of scope for the local loop.
- TASK-0330.05 (run the playbook end to end on a real file) was never run. Its function is superseded by the loop skill; the first executed loop iteration is the #648 re-check, and twin translations remain in the cursor.
- TASK-0340 (the type-position import hole) is carried into the ledger finding row, the committed probe `server/scripts/gala/probes/type-position-import/`, and TASK-0342. Its durable facts: `struct Holder(F Future[int])`, `var holder Future[int]`, and `func Await2(f Future[int]) int = 1` all transpile with no diagnostic, the generated Go keeps the bare name, and `go build` reports `undefined: Future` at the `//line` positions; upstream #648 was filed 2026-09-30 and closed 2026-10-03, after the pinned rev.
- The old probe corpus is preserved outside the repository at `/tmp/opencode/backup-20260923_gala_translation_probes/` (84 MB) for the remainder of this session; it is not part of the deliverable.

## Verification evidence (2026-10-04)

- `server/scripts/gala/verify.sh` default mode: provenance printer showed `GALA version 0.84.1`, `which gala` at `/nix/store/k7lxw6inxv6zydnylkh0xwbc1llcskmm-gala-0.84.1/bin/gala`, flake rev `e2e28c3...`, and the four `VENDORED_FROM` fields; result `verify.sh: OK (15 twins)`, exit 0.
- `server/scripts/gala/verify.sh --write`: wrote the 4 drifted twins (`eventid/eventid.go`, `middleware/auth.go`, `observability/redact.go`, `postgres/dailylogs.go`), then `OK (15 twins)`. The comment changes are #619 output; the `StructMeta_DailyUserLogMember` block is the new codec codegen and builds against the synced `std` package.
- `eventid/doc.go` was retired because the emitted package comment duplicated it in `go doc`; the source-location sentence moved into `eventid.gala`. `middleware/auth.gala` and `middleware/doc.go` lost their stale "transpiler emits no comments" statements.
- Canonical backend test sequence: all packages `ok`, including `postgres` (7.172s) and `routes` (3.742s), with no failures.
- `npm run format:markdown:check`: exit 0 after `npm run format:markdown`.
- `npm run fmt:check`: `All matched files use the correct format.`
- `go build ./...` in `server/`: exit 0; `go build ./...` inside `server/third_party/gala/`: exit 0.
- Dangling-reference sweep outside `backlog/` for `gala-translation`, `gala-rewrite-playbook`, `gala-rewrite`, and `20260923_gala_translation_probes`: clean. The only remaining hits are historical task records under the Backlog archive and completed tasks.
- Skill sweep for `server/`, `docs/`, `backlog/`, probe names, and release pins under `.agents/skills/gala-from-go/`: clean; the three repository-relative links are inlined (the plan counted two, and the third was the `go_interop` source link noted in the deviations).
- #648 probe: transpile exit 0 with no diagnostic; `go build` exit 1 with `main.gala:4,5,7,13,20: undefined: Future`; recorded in `server/scripts/gala/probes/type-position-import/` and reported as TASK-0342.
- codebase-memory refresh: project `home-eyjafjallajokull-Desktop-gh-timeful` indexed with 11317 nodes and 46314 edges, status `indexed`.
- Superseded tasks archived under `backlog/archive/tasks/`: TASK-0339, 0339.01-.05, TASK-0323, TASK-0330.05, TASK-0340. TASK-0330 is `status: Done` with the `.05` supersession comment and final summary; its `.01-.04` remain recorded.

## Archived-task supersession records (TASK-0343, AC #6)

The superseded tasks archived by this task carry no supersession comment of their own: the Backlog MCP resolves archived tasks for viewing but refuses edits with `Task not found`, and `BACKLOG_WORKFLOW.md` forbids editing their generated Markdown directly. This section is therefore their supersession record. TASK-0339 and its subtasks TASK-0339.01 through TASK-0339.05, TASK-0323, and TASK-0340 were all superseded by TASK-0341 on 2026-10-04 and archived under `backlog/archive/tasks/`; their durable facts are carried in the 'Carried facts from the superseded tasks (AC #9)' section above and in `server/GALA.md`. TASK-0330.05 is the exception: it is annotated by the supersession comment on TASK-0330, its parent.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-10-04 12:57
---
Review of the staged TASK-0341 work (2026-10-04). Verified: `server/scripts/gala/verify.sh` passes with `verify.sh: OK (15 twins)` on flake rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f`; the type-position-import probe reproduces exactly (`main.gala:4,5,7,13,20: undefined: Future`, clean transpile); `server/third_party/gala/` is byte-identical to `~/.gala/stdlib/v0.84.1` apart from `VENDORED_FROM` and the flattened module files; no references to the deleted documents, skill, or corpus survive outside `backlog/`; `.agents/skills/gala-from-go/` has no repository-relative links; the registry counts and the README, Dockerfile, and `server/go.mod` claims match; `npm run format:markdown:check`, root `npm run fmt:check`, and `git diff --cached --check` pass. Findings: the loop skill step 4 and `server/GALA.md`'s twin-registration sentence instruct adding a `server/README.md` pointer row that no longer exists; the re-vendor procedure copies `~/.gala/stdlib/v<version>` without forcing or verifying that the extraction came from the pinned compiler, although the extraction carries a `.stdlib-extracted` version-and-hash marker and `verify.sh` never checks the recorded `compiler:` line; `verify.sh` exits 0 when `VENDORED_FROM` is missing and has no orphan-twin check; the `gaps.md` report template lacks the `defect` class that TASK-0342 uses; the #648 workaround wording is unclear; TASK-0330 is Done with all DoD boxes unchecked; the archived superseded tasks carry no supersession note; and the `gala-rewrite-consolidation` milestone has no milestone file. Decisions: TASK-0343 is filed for the corrections; TASK-0342 is closed as the report handoff; TASK-0344 is filed for the flake bump that retires #648 and re-checks #621, depending on TASK-0343. No change to this task's Done state or artifacts.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Consolidated the overlapping GALA documents into one ledger (`server/GALA.md`, 154 lines), one project-independent translation reference (`.agents/skills/gala-from-go/`), one loop skill (`.agents/skills/gala-loop/SKILL.md`), and one verification script (`server/scripts/gala/verify.sh`). Deleted `docs/gala-translation.md`, `docs/gala-rewrite-playbook.md`, `.agents/skills/gala-rewrite/`, and the 84 MB probe corpus, and updated `server/README.md` to point at the ledger, the loop, and the script.

Re-vendored `server/third_party/gala/` from the pinned compiler's own extraction and recorded `VENDORED_FROM` (0.84.1, rev `e2e28c318ff1eb606f7f607199b629d93b78ab1f`, source path, date). The re-baseline touched 4 of 15 twins: comment output in `eventid`, `middleware/auth`, and `observability/redact`, and new `StructMeta_*` codec output in `postgres/dailylogs`. All 15 twins regenerate byte-identically, `go build ./...` passes, and the canonical backend suite is green.

First loop iteration: the type-position import hole persists on the pinned rev (upstream #648 closed after the lock), so `server/scripts/gala/probes/type-position-import/` and report TASK-0342 were created. Archived TASK-0339 and its five subtasks, TASK-0323, TASK-0330.05, and TASK-0340 with their durable facts carried into TASK-0341 and the ledger, and finalized TASK-0330 Done with the `.05` supersession comment.

Evidence: `verify.sh` OK (15 twins), `npm run format:markdown:check` and root `npm run fmt:check` pass, the dangling-reference sweep is clean, the complete backend test sequence passes, and the codebase-memory graph was refreshed (11317 nodes, 46314 edges).
<!-- SECTION:FINAL_SUMMARY:END -->
