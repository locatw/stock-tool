#!/bin/bash
set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
REVIEW_OUTPUT="$REPO_ROOT/.codex/review_result.md"
WORKTREE_DIR=$(mktemp -d)

cleanup() {
  cd "$REPO_ROOT"
  git worktree remove "$WORKTREE_DIR" --force 2>/dev/null || true
}
trap cleanup EXIT

# Parse arguments
MODE="uncommitted"
BRANCH=""
COMMIT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --branch)
      MODE="branch"
      BRANCH="$2"
      shift 2
      ;;
    --commit)
      MODE="commit"
      COMMIT="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [ "$MODE" = "branch" ]; then
  BASE=$(git merge-base main "$BRANCH")
  git worktree add "$WORKTREE_DIR" "$BASE" --detach

  PATCH_FILE=$(mktemp)
  git diff "$BASE" "$BRANCH" > "$PATCH_FILE" 2>/dev/null || true
  if [ -s "$PATCH_FILE" ]; then
    cd "$WORKTREE_DIR" && git apply "$PATCH_FILE" 2>/dev/null || true
    cd "$REPO_ROOT"
  fi
  rm -f "$PATCH_FILE"

elif [ "$MODE" = "commit" ]; then
  BASE=$(git rev-parse "$COMMIT^")
  git worktree add "$WORKTREE_DIR" "$BASE" --detach

  PATCH_FILE=$(mktemp)
  git diff "$BASE" "$COMMIT" > "$PATCH_FILE" 2>/dev/null || true
  if [ -s "$PATCH_FILE" ]; then
    cd "$WORKTREE_DIR" && git apply "$PATCH_FILE" 2>/dev/null || true
    cd "$REPO_ROOT"
  fi
  rm -f "$PATCH_FILE"

else
  # Default: review uncommitted changes
  HEAD_COMMIT=$(git rev-parse HEAD)

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
fi

# Run Codex review in the worktree
mkdir -p "$(dirname "$REVIEW_OUTPUT")"
cd "$WORKTREE_DIR"
codex review --uncommitted > "$REVIEW_OUTPUT" 2>&1

echo "Review result saved to: $REVIEW_OUTPUT" >&2
