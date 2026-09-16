package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type FileGuard struct {
	repo *repository.FileRepository
}

func NewFileGuard(repo *repository.FileRepository) *FileGuard {
	return &FileGuard{repo: repo}
}

func (g *FileGuard) EnsureFileExists(ctx context.Context, fileID pgtype.UUID) (database.File, error) {
	file, err := g.repo.GetFileByID(ctx, fileID)
	if err != nil {
		if isNoRows(err) {
			return database.File{}, apperrors.NotFoundError("file not found")
		}
		return database.File{}, apperrors.InternalError("failed to fetch file", err)
	}
	return file, nil
}

func (g *FileGuard) EnsureFolderExists(ctx context.Context, folderID pgtype.UUID) (database.Folder, error) {
	folder, err := g.repo.GetFolderByID(ctx, folderID)
	if err != nil {
		if isNoRows(err) {
			return database.Folder{}, apperrors.NotFoundError("folder not found")
		}
		return database.Folder{}, apperrors.InternalError("failed to fetch folder", err)
	}
	return folder, nil
}

func (g *FileGuard) EnsureProjectFileExists(ctx context.Context, projectID, fileID pgtype.UUID) (database.File, error) {
	file, err := g.repo.GetProjectFileByID(ctx, database.GetProjectFileByIDParams{
		ID:        fileID,
		ProjectID: projectID,
	})
	if err != nil {
		if isNoRows(err) {
			return database.File{}, apperrors.NotFoundError("project file not found")
		}
		return database.File{}, apperrors.InternalError("failed to fetch project file", err)
	}
	return file, nil
}

func (g *FileGuard) EnsureProjectFolderExists(ctx context.Context, projectID, folderID pgtype.UUID) (database.Folder, error) {
	folder, err := g.repo.GetProjectFolderByID(ctx, database.GetProjectFolderByIDParams{
		ID:        folderID,
		ProjectID: projectID,
	})
	if err != nil {
		if isNoRows(err) {
			return database.Folder{}, apperrors.NotFoundError("project folder not found")
		}
		return database.Folder{}, apperrors.InternalError("failed to fetch project folder", err)
	}
	return folder, nil
}

func (g *FileGuard) EnsureFileBelongsToProject(ctx context.Context, projectID, fileID pgtype.UUID) error {
	_, err := g.EnsureProjectFileExists(ctx, projectID, fileID)
	return err
}

func (g *FileGuard) EnsureFolderBelongsToProject(ctx context.Context, projectID, folderID pgtype.UUID) error {
	_, err := g.EnsureProjectFolderExists(ctx, projectID, folderID)
	return err
}

func (g *FileGuard) EnsureFileNameAvailable(ctx context.Context, projectID, folderID pgtype.UUID, name string) error {
	exists, err := g.repo.CheckFileNameExists(ctx, database.CheckFileNameExistsParams{
		FolderID: folderID,
		Name:     name,
	})
	if err != nil {
		return apperrors.InternalError("failed to check file name", err)
	}
	if exists {
		return apperrors.ConflictError("file name already exists")
	}
	return nil
}

func (g *FileGuard) EnsureFolderNameAvailable(ctx context.Context, projectID, parentFolderID pgtype.UUID, name string) error {
	exists, err := g.repo.CheckFolderNameExists(ctx, database.CheckFolderNameExistsParams{
		ProjectID:      projectID,
		ParentFolderID: parentFolderID,
		Name:           name,
	})
	if err != nil {
		return apperrors.InternalError("failed to check folder name", err)
	}
	if exists {
		return apperrors.ConflictError("folder name already exists")
	}
	return nil
}

func (g *FileGuard) EnsureFileShareExists(ctx context.Context, fileID, sharedWith pgtype.UUID) (database.FileShare, error) {
	share, err := g.repo.GetFileShareByFileAndSharedWith(ctx, database.GetFileShareByFileAndSharedWithParams{
		FileID:     fileID,
		SharedWith: sharedWith,
	})
	if err != nil {
		if isNoRows(err) {
			return database.FileShare{}, apperrors.NotFoundError("file share not found")
		}
		return database.FileShare{}, apperrors.InternalError("failed to fetch file share", err)
	}
	return share, nil
}
