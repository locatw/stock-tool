# Project Overview

Stock-tool is a data extraction platform for stock market data, integrating with the J-Quants API.
It stores stock brands and price data in PostgreSQL, tracking extraction tasks and their execution history.

## Entry Points

Two CLI applications built with Cobra:

- `backend/cmd/cli/` — Database initialization and migration management
- `backend/cmd/task/` — Data extraction worker (`extract jquants` subcommand)

## External Services

- **J-Quants API** — Source for Japanese stock market data (brands, daily quotes)
- **PostgreSQL** — Primary data store (schema: `stock`)