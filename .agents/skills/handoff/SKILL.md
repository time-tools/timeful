---
name: handoff
description: Session handoff, handoff notes, or handoff file: use when the user explicitly asks to capture durable context from the current session for later work.
---

# Session Handoff

Create a handoff only when the user explicitly requests one. Do not create one automatically at task completion.

<!--
Content guidance adapted from Matt Pocock's handoff skill:
https://github.com/mattpocock/skills/blob/main/skills/productivity/handoff/SKILL.md
Licensed under MIT.
-->

## Create The Handoff

1. Run the repository-owned creator immediately before completing the handoff:

   ```sh
   scripts/handoff/create-handoff.sh
   ```

   The script prints the created `backlog/handoffs/handoff-<UTC timestamp>.md` path, captures the most recently updated OpenCode session, and writes the standard template. The OpenCode CLI has no explicit current-session command or environment variable. If another session in the same directory may have updated concurrently, stop and ask the user to confirm the session ID before running the script.

2. Fill the created template with facts established during the current session, removing empty sections when appropriate. Do not overwrite or edit older archive entries.
3. Report the resulting path.

## Content

Capture only facts established during the current session. The handoff must be useful without the chat transcript and must not invent completion, verification, risks, or decisions.

Do not duplicate material already captured in a specification, plan, ADR, issue, commit, diff, or test. Reference the artifact by repository path or URL and record only the current status or implication. Redact secrets and personal data; never include credentials, tokens, or other sensitive values. The repository script owns the standard section template.

## Section Guidance

The script's section headings are canonical; keep the list below in sync with `scripts/handoff/create-handoff.sh`.

- **Session Objective**: State the original problem, intended outcome, and its current state: completed, in progress, blocked, or not started. Describe the user's goal rather than only the implementation task.
- **Motivation And Impact**: Explain why the work is being done and the practical consequence of success or inaction. Include affected users, workflows, reliability, performance, or maintenance concerns only when established.
- **Current State**: Describe the relevant state at handoff time, especially for partial work: implemented but unvalidated, failing test behavior, active worktree changes, deployed state, or known runtime behavior. Do not duplicate **Completed Work**.
- **Completed Work**: List concrete work completed during the session. Reference the implementation, test, specification, issue, commit, or other durable artifact rather than duplicating its detail.
- **Important Decisions**: Record decisions that govern future work, including the selected approach and its consequence. Include only decisions made during the session or confirmed by the user.
- **Decision Rationale**: Explain why an important decision was made. Include rejected alternatives only when that context prevents a likely reversal or repeated investigation. Reference an ADR, issue, or plan instead of restating it.
- **Constraints And Assumptions**: Record scope boundaries, compatibility requirements, operational limitations, and assumptions that still need validation. Clearly distinguish confirmed constraints from assumptions.
- **Findings And Risks**: Capture discoveries that may affect correctness, delivery, or future investigation, plus concrete remaining risks. Avoid speculative risks without an actionable basis.
- **Dependencies And Prerequisites**: List external services, configuration, migrations, approvals, data state, related branches or issues, and ordering requirements needed to proceed. Redact sensitive values.
- **Validation**: Include commands actually run, their outcomes, and meaningful checks not run. Include exact commands when they materially establish confidence in the work.
- **Remaining Work**: Describe uncompleted work as concrete, actionable tasks. Include blockers and prerequisites. Tailor it to the user's stated next-session focus when one is provided, and do not repeat optional work already covered by a tracked issue without linking it.
- **Next Recommended Action**: State the single most useful first action for the next session, particularly when sequencing matters. Omit when **Remaining Work** is already unambiguous.
- **Handoff Relationships**: Describe how this handoff relates to earlier handoffs or durable planning artifacts. Use `Continues` for still-active prior work and `Supersedes` only when prior conclusions or next steps are no longer current. Omit when there is no meaningful relationship.
- **Suggested Skills**: Name only repository skills that are known to exist and directly useful for the remaining work.
- **Relevant Files**: Include changed or inspected paths that a future session needs to understand or continue the work. Add a brief reason when the path alone is not self-explanatory.

Keep entries concise, factual, and actionable.
