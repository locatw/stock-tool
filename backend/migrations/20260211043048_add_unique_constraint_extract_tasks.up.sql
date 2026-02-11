ALTER TABLE stock.extract_tasks
    ADD CONSTRAINT extract_tasks_source_data_type_timing_key UNIQUE (source, data_type, timing);
