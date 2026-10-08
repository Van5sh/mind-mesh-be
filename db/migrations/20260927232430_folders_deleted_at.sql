-- +goose up
-- Folders have no soft-delete concept at all today (unlike files, which at
-- least have an unused file_properties.deleted_at column) - needed so
-- trashFolder/restoreFolder have somewhere to write.
ALTER TABLE folders ADD COLUMN deleted_at TIMESTAMPTZ;

-- +goose down
ALTER TABLE folders DROP COLUMN deleted_at;
