---
name: codex-review
description: Run Codex review on uncommitted changes and save result to .codex/review_result.md
disable-model-invocation: true
allowed-tools: Bash
---

Run the Codex review script to review all uncommitted changes.

Steps:
1. Run `.claude/skills/codex-review/codex-review.sh` from the project root using Bash
2. Report the result path `.codex/review_result.md` to the user
