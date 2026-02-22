BEGIN;

--
-- data_sources
--
CREATE TABLE stock.data_sources (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    timezone TEXT NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT data_sources_name_key UNIQUE (name)
);

CREATE INDEX ON stock.data_sources (name);
CREATE INDEX ON stock.data_sources (created_at);
CREATE INDEX ON stock.data_sources (updated_at);

--
-- data_types
--
CREATE TABLE stock.data_types (
    id UUID PRIMARY KEY,
    data_source_id UUID NOT NULL REFERENCES stock.data_sources(id),
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    update_frequency TEXT NOT NULL,
    update_times JSONB NOT NULL DEFAULT '[]',
    backfill_enabled BOOLEAN NOT NULL DEFAULT true,
    stale_timeout_minutes INTEGER NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT data_types_data_source_id_name_key UNIQUE (data_source_id, name)
);

CREATE INDEX ON stock.data_types (data_source_id);
CREATE INDEX ON stock.data_types (name);
CREATE INDEX ON stock.data_types (created_at);
CREATE INDEX ON stock.data_types (updated_at);

COMMIT;
