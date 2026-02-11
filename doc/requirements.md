# Requirements

## Runtime

- Go (see version in `go.mod`)
- PostgreSQL (see version in `compose.yaml`)
- Docker (required for running repository tests and local database)

## Environment Setup

Start local infrastructure:

```bash
docker compose up db seaweedfs
```

Three `.env` files are needed (copy from `.env.template` in each location):

| File | Purpose |
|---|---|
| `.env` | DB_USER, DB_PASSWORD for Docker Compose |
| `backend/cmd/cli/.env` | DB connection for CLI tool |
| `backend/cmd/task/.env` | DB connection + J-Quants credentials for task worker |
