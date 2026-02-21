# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation

- [doc/project-overview.md](doc/project-overview.md) — What this project does and its entry points
- [doc/architecture.md](doc/architecture.md) — Clean architecture layers, domain model, database design
- [doc/proposals/data-persistence-architecture.md](doc/proposals/data-persistence-architecture.md) — Lakehouse data persistence design proposal (Iceberg, DuckDB, Ceph)
- [doc/proposals/local-development-environment.md](doc/proposals/local-development-environment.md) — Repository strategy and local dev environment proposal (Docker Compose, Kind)
- [doc/proposals/data-acquisition-requirements.md](doc/proposals/data-acquisition-requirements.md) — General data acquisition requirements (storage, backfill, configuration)
- [doc/proposals/jquants-data-acquisition.md](doc/proposals/jquants-data-acquisition.md) — J-Quants specific data acquisition requirements and constraints
- [doc/proposals/data-lineage-design.md](doc/proposals/data-lineage-design.md) — Data lineage design (batch-level execution records over custom data IDs)
- [doc/requirements.md](doc/requirements.md) — Prerequisites, environment setup, `.env` configuration
- [doc/development-guideline.md](doc/development-guideline.md) — Common commands, testing
- [doc/documentation-guideline.md](doc/documentation-guideline.md) — Documentation policies and principles
- [doc/coding-guidelines/go.md](doc/coding-guidelines/go.md) — Go coding conventions (formatting, GORM, repository rules)
- [doc/coding-guidelines/markdown.md](doc/coding-guidelines/markdown.md) — Markdown formatting rules
- [doc/spec/](doc/spec/) — Specifications with use case scenarios

## Quick Reference

All Go commands run from `backend/`:

```bash
cd backend && go test ./...                    # Run all tests (requires Docker daemon)
cd backend && go run ./cmd/cli/ migrate up     # Apply migrations
make lint                                      # Run golangci-lint
make lint-fix                                  # Run golangci-lint with auto-fix
```

## Hooks

A Stop hook (`.claude/hooks/lint-go.sh`) runs `make lint` automatically when Claude finishes a task. If any `.go` files were changed and lint violations are found, the hook blocks Claude from stopping and feeds the violations back so Claude can fix them before completing.
