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

COMMIT;
