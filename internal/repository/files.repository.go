package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type FileRepository struct {
	q *database.Queries
}

func NewFileRepository(q *database.Queries) *FileRepository {
	return &FileRepository{
		q: q,
	}
}

func (r *FileRepository) CheckFileNameExists(ctx context.Context, params database.CheckFileNameExistsParams) (bool, error) {
	return r.q.CheckFileNameExists(ctx, params)
}

func (r *FileRepository) CheckFolderNameExists(ctx context.Context, params database.CheckFolderNameExistsParams) (bool, error) {
	return r.q.CheckFolderNameExists(ctx, params)
}

func (r *FileRepository) CheckProjectContainsFile(ctx context.Context, params database.CheckProjectContainsFileParams) (bool, error) {
	return r.q.CheckProjectContainsFile(ctx, params)
}

func (r *FileRepository) CheckProjectFileNameExists(ctx context.Context, params database.CheckProjectFileNameExistsParams) (bool, error) {
	return r.q.CheckProjectFileNameExists(ctx, params)
}

func (r *FileRepository) CountProjectFiles(ctx context.Context, projectID pgtype.UUID) (int64, error) {
	return r.q.CountProjectFiles(ctx, projectID)
}

func (r *FileRepository) CreateFile(ctx context.Context, params database.CreateFileParams) (database.File, error) {
	return r.q.CreateFile(ctx, params)
}

func (r *FileRepository) CreateFileAIMetadata(ctx context.Context, params database.CreateFileAIMetadataParams) (database.FileAiMetadatum, error) {
	return r.q.CreateFileAIMetadata(ctx, params)
}

func (r *FileRepository) CreateFileProperties(ctx context.Context, params database.CreateFilePropertiesParams) (database.FileProperty, error) {
	return r.q.CreateFileProperties(ctx, params)
}

func (r *FileRepository) CreateFileStorage(ctx context.Context, params database.CreateFileStorageParams) (database.FileStorage, error) {
	return r.q.CreateFileStorage(ctx, params)
}

func (r *FileRepository) CreateFolder(ctx context.Context, params database.CreateFolderParams) (database.Folder, error) {
	return r.q.CreateFolder(ctx, params)
}

func (r *FileRepository) CreateMessageFileReference(ctx context.Context, params database.CreateMessageFileReferenceParams) (database.MessageFileReference, error) {
	return r.q.CreateMessageFileReference(ctx, params)
}

func (r *FileRepository) CreateProjectFile(ctx context.Context, params database.CreateProjectFileParams) (database.ProjectFile, error) {
	return r.q.CreateProjectFile(ctx, params)
}

func (r *FileRepository) CreateUserFilePreference(ctx context.Context, params database.CreateUserFilePreferenceParams) (database.UserFilePreference, error) {
	return r.q.CreateUserFilePreference(ctx, params)
}

func (r *FileRepository) DeleteFile(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteFile(ctx, id)
}

func (r *FileRepository) DeleteFileAIMetadata(ctx context.Context, fileID pgtype.UUID) error {
	return r.q.DeleteFileAIMetadata(ctx, fileID)
}

func (r *FileRepository) DeleteFileProperties(ctx context.Context, fileID pgtype.UUID) error {
	return r.q.DeleteFileProperties(ctx, fileID)
}

func (r *FileRepository) DeleteFileShare(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteFileShare(ctx, id)
}

func (r *FileRepository) DeleteFileStorage(ctx context.Context, fileID pgtype.UUID) error {
	return r.q.DeleteFileStorage(ctx, fileID)
}

func (r *FileRepository) DeleteFolder(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteFolder(ctx, id)
}

func (r *FileRepository) DeleteMessageFileReferences(ctx context.Context, messageID pgtype.UUID) error {
	return r.q.DeleteMessageFileReferences(ctx, messageID)
}

func (r *FileRepository) DeleteUserFilePreference(ctx context.Context, params database.DeleteUserFilePreferenceParams) error {
	return r.q.DeleteUserFilePreference(ctx, params)
}

func (r *FileRepository) FileShare(ctx context.Context, params database.FileShareParams) (database.FileShare, error) {
	return r.q.FileShare(ctx, params)
}

func (r *FileRepository) GetChildFolders(ctx context.Context, parentFolderID pgtype.UUID) ([]database.Folder, error) {
	return r.q.GetChildFolders(ctx, parentFolderID)
}

func (r *FileRepository) GetDeletedFiles(ctx context.Context, projectID pgtype.UUID) ([]database.File, error) {
	return r.q.GetDeletedFiles(ctx, projectID)
}

func (r *FileRepository) GetFavoritesFiles(ctx context.Context, params database.GetFavoritesFilesParams) ([]database.File, error) {
	return r.q.GetFavoritesFiles(ctx, params)
}

func (r *FileRepository) GetFileAIMetadata(ctx context.Context, fileID pgtype.UUID) (database.FileAiMetadatum, error) {
	return r.q.GetFileAIMetadata(ctx, fileID)
}

func (r *FileRepository) GetFileByFolderAndName(ctx context.Context, params database.GetFileByFolderAndNameParams) (database.File, error) {
	return r.q.GetFileByFolderAndName(ctx, params)
}

func (r *FileRepository) GetFileByID(ctx context.Context, id pgtype.UUID) (database.File, error) {
	return r.q.GetFileByID(ctx, id)
}

func (r *FileRepository) GetFileProperties(ctx context.Context, fileID pgtype.UUID) (database.FileProperty, error) {
	return r.q.GetFileProperties(ctx, fileID)
}

func (r *FileRepository) GetFileShareByFileAndSharedWith(ctx context.Context, params database.GetFileShareByFileAndSharedWithParams) (database.FileShare, error) {
	return r.q.GetFileShareByFileAndSharedWith(ctx, params)
}

func (r *FileRepository) GetFileShareByID(ctx context.Context, id pgtype.UUID) (database.FileShare, error) {
	return r.q.GetFileShareByID(ctx, id)
}

func (r *FileRepository) GetFileSharesByFileID(ctx context.Context, fileID pgtype.UUID) ([]database.FileShare, error) {
	return r.q.GetFileSharesByFileID(ctx, fileID)
}

func (r *FileRepository) GetFileSharesBySharedWith(ctx context.Context, sharedWith pgtype.UUID) ([]database.FileShare, error) {
	return r.q.GetFileSharesBySharedWith(ctx, sharedWith)
}

func (r *FileRepository) GetFileStorage(ctx context.Context, fileID pgtype.UUID) (database.FileStorage, error) {
	return r.q.GetFileStorage(ctx, fileID)
}

func (r *FileRepository) GetFilesByFolderID(ctx context.Context, folderID pgtype.UUID) ([]database.File, error) {
	return r.q.GetFilesByFolderID(ctx, folderID)
}

func (r *FileRepository) GetFilesByIDs(ctx context.Context, ids []pgtype.UUID) ([]database.File, error) {
	return r.q.GetFilesByIDs(ctx, ids)
}

func (r *FileRepository) GetFilesByProjectID(ctx context.Context, projectID pgtype.UUID) ([]database.File, error) {
	return r.q.GetFilesByProjectID(ctx, projectID)
}

func (r *FileRepository) GetFilesPendingEmbedding(ctx context.Context) ([]database.File, error) {
	return r.q.GetFilesPendingEmbedding(ctx)
}

func (r *FileRepository) GetFilesSharedByUser(ctx context.Context, sharedBy pgtype.UUID) ([]database.FileShare, error) {
	return r.q.GetFilesSharedByUser(ctx, sharedBy)
}

func (r *FileRepository) GetFolderByID(ctx context.Context, id pgtype.UUID) (database.Folder, error) {
	return r.q.GetFolderByID(ctx, id)
}

func (r *FileRepository) GetFolderContents(ctx context.Context, parentFolderID pgtype.UUID) ([]database.GetFolderContentsRow, error) {
	return r.q.GetFolderContents(ctx, parentFolderID)
}

func (r *FileRepository) GetFoldersByProjectID(ctx context.Context, projectID pgtype.UUID) ([]database.Folder, error) {
	return r.q.GetFoldersByProjectID(ctx, projectID)
}

func (r *FileRepository) GetIndexedFiles(ctx context.Context, projectID pgtype.UUID) ([]database.File, error) {
	return r.q.GetIndexedFiles(ctx, projectID)
}

func (r *FileRepository) GetMessageFileReferences(ctx context.Context, messageID pgtype.UUID) ([]database.File, error) {
	return r.q.GetMessageFileReferences(ctx, messageID)
}

func (r *FileRepository) GetMessagesReferencingFile(ctx context.Context, fileID pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetMessagesReferencingFile(ctx, fileID)
}

func (r *FileRepository) GetProjectFile(ctx context.Context, params database.GetProjectFileParams) (database.ProjectFile, error) {
	return r.q.GetProjectFile(ctx, params)
}

func (r *FileRepository) GetProjectFileByFolderAndName(ctx context.Context, params database.GetProjectFileByFolderAndNameParams) (database.File, error) {
	return r.q.GetProjectFileByFolderAndName(ctx, params)
}

func (r *FileRepository) GetProjectFileByID(ctx context.Context, params database.GetProjectFileByIDParams) (database.File, error) {
	return r.q.GetProjectFileByID(ctx, params)
}

func (r *FileRepository) GetProjectFilesByFolderID(ctx context.Context, params database.GetProjectFilesByFolderIDParams) ([]database.File, error) {
	return r.q.GetProjectFilesByFolderID(ctx, params)
}

func (r *FileRepository) GetProjectFolderByID(ctx context.Context, params database.GetProjectFolderByIDParams) (database.Folder, error) {
	return r.q.GetProjectFolderByID(ctx, params)
}

func (r *FileRepository) GetProjectFolderContents(ctx context.Context, params database.GetProjectFolderContentsParams) ([]database.GetProjectFolderContentsRow, error) {
	return r.q.GetProjectFolderContents(ctx, params)
}

func (r *FileRepository) GetRootFiles(ctx context.Context, projectID pgtype.UUID) ([]database.File, error) {
	return r.q.GetRootFiles(ctx, projectID)
}

func (r *FileRepository) GetRootFolders(ctx context.Context, projectID pgtype.UUID) ([]database.Folder, error) {
	return r.q.GetRootFolders(ctx, projectID)
}

func (r *FileRepository) GetStandaloneRootFolders(ctx context.Context) ([]database.Folder, error) {
	return r.q.GetStandaloneRootFolders(ctx)
}

func (r *FileRepository) GetUserFilePreference(ctx context.Context, params database.GetUserFilePreferenceParams) (database.UserFilePreference, error) {
	return r.q.GetUserFilePreference(ctx, params)
}

func (r *FileRepository) MarkFileEmbeddingSynced(ctx context.Context, params database.MarkFileEmbeddingSyncedParams) (database.FileAiMetadatum, error) {
	return r.q.MarkFileEmbeddingSynced(ctx, params)
}

func (r *FileRepository) MarkFileIndexed(ctx context.Context, params database.MarkFileIndexedParams) error {
	return r.q.MarkFileIndexed(ctx, params)
}

func (r *FileRepository) MoveFile(ctx context.Context, params database.MoveFileParams) (database.File, error) {
	return r.q.MoveFile(ctx, params)
}

func (r *FileRepository) MoveFolder(ctx context.Context, params database.MoveFolderParams) (database.Folder, error) {
	return r.q.MoveFolder(ctx, params)
}

func (r *FileRepository) MoveProjectFile(ctx context.Context, params database.MoveProjectFileParams) (database.ProjectFile, error) {
	return r.q.MoveProjectFile(ctx, params)
}

func (r *FileRepository) RemoveProjectFile(ctx context.Context, params database.RemoveProjectFileParams) error {
	return r.q.RemoveProjectFile(ctx, params)
}

func (r *FileRepository) RenameFile(ctx context.Context, params database.RenameFileParams) (database.File, error) {
	return r.q.RenameFile(ctx, params)
}

func (r *FileRepository) RenameFolder(ctx context.Context, params database.RenameFolderParams) (database.Folder, error) {
	return r.q.RenameFolder(ctx, params)
}

func (r *FileRepository) RestoreFile(ctx context.Context, fileID pgtype.UUID) error {
	return r.q.RestoreFile(ctx, fileID)
}

func (r *FileRepository) SearchFiles(ctx context.Context, params database.SearchFilesParams) ([]database.File, error) {
	return r.q.SearchFiles(ctx, params)
}

func (r *FileRepository) SearchFolders(ctx context.Context, params database.SearchFoldersParams) ([]database.Folder, error) {
	return r.q.SearchFolders(ctx, params)
}

func (r *FileRepository) SearchStandaloneFolders(ctx context.Context, search pgtype.Text) ([]database.Folder, error) {
	return r.q.SearchStandaloneFolders(ctx, search)
}

func (r *FileRepository) SetFavorite(ctx context.Context, params database.SetFavoriteParams) (database.UserFilePreference, error) {
	return r.q.SetFavorite(ctx, params)
}

func (r *FileRepository) SoftDeleteFile(ctx context.Context, fileID pgtype.UUID) error {
	return r.q.SoftDeleteFile(ctx, fileID)
}

func (r *FileRepository) UpdateFileAIMetadata(ctx context.Context, params database.UpdateFileAIMetadataParams) (database.FileAiMetadatum, error) {
	return r.q.UpdateFileAIMetadata(ctx, params)
}

func (r *FileRepository) UpdateFileProperties(ctx context.Context, params database.UpdateFilePropertiesParams) (database.FileProperty, error) {
	return r.q.UpdateFileProperties(ctx, params)
}

func (r *FileRepository) UpdateFilePropertiesOriginalName(ctx context.Context, params database.UpdateFilePropertiesOriginalNameParams) (database.FileProperty, error) {
	return r.q.UpdateFilePropertiesOriginalName(ctx, params)
}

func (r *FileRepository) UpdateFileSharePermission(ctx context.Context, params database.UpdateFileSharePermissionParams) (database.FileShare, error) {
	return r.q.UpdateFileSharePermission(ctx, params)
}

func (r *FileRepository) UpdateFileSize(ctx context.Context, params database.UpdateFileSizeParams) (database.File, error) {
	return r.q.UpdateFileSize(ctx, params)
}

func (r *FileRepository) UpdateFileStorage(ctx context.Context, params database.UpdateFileStorageParams) (database.FileStorage, error) {
	return r.q.UpdateFileStorage(ctx, params)
}
