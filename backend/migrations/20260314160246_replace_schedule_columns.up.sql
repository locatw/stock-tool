BEGIN;

ALTER TABLE stock.data_types ADD COLUMN schedule JSONB;

UPDATE stock.data_types
SET schedule = jsonb_build_object('type', 'daily', 'times', update_times);

ALTER TABLE stock.data_types ALTER COLUMN schedule SET NOT NULL;

ALTER TABLE stock.data_types DROP COLUMN update_frequency;
ALTER TABLE stock.data_types DROP COLUMN update_times;

COMMIT;
