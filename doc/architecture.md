# Architecture

## Clean Architecture Layers

All internal packages live under `backend/internal/`:

```text
infra/                  Infrastructure implementations
├── api/jquants/        J-Quants API client (auth, token refresh, brand/price endpoints)
├── repository/         GORM-based PostgreSQL access (stock schema)
└── registry/           Dependency wiring (WIP)
    ↓
usecase/task/           Business logic orchestrating API calls, repository, and execution tracking
    ↓
domain/extract/         Immutable entities (private fields, getters, New*/New*Directly constructors)
```

## Dependency Injection

Uses `samber/do` for service registration. Wiring happens in each `cmd/*/main.go`.

## Domain Model

`ExtractTask` → has many `ExtractTaskExecution` → has many `ExtractedDataS3`

- ExtractTask — Defines what to extract (source, dataType, timing)
- ExtractTaskExecution — Tracks an individual run (status, timestamps, errors)
- ExtractedDataS3 — Tracks output artifacts (S3 object keys)

All domain entities are immutable with private fields and getter methods. Two constructor patterns:

- `New*()` — Creates a new instance with timestamps
- `New*Directly()` — Reconstructs from persisted data

## Database

- Schema: `stock` (not the default `public` schema)
- Migrations in `backend/migrations/` using golang-migrate
