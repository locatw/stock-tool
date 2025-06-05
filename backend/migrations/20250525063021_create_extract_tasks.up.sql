BEGIN;

--
-- extract_tasks
--
CREATE TABLE stock.extract_tasks (
    id SERIAL PRIMARY KEY,
    source TEXT NOT NULL,
    data_type TEXT NOT NULL,
    timing TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ON stock.extract_tasks (source);
CREATE INDEX ON stock.extract_tasks (data_type);
CREATE INDEX ON stock.extract_tasks (timing);
CREATE INDEX ON stock.extract_tasks (created_at);
CREATE INDEX ON stock.extract_tasks (updated_at);

--
-- extract_task_executions
--

CREATE TABLE stock.extract_task_executions (
    id SERIAL PRIMARY KEY,
    extract_task_id INTEGER NOT NULL REFERENCES stock.extract_tasks(id),
    target_date_time TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL,
    error_info TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ON stock.extract_task_executions (extract_task_id);
CREATE INDEX ON stock.extract_task_executions (target_date_time);
CREATE INDEX ON stock.extract_task_executions (status);
CREATE INDEX ON stock.extract_task_executions (started_at);
CREATE INDEX ON stock.extract_task_executions (finished_at);
CREATE INDEX ON stock.extract_task_executions (created_at);
CREATE INDEX ON stock.extract_task_executions (updated_at);

--
-- extracted_data_s3s
--
CREATE TABLE stock.extracted_data_s3s (
    id SERIAL PRIMARY KEY,
    extract_task_execution_id INTEGER NOT NULL REFERENCES stock.extract_task_executions(id),
    key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ON stock.extracted_data_s3s (extract_task_execution_id);
CREATE INDEX ON stock.extracted_data_s3s (created_at);
CREATE INDEX ON stock.extracted_data_s3s (updated_at);

COMMIT;
