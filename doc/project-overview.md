# Project Overview

Stock-tool is an investment support application:

- Fetches investment-related data from external sources
- Persists data in a local data store
- Provides a foundation for analyzing stored data to inform investment decisions

Data sources and supported data types are configured at runtime; see the ingestion domain model for details.

## Entry Points

Applications built with Cobra (see each `main.go` for available subcommands):

- `backend/cmd/cli/` — Database initialization and migration management
- `backend/cmd/task/` — Data extraction worker
- `backend/cmd/api/` — HTTP API server

## External Services

- External data APIs — Configured as data sources at runtime
- PostgreSQL — Primary data store (schema: `stock`)
- S3-compatible object storage — Landing zone for extracted raw data