#!/bin/bash
set -euo pipefail

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Skip if no file path or not a .go file
[[ -z "$FILE_PATH" || "$FILE_PATH" != *.go ]] && exit 0

# Skip if not under backend/
BACKEND_DIR="$CLAUDE_PROJECT_DIR/backend"
[[ "$FILE_PATH" != "$BACKEND_DIR"/* ]] && exit 0

# Determine guideline files
GUIDELINES="doc/coding-guidelines/go/coding.md"
if [[ "$FILE_PATH" == *_test.go ]]; then
  GUIDELINES="$GUIDELINES and doc/coding-guidelines/go/testing.md"
fi

PROMPT="Review the Go file at \`$FILE_PATH\` against the coding guidelines at $GUIDELINES.
Read both the file and the guidelines. Then check for any violations.
Output exactly \`PASS\` if the file complies with all guidelines.
If there are violations, list them concisely (one per line) without outputting PASS."

cd "$CLAUDE_PROJECT_DIR"
OUTPUT=$(claude -p --tools "Read" --max-turns 1 --output-format text "$PROMPT" 2>&1)

if echo "$OUTPUT" | grep -q "^PASS"; then
  exit 0
fi

echo "$OUTPUT" >&2
exit 2
