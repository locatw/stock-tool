-- +migrate Up
CREATE TABLE prices (
    id bigserial PRIMARY KEY,
    date text NOT NULL,
    code text NOT NULL,
    "open" numeric(11, 1),
    high numeric(11, 1),
    low numeric(11, 1),
    close numeric(11, 1),
    volume numeric(11, 1),
    turnover_value numeric(20, 1),
    adjustment_factor numeric(11, 1),
    adjustment_open numeric(11, 1),
    adjustment_high numeric(11, 1),
    adjustment_low numeric(11, 1),
    adjustment_close numeric(11, 1),
    adjustment_volume numeric(11, 1),
    created_at timestamp WITH time zone NOT NULL,
    updated_at timestamp WITH time zone NOT NULL,
    deleted_at timestamp WITH time zone,
    UNIQUE(date, code)
);

CREATE INDEX ON prices (created_at);

CREATE INDEX ON prices (updated_at);

-- +migrate Down
DROP TABLE prices;
