# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation

- [doc/principles.md](doc/principles.md) — Cross-cutting engineering principles (DRY, YAGNI); apply to all changes
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
- [doc/coding-guidelines/go/coding.md](doc/coding-guidelines/go/coding.md) — Go coding conventions
- [doc/coding-guidelines/go/doc-comments.md](doc/coding-guidelines/go/doc-comments.md) — Go doc comment conventions
- [doc/coding-guidelines/go/testing.md](doc/coding-guidelines/go/testing.md) — Go testing conventions
- [doc/coding-guidelines/markdown/style.md](doc/coding-guidelines/markdown/style.md) — Markdown formatting rules
- [doc/review-guidelines/overview.md](doc/review-guidelines/overview.md) — Review guidelines (code, security, test coverage, architecture)
- [doc/specification-guideline.md](doc/specification-guideline.md) — Where to write specifications (doc/spec/ vs code doc comments)
- [doc/spec/](doc/spec/) — Specifications with use case scenarios
