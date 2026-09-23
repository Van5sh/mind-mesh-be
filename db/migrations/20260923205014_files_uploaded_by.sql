-- +goose up
-- Who created a file, set synchronously at INSERT time - unlike
-- file_storage.uploaded_by (that table is never actually populated by any
-- current code path). This is what personal (project-less) files use for
-- ownership; project files get it too, for free, as useful metadata.
ALTER TABLE files ADD COLUMN uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by);

-- +goose down
DROP INDEX IF EXISTS idx_files_uploaded_by;
ALTER TABLE files DROP COLUMN uploaded_by;
