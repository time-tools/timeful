---
id: TASK-0360
title: Bump the GALA compiler pin to 0.92.0 and re-baseline the twins
status: In Progress
assignee: []
created_date: '2026-10-21 20:02'
updated_date: '2026-10-24 06:44'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - 'https://github.com/martianoff/gala/releases/tag/0.92.0'
  - 'https://github.com/martianoff/gala/pull/832'
  - 'https://github.com/martianoff/gala/pull/833'
priority: medium
type: chore
ordinal: 364300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The GALA flake lock already sits on `a5206156` (0.91.0 + 2) from the lock-only update commit `5ddd49d8`, but the bump iteration it requires never ran: `server/GALA_COMPILER` still records `014c2f23` (0.90.0), `server/go.mod` still requires `go.gala.fyi/stdlib v0.90.0`, and `scripts/gala/verify.sh` now fails its compiler, rev, and extraction provenance checks. The ledger's rule is that a rev mismatch is resolved by a bump iteration and never by translating across it.

Upstream `master` is now `4bd2d4e8` (`0.92.0` + 5, released 2026-10-20), a pointer-safety release: `Ptr[T]` and `OptionPtr[T]` replace GALA-native nullable pointers, `Ptr` parameters are nil-checked at the boundary, local `val`s lower to plain Go variables, a function that returns an if-expression emits a plain `if` instead of an immediately invoked closure, and the post-release `#832`/`#833` fixes land the zero-argument constructor rules. The release also moves the published `go.gala.fyi/stdlib` module to `v0.92.0`.

Bump the lock to the newest recorded rev, refresh the extraction provenance, move the stdlib module, re-baseline all committed twins under the new compiler, re-run the committed probes, and record every shape change in the ledger.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 flake.lock's gala input is updated to 4bd2d4e8 (or a recorded newer rev), and nix develop provides GALA 0.92.0 with a working gala-local.
- [ ] #2 The extraction marker is refreshed by the pinned compiler, and server/GALA_COMPILER records the 0.92.0 version, flake rev, fingerprint, and date.
- [ ] #3 server/go.mod and server/go.sum require the paired go.gala.fyi/stdlib v0.92.0, and the module builds against it.
- [ ] #4 scripts/gala/verify.sh --write regenerates all committed twins, the per-twin diff is reviewed, and every shape change is recorded with no regression papered over.
- [ ] #5 Every committed probe is re-run on the new rev and the cursor records the rev the blockers were re-verified on.
- [ ] #6 go build ./... and the canonical backend test sequence pass under the bumped compiler.
- [ ] #7 server/GALA.md, the gala-from-go construct reference, and the probe notes record the new provenance and shape changes.
- [ ] #8 Changed Markdown is formatted and the code knowledge graph is refreshed.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
- [ ] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [ ] #6 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
- [ ] #7 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [ ] #8 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Record baseline provenance and the current provenance failure (already observed on 2026-10-21).\n2. Update flake.lock with nix flake update gala and build the dev shell.\n3. Refresh the extraction marker and write server/GALA_COMPILER.\n4. Move go.mod/go.sum to go.gala.fyi/stdlib v0.92.0.\n5. Run verify.sh --write, review the per-twin diff, then verify.sh check.\n6. Re-run the committed probes and update their notes and the cursor rev.\n7. Run go build ./... and the canonical backend test sequence.\n8. Update the ledger and construct references, format Markdown, refresh the code graph, and finalize the task.
<!-- SECTION:PLAN:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: DeepSeek Harness
created: 2026-10-24 06:44
---
The pin moved past the recorded 0.92.0 rev to the 0.92.1 release: flake.lock now locks martianoff/gala at 911cc059 (0.92.1 plus the upstream nix pin of the release binary, PR #835), which is version-only relative to 4bd2d4e8 (gala.mod, MODULE.bazel, and nix/gala.nix). server/GALA_COMPILER records 0.92.1/911cc059 with the same extraction content hash (2c6059ca..., version prefix changed), server/go.mod requires go.gala.fyi/stdlib v0.92.1, all 30 twins regenerate byte-identically, the six committed probes were re-run on 2026-10-24, and go build ./... plus the canonical Compose backend suite pass (2026-10-24).

Observation outside this change: server/tmp-alias-probe and server/tmp-optmap-probe are staged scratch trees whose .gala files have no committed twins, so verify.sh --write reports them as MISSING; the run above passed with those two trees set aside. They need to be removed from the index or the verify contract before CI runs it.
---
<!-- COMMENTS:END -->
