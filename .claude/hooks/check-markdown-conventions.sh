#!/bin/bash
set -euo pipefail

CHANGED_FILES=$(git diff --name-only HEAD -- '*.md' '**/*.md' 2>/dev/null || true)

[[ -z "$CHANGED_FILES" ]] && exit 0

cd "$CLAUDE_PROJECT_DIR"

BATCH_SIZE=5
GUIDELINES=$(cat doc/coding-guidelines/markdown/style.md)
mapfile -t FILES_ARRAY <<< "$CHANGED_FILES"
TOTAL=${#FILES_ARRAY[@]}
ERRORS=""

for (( i=0; i<TOTAL; i+=BATCH_SIZE )); do
  BATCH=("${FILES_ARRAY[@]:i:BATCH_SIZE}")

  PROMPT="You are a Markdown reviewer. Check the following files against the style guidelines below.

=== STYLE GUIDELINES ===
$GUIDELINES

=== FILES TO REVIEW ===
"
  for f in "${BATCH[@]}"; do
    if [[ -f "$f" ]]; then
      PROMPT+="
--- $f ---
$(cat "$f")
"
    fi
  done

  PROMPT+="
Review each file against the guidelines. If all files comply, output exactly PASS on a single line.
If there are violations, list them concisely (file:line description) without outputting PASS."

  BATCH_NUM=$(( i/BATCH_SIZE + 1 ))
  BATCH_TOTAL=$(( (TOTAL + BATCH_SIZE - 1) / BATCH_SIZE ))
  echo "Checking batch $BATCH_NUM/$BATCH_TOTAL (files $((i+1))-$((i+${#BATCH[@]})) of $TOTAL)..." >&2

  OUTPUT=$(claude -p --output-format text "$PROMPT" 2>&1)

  if ! echo "$OUTPUT" | grep -qx "PASS"; then
    ERRORS+="$OUTPUT"$'\n'
  fi
done

if [[ -n "$ERRORS" ]]; then
  echo "$ERRORS" >&2
  exit 2
fi

exit 0
