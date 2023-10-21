-- +migrate Up
CREATE TABLE brands (
    id bigserial PRIMARY KEY,
    date text NOT NULL,
    code text UNIQUE NOT NULL,
    company_name text NOT NULL,
    company_name_english text NOT NULL,
    sector_17_code text NOT NULL,
    sector_17_code_name text NOT NULL,
    sector_33_code text NOT NULL,
    sector_33_code_name text NOT NULL,
    scale_category text NOT NULL,
    market_code text NOT NULL,
    market_code_name text NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    updated_at timestamp WITH time zone NOT NULL,
    deleted_at timestamp WITH time zone
);

CREATE INDEX ON brands (date);

CREATE INDEX ON brands (code);

CREATE INDEX ON brands (created_at);

CREATE INDEX ON brands (updated_at);

-- +migrate Down
DROP TABLE brands;
