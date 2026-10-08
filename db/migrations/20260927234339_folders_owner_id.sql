-- +goose up
-- Folders have no creator/owner tracking at all today - fine while every
-- folder is project-scoped (requireProjectMember covers authorization), but
-- rootFolders/folders both require a projectId, so there's no such thing as
-- a personal folder yet. This is what a personal folder's authorization
-- (requireFolderAccess) and its listing queries (GetPersonalFolders,
-- GetTrashedPersonalFolders, ...) key off, same as files.uploaded_by.
ALTER TABLE folders ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_folders_owner_id ON folders(owner_id);

-- +goose down
DROP INDEX IF EXISTS idx_folders_owner_id;
ALTER TABLE folders DROP COLUMN owner_id;
