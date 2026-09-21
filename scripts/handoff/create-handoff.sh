#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
archive_dir="${HANDOFF_ARCHIVE_DIR:-$repo_root/backlog/handoffs}"

session_id="$(
  opencode session list --format json --max-count 1 | python3 -c '
import json
import sys

sessions = json.load(sys.stdin)
if len(sessions) != 1:
    raise SystemExit("Could not determine the most recently updated OpenCode session.")

session_id = sessions[0].get("id")
if not isinstance(session_id, str) or not session_id.startswith("ses_"):
    raise SystemExit("Could not determine the most recently updated OpenCode session.")

print(session_id)
'
)"
created_at="$(date -u +%Y-%m-%dT%H-%M-%SZ)"
handoff_path="$archive_dir/handoff-$created_at.md"

mkdir -p "$archive_dir"
temporary_path="$(mktemp "$archive_dir/.handoff.XXXXXX")"
trap 'rm -f "$temporary_path"' EXIT

# The section headings below are canonical for the handoff template.
# Keep them in sync with `.agents/skills/handoff/SKILL.md` (Section Guidance).
{
  printf '%s\n' \
    '# Session Handoff' \
    '' \
    "- Created session: \`$session_id\`" \
    '' \
    '## Session Objective' \
    '' \
    '## Motivation And Impact' \
    '' \
    '## Current State' \
    '' \
    '## Completed Work' \
    '' \
    '## Important Decisions' \
    '' \
    '## Decision Rationale' \
    '' \
    '## Constraints And Assumptions' \
    '' \
    '## Findings And Risks' \
    '' \
    '## Dependencies And Prerequisites' \
    '' \
    '## Validation' \
    '' \
    '## Remaining Work' \
    '' \
    '## Next Recommended Action' \
    '' \
    '## Handoff Relationships' \
    '' \
    '## Suggested Skills' \
    '' \
    '## Relevant Files'
} >"$temporary_path"
chmod 644 "$temporary_path"

# A hard link atomically reserves the destination without replacing an archive entry.
if ! ln "$temporary_path" "$handoff_path"; then
  printf 'Handoff already exists: %s\n' "$handoff_path" >&2
  exit 1
fi

printf '%s\n' "$handoff_path"
