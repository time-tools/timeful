---
id: TASK-0358
title: Bump the GALA compiler pin to 0.90.0 and re-baseline the twins
status: Done
assignee: []
created_date: '2026-10-17 22:41'
updated_date: '2026-10-17 22:41'
labels:
  - gala
  - tooling
dependencies: []
references:
  - flake.lock
  - server/GALA.md
  - server/GALA_COMPILER
  - server/scripts/gala/verify.sh
  - 'https://github.com/martianoff/gala/releases/tag/0.90.0'
  - 'https://github.com/martianoff/gala/issues/749'
  - 'https://github.com/martianoff/gala/issues/750'
  - 'https://github.com/martianoff/gala/pull/769'
  - 'https://github.com/martianoff/gala/pull/770'
  - 'https://github.com/martianoff/gala/pull/778'
documentation:
  - server/README.md
  - .agents/skills/gala-loop/SKILL.md
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
modified_files:
  - flake.lock
  - server/GALA_COMPILER
  - server/GALA.md
  - server/go.mod
  - server/go.sum
  - server/discord_bot/commands/active_users.go
  - server/discord_bot/commands/help.go
  - server/middleware/auth.go
  - server/models/uuid.gala
  - server/models/uuid.go
  - server/slackbot/commands/active_users.go
  - .agents/skills/gala-from-go/SKILL.md
  - .agents/skills/gala-from-go/references/constructs.md
  - server/scripts/gala/probes/models-defined-scalar-methods/notes.md
  - server/scripts/gala/probes/models-empty-struct/notes.md
  - server/scripts/gala/probes/models-fixed-array/notes.md
  - server/scripts/gala/probes/models-struct-tag/notes.md
  - server/scripts/gala/probes/routes-group-chan-recover/notes.md
  - server/scripts/gala/probes/routes-group-embedded-field/notes.md
priority: medium
type: chore
ordinal: 362300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Upstream master moved from the locked `cd2fdcb5` (0.87.1 + five commits) to `014c2f23` (0.90.0 + three commits), carrying the 0.88.0-0.90.0 releases and the fixes that close this repository's [#749](https://github.com/martianoff/gala/issues/749) alias-signature and [#750](https://github.com/martianoff/gala/issues/750) OptionFromMap-lambda reports, plus default values (`GALA-E0075`), the `ConstPtr` retirement and immutable-address refusal (`GALA-E0076`), sealed `default case`, and the removal of the Hindley-Milner inference fallback. The GALA ledger treats a flake rev mismatch as a stop condition and requires a bump as its own re-baseline iteration: every committed twin must regenerate under the new compiler, the adopted `go.gala.fyi/stdlib` module must move to the paired release, and the committed probes must be re-run because the rev moved. The bumped compiler refuses `models/uuid.gala`'s implicit opaque-type conversion with `GALA-E0064`, so the bump also carries that source fix.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 flake.lock's gala input is updated to `014c2f23` (or a recorded newer rev), and `nix develop` provides GALA 0.90.0 with a working `gala-local`.
- [x] #2 The extraction marker is refreshed by the pinned compiler, and `server/GALA_COMPILER` records the 0.90.0 version, flake rev, fingerprint, and date.
- [x] #3 `server/go.mod` and `server/go.sum` require the paired `go.gala.fyi/stdlib v0.90.0`, and the module builds against it.
- [x] #4 `verify.sh --write` regenerates all 29 committed twins, the per-twin diff is reviewed, and every shape change is recorded with no regression papered over.
- [x] #5 Every committed probe is re-run on the new rev; #749 and #750 are verified fixed, and the cursor records the rev the blockers were re-verified on.
- [x] #6 `go build ./...` and the canonical backend test sequence pass under the bumped compiler.
- [x] #7 `server/GALA.md`, the gala-from-go construct reference, and the probe notes record the new provenance, shape changes, closed reports, and re-verification rev.
- [x] #8 Changed Markdown is formatted and the code knowledge graph is refreshed.
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Provenance before (baseline)
- `nix develop -c gala version` = `GALA version 0.87.1`; `which gala` = `/nix/store/mib0p8skl8rh79alw3hp2nza8qbvhdwd-gala-0.87.1/bin/gala`; `flake.lock` rev `cd2fdcb50cf1bc988ed03c3f6cdc0c485403576c`; extraction `0.87.1 7b1dd2080a304c06a04eeb7240937109fd3d3f02e19b103f4755be79fb71d7fa`; `verify.sh` OK (29 twins).
- Precondition note: the worktree carried a known unrelated unstaged edit to `backlog/backlog.md`; it was left untouched. Both halves of every twin and the ledger were committed.

## Upstream delta
- Lock moved `cd2fdcb5` -> `014c2f23` (0.90.0 + 3 commits). The range carries 0.88.0, 0.89.0, and 0.90.0 plus #747, #751, #753 (fixes #749), #754 (fixes #750), #760, #761, #763, #764, #765, #767, #769, #770, #771, #775, #776, #778, #779, #784, #786, and #790.
- The two reports this repository filed, #749 (aliased import loses a Go call signature) and #750 (OptionFromMap lambda leaks `T`), were closed upstream on 2026-10-10 and are now in the pinned rev.
- Compiler identity: `gala version` reports `0.90.0`; the flake rev `014c2f23e45a2cd631b73cab38c0bdcb5e773347` is `0.90.0-3-g014c2f23`. The local `tmp/gala` checkout is at `268669aa` (`0.90.0-2`), one commit behind the lock.

## Bump
- Anonymous GitHub API access returned HTTP 403, so the lock was updated with `NIX_CONFIG="access-tokens = github.com=$(gh auth token)" nix flake update gala`; the token was never printed.
- Locked rev `014c2f23e45a2cd631b73cab38c0bdcb5e773347`, lastModified `1792273007`, narHash `sha256-p6WJqxbrkkVm3Ns0yDEpF6QGKHKXC2z1/rnaqy+Q8=`; the `github:martianoff/gala` input type is preserved.
- `nix develop` rebuilt the dev shell and provided `/nix/store/0rd7iwl6ffmvxllzkd40k59sr71ly1z8-gala-0.90.0/bin/gala`.
- Extraction refresh: deleted `~/.gala/stdlib/v0.90.0/.stdlib-extracted`, transpiled `server/eventid/eventid.gala` to a scratch path, and read `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`. The fingerprint changed with the embedded stdlib snapshot; this proves the pinned compiler wrote it.
- `server/GALA_COMPILER` records version, rev, extraction, and date 2026-10-18.
- `server/go.mod` and `server/go.sum` moved to `go.gala.fyi/stdlib v0.90.0` via `go get` then `go mod tidy`; the v0.87.1 sums were removed.

## Re-baseline and review
- First `verify.sh --write` failed: `models/uuid.gala` was refused with `GALA-E0064: cannot use zeroUUIDValue (string) as UUID: an opaque type never converts implicitly`. The source now spells `UUID(zeroUUIDValue)`, preserving the Go surface and behavior.
- Second `verify.sh --write`: five WRITE lines, all reviewed; every other twin was byte-identical.
  - `discord_bot/commands/active_users.go`, `middleware/auth.go`, and `slackbot/commands/active_users.go`: only the unreachable arm of an exhaustive `std.Try` match now panics with the compiler's descriptive zero-value message.
  - `discord_bot/commands/help.go`: `%v` -> `%s` for two string fields; output is identical because both are `string`.
  - `models/uuid.go`: the explicit `UUID(zeroUUIDValue)` conversion above.
- Final `verify.sh` check: provenance matches, `go build ./...` green, `verify.sh: OK (29 twins)`.

## Probes
- Re-ran all six committed probes on `014c2f23`:
  - `models-defined-scalar-methods`: default and suppression runs still reproduce the documented runtime-backed and runtime-free outputs.
  - `models-empty-struct`, `models-fixed-array`, `models-struct-tag`, `routes-group-chan-recover`, `routes-group-embedded-field`: the same refusals; the expected-token lists now include the new `default` keyword.
- #749 verification: a scratch `pgstore "timeful/server/postgres"` call now emits `std.GoTry(pgstore.DefaultRepository())`.
- #750 verification: an unannotated `go_interop.OptionFromMap(...).ForEach` lambda over `map[string]other.Thing` now emits `func(thing other.Thing)`.
- Cursor entry 1 now records blockers last re-verified on rev `014c2f23` (2026-10-18).

## Tests
- Canonical Compose backend sequence from `server/README.md` (existing `.env.test` and cache volumes): `up -d postgres-test postgres-test-bootstrap postgres-test-migrate`, then `run --rm server-route-test` exited 0 with every package `ok` and no `FAIL`, including `server/routes`, `server/postgres`, and `server/services/*`.

## Docs and checks
- `server/GALA.md`: compiler paragraph, provenance table, runtime version, report index for #749/#750, probe re-verification, and cursor rev updated. The registry is unchanged because no candidate was translated.
- `.agents/skills/gala-from-go/SKILL.md` and `references/constructs.md`: recorded the explicit opaque-type conversion, default values (`GALA-E0075`), immutable-address refusal (`GALA-E0076`), sealed `default case` and Go constructors, and removed the stale #749/#750 defects.
- The six probe notes gained a 2026-10-18 re-verification section with the new provenance.
- `npm run format:markdown:check` clean. `npm run lint:markdown` still reports one pre-existing error at `docs/environments.md:95`, untouched by this bump.
- Code graph refreshed: 9028 nodes / 28706 edges.
- Root `npm run fmt:check` not run: `scripts/` and `prettier/` were unchanged. e2e and Swagger not run: no frontend surface and no route annotation changed.
- No commit made: the user did not ask for one.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Bumped the GALA flake pin from `cd2fdcb5` (0.87.1) to `014c2f23` (0.90.0 + 3) and re-baselined all 29 committed twins against it.

Delivered:
- `flake.lock` gala node locked to `014c2f23e45a2cd631b73cab38c0bdcb5e773347`; the dev shell provides GALA 0.90.0; input type preserved.
- Extraction refreshed to `0.90.0 bd0c1591b19ce9971d64efd52406bab21723db64807b526c0f716102453bdd90`; `server/GALA_COMPILER` records version, rev, fingerprint, and date.
- `server/go.mod`/`go.sum` moved to `go.gala.fyi/stdlib v0.90.0`.
- `verify.sh --write` changed five twins, all reviewed: three descriptive panic-message changes in exhaustive `std.Try` matches, `help.go` `%v` -> `%s` on string fields, and `models/uuid.gala` `UUID(zeroUUIDValue)` for the new explicit opaque-type conversion rule. No regression was papered over.
- All six probes re-ran; existing blockers persist with only the `default` keyword added to expected-token lists. #749 now emits `std.GoTry` through an aliased import and #750 now infers the concrete map value type; both are recorded fixed.
- `server/GALA.md`, the gala-from-go reference, and the probe notes record the new provenance, shape changes, closed reports, and re-verification rev.

Evidence:
- `verify.sh` check: `OK (29 twins)` with `go build ./...` green.
- Canonical Compose backend suite passed, every package `ok`, no `FAIL`.
- `npm run format:markdown:check` clean; code graph refreshed to 9028 nodes / 28706 edges.
- e2e and Swagger not required: no frontend surface and no route annotation changed. The pre-existing unrelated `backlog/backlog.md` edit was left untouched, and nothing was committed (not requested).
<!-- SECTION:FINAL_SUMMARY:END -->
