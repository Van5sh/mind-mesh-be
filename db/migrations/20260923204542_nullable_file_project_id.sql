-- +goose up
-- A file can now exist outside any project ("personal" files, uploaded
-- straight to a user's own drive space rather than a project). Ownership
-- for those is file_storage.uploaded_by (already NOT NULL, already
-- indexed) - no new column needed. Folders remain project-scoped only;
-- personal files live at the root, with no folder.
ALTER TABLE files ALTER COLUMN project_id DROP NOT NULL;

-- +goose down
-- Down-migration note: this fails if any personal (project_id IS NULL)
-- files exist by the time it runs - delete or reassign them first.
ALTER TABLE files ALTER COLUMN project_id SET NOT NULL;
