# Project Overview

Stock-tool is an investment support application designed to assist with investment decision-making.
It fetches investment-related data from external sources, persists it in a local data store, and provides a foundation for analyzing the stored data to inform investment decisions.

Currently, the platform integrates with the J-Quants API to extract Japanese stock market data (brands and daily quotes) and stores it in PostgreSQL, tracking extraction tasks and their execution history.

## Entry Points

Two CLI applications built with Cobra:

- `backend/cmd/cli/` — Database initialization and migration management
- `backend/cmd/task/` — Data extraction worker (`extract jquants` subcommand)

## External Services

- **J-Quants API** — Source for Japanese stock market data (brands, daily quotes)
- **PostgreSQL** — Primary data store (schema: `stock`)