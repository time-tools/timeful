---
id: TASK-0232
title: Exclude the local Go build cache from the server Docker build context
status: Done
assignee:
  - opencode
created_date: '2026-09-14 12:28'
updated_date: '2026-09-14 12:59'
labels:
  - docker
  - build
dependencies: []
references:
  - server/.dockerignore
  - server/.gitignore
  - server/Dockerfile
  - compose.yaml
priority: low
type: chore
ordinal: 234000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`server/.dockerignore` excludes `.go-cache/`, but the local Go build cache on this development machine lives at `server/.gocache` (no hyphen), which `server/.gitignore` also lists as ignored. `server/Dockerfile` runs `COPY . .` in the builder stage, so a server image build transfers the whole ~317 MB `.gocache` directory and copies it into the builder layer. The runtime image is unaffected because it copies only the compiled binary from the builder stage; the observed cost is a ~315 MB build-context transfer and a bloated cached layer.

Evidence (2026-09-14 during TASK-0231 verification): a `docker compose ... build server` run reported `transferring context: 315.09MB`, while the same context excluding `.gocache` was under 500 KB. `server/.gitignore` already ignores both `.gocache` and `.go-cache`; only `.dockerignore` disagrees.

Outcome: server image builds never include gitignored local Go build cache directories, the reported build-context transfer for the `server` service is source-only size, and server and migrator builds keep working unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 With local Go cache directories (for example `server/.gocache` and `server/.go-cache`) populated on disk, `docker compose --env-file .env.development -f compose.yaml build server` succeeds and the reported build-context transfer size excludes those directories (source-only, on the order of the server sources rather than hundreds of megabytes).
- [x] #2 `server/.dockerignore` excludes every local Go build cache directory that `server/.gitignore` ignores, so a future cache path or tool rename cannot silently reintroduce the cache in the build context.
- [x] #3 `server` and `postgres-migrate` images build from the same Dockerfile stages, and base `compose.yaml` plus the staging, production, and test overlays resolve unchanged.
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
## Plan

1. Confirm the divergence: `server/.gitignore` ignores `.gocache` and `.go-cache`, while `server/.dockerignore` only excludes `.go-cache/`.
2. Edit `server/.dockerignore` so it excludes both local Go build cache directories, with a comment tying the list to the Git-ignored cache paths. No Dockerfile or Compose changes are needed because the builder stage's `COPY . .` is the transfer source.
3. Verify AC#1: populate `server/.go-cache` with dummy bytes alongside the existing ~317 MB `server/.gocache`, run `docker compose --env-file .env.development -f compose.yaml build server`, and confirm the build succeeds with a source-only `transferring context` size. Move the dummy `.go-cache` aside afterward.
4. Verify AC#2: compare the cache entries in `server/.gitignore` and `server/.dockerignore` for set parity.
5. Verify AC#3: run `docker compose config --quiet` for base, development, staging, production, and test overlay combinations, and build both the `server` and `postgres-migrate` services to confirm the Dockerfile stages still work.
6. Record per-criterion evidence in implementation notes and finalize. No Markdown changes are expected, so markdown formatting checks are not applicable.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation (2026-09-14): added a `# Local Go build caches (mirror the cache paths ignored in .gitignore)` block to `server/.dockerignore` listing `.gocache/` and `.go-cache/`; `.gocache/` was the missing entry. No Dockerfile or Compose changes were needed.

AC#1 evidence: with `server/.gocache` at 317M and `server/.go-cache` populated with 26M of dummy data, `docker compose --env-file .env.development -f compose.yaml build server` exited 0 and reported `transferring context: 8.05kB` (full log: /tmp/opencode/task-0232-build-server.log). A scratch probe Dockerfile (`COPY . /probe`, same context and real `.dockerignore`) printed `CACHE_DIRS_ABSENT` and `1.6M /probe` (log: /tmp/opencode/task-0232-context-probe.log), directly proving both cache directories are absent from the copied context. A measured filtered source tar is ~0.9MB actual bytes, versus ~343MB for the two cache directories combined. BuildKit reports only changed blobs on a warm builder, so the reported number is small even for real sources; the probe confirms the filtered copy.

AC#2 evidence: set-parity check shows every `server/.gitignore` Go cache entry (`.gocache`, `.go-cache`) is covered by `server/.dockerignore` (`.gocache/`, `.go-cache/`). The comment instructs mirroring `.gitignore` cache paths so a future rename updates both files together.

AC#3 evidence: `server/Dockerfile` is untouched; `postgres-migrate` (migrator stage) and `server` (full builder chain) both built successfully in this session; `docker compose config --quiet` exits 0 for base, development, test, staging, and production. Staging/production used `.env.staging.example` and `.env.production.example` copied to /tmp/opencode with blank required values substituted only for interpolation during `config` resolution.

Cleanup: the dummy `server/.go-cache` was moved aside to /tmp/opencode/task-0232-dummy-go-cache; it did not exist before this task. The pre-existing `server/.gocache` was left in place. No unit or e2e tests apply because no runtime code path changed, and no Markdown files changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the missing `.gocache/` entry to `server/.dockerignore` and grouped both local Go build cache paths under a comment that ties them to the cache paths ignored in `server/.gitignore`. Before this change the builder stage's `COPY . .` sent the whole ~317 MB `server/.gocache` directory through the build context; now `docker compose --env-file .env.development -f compose.yaml build server` succeeds and reports `transferring context: 8.05kB` with `server/.gocache` (317M) and a populated `server/.go-cache` (26M dummy data) on disk. A scratch `COPY . /probe` build using the real `.dockerignore` printed `CACHE_DIRS_ABSENT` and a source-only copied tree of 1.6M. `postgres-migrate` and `server` both build from the unchanged Dockerfile stages, and base, development, test, staging, and production Compose configurations resolve unchanged with `docker compose config --quiet`. No runtime code, Markdown, unit, or e2e changes apply; the only modified file is `server/.dockerignore`.
<!-- SECTION:FINAL_SUMMARY:END -->
