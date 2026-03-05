# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation

- [doc/project-overview.md](doc/project-overview.md) — What this project does and its entry points
- [doc/architecture.md](doc/architecture.md) — Clean architecture layers, domain model, database design
- [doc/proposals/data-persistence-architecture.md](doc/proposals/data-persistence-architecture.md) — Lakehouse data persistence design proposal (Iceberg, DuckDB, Ceph)
- [doc/proposals/local-development-environment.md](doc/proposals/local-development-environment.md) — Repository strategy and local dev environment proposal (Docker Compose, Kind)
- [doc/proposals/data-ingestion-requirements.md](doc/proposals/data-ingestion-requirements.md) — General data ingestion requirements (storage, backfill, configuration)
- [doc/proposals/jquants-data-ingestion.md](doc/proposals/jquants-data-ingestion.md) — J-Quants specific data ingestion requirements and constraints
- [doc/proposals/data-lineage-design.md](doc/proposals/data-lineage-design.md) — Data lineage design (batch-level execution records over custom data IDs)
- [doc/requirements.md](doc/requirements.md) — Prerequisites, environment setup, `.env` configuration
- [doc/development-guideline.md](doc/development-guideline.md) — Common commands, migration file creation, testing
- [doc/documentation-guideline.md](doc/documentation-guideline.md) — Documentation policies and principles
- [doc/coding-guidelines/go.md](doc/coding-guidelines/go.md) — Go coding conventions (when updating, also sync .claude/rules/go-coding.md and go-testing.md)
- [doc/coding-guidelines/markdown.md](doc/coding-guidelines/markdown.md) — Markdown formatting rules
- [doc/spec/](doc/spec/) — Specifications with use case scenarios
