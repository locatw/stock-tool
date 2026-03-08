---
name: verify-api
description: Start the API server and verify all endpoints with test requests
disable-model-invocation: true
allowed-tools: Bash, mcp__postgres__run_dql_query, mcp__postgres__run_dml_query, mcp__postgres__run_ddl_query
---

Start the API server with a disposable verification DB and run curl-based endpoint
tests. The `stock` DB is never touched.

Important: On any failure, always execute steps 5 and 6 (stop server, drop DB) before
reporting the error.

## Steps

### 1. Create verification DB

Read DB credentials from `.env` (same method as `scripts/postgres-mcp.sh`):

```bash
REPO_ROOT="$(git rev-parse --show-toplevel)"
set -a && source "$REPO_ROOT/.env" && set +a
docker compose exec db createdb -U "$DB_USER" stock_verify
```

### 2. Initialize and migrate

```bash
cd "$REPO_ROOT/backend"
DB_NAME=stock_verify go run ./cmd/cli/ initdb
DB_NAME=stock_verify go run ./cmd/cli/ migrate up
```

### 3. Start API server

```bash
DB_NAME=stock_verify PORT=18080 "$REPO_ROOT/scripts/api-server.sh" start
PORT=18080 "$REPO_ROOT/scripts/api-server.sh" wait
```

If wait fails, run `scripts/api-server.sh stop` and drop the DB, then report the error.

### 4. Send test requests

Use `BASE=http://localhost:18080` for all requests.

data-sources CRUD cycle:

```bash
# List (empty)
curl -sf "$BASE/api/v1/data-sources" | jq .

# Create
DS_RESP=$(curl -sf -X POST "$BASE/api/v1/data-sources" \
  -H 'Content-Type: application/json' \
  -d '{"name":"__verify_test__","enabled":true,"timezone":"UTC","settings":{}}' \
  -w '\n%{http_code}')
DS_ID=$(echo "$DS_RESP" | head -1 | jq -r '.id')

# Get by ID
curl -sf "$BASE/api/v1/data-sources/$DS_ID" | jq .

# Update
curl -sf -X PUT "$BASE/api/v1/data-sources/$DS_ID" \
  -H 'Content-Type: application/json' \
  -d '{"name":"__verify_updated__","enabled":true,"timezone":"UTC","settings":{}}' | jq .

# Delete
curl -sf -X DELETE "$BASE/api/v1/data-sources/$DS_ID" -w '%{http_code}'
# Expect 204

# Confirm 404
curl -s -o /dev/null -w '%{http_code}' "$BASE/api/v1/data-sources/$DS_ID"
# Expect 404
```

data-types CRUD cycle:

- Create a parent data-source first
- Then run the same create/get/update/delete/get-404 cycle for data-types under that
  data-source
- Clean up the parent data-source afterward

For each request, verify the HTTP status code matches the expected value (200, 201,
204, 404). Report any mismatches as failures.

### 5. Stop API server

```bash
PORT=18080 "$REPO_ROOT/scripts/api-server.sh" stop
```

### 6. Drop verification DB

```bash
docker compose exec db dropdb -U "$DB_USER" stock_verify
```

### 7. Report results

Print a summary table of each endpoint tested with its expected and actual status code.
