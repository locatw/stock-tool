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

Each `backend/cmd/*/` directory and the project root contain `.env.template` files. Copy them to `.env` and fill in the values. See each template for required variables.
