BEGIN;

--
-- extract_tasks
--
CREATE TABLE stock.extract_tasks (
    id SERIAL PRIMARY KEY,
    source TEXT NOT NULL,
    data_type TEXT NOT NULL,
    STATUS TEXT NOT NULL,
    error_info TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ON stock.extract_tasks (source);

CREATE INDEX ON stock.extract_tasks (data_type);

CREATE INDEX ON stock.extract_tasks (STATUS);

CREATE INDEX ON stock.extract_tasks (started_at);

CREATE INDEX ON stock.extract_tasks (finished_at);

CREATE INDEX ON stock.extract_tasks (created_at);

CREATE INDEX ON stock.extract_tasks (updated_at);

--
-- extracted_data_s3s
--
CREATE TABLE stock.extracted_data_s3s (
    id SERIAL PRIMARY KEY,
    extract_task_id INTEGER NOT NULL REFERENCES stock.extract_tasks(id),
    target_date_time TIMESTAMPTZ NOT NULL,
    bucket TEXT NOT NULL,
    KEY TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ON stock.extracted_data_s3s (extract_task_id);

CREATE INDEX ON stock.extracted_data_s3s (target_date_time);

CREATE INDEX ON stock.extracted_data_s3s (bucket);

CREATE INDEX ON stock.extracted_data_s3s (created_at);

CREATE INDEX ON stock.extracted_data_s3s (updated_at);

COMMIT;
