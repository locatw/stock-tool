#!/bin/bash
set -euo pipefail

# API 定義ファイルの変更がなければスキップ
if ! git diff --name-only HEAD 2>/dev/null | grep -q '^backend/api/definition/'; then
  exit 0
fi

cd "$CLAUDE_PROJECT_DIR"
OUTPUT=$(make gen-api 2>&1) && exit 0

echo "$OUTPUT" >&2
exit 2
