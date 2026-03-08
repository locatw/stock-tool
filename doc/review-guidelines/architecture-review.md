# Architecture Review Guidelines

Verify alignment with the clean architecture defined in [doc/architecture.md](../architecture.md).

## Layer Dependency Direction

The dependency direction must flow inward only:

- `infra/` depends on `usecase/` and `domain/`.
- `usecase/` depends on `domain/` only.
- `domain/` has no dependencies on other internal packages.

Flag any import that crosses a layer boundary in the wrong direction.

## Domain Entities

- Domain entities must have private fields with getter methods.
- Two constructor patterns are required:
  - `New*()` for creating new instances (sets timestamps).
  - `New*Directly()` for reconstructing from persisted data.
- No domain entity may expose a public field.

## Repository Rules

- Repository packages must provide only concrete implementations.
- Interfaces must be defined by the consumer (usecase layer), not the repository package.
- Transaction control must not be inside repository methods; it belongs in the upper layer.
- All repository methods must accept `context.Context` as their first parameter.

## Dependency Injection

- Service wiring must happen in `cmd/*/main.go` using `samber/do`.
- No manual dependency passing should appear outside of `main.go` or registry files.

## Database

- All tables must use the `stock` schema, not the default `public` schema.
- Schema changes must be managed via migration files in `backend/migrations/`.
