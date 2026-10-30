---
id: TASK-0370
title: Bump the GALA compiler pin to 0.95.0 and re-baseline the twins
status: Done
assignee:
  - DeepSeek Harness
created_date: '2026-10-29 13:50'
updated_date: '2026-10-29 15:49'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - 'https://github.com/martianoff/gala/releases/tag/0.93.1'
  - 'https://github.com/martianoff/gala/releases/tag/0.94.0'
  - 'https://github.com/martianoff/gala/releases/tag/0.95.0'
  - 'https://github.com/martianoff/gala/pull/902'
documentation:
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - .agents/skills/gala-loop/SKILL.md
  - server/README.md
priority: medium
type: chore
ordinal: 374300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The GALA flake input is locked at `8ba02b2b` (0.93.0) and `server/go.mod` requires `go.gala.fyi/stdlib v0.93.0`, while upstream has released 0.95.0. The release tag is `6701c03b`; `master` head `173f7d77` is the 0.95.0 post-release Nix pin that makes `gala-local` point at the release binary, which is the same shape the 0.92.1 pin took. `go.gala.fyi/stdlib v0.95.0` is published.

`server/GALA.md` requires a bump iteration whenever the flake rev and `server/GALA_COMPILER` disagree: refresh the extraction provenance, move the paired stdlib module, regenerate all committed twins, re-run the committed probes, and run the build and the canonical Compose backend suite.

The 0.93.0-to-0.95.0 range includes behavioral changes that must not be papered over. 0.93.1 fixes diagnostic delivery for unparseable siblings and generated-import leakage; 0.93.2 fixes tuple unification and aliased-import types; 0.94.0 makes `Tuple`/`Option`/`Either` hashable and rejects `==` on non-comparable values as GALA-E0084; 0.95.0 makes `galop` the default parser, makes `List`/`HashSet`/`Array` hashable, and breaks the previous behavior where a block-form struct field could carry a Go struct tag: such tags are now GALA-E0085, so any `.gala` source relying on them must move to a handwritten Go sibling. This repository already recorded the block-form tag capability and filed upstream issue #885, so the new rejection is expected and must be reflected in the ledger and the construct reference.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 flake.lock locks martianoff/gala at 173f7d77 (or a recorded newer rev that reports 0.95.0), and `nix develop` provides `GALA version 0.95.0` with a working `gala-local`.
- [x] #2 server/GALA_COMPILER records the 0.95.0 version, flake rev, extraction fingerprint, and date, written from the pinned compiler under the documented GALA_HOME.
- [x] #3 server/go.mod and server/go.sum require go.gala.fyi/stdlib v0.95.0, and `go build ./...` in server/ succeeds against it.
- [x] #4 `scripts/gala/verify.sh --write` regenerates every committed twin; the per-twin diff is reviewed, every shape change is recorded rather than papered over, and `verify.sh --check` reports OK for all twins with no RUNTIME, DRIFT, or MISSING entries.
- [x] #5 Every committed probe is re-run on the new rev, its notes record a re-verification section with the new compiler version, and the ledger cursor records the rev the blockers were re-verified on.
- [x] #6 `go build ./...` and the canonical Compose backend test sequence pass under the bumped compiler.
- [x] #7 server/GALA.md, the gala-from-go construct reference, and the probe notes record the new provenance and any shape or construct changes, including the 0.95.0 struct-tag rejection (GALA-E0085) wherever the tag guidance appears.
- [x] #8 Changed Markdown is formatted with `npm run format:markdown` and the code knowledge graph is refreshed.
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
1. Record baseline provenance: `gala version`, flake rev, `GALA_COMPILER`, worktree state, twin count.
2. `nix flake update gala`; confirm the lock lands on 173f7d77 (0.95.0 + post-release Nix pin) and `nix develop -c gala version` reports 0.95.0 with a working `gala-local`.
3. Force the pinned compiler's stdlib extraction under `$GALA_HOME`, read the `.stdlib-extracted` marker, and write `server/GALA_COMPILER` (0.95.0, rev, marker, date).
4. Move `server/go.mod`/`go.sum` to `go.gala.fyi/stdlib v0.95.0` (`go get` + `go mod tidy`), then `go build ./...`.
5. Run `scripts/gala/verify.sh --write`, review every per-twin change, and re-run `--check`. If 0.95.0's GALA-E0085 struct-tag rejection breaks a twin, move the tagged shape to a handwritten sibling (or apply the construct reference's replacement) and record the migration in the ledger.
6. Re-run the six committed probes on the new rev; append a re-verification section to each notes file and update the ledger cursor rev.
7. Update `server/GALA.md` (provenance table, 0.93.0-to-0.95.0 range changes, cursor rev, probe dates) and `.agents/skills/gala-from-go/references/constructs.md` (struct tag row now GALA-E0085) plus the tag guidance in `SKILL.md`/`gaps.md` where it appears.
8. Run `go build ./...`, `go vet ./...`, and the canonical Compose backend sequence from `server/README.md`.
9. Format Markdown, refresh the code knowledge graph, verify all acceptance criteria, and finalize the task.

Risks: the 0.95.0 struct-tag rejection may require a source migration; the galop-based Nix build may need a fresh dev-shell evaluation; the new stdlib extraction must be written before the verifier can pass its provenance check.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Baseline provenance: `GALA version 0.93.0`, flake rev `8ba02b2b`, extraction `0.93.0 6368c093…`, 78 twins, clean worktree. `nix flake update gala` moved the lock to `173f7d77` (0.95.0 + post-release Nix pin, PR #902), and `nix develop --command gala version` reports 0.95.0. Forced extraction produced `0.95.0 7b68112a768d6d4776444c8ab561ce629c219d86748f93f0ca1f485daf7d1648`; `server/GALA_COMPILER` records it. `go get go.gala.fyi/stdlib@v0.95.0` plus `go mod tidy` moved the module (GOPATH/GOMODCACHE stay in the checkout).

`verify.sh --write` regenerated 15 of 78 twins and left the other 63 byte-identical; the 15 diffs are codegen-only: match IIFEs become bare blocks with `obj := <subject>`, and `std.<Variant>[T]{}.Unapply` calls become generated per-file `_unapply_std_<Variant>_<hash>` generic helpers (so `middleware/auth.go` drops its `models` and `postgres` imports). `verify.sh --check` then reported `78 discovered, 78 transpiled, 78 ok` with gofmt, orphan scan, and `go build ./...` green. `go build ./...` and `go vet ./...` also pass in server/.

Re-ran all six committed probes on 0.95.0. `models-struct-tag` changed status: the block-form tag workaround is gone and the probe now produces seven `GALA-E0085` diagnostics; #885 closed completed via PR #894, so the construct is a documented answer whose hint names `Codec[T]` and a handwritten `.go` file. The other five reproduced their refusals (expected-token lists now name `default`); the `opaque type` default/alias/suppression variants are unchanged. Updated all six probe notes, `constructs.md`'s struct-tag row, `gaps.md` (tag removed from never-a-gap, added to the answered list, history note), `SKILL.md`'s boundary sentence, and `server/GALA.md` (provenance, 0.93.1-to-0.95.0 range, 15-twin shape changes, findings table, #885/#528 rows, cursor rev and tag rationale). `npm run format:markdown` and `format:markdown:check` pass.

Final verification. The canonical Compose sequence's `postgres-test-bootstrap` still hit the known Docker-bridge timeout (`psql` to `postgres-test` timed out), so a temporary gitignored override, `tmp/compose.test.host.yaml`, ran the bootstrap, migrate, and `server-route-test` containers with host networking against the published `127.0.0.1:5433`. The bootstrap created and granted the test database, the migrate image reported `no migrations to run` at `20260919000000` and re-applied the grants, and `server-route-test` passed every package (`postgres` 17.826s, `routes` 5.218s) after downloading `go.gala.fyi/stdlib v0.95.0`. Also: `python3 server/scripts/gala/test_verify.py` 8/8 ok, `npm run fmt:check` pass, `npm run format:markdown` and `format:markdown:check` pass, and `codebase-memory-mcp cli index_repository --repo-path .` indexed 9511 nodes / 30822 edges. No swagger annotations, frontend files, or environment-contract documents changed, so `swag init` and e2e were not required.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved the GALA compiler pin to the 0.95.0 release line and completed the ledger's bump iteration.

Pin: `flake.lock` now locks `martianoff/gala` at `173f7d77`, the post-release Nix pin for 0.95.0 (PR #902), and `nix develop` provides `GALA version 0.95.0` with a working `gala-local`. `server/GALA_COMPILER` records 0.95.0, rev `173f7d77`, extraction `0.95.0 7b68112a…`, and 2026-10-29. `server/go.mod`/`go.sum` require `go.gala.fyi/stdlib v0.95.0`, verified by the container suite downloading it.

Re-baseline: `verify.sh --write` regenerated 15 of the 78 twins (63 byte-identical); every change is codegen-only. The affected `match` sites now lower to a bare block (`obj := <subject>`) instead of an immediately invoked closure, and `Unapply` calls route through generated per-file `_unapply_std_<Variant>_<hash>` helpers, which let `middleware/auth.go` drop two now-unused imports. No Go-facing signature, wire shape, or behavior changed; `verify.sh --check` reports `78 discovered, 78 transpiled, 78 ok` with gofmt, orphan scan, and `go build ./...` green. `go build ./...` and `go vet ./...` also pass directly.

Construct status change: 0.95.0 rejects Go struct tags as `GALA-E0085`, so the 0.93.0 block-form tag workaround is gone. `models/location.go` and `models/event.go` stay handwritten, and the struct tag is now a documented answer whose hint names `Codec[T]` and a handwritten `.go` file. #885 closed completed via PR #894. Updated `server/GALA.md` (provenance, 0.93.1-to-0.95.0 range, 15-twin shape changes, findings table, #528/#885 rows, cursor rev), `constructs.md`, `gaps.md`, `SKILL.md`, and all six probe notes with 2026-10-29 re-verification; the five other probes reproduced their refusals, with `default` now present in the expected-token lists.

Evidence: `verify.sh` OK (78 twins); `go test ./...` in the isolated stack passed every package (`postgres` 17.826s, `routes` 5.218s); verifier regression tests 8/8; Markdown format and root `fmt:check` pass; code graph refreshed (9511 nodes / 30822 edges). No swagger, frontend, or environment-contract changes were required. The Compose bridge issue from earlier sessions recurred, so the bootstrap/migrate/route-test containers ran with host networking against the isolated database, matching the documented workaround.
<!-- SECTION:FINAL_SUMMARY:END -->
