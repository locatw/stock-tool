BEGIN;

--
-- data_types
--
CREATE TABLE stock.data_types (
    id UUID PRIMARY KEY,
    data_source_id UUID NOT NULL REFERENCES stock.data_sources(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    backfill_enabled BOOLEAN NOT NULL DEFAULT true,
    stale_timeout_minutes INTEGER NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    schedule JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT data_types_data_source_id_name_key UNIQUE (data_source_id, name)
);

CREATE INDEX ON stock.data_types (data_source_id);
CREATE INDEX ON stock.data_types (name);
CREATE INDEX ON stock.data_types (created_at);
CREATE INDEX ON stock.data_types (updated_at);

COMMIT;
