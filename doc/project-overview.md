# Project Overview

Stock-tool is an investment support application:

- Fetches investment-related data from external sources
- Persists data in a local data store
- Provides a foundation for analyzing stored data to inform investment decisions

Current scope: J-Quants API for Japanese stock market data (brands, daily quotes), stored in PostgreSQL with extraction task and execution history tracking.

## Entry Points

Applications built with Cobra:

- `backend/cmd/cli/` — Database initialization and migration management
- `backend/cmd/task/` — Data extraction worker (`extract jquants` subcommand)
- `backend/cmd/api/` — HTTP API server (Echo, health check endpoint)

## External Services

- J-Quants API — Source for Japanese stock market data (brands, daily quotes)
- PostgreSQL — Primary data store (schema: `stock`)