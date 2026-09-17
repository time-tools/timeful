---
id: TASK-0253
title: Point the flake Backlog.md input at the time-tools fork
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-17 14:48'
updated_date: '2026-09-17 15:08'
labels:
  - nix
  - backlog-md
  - tooling
dependencies: []
references:
  - flake.nix
  - flake.lock
  - .opencode/opencode.jsonc
  - 'https://github.com/time-tools/backlog.md'
modified_files:
  - flake.nix
  - flake.lock
priority: medium
type: chore
ordinal: 244000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Switch the Nix flake's Backlog.md input from upstream MrLesk/Backlog.md to the time-tools/Backlog.md fork, and regenerate flake.lock.

Current state:
- flake.nix line 11 pins `backlog-md.url = "github:MrLesk/Backlog.md/583f928dfa65266df994a4323566eb426446ad55"` (upstream commit from 2026-08-15, builds as Backlog.md 1.50.1).
- flake.lock locks that same MrLesk rev; the URL carries an explicit rev, so `nix flake update backlog-md` alone cannot move it.
- The fork's default branch is `main`; at task creation its HEAD is `aded8e254e6a0205b878cf07e631d1a592782040` (2026-09-17, package.json version 1.52.0) and it is currently identical to upstream main with no fork-only commits.
- The devShell-provided `backlog` binary is also the Backlog MCP server: .opencode/opencode.jsonc launches `backlog mcp start` from PATH.

Scope:
- Update the `backlog-md` input URL to the canonical fork name `github:time-tools/Backlog.md` pinned to the fork main rev, keeping the explicit-rev convention used by the nixpkgs and flake-parts inputs.
- Regenerate flake.lock from the changed input; do not hand-edit it.
- Keep the rest of the flake (packages, apps, devShell packages, shellHook) unchanged.
- No other tracked files reference MrLesk/Backlog.md, so no doc updates are expected.

Out of scope: changing how the fork is consumed (tag refs, branch tracking, overlays) or patching fork content.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 flake.nix's backlog-md input sources github:time-tools/Backlog.md at an explicit rev, and no tracked file references MrLesk/Backlog.md
- [x] #2 flake.lock regenerates cleanly: nix flake metadata succeeds and the backlog-md node's owner is time-tools with the pinned rev and an updated narHash
- [x] #3 nix develop --command backlog --version reports 1.52.0 (fork HEAD version) instead of 1.50.1
- [x] #4 The devShell binary still serves this repo: nix develop --command backlog task list succeeds
- [x] #5 nix develop --command backlog mcp start remains the launchable MCP entry point referenced by .opencode/opencode.jsonc
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
1. Change flake.nix `backlog-md.url` to `github:time-tools/Backlog.md/aded8e254e6a0205b878cf07e631d1a592782040` (fork main HEAD, v1.52.0).
2. Regenerate flake.lock with `nix flake lock` (the input ref changes, so no --update-input needed) and confirm the backlog-md node records owner time-tools, the new rev, and a new narHash.
3. Verify inside the dev shell: `nix flake metadata` evaluation, `backlog --version` reports 1.52.0, `backlog task list` reads this repo, and `backlog mcp start` remains a valid entry point.
4. Confirm no tracked file still references MrLesk/Backlog.md and record evidence in the final summary.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-17: flake.nix line 11 changed to `backlog-md.url = "github:time-tools/Backlog.md/aded8e254e6a0205b878cf07e631d1a592782040"`. `nix flake lock` output: Updated input 'backlog-md': 'github:MrLesk/Backlog.md/583f928' (2026-08-15) -> 'github:time-tools/Backlog.md/aded8e2' (2026-09-17). Diff touches only the backlog-md node (`owner: time-tools`, new `rev`, new `narHash` sha256-eDWI/ubglLHpS8LU+IC3dZu8fVsQx7z8brsbBkl6RD0=); bun2nix/flake-utils/nixpkgs nodes were unchanged.

Evidence: `nix flake metadata` shows backlog-md: github:time-tools/Backlog.md/aded8e2 (2026-09-17 06:45:46). `nix develop --command backlog --version` prints 1.52.0. `nix develop --command backlog task list` lists tasks but exits 1 from the pre-existing duplicate TASK-238 ID collision; the 1.50.1 binary at /nix/store/vff582njpwbh5slh3s5lxxj5sn1jjbvx-backlog-1.50.1/bin/backlog exits 1 with the same warning, establishing the baseline. Worktree `git status --porcelain` was identical before and after running the 1.52.0 binary.

Evidence: `nix develop --command backlog mcp start < /dev/null` exits 0 with no output (clean stdio shutdown), so the .opencode/opencode.jsonc launch command remains valid. `git grep MrLesk` matches only TASK-0253's own description lines; flake.nix, flake.lock, docs, and scripts contain no upstream reference.

Checks note: the change is Nix build configuration only. No frontend files changed, so the frontend lint/fmt/typecheck/build/unit list does not apply; no new unit or e2e test covers this. Relevant Nix checks all ran: lock resolution, flake metadata evaluation, devShell build, and binary smoke tests. `scripts/markdown.mjs` excludes `backlog/**` and no other Markdown changed. `graphify update .` ran after the edit (graphify-out refreshed).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Switched the flake's Backlog.md input from upstream MrLesk/Backlog.md to the time-tools/Backlog.md fork and regenerated flake.lock.

Change: flake.nix now pins `github:time-tools/Backlog.md/aded8e254e6a0205b878cf07e631d1a592782040` (fork main HEAD, 2026-09-17, Backlog.md 1.52.0). `nix flake lock` updated only the backlog-md node (narHash sha256-eDWI/ubglLHpS8LU+IC3dZu8fVsQx7z8brsbBkl6RD0=); its transitive inputs were already identical, so no other lock nodes changed.

Verification:
- `nix flake metadata` resolves backlog-md to github:time-tools/Backlog.md/aded8e2 and evaluates the flake.
- `nix develop --command backlog --version` prints 1.52.0 (previous devShell binary was 1.50.1), so the devShell builds the fork revision.
- `nix develop --command backlog task list` lists this repo's tasks. It exits 1 only because of the pre-existing duplicate TASK-238 ID collision; the previous 1.50.1 binary shows the same warning and exit code, so this is repo data, not a regression. The 1.52.0 binary left the worktree unchanged (identical `git status --porcelain` before/after).
- `nix develop --command backlog mcp start < /dev/null` exits 0 cleanly, so the MCP entry point launched by .opencode/opencode.jsonc remains valid.
- `git grep MrLesk` finds the name only inside TASK-0253's own historical description; no config, docs, or scripts reference upstream.

Scope notes: no frontend or server code changed, so the frontend check list and the unit/e2e suites were not applicable; the relevant checks for a flake input change are Nix evaluation plus the binary smoke tests above, all run. The repo Markdown formatter excludes `backlog/**`, and no tracked Markdown outside that directory changed, so no format:markdown work applied. `graphify update .` was run after the flake change.

Unrelated pre-existing worktree state: `backlog/backlog.md` and `backlog/config.yml` were already modified before this session (mtimes predate TASK-0253 creation).
<!-- SECTION:FINAL_SUMMARY:END -->
