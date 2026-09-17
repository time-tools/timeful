---
id: TASK-0254
title: Enable cross-branch task ID collision checks in the Backlog configuration
status: Done
assignee:
  - '@OpenCode'
created_date: '2026-09-17 15:20'
updated_date: '2026-09-17 15:24'
labels:
  - backlog-md
  - tooling
dependencies: []
references:
  - backlog/config.yml
  - 'https://github.com/time-tools/timeful/pull/50'
documentation:
  - 'https://github.com/MrLesk/Backlog.md/blob/main/ADVANCED-CONFIG.md'
modified_files:
  - backlog/config.yml
priority: medium
type: chore
ordinal: 254000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog reads task IDs only from the current branch while remote_operations and check_active_branches are disabled, so agents working on parallel branches each allocate the same next ID and duplicate IDs land in the repository; the TASK-0238 collision is existing evidence. Enabling remote operations and active-branch checking, with the active window extended to 365 days, makes ID allocation consider task files on local and origin branches whose tip commit is within the last year, so newly created tasks cannot silently reuse an ID that another active branch already committed. The change is limited to backlog/config.yml; existing duplicates are out of scope.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 backlog/config.yml sets remote_operations: true, check_active_branches: true, and active_branch_days: 365, and backlog config get remoteOperations, checkActiveBranches, and activeBranchDays report true, true, and 365
- [x] #2 New task IDs are allocated against task files on local and origin branches whose tip commit is within the last 365 days, so a task created on one branch does not reuse an ID already present on another active branch
- [x] #3 The existing duplicate TASK-0238 ID collision is neither silently resolved nor deleted, and the task notes record its state with cross-branch checks enabled
- [x] #4 Verification evidence is recorded in the task notes, including the cross-branch task state Backlog reports and a probe or listing confirming no new ID collision
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
1. Confirm the effective configuration: read backlog/config.yml and `backlog config get remoteOperations`, `checkActiveBranches`, and `activeBranchDays`, expecting true, true, and 365.
2. Verify cross-branch awareness and the existing TASK-0238 collision: enumerate task IDs per ref whose tip commit falls within 365 days, locate the colliding TASK-0238 files (`backlog/completed/task-0238 - Add-the-OpenObserve...` on main/update-backlog/set-up-openobserve/origin/main vs `backlog/tasks/task-0238 - Replace-the-vendored-undo-icon...` on optimize-frontend-size), and record that the collision is neither resolved nor deleted.
3. Probe cross-branch ID allocation: create a temporary branch with a committed task file carrying ID 0300 (above the current local max 0253), confirm Backlog can resolve the cross-branch task, create a probe task on the working branch, and confirm it receives 0301 rather than 0255; then remove the probe task and the temporary branch.
4. Record all evidence in the task notes, check acceptance criteria and Definition of Done, write the final summary, and move the task to Done.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-09-17 configuration evidence: backlog/config.yml lines 10, 15, and 16 set remote_operations: true, check_active_branches: true, and active_branch_days: 365; `backlog config get remoteOperations|checkActiveBranches|activeBranchDays` returns true, true, 365. Active refs whose tip commit falls within 365 days (git for-each-ref plus `git log -1 --format=%ct` per ref): 30 refs including refs/heads and refs/remotes/origin/*; the highest task ID across them is TASK-0253 on update-backlog (main and origin/main TASK-0252, set-up-openobserve TASK-0252, optimize-frontend-size TASK-0241).

2026-09-17 isolation probe: created temporary worktree branch tmp/task-0254-probe at HEAD 4f9022c1 with committed marker `backlog/tasks/task-0300 - Probe cross-branch task ID marker.md`, then `backlog task create "Probe cross-branch ID allocation"` allocated TASK-0301; local-only allocation would have produced TASK-0255 because the working-copy maximum is TASK-0254, so the allocator read the ID occupied on the temporary branch. Cleanup: `git worktree remove --force`, `git branch -D tmp/task-0254-probe`, probe task file moved aside to /tmp/opencode/probe-task-0301-backup.md, and `backlog task list -s "To Do" --plain` no longer contains TASK-0300 or TASK-0301. Remote-tracking refs participate as well: `backlog doctor` reports collisions against origin/main entries while remote_operations is enabled.

2026-09-17 TASK-0238 collision state: `backlog doctor` (read-only, no --fix) reports `Possible cross-branch ID collisions (diagnostic only)` and lists TASK-238 with main, origin/main, set-up-openobserve, and update-backlog carrying the completed `backlog/completed/task-0238 - Add-the-OpenObserve-observability-ADR-and-quality-requirements.md`, while optimize-frontend-size carries the active `backlog/tasks/task-0238 - Replace-the-vendored-undo-icon-with-the-project-icon-sets-backup-restore-icon-in-the-timezone-selector.md`. The report states Backlog.md will not edit another branch, and enabling the checks created, renamed, or deleted nothing; the local completed file is untouched (`git status --short -- backlog/completed/task-0238*` is empty). The same report shows the pre-existing TASK-178 and TASK-188 collisions, so the existing duplicates are surfaced, not silently resolved.

2026-09-17 checks: the change is configuration-only with no runtime code, so no unit or e2e test covers it and none is required; `npm run format:markdown:check` exits 0, and the root Markdown formatter excludes backlog/**, so the changed task record is outside the formatted set.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Enabled cross-branch task ID collision checks in backlog/config.yml so parallel-branch task creation cannot reuse an ID.

- backlog/config.yml now sets remote_operations: true, check_active_branches: true, and active_branch_days: 365 (previously false, false, 30); `backlog config get remoteOperations`, `checkActiveBranches`, and `activeBranchDays` return true, true, 365.
- Probe on a temporary branch (tmp/task-0254-probe, marker TASK-0300) made a new probe task allocate TASK-0301 instead of the local-only TASK-0255, proving allocation reads IDs occupied on active branches; the marker branch/worktree and probe task were removed afterward. Remote-tracking refs participate too, as `backlog doctor` reports collisions against origin/main.
- `backlog doctor` now surfaces the existing cross-branch collisions (TASK-238, TASK-178, TASK-188) as diagnostic-only and never edits other branches; the TASK-0238 duplicate remains untouched on its branches.
- Configuration-only change: no unit or e2e tests apply and none are required; `npm run format:markdown:check` exits 0. Modified file: backlog/config.yml.

Follow-up: the existing cross-branch duplicate IDs (TASK-238, TASK-178, TASK-188) still need manual reconciliation on their branches.
<!-- SECTION:FINAL_SUMMARY:END -->
