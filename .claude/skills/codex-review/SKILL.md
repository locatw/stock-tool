---
name: codex-review
description: Run Codex review with optional target support (branch, commit, or uncommitted changes), analyze results, and fix issues
allowed-tools: Bash, Read, Edit, Glob, Grep
---

Run the Codex review script to review changes, then analyze the results and fix any issues found.

Steps:

1. Run `.claude/skills/codex-review/codex-review.sh` from the project root using Bash.
   If the user specified a branch (e.g. `/codex-review --branch feature`), pass `--branch <name>`.
   If the user specified a commit (e.g. `/codex-review --commit abc1234`), pass `--commit <ref>`.
   If no target was specified, run with no arguments to review uncommitted changes.
2. Read `.codex/review_result.md` using the Read tool
3. In the review result, find the last `codex` line (the final evaluation) and extract any issues raised
4. If there are no issues, report that the review passed with no issues
5. If there are issues, fix each one by reading and editing the relevant source files
6. Report a summary of the review result and any fixes applied to the user
