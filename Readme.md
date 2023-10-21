# stock-tool

## Initial Set up

    $ cp .env.template .env
    $ vi .env

## Migration

Show migration status.

    $ docker compose run --rm migration sql-migrate status

Generate new migration file.

    $ docker compose run --rm migration sql-migrate new MIGRATION_NAME
