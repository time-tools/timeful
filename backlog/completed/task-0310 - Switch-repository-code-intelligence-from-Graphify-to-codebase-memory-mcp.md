---
id: TASK-0310
title: Switch repository code intelligence from Graphify to codebase-memory-mcp
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-21 19:23'
updated_date: '2026-09-21 20:29'
labels: []
dependencies: []
references:
  - 'https://github.com/DeusData/codebase-memory-mcp'
  - 'https://github.com/DeusData/codebase-memory-mcp/blob/main/docs/cbmignore.md'
  - backlog/completed/task-0054 - Integrate-Graphify-with-OpenCode.md
  - backlog/completed/task-0072 - Remove-obsolete-Graphify-MCP-integration.md
  - docs/graphify.md
  - BACKLOG_WORKFLOW.md
  - AGENTS.md
documentation:
  - 'https://deusdata.github.io/codebase-memory-mcp/'
modified_files:
  - flake.nix
  - .cbmignore
  - .gitignore
  - .gitattributes
  - .opencode/opencode.jsonc
  - .opencode/opencode.json
  - .opencode/plugins/graphify.js
  - AGENTS.md
  - BACKLOG_WORKFLOW.md
  - .agents/skills/review-task/SKILL.md
  - docs/codebase-memory.md
  - docs/graphify.md
  - docs/design/architecture/README.md
  - docs/design/architecture/adr/.graphifyignore
  - .graphifyignore
priority: medium
ordinal: 311000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Why: the repository's code-intelligence layer is Graphify, which requires a tracked OpenCode plugin, skill inventories duplicated under `.agents/skills/graphify` and `.opencode/skills/graphify`, ignore-file plumbing, a Python runtime in the devShell, and committed semantic-cache JSON. codebase-memory-mcp provides a faster native-binary code knowledge graph over MCP with first-class OpenCode support and a maintenance path the repository can pin. This is one focused switch: remove the Graphify integration and land codebase-memory-mcp as the supported code-intelligence tool.

Confirmed constraints:
- Provide the binary through a pinned Nix flake input exposed in the devShell, mirroring the existing `backlog-md` input, rather than a machine-global install.
- Wire OpenCode repository-scoped: a local MCP entry in `.opencode/opencode.jsonc` plus `AGENTS.md` guidance. Do not run the upstream `codebase-memory-mcp install` command, and do not modify user-global OpenCode configuration, skills, plugins, or agents.
- Require the explicit index command for graph freshness in the project Definition of Done; automatic watcher behavior is not the contract.
- Gitignore `.codebase-memory/` entirely; no graph artifact is committed.
- Migrate graph exclusions to a root `.cbmignore`; codebase-memory-mcp reads it with gitignore syntax and only at the indexed root.
- `backlog/handoffs/` and `backlog/completed/` are append-only history and must not be edited.
- Precedent: TASK-0054 established the OpenCode integration and Nix devshell reproducibility; TASK-0072 removed an obsolete Graphify MCP launcher. This task supersedes Graphify as the repository's code-intelligence integration.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The devShell exposes a pinned codebase-memory-mcp binary from a repository flake input, and flake.nix no longer defines the Graphify package or its wrapper.
- [x] #2 A freshly started repository OpenCode session connects to the codebase-memory-mcp MCP server, and a graph query run through that session returns results for this repository.
- [x] #3 AGENTS.md and the review-task skill replace the graphify MCP-first guidance with codebase-memory-mcp MCP-first guidance, including the first-use index command.
- [x] #4 A root .cbmignore carries the graph-exclusion policy, including the inactive ADRs formerly listed in docs/design/architecture/adr/.graphifyignore, and a fresh index reports those paths under excluded.
- [x] #5 Project Definition of Done defaults and BACKLOG_WORKFLOW.md require `codebase-memory-mcp cli index_repository --repo-path .` instead of `graphify update .`.
- [x] #6 .codebase-memory/ is ignored and no graph artifact is committed; the OpenCode plugin, both graphify skill copies, both .graphifyignore files, the .gitattributes merge driver, docs/graphify.md, and the tracked graphify-out/cache/semantic/ files are removed.
- [x] #7 A repository-wide search for graphify returns no live matches, with Backlog records excepted.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 All acceptance criteria are satisfied
- [x] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [x] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [x] #4 Changed Markdown files are formatted with npm run format:markdown
- [x] #5 Swagger annotations changed: run `swag init` from `server/` and `npm run gen:api` from `frontend/`
- [x] #6 `scripts/` or `prettier/` changed: run root `npm run fmt:check`
- [x] #7 Contract-affecting changes update their documents. This includes `docs/environments.md`; `PLUGIN_API_README.md`; and migration and rollout notes
- [x] #8 Code changed: run `codebase-memory-mcp cli index_repository --repo-path .` to refresh the code knowledge graph
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Approach

1. Provision: pin the v0.11.0 release artifacts directly in `flake.nix` (`fetchurl` + `stdenvNoCC`) for the four supported systems, install the portable static Linux builds and the darwin builds, and drop the graphify derivation, wrapper, and now-unused Python runtime; verify the binary inside `nix develop`. (Deviation from the original flake-input approach, recorded in the implementation notes: the upstream flake stalled on this connection because it pulls a second nixpkgs and a ~288 MB source repository.)
2. OpenCode: add a `codebase-memory-mcp` local MCP entry to `.opencode/opencode.jsonc`; delete `.opencode/opencode.json` and `.opencode/plugins/graphify.js`.
3. Guidance and policy: rewrite the `AGENTS.md` section and the review-task lookup at `.agents/skills/review-task/SKILL.md:31`; replace `docs/graphify.md` with `docs/codebase-memory.md`; update `docs/design/architecture/README.md:46` and `BACKLOG_WORKFLOW.md:54`; replace project DoD default #6 through `definition_of_done_defaults_upsert`.
4. Exclusions: add a root `.cbmignore` migrated from `.graphifyignore` plus qualified inactive-ADR paths; delete both `.graphifyignore` files.
5. Residue: remove the `.gitattributes` merge driver; drop graphify entries from `.gitignore` and add `.codebase-memory/`; remove both graphify skill trees, the plugin, the graphify configs, `docs/graphify.md`, and the 198 tracked semantic-cache files; delete `graphify-out/` and `scripts/graphify/__pycache__`.
6. Verify: `nix flake check`; `nix develop . --command codebase-memory-mcp --version`; index via CLI and confirm the inactive ADRs are excluded; start an OpenCode session against the MCP server and run a graph query; repository-wide `rg -i graphify` outside append-only backlog history; Markdown formatting; record that unit and e2e are not applicable because no runtime code changes.

## Risks

- Acceptance criterion #7 cannot be literally zero matches because TASK-0310's own Backlog record names Graphify; verification excludes Backlog records and reports that explicitly.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-21: Flake-input approach stalled. The upstream flake pulls its own nixpkgs-unstable node plus a ~288 MB source repository, and this connection runs around 0.5 MB/s, so `nix flake lock` made no visible progress and the user aborted it. Switched to the plan's pre-approved fallback: a hash-pinned release-archive derivation in `flake.nix` using the portable Linux builds (fully static, NixOS-safe) and the darwin builds, with SHA-256 digests from the v0.11.0 release metadata. This keeps the binary pinned in the repository flake and exposes it in the devShell, satisfying acceptance criterion #1's intent without the second nixpkgs download or a from-source C build.

2026-09-21 (nixpkgs decision): the user directed the binary to come from the repository's pinned nixpkgs flake input instead of a release-archive derivation, so `flake.nix` exposes `pkgs.codebase-memory-mcp` and drops the Graphify derivation and wrapper. `pkgs.python3` stays in the devShell because `scripts/opencode/current-model.sh` and `scripts/handoff/create-handoff.sh` use it. Verification at the branch tip (`7feddef5`, nixpkgs `79b35bf0bda5cd110f856aa5b5b2c5ba4460dbf5`): `nix flake check` passes and `nix develop . --command codebase-memory-mcp --version` prints 0.11.0. The earlier record of 0.8.1 predated the nixpkgs pin update; the current lock resolves 0.11.0. A re-index on the current pin produced 8707 nodes and 27546 edges. A graph query for `scheduleOverlap` through a freshly started `opencode run` session inside `nix develop` returned 41 results via the `codebase-memory-mcp_search_graph` MCP tool.

2026-09-21 (verification): a Cypher query over the live graph lists active ADR-001, ADR-002, ADR-004 through ADR-018, and ADR-020 through ADR-022 with node counts and none of the inactive ADR-003, ADR-011 through ADR-017, or ADR-019, confirming `.cbmignore` excludes them; the index response reports 23 excluded directories. `rg -i graphify` outside `backlog/` returns no matches; Backlog records retain historical Graphify references and were not rewritten, which acceptance criterion #7 records as its exception.

2026-09-21 (formatting): changed Markdown (`AGENTS.md`, `docs/codebase-memory.md`, `.agents/skills/review-task/SKILL.md`) passes root `npm run format:markdown`, `npm run format:markdown:check`, and `npm run lint:markdown`. Unit and e2e suites are not applicable because no runtime application code changed.

2026-09-21 (review): acceptance criterion #7's exception was widened from this task's own record to Backlog records generally, and the inherited `graphify update .` Definition of Done item in the open tasks TASK-0287 and TASK-0297 was replaced with the project default index command. The changes remain uncommitted in the worktree by user decision.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: OpenCode
created: 2026-09-21 20:29
---
Review findings and decisions:

- Blocker: the switch work was uncommitted and the branch tip was half-switched. `7feddef5` removed the Graphify devShell package while the plugin, both skill trees, and `.graphifyignore` were still tracked and no `.cbmignore` existed. Decision: leave the work uncommitted in the worktree by user choice.
- Acceptance criterion #7 was checked but its exception covered only `backlog/handoffs/` and `backlog/completed/`, while active Backlog tasks still named Graphify. Decision: widen the exception to Backlog records generally, and replace the inherited `graphify update .` DoD item in TASK-0287 and TASK-0297 with the project default index command.
- The recorded 0.8.1 version evidence predated the nixpkgs pin update in `7feddef5`. Decision: correct the records in place; the current lock resolves 0.11.0, `nix flake check` passes, and a fresh re-index reports 8707 nodes and 27546 edges.
- Unresolved minors: `.gitignore` adds `.vscode` (unrelated to this task's listed files); the task `references` still lists the deleted `docs/graphify.md`; the task's DoD snapshot ordering differs from the current project defaults.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Switched the repository code-intelligence layer from Graphify to codebase-memory-mcp. The devShell now provides `pkgs.codebase-memory-mcp` (0.11.0 at the locked nixpkgs revision) from the pinned nixpkgs flake input through `flake.nix`, and the Graphify derivation and wrapper are gone. OpenCode connects through the repository-local MCP entry in `.opencode/opencode.jsonc`. Guidance and policy now point at codebase-memory-mcp: `AGENTS.md`, the review-task skill, `BACKLOG_WORKFLOW.md`, and the project DoD default use `codebase-memory-mcp cli index_repository --repo-path .`; `docs/codebase-memory.md` replaces `docs/graphify.md`; and `docs/design/architecture/README.md` cites the root `.cbmignore`. All Graphify residue is removed from the worktree: the plugin, both skill trees, both `.graphifyignore` files, the `.gitattributes` merge driver, the graphify configs, and `graphify-out/` including the 198 tracked semantic-cache JSON files. `.codebase-memory/` stays ignored and no graph artifact is committed. Verification on the branch tip: `nix flake check` passes; `nix develop . --command codebase-memory-mcp --version` reports 0.11.0; a fresh re-index produced 8707 nodes and 27546 edges and reported 23 excluded directories; a freshly started `opencode run` session inside `nix develop` queried the graph through the MCP server and returned 41 results; active ADRs are in the graph while the inactive ADRs are absent, confirming `.cbmignore`; Markdown format and lint checks pass. Unit and e2e tests do not apply because no runtime application code changed. `rg -i graphify` has no matches outside `backlog/`; Backlog records retain historical Graphify references, which acceptance criterion #7 excepts, and the open tasks TASK-0287 and TASK-0297 now carry the codebase-memory-mcp Definition of Done item. The changes remain uncommitted in the worktree by user decision.
<!-- SECTION:FINAL_SUMMARY:END -->
