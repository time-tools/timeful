# codebase-memory-mcp Policy

`codebase-memory-mcp` serves this repository's code knowledge graph over MCP.
The devShell provides the `codebase-memory-mcp` binary from the pinned nixpkgs flake input, and OpenCode connects to it through the local MCP entry in `.opencode/opencode.jsonc`.
This document records the artifact and exclusion policy; `.cbmignore` is the machine-readable exclude list.

## Indexing

Refresh the graph after code changes with `codebase-memory-mcp cli index_repository --repo-path .`.
The project Definition of Done requires that command for code changes.
The first index on a machine is a full index; later runs are incremental.

## Untracked Index State

The live index lives outside the repository under the tool's user cache, and the optional shared artifact under `.codebase-memory/` stays untracked.
Do not commit `.codebase-memory/`.
Collaborators index locally instead of sharing a graph artifact, so no binary graph enters repository history.

## Excludes

`.cbmignore` at the repository root controls which files the indexer sees.
It uses gitignore syntax, and codebase-memory-mcp reads it only from the indexed root.
A path matched by `.cbmignore` never enters the graph, and a change to the file takes effect on the next re-index.
Keep the file current when the repository layout changes, and do not rely on nested ignore files, because the tool does not read them.

## Verification

A re-index reports skipped paths under `excluded` in the `index_repository` response.
