BEGIN;

CREATE TABLE stock.prices (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    date TEXT NOT NULL,
    "open" NUMERIC(11, 1),
    high NUMERIC(11, 1),
    low NUMERIC(11, 1),
    close NUMERIC(11, 1),
    volume NUMERIC(11, 1),
    turnover_value NUMERIC(20, 1),
    adjustment_factor NUMERIC(11, 1),
    adjustment_open NUMERIC(11, 1),
    adjustment_high NUMERIC(11, 1),
    adjustment_low NUMERIC(11, 1),
    adjustment_close NUMERIC(11, 1),
    adjustment_volume NUMERIC(11, 1),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    UNIQUE(code, date)
);

CREATE INDEX ON stock.prices (code);

CREATE INDEX ON stock.prices (date);

CREATE INDEX ON stock.prices (created_at);

CREATE INDEX ON stock.prices (updated_at);

CREATE INDEX ON stock.prices (deleted_at);

COMMIT;
