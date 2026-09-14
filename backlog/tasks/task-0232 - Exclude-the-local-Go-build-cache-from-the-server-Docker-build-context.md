---
id: TASK-0232
title: Exclude the local Go build cache from the server Docker build context
status: To Do
assignee: []
created_date: '2026-09-14 12:28'
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
- [ ] #1 With local Go cache directories (for example `server/.gocache` and `server/.go-cache`) populated on disk, `docker compose --env-file .env.development -f compose.yaml build server` succeeds and the reported build-context transfer size excludes those directories (source-only, on the order of the server sources rather than hundreds of megabytes).
- [ ] #2 `server/.dockerignore` excludes every local Go build cache directory that `server/.gitignore` ignores, so a future cache path or tool rename cannot silently reintroduce the cache in the build context.
- [ ] #3 `server` and `postgres-migrate` images build from the same Dockerfile stages, and base `compose.yaml` plus the staging, production, and test overlays resolve unchanged.
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
- [ ] #4 Changed Markdown files are formatted with npm run format:markdown
<!-- DOD:END -->
