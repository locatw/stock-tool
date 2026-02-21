# Development Guideline

## Common Commands

All Go commands run from `backend/`:

```bash
# Run all tests (requires Docker daemon for repository tests)
cd backend && go test ./...

# Run a single test
cd backend && go test ./internal/infra/repository/ -run TestExtractTaskRepository/TestCreate

# Database migrations
# Always use `migrate create` to generate migration files — never create them manually.
cd backend && go run ./cmd/cli/ migrate version
cd backend && go run ./cmd/cli/ migrate up
cd backend && go run ./cmd/cli/ migrate create MIGRATION_NAME

# Run extract task
cd backend && go run ./cmd/task/ extract jquants --type brand --code 86970 --dest-url file://./output
```

## Linting

```bash
make lint       # Run golangci-lint
make lint-fix   # Run golangci-lint with auto-fix
```

- Configuration: `backend/.golangci.yml`
- Managed as a `go tool` dependency — no separate installation needed

## Testing

- Repository tests use `ory/dockertest` to spin up a Postgres container per test suite
- `testutil.DBTest` base suite (`backend/internal/util/testutil/`) handles container lifecycle and migration
- Requires a running Docker daemon
- Test framework: `testify/suite` with `testify/assert`
- Deep comparisons: `google/go-cmp`
