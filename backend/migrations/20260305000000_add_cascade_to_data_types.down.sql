BEGIN;

ALTER TABLE stock.data_types
    DROP CONSTRAINT data_types_data_source_id_fkey,
    ADD CONSTRAINT data_types_data_source_id_fkey
        FOREIGN KEY (data_source_id) REFERENCES stock.data_sources(id);

COMMIT;
