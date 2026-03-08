#!/bin/bash
set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
REVIEW_OUTPUT="$REPO_ROOT/.codex/review_result.md"
WORKTREE_DIR=$(mktemp -d)
HEAD_COMMIT=$(git rev-parse HEAD)

cleanup() {
  cd "$REPO_ROOT"
  git worktree remove "$WORKTREE_DIR" --force 2>/dev/null || true
}
trap cleanup EXIT

# 1. Create clean worktree (committed files only, no .env)
git worktree add "$WORKTREE_DIR" "$HEAD_COMMIT" --detach

# 2. Apply uncommitted changes of tracked files via patch
PATCH_FILE=$(mktemp)
git diff HEAD > "$PATCH_FILE" 2>/dev/null || true
if [ -s "$PATCH_FILE" ]; then
  cd "$WORKTREE_DIR" && git apply "$PATCH_FILE" 2>/dev/null || true
  cd "$REPO_ROOT"
fi
rm -f "$PATCH_FILE"

# 3. Copy untracked files (.gitignore-respected, so .env is excluded)
git ls-files --others --exclude-standard | while IFS= read -r file; do
  mkdir -p "$WORKTREE_DIR/$(dirname "$file")"
  cp "$file" "$WORKTREE_DIR/$file"
done

# 4. Run Codex review in the worktree
mkdir -p "$(dirname "$REVIEW_OUTPUT")"
cd "$WORKTREE_DIR"
codex review --uncommitted > "$REVIEW_OUTPUT" 2>&1

echo "Review result saved to: $REVIEW_OUTPUT" >&2
