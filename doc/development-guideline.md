# Development Guideline

## Common Commands

All Go commands run from `backend/`:

```bash
# Run all tests (requires Docker daemon for repository tests)
cd backend && go test ./...

# Run a single test
cd backend && go test ./internal/infra/repository/ -run TestExtractTaskRepository/TestCreate

# Database migrations
cd backend && go run ./cmd/cli/ migrate version
cd backend && go run ./cmd/cli/ migrate up
cd backend && go run ./cmd/cli/ migrate create MIGRATION_NAME

# Run extract task
cd backend && go run ./cmd/task/ extract jquants --type brand --code 86970 --dest-url file://./output
```

## Testing

Repository tests use `ory/dockertest` to spin up a Postgres container per test suite. The `testutil.DBTest` base suite (in `backend/internal/util/testutil/`) handles container lifecycle and migration application. Tests require a running Docker daemon.

- Test framework: `testify/suite` with `testify/assert`
- Deep comparisons: `google/go-cmp`
