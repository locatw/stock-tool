BEGIN;

ALTER TABLE stock.data_types ADD COLUMN update_frequency TEXT;
ALTER TABLE stock.data_types ADD COLUMN update_times JSONB;

UPDATE stock.data_types
SET update_frequency = schedule->>'type',
    update_times = schedule->'times';

ALTER TABLE stock.data_types ALTER COLUMN update_frequency SET NOT NULL;
ALTER TABLE stock.data_types ALTER COLUMN update_times SET NOT NULL;

ALTER TABLE stock.data_types DROP COLUMN schedule;

COMMIT;
