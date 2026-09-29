#!/usr/bin/env bash
set -euo pipefail

data_dir="${OPENCODE_DATA_DIR:-$HOME/.local/share/opencode}"
database_path="$data_dir/opencode.db"

session_id="$(
  opencode session list --format json --max-count 1 | python3 -c '
import json
import sys

payload = sys.stdin.read().strip()
if not payload:
    raise SystemExit(
        "No OpenCode session found for the current directory, because"
        " `opencode session list` scopes to it. Start opencode here and retry."
    )

try:
    sessions = json.loads(payload)
except json.JSONDecodeError as error:
    raise SystemExit(f"Unexpected output from `opencode session list`: {error}")

if not isinstance(sessions, list) or len(sessions) != 1:
    raise SystemExit("Could not determine the most recently updated OpenCode session.")

session_id = sessions[0].get("id") if isinstance(sessions[0], dict) else None
if not isinstance(session_id, str) or not session_id.startswith("ses_"):
    raise SystemExit("Could not determine the most recently updated OpenCode session.")

print(session_id)
'
)"

if [ ! -f "$database_path" ]; then
  printf 'OpenCode session database not found: %s\n' "$database_path" >&2
  exit 1
fi

current_model="$(
  python3 - "$session_id" "$database_path" <<'PYEOF'
import json
import sqlite3
import sys

session_id = sys.argv[1]
database_path = sys.argv[2]

try:
    connection = sqlite3.connect(f"file:{database_path}?mode=ro", uri=True)
except sqlite3.Error as error:
    raise SystemExit(f"Cannot open session database {database_path}: {error}")

row = connection.execute(
    "SELECT data FROM message"
    " WHERE session_id = ? AND json_extract(data, '$.role') = 'assistant'"
    " ORDER BY rowid DESC LIMIT 1",
    (session_id,),
).fetchone()
connection.close()

if row is None:
    raise SystemExit(f"No assistant message found for session {session_id} yet.")

message = json.loads(row[0])
provider_id = message.get("providerID")
model_id = message.get("modelID")
if not provider_id or not model_id:
    raise SystemExit("Latest assistant message has no provider/model metadata.")

print(f"{provider_id}/{model_id}")
PYEOF
)"

printf '%s\n' "$current_model"
