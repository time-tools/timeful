---
id: TASK-0057
title: Migrate JavaScript tooling to Bun
status: To Do
assignee: []
created_date: '2026-08-24 15:32'
updated_date: '2026-09-13 15:22'
labels:
  - tooling
  - bun
dependencies: []
references:
  - 'https://github.com/oven-sh/bun/blob/main/flake.nix'
modified_files:
  - flake.nix
  - frontend/package.json
  - frontend/bun.lock
  - frontend/package-lock.json
  - .opencode/bun.lock
  - .opencode/.gitignore
  - frontend/Dockerfile
  - .github/workflows/frontend-ci.yml
  - .github/workflows/backend-ci.yml
  - frontend/config/tooling.ts
  - AGENTS.md
  - frontend/AGENTS.md
  - frontend/README.md
  - e2e/inspect/AGENTS.md
  - server/routes/README.md
  - docs/environments.md
priority: medium
type: chore
ordinal: 59000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Standardize active JavaScript dependency installation and script execution on Bun so local development, OpenCode plugins, container builds, CI, and contributor instructions use one reproducible toolchain.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Bun is the JavaScript toolchain in the project devshell
- [ ] #2 Frontend and OpenCode dependencies have tracked Bun lockfiles
- [ ] #3 Active frontend tooling and container builds install and run scripts with Bun
- [ ] #4 CI installs and runs frontend tooling with Bun
- [ ] #5 Active contributor and operational guidance uses Bun commands
- [ ] #6 Frontend quality checks and isolated browser E2E pass with Bun
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 All acceptance criteria are satisfied
- [ ] #2 All required unit tests pass. Documentation-only changes are exempt unless the user requests unit tests
- [ ] #3 All required e2e tests pass. Documentation-only changes are exempt unless the user requests e2e tests
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: opencode
created: 2026-09-13 15:22
---
2026-09-13: updated paths for the root e2e package move (TASK-0164); the inspect guide now lives at e2e/inspect/AGENTS.md.
---
<!-- COMMENTS:END -->
