# stock-tool

## Initial Set up

Make .env in repository root.

    $ cp .env.template .env
    $ vi .env

Make .env in backend.

    $ cp ./backend/.env.template ./backend/.env
    $ vi ./backend/.env

## Migration

Show migration status.

    $ go run ./cmd/cli/ migrate version

Generate new migration file.

    $ go run ./cmd/cli/ migrate create MIGRATION_NAME

Apply migrations.

    $ go run ./cmd/cli/ migrate up
