-- +goose up
-- CreateFile never inserted a file_properties row (only files +
-- file_ai_metadata), so every existing file is missing one. properties is a
-- non-null GraphQL field and SoftDeleteFile/RestoreFile write to this table,
-- so every file needs a row before either can mean anything.
INSERT INTO file_properties (file_id, original_name, is_indexed, deleted_at)
SELECT f.id, f.name, FALSE, NULL
FROM files f
LEFT JOIN file_properties fp ON fp.file_id = f.id
WHERE fp.file_id IS NULL;

-- +goose down
-- No-op: these rows are indistinguishable from ones CreateFile will insert
-- going forward, so there's nothing safe to remove on rollback.
