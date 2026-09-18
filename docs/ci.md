# Continuous Integration

GitHub-hosted CI runs as GitHub Actions workflows defined in `.github/workflows/`.
Each workflow validates one area of the repository, and Dependabot opens weekly pull requests to keep actions and npm dependencies up to date.

## Workflows

| Workflow               | File                         | Purpose                                                                                                                                                                                                        |
| ---------------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Markdown CI            | `markdown-ci.yml`            | Checks Markdown formatting and linting, and lints all workflow files with `actionlint`                                                                                                                         |
| Frontend CI            | `frontend-ci.yml`            | Lints, type-checks, unit-tests, and builds the frontend                                                                                                                                                        |
| Backend CI             | `backend-ci.yml`             | Runs the Go server tests against an isolated PostgreSQL Compose stack                                                                                                                                          |
| E2E CI                 | `e2e-ci.yml`                 | Runs the browser E2E suites as parallel per-suite matrix jobs against isolated Compose test stacks, covering the Chromium desktop, mobile, and production projects plus the Firefox desktop and touch projects |
| Backlog Weekly Cleanup | `backlog-weekly-cleanup.yml` | Moves tasks in the terminal Done status to `backlog/completed/` weekly and opens or refreshes the cleanup pull request                                                                                         |

Markdown CI triggers on Markdown changes and on changes to any file matching `.github/workflows/*.yml`, so every workflow edit is validated in CI.
Dependabot is configured in `.github/dependabot.yml` with weekly updates for the `github-actions` ecosystem and the npm ecosystems for the root, frontend, and e2e directories.

## Browser E2E parallelism

E2E CI runs the browser suites as three independent `browser-e2e` matrix jobs, one per suite, each on its own GitHub-hosted runner with its own isolated `postgres-test` and `server-test` Compose stack on the fixed isolated ports.
Because the jobs are independent, wall-clock time approaches job setup plus the slowest single suite instead of the sum of all suites.
Parallel jobs duplicate setup compute on purpose, which is the accepted trade-off for the lower wall-clock time.

- The Chromium job runs `chromium-desktop`, `chromium-mobile`, `chromium-production-desktop`, and `chromium-production-mobile` at two Playwright workers.
- The Firefox desktop job runs `firefox-desktop` at two Playwright workers and sets `E2E_FRONTEND=bundled` so the recorded access-transfer journeys stay within budget.
- The Firefox touch job runs `firefox-touch` at one worker because it matches a single serial spec file that cannot parallelize further.

Every matrix job restores the Nix, Go, npm, and migrator-image caches; only the Chromium job saves them, so concurrent same-key saves cannot race.
Each job uploads Playwright failure artifacts under `playwright-failure-artifacts-<suite>` so parallel uploads do not collide.

## Validating workflow files locally

`actionlint` validates workflow syntax, shell snippets in `run:` steps, and common workflow mistakes before CI does.
The Nix development shell defined in `flake.nix` provides `actionlint`, so it is available on `PATH` after entering the shell with `nix develop`.
Run `actionlint` from the repository root; with no arguments it checks every workflow file in `.github/workflows/`.
Run it before pushing any workflow change, because Markdown CI runs the same check on GitHub-hosted runners.

## Action reference pinning

Every `uses:` reference to a GitHub-hosted action is pinned to a full-length commit SHA with a same-line version comment, such as `actions/checkout@<full-length commit SHA> # v7.0.1`.
Full-length SHA pinning prevents tag-move supply-chain attacks because a moved tag cannot change what a pinned workflow executes.
Docker image references in `uses:` are pinned by image digest with a same-line comment recording the version tag, such as the `actionlint` step in `markdown-ci.yml`.
Dependabot updates pinned SHA references and their version comments together in its weekly `github-actions` pull requests.
When adding a new action reference, resolve the SHA from the official action repository with `git ls-remote`, and never guess it or copy it from a fork.

## Backlog weekly cleanup

The `Backlog Weekly Cleanup` workflow (`backlog-weekly-cleanup.yml`) runs every Sunday at 00:00 UTC and on manual dispatch.
It moves every task in the terminal Done status from `backlog/tasks/` to `backlog/completed/` with `backlog task complete`, commits the moves on the workflow-owned `chore/backlog-weekly-completed` branch, and opens or refreshes one pull request against `main`.
A run with no Done task commits nothing and opens no pull request.
This automation performs the periodic cleanup that `BACKLOG_WORKFLOW.md` keeps separate from task finalization and the completing pull request.

`GH_TOKEN` prefers the repository `BACKLOG_CLEANUP_TOKEN` secret and falls back to `github.token`.
Pull requests opened with the repository `GITHUB_TOKEN` do not start pull-request workflows, so the secret exists to get CI on generated cleanup pull requests when it is configured.

### Enabling pull request creation

Workflows using the repository `GITHUB_TOKEN` cannot create pull requests until the "Allow GitHub Actions to create and approve pull requests" setting is enabled.
The cleanup workflow needs this setting for its `github.token` fallback, because the fallback cannot open the cleanup pull request without it.
The `BACKLOG_CLEANUP_TOKEN` path is not gated by the setting, because a personal access token acts as its user rather than as the repository `GITHUB_TOKEN`.

1. As an organization owner, open the organization's **Settings**.
2. In the sidebar, select **Actions**, then **General**.
3. Under "Workflow permissions", enable **Allow GitHub Actions to create and approve pull requests**.
4. Select **Save**.
5. In the repository, open **Settings**.
6. In the sidebar, select **Actions**, then **General**.
7. Under "Workflow permissions", enable **Allow GitHub Actions to create and approve pull requests**.
8. Select **Save**.

If an organization or enterprise policy leaves the setting disabled, configure the token below so the workflow does not depend on the fallback.

### Backlog cleanup token

`BACKLOG_CLEANUP_TOKEN` is a fine-grained personal access token that lets the workflow push its branch and open pull requests, and that makes pull-request workflows run on the generated cleanup pull requests.
Store it as an Actions secret named exactly `BACKLOG_CLEANUP_TOKEN`, either as a repository secret or as an organization secret available to the repository.

Create the token:

1. Open your profile picture, then select **Settings**.
2. Select **Developer settings**.
3. Select **Personal access tokens**, then **Fine-grained tokens**.
4. Select **Generate new token**.
5. Enter a token name and description, select the resource owner, and set an expiration.
6. Under "Repository access", select **Only select repositories** and select `timeful`.
7. Add the repository permission **Contents** with **Read and write**, because the workflow creates and force-updates `chore/backlog-weekly-completed`.
8. Add the repository permission **Pull requests** with **Read and write**, because the workflow lists, creates, and edits the cleanup pull request; **Metadata** with **Read-only** is added automatically.
9. Select **Generate token** and copy the value.

Store and rotate the secret:

1. In the repository, open **Settings**.
2. In the sidebar, select **Secrets and variables**, then **Actions**.
3. Select **New repository secret**.
4. Enter `BACKLOG_CLEANUP_TOKEN` as the name and paste the token as the value.
5. Select **Add secret**.
6. Replace the secret before the token expires, because an expired token fails the workflow steps that push the branch and open or refresh the pull request.

## Actions secrets

This repository's workflows read these Actions secrets; configure them in the repository's **Settings**, **Secrets and variables**, then **Actions**.

| Secret                  | Used by                      | Permissions                                                                                                | Purpose                                                                                                                                                                                                         |
| ----------------------- | ---------------------------- | ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BACKLOG_CLEANUP_TOKEN` | `backlog-weekly-cleanup.yml` | Fine-grained personal access token with **Contents: Read and write** and **Pull requests: Read and write** | Pushes `chore/backlog-weekly-completed` and opens or refreshes the cleanup pull request, and lets pull-request workflows run on it; the workflow falls back to `github.token` when the secret is not configured |

## Recommended repository settings

Enable "Require actions to be pinned to full-length commit SHA" under Settings > Actions > General in the GitHub repository settings.
This setting is the enforcement backstop for the pinning policy above: GitHub rejects any workflow change that references an action by mutable tag, even when a review misses it.
The Dependabot configuration already opens pull requests with SHA-pinned references, so enabling the setting does not block dependency updates.
