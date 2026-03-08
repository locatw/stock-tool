#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../.env
set -a
source "$REPO_ROOT/.env"
set +a

exec uvx pgsql-mcp-server --dsn "postgresql://$DB_USER:$DB_PASSWORD@localhost:5432/stock"
