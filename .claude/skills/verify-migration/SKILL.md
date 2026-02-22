---
name: verify-migration
description: Review new migration files and run migrate up against the local DB with user confirmation
disable-model-invocation: true
allowed-tools: Read, Bash, Glob, AskUserQuestion
---

Review new or modified migration files, obtain user approval, then run `migrate up`
against the local DB container.

Steps:
1. Run `git status --short` and identify new/changed files under `migrations/`
2. Display the contents of the relevant `.up.sql` files using the Read tool
3. Use AskUserQuestion to confirm with the user whether to proceed with `migrate up`
4. If approved, run the following in order:
   a. `docker compose up db -d --wait` from the project root
   b. `cd backend && go run ./cmd/cli/ migrate up`
5. Report the result (success or error)

Note: Do not run `migrate down`.
