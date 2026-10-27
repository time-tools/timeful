#!/usr/bin/env bash
# Entry point for the GALA twin verifier.
#
# The implementation lives in verify.py; this wrapper keeps the documented
# path (server/scripts/gala/verify.sh) and passes every argument through.
# The verifier writes a timestamped log to server/.gala/verify.log by default;
# see `server/scripts/gala/verify.sh --help` for --write, --jobs, --verbose,
# --log, and --no-log.
set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

if ! command -v python3 >/dev/null 2>&1; then
  printf 'verify.sh: python3 not found; run inside `nix develop`\n' >&2
  exit 1
fi

exec python3 "$script_dir/verify.py" "$@"
