# Development Guideline

## Migrations

```bash
# Always use `migrate create` to generate migration files — never create them manually.
cd backend && go run ./cmd/cli/ migrate version
cd backend && go run ./cmd/cli/ migrate up
cd backend && go run ./cmd/cli/ migrate create MIGRATION_NAME
```

After creating or editing migration files, run `/verify-migration` to verify that the changes apply cleanly to the local DB.

## Testing

```bash
# Run all tests (requires Docker daemon for repository tests)
cd backend && go test ./...

# Run a single test
cd backend && go test ./internal/infra/repository/ -run TestExtractTaskRepository/TestCreate
```

- Repository tests use `ory/dockertest` to spin up a Postgres container per test suite
- `testutil.DBTest` base suite (`backend/internal/util/testutil/`) handles container lifecycle and migration
- Requires a running Docker daemon
- Test framework: `testify/suite` with `testify/assert`
- Deep comparisons: `google/go-cmp`

## Linting

```bash
make lint       # Run golangci-lint
make lint-fix   # Run golangci-lint with auto-fix
make lint-api   # Run Redocly lint on OpenAPI definition
```

- Go: Configuration `backend/.golangci.yml` — managed as a `go tool` dependency, no separate installation needed
- OpenAPI: Configuration `backend/api/config/redocly.yaml` — runs via Docker, no separate installation needed

## API Docs

```bash
docker compose --profile doc up api-doc
```

- Opens at `http://localhost:8080` (change port via `API_DOC_PORT` in `.env`)
- Definition files are served directly — reload the browser to see changes
