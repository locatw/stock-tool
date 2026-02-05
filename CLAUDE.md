# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Stock-tool is a data extraction platform for Japanese stock market data, integrating with the J-Quants API.
It stores stock brands and price data in PostgreSQL, tracking extraction tasks and their execution history.

## Development Environment

- **Go 1.24**, **PostgreSQL 17.5**
- Docker Compose for local Postgres: `docker compose up db`
- Three `.env` files are needed (copy from `.env.template` in each location):
  - Root `.env` (DB_USER, DB_PASSWORD for compose)
  - `backend/cmd/cli/.env` (DB connection for CLI)
  - `backend/cmd/task/.env` (DB connection + J-Quants credentials for task worker)

## Common Commands

All Go commands run from `backend/`:

```bash
# Run tests (requires Docker daemon for repository tests)
cd backend && go test ./...

# Run a single test
cd backend && go test ./internal/infra/repository/ -run TestExtractTaskRepository/TestCreate

# Database migrations
cd backend && go run ./cmd/cli/ migrate version
cd backend && go run ./cmd/cli/ migrate up
cd backend && go run ./cmd/cli/ migrate create MIGRATION_NAME

# Run extract task
cd backend && go run ./cmd/task/ extract jquants --type brand --code 86970 --dest-url file://./output
```

## Architecture

Clean Architecture with three layers:

```
domain/extract/         Immutable entities (private fields, getters, New*/New*Directly constructors)
    ↑
usecase/task/           Business logic orchestrating API calls, repository, and execution tracking
    ↑
infra/                  Infrastructure implementations
├── api/jquants/        J-Quants API client (auth, token refresh, brand/price endpoints)
├── repository/         GORM-based PostgreSQL access (stock schema)
└── registry/           Dependency wiring (WIP)
```

**Two CLI entry points** (Cobra-based):
- `cmd/cli/` — Database init and migration management
- `cmd/task/` — Data extraction worker (extract jquants subcommand)

**Dependency injection** uses `samber/do`.

## Key Domain Model

`ExtractTask` → has many `ExtractTaskExecution` → has many `ExtractedDataS3`

Tasks track what to extract (source, dataType, timing). Executions track individual runs (status, timestamps, errors). S3 files track output artifacts.

## Database

- Schema: `stock` (not public)
- Tables: `extract_tasks`, `extract_task_executions`, `extracted_data_s3s`, `brands`, `prices`
- Migrations in `backend/migrations/` using golang-migrate

## Testing

Repository tests use `ory/dockertest` to spin up a Postgres 17 container per suite. The `testutil.DBTest` base suite handles container lifecycle and migration application. Tests require a running Docker daemon.

Test framework: `testify/suite` with `testify/assert`. Use `google/go-cmp` for deep comparisons.

## Coding Conventions

See `.clinerules/` for full rules. Key points:

- Use `samber/lo` for slice transformations (lo.Map, etc.); keep explicit loops for DB operations
- No interfaces in the repository package — consumers define their own interfaces
- Line length max 120 chars; break long method signatures across lines
- Use `:=` (not `=`) in single-line `if err :=` checks
- GORM models: minimal tags (only override defaults), singular struct names, pointer slices for relations
- Repository models implement `ToEntity()` (DB→domain); private `toXxx()` functions for domain→DB
- All repository methods take `context.Context` first and use `WithContext()` on GORM
- Transaction control belongs in the upper layer, not in repositories
- Use "Repository" not "Repo" in type names
- No parent object references in child structs (only parent ID)