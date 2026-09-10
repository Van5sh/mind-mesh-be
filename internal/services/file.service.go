package services

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/utils"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

type FileService struct {
	repo   *repository.FileRepository
	guards *guards.FileGuard
}

func NewFileService(repo *repository.FileRepository, guard *guards.FileGuard) *FileService {
	return &FileService{
		repo:   repo,
		guards: guard,
	}
}

func (s *FileService) CreateFile(
	ctx context.Context,
	params database.CreateFileParams,
) (database.File, error) {
	if err := validators.ValidateUUID("file id", params.ID); err != nil {
		return database.File{}, err
	}

	if err := validators.ValidateFileName(params.Name); err != nil {
		return database.File{}, err
	}

	if params.FolderID.Valid {
		if err := validators.ValidateUUID(
			"folder id",
			params.FolderID,
		); err != nil {
			return database.File{}, err
		}

		if _, err := s.guards.EnsureFolderExists(
			ctx,
			params.FolderID,
		); err != nil {
			return database.File{},
				apperrors.NotFoundError("destination folder not found")
		}
	}

	name, err := s.resolveFileName(
		ctx,
		params.FolderID,
		params.Name,
	)
	if err != nil {
		return database.File{}, err
	}

	params.Name = name

	file, err := s.repo.CreateFile(ctx, params)
	if err != nil {
		return database.File{}, apperrors.InternalError("failed to create file", err)
	}

	return file, nil
}

func (s *FileService) GetFileByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.File, error) {
	if err := validators.ValidateUUID("file id", id); err != nil {
		return database.File{}, err
	}

	file, err := s.guards.EnsureFileExists(ctx, id)
	if err != nil {
		return database.File{},
			apperrors.NotFoundError("file not found")
	}

	return file, nil
}

func (s *FileService) GetFilesByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID(
		"project id",
		projectID,
	); err != nil {
		return nil, err
	}

	files, err := s.repo.GetFilesByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get files by project", err)
	}

	return files, nil
}

func (s *FileService) GetFilesByFolderID(
	ctx context.Context,
	folderID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("folder id", folderID); err != nil {
		return nil, err
	}

	if _, err := s.guards.EnsureFolderExists(ctx, folderID); err != nil {
		return nil, apperrors.NotFoundError("folder not found")
	}

	files, err := s.repo.GetFilesByFolderID(ctx, folderID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to get files by folder",
			err,
		)
	}

	return files, nil
}

func (s *FileService) DeleteFile(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("file id", id); err != nil {
		return err
	}

	if _, err := s.guards.EnsureFileExists(ctx, id); err != nil {
		return apperrors.NotFoundError("file not found")
	}

	if err := s.repo.DeleteFile(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete file", err)
	}

	return nil
}

func (s *FileService) resolveFileName(
	ctx context.Context,
	folderID pgtype.UUID,
	name string,
) (string, error) {
	originalName := name
	candidate := name

	for i := 1; ; i++ {
		exists, err := s.repo.FileNameExistsInFolder(
			ctx,
			database.FileNameExistsInFolderParams{
				FolderID: folderID,
				Name:     candidate,
			},
		)
		if err != nil {
			return "", apperrors.InternalError("failed to resolve file name", err)
		}

		if !exists {
			return candidate, nil
		}

		candidate = utils.IncrementFileName(
			originalName,
			i,
		)
	}
}

func (s *FileService) CreateFolder(
	ctx context.Context,
	params database.CreateFolderParams,
) (database.Folder, error) {
	if err := validators.ValidateUUID(
		"folder id",
		params.ID,
	); err != nil {
		return database.Folder{}, err
	}

	if err := validators.ValidateFolderName(
		params.Name,
	); err != nil {
		return database.Folder{}, err
	}

	if params.ProjectID.Valid {
		if err := validators.ValidateUUID(
			"project id",
			params.ProjectID,
		); err != nil {
			return database.Folder{}, err
		}
	}

	if params.ParentFolderID.Valid {
		if err := validators.ValidateUUID(
			"parent folder id",
			params.ParentFolderID,
		); err != nil {
			return database.Folder{}, err
		}

		parent, err := s.guards.EnsureFolderExists(
			ctx,
			params.ParentFolderID,
		)
		if err != nil {
			return database.Folder{},
				apperrors.NotFoundError("parent folder not found")
		}

		if params.ProjectID.Valid &&
			parent.ProjectID.Valid &&
			parent.ProjectID.Bytes != params.ProjectID.Bytes {
			return database.Folder{},
				apperrors.ConflictError(
					"parent folder belongs to another project",
				)
		}
	}

	name, err := s.resolveFolderName(
		ctx,
		params.ProjectID,
		params.ParentFolderID,
		params.Name,
	)
	if err != nil {
		return database.Folder{}, err
	}

	params.Name = name

	folder, err := s.repo.CreateFolder(ctx, params)
	if err != nil {
		return database.Folder{}, apperrors.InternalError("failed to create folder", err)
	}

	return folder, nil
}

func (s *FileService) GetFolderByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Folder, error) {
	if err := validators.ValidateUUID(
		"folder id",
		id,
	); err != nil {
		return database.Folder{}, err
	}

	folder, err := s.guards.EnsureFolderExists(ctx, id)
	if err != nil {
		return database.Folder{},
			apperrors.NotFoundError("folder not found")
	}

	return folder, nil
}

func (s *FileService) DeleteFolder(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID(
		"folder id",
		id,
	); err != nil {
		return err
	}

	if _, err := s.guards.EnsureFolderExists(
		ctx,
		id,
	); err != nil {
		return apperrors.NotFoundError("folder not found")
	}

	if err := s.repo.DeleteFolder(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete folder", err)
	}

	return nil
}

func (s *FileService) resolveFolderName(
	ctx context.Context,
	projectID pgtype.UUID,
	parentFolderID pgtype.UUID,
	name string,
) (string, error) {
	originalName := name
	candidate := name

	for i := 1; ; i++ {
		exists, err := s.repo.FolderNameExists(
			ctx,
			database.FolderNameExistsParams{
				ProjectID:      projectID,
				ParentFolderID: parentFolderID,
				Name:           candidate,
			},
		)
		if err != nil {
			return "", apperrors.InternalError("failed to resolve folder name", err)
		}

		if !exists {
			return candidate, nil
		}

		candidate = utils.IncrementFolderName(
			originalName,
			i,
		)
	}
}

func (s *FileService) GetFoldersByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID(
		"project id",
		projectID,
	); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetFoldersByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get folders by project", err)
	}

	return folders, nil
}

func (s *FileService) SetFavorite(
	ctx context.Context,
	params database.SetFavoriteParams,
) (database.UserFilePreference, error) {
	if err := validators.ValidateUUID(
		"file id",
		params.FileID,
	); err != nil {
		return database.UserFilePreference{}, err
	}

	if err := validators.ValidateUUID(
		"user id",
		params.UserID,
	); err != nil {
		return database.UserFilePreference{}, err
	}

	if _, err := s.guards.EnsureFileExists(
		ctx,
		params.FileID,
	); err != nil {
		return database.UserFilePreference{},
			apperrors.NotFoundError("file not found")
	}

	favorite, err := s.repo.SetFavorite(ctx, params)
	if err != nil {
		return database.UserFilePreference{},
			apperrors.InternalError("failed to set favorite", err)
	}

	return favorite, nil
}

func (s *FileService) GetFavorite(
	ctx context.Context,
	params database.GetUserFilePreferenceParams,
) (database.UserFilePreference, error) {
	if err := validators.ValidateUUID(
		"file id",
		params.FileID,
	); err != nil {
		return database.UserFilePreference{}, err
	}

	if err := validators.ValidateUUID(
		"user id",
		params.UserID,
	); err != nil {
		return database.UserFilePreference{}, err
	}

	if _, err := s.guards.EnsureFileExists(
		ctx,
		params.FileID,
	); err != nil {
		return database.UserFilePreference{},
			apperrors.NotFoundError("file not found")
	}

	favorite, err := s.repo.GetUserFilePreference(ctx, params)
	if err != nil {
		return database.UserFilePreference{},
			apperrors.InternalError("failed to get favorite", err)
	}

	return favorite, nil
}

func (s *FileService) MoveFile(
	ctx context.Context,
	params database.MoveFileParams,
) (database.File, error) {
	if err := validators.ValidateUUID(
		"file id",
		params.ID,
	); err != nil {
		return database.File{}, err
	}

	if err := validators.ValidateUUID(
		"destination folder id",
		params.FolderID,
	); err != nil {
		return database.File{}, err
	}

	if _, err := s.guards.EnsureFileExists(
		ctx,
		params.ID,
	); err != nil {
		return database.File{},
			apperrors.NotFoundError("file not found")
	}

	if _, err := s.guards.EnsureFolderExists(
		ctx,
		params.FolderID,
	); err != nil {
		return database.File{},
			apperrors.NotFoundError("destination folder not found")
	}

	file, err := s.repo.MoveFile(ctx, params)
	if err != nil {
		return database.File{},
			apperrors.InternalError("failed to move file", err)
	}

	return file, nil
}

func (s *FileService) RenameFile(
	ctx context.Context,
	params database.RenameFileParams,
) (database.File, error) {
	if err := validators.ValidateUUID(
		"file id",
		params.ID,
	); err != nil {
		return database.File{}, err
	}

	if err := validators.ValidateFileName(params.Name); err != nil {
		return database.File{}, err
	}

	file, err := s.guards.EnsureFileExists(ctx, params.ID)
	if err != nil {
		return database.File{},
			apperrors.NotFoundError("file not found")
	}

	name, err := s.resolveFileName(
		ctx,
		file.FolderID,
		params.Name,
	)
	if err != nil {
		return database.File{}, err
	}

	params.Name = name

	renamed, err := s.repo.RenameFile(ctx, params)
	if err != nil {
		return database.File{},
			apperrors.InternalError("failed to rename file", err)
	}

	return renamed, nil
}

func (s *FileService) MoveFolder(
	ctx context.Context,
	params database.MoveFolderParams,
) (database.Folder, error) {
	if err := validators.ValidateUUID(
		"folder id",
		params.ID,
	); err != nil {
		return database.Folder{}, err
	}

	if params.ParentFolderID.Valid {
		if err := validators.ValidateUUID(
			"parent folder id",
			params.ParentFolderID,
		); err != nil {
			return database.Folder{}, err
		}

		if _, err := s.guards.EnsureFolderExists(
			ctx,
			params.ParentFolderID,
		); err != nil {
			return database.Folder{},
				apperrors.NotFoundError("parent folder not found")
		}
	}

	if _, err := s.guards.EnsureFolderExists(
		ctx,
		params.ID,
	); err != nil {
		return database.Folder{},
			apperrors.NotFoundError("folder not found")
	}

	folder, err := s.repo.MoveFolder(ctx, params)
	if err != nil {
		return database.Folder{},
			apperrors.InternalError("failed to move folder", err)
	}

	return folder, nil
}

func (s *FileService) RenameFolder(
	ctx context.Context,
	params database.RenameFolderParams,
) (database.Folder, error) {
	if err := validators.ValidateUUID(
		"folder id",
		params.ID,
	); err != nil {
		return database.Folder{}, err
	}

	if err := validators.ValidateFolderName(params.Name); err != nil {
		return database.Folder{}, err
	}

	folder, err := s.guards.EnsureFolderExists(ctx, params.ID)
	if err != nil {
		return database.Folder{},
			apperrors.NotFoundError("folder not found")
	}

	name, err := s.resolveFolderName(
		ctx,
		folder.ProjectID,
		folder.ParentFolderID,
		params.Name,
	)
	if err != nil {
		return database.Folder{}, err
	}

	params.Name = name

	renamed, err := s.repo.RenameFolder(ctx, params)
	if err != nil {
		return database.Folder{},
			apperrors.InternalError("failed to rename folder", err)
	}

	return renamed, nil
}

func (s *FileService) GetRootFiles(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID(
		"project id",
		projectID,
	); err != nil {
		return nil, err
	}

	files, err := s.repo.GetRootFiles(ctx, projectID)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get root files", err)
	}

	return files, nil
}

func (s *FileService) GetRootFolders(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID(
		"project id",
		projectID,
	); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetRootFolders(ctx, projectID)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get root folders", err)
	}

	return folders, nil
}

func (s *FileService) GetFoldersByParentFolderID(
	ctx context.Context,
	parentFolderID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID(
		"parent folder id",
		parentFolderID,
	); err != nil {
		return nil, err
	}

	if _, err := s.guards.EnsureFolderExists(
		ctx,
		parentFolderID,
	); err != nil {
		return nil, apperrors.NotFoundError("parent folder not found")
	}

	folders, err := s.repo.GetChildFolders(
		ctx,
		parentFolderID,
	)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get child folders", err)
	}

	return folders, nil
}

func (s *FileService) GetFolderPath(
	ctx context.Context,
	folderID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID(
		"folder id",
		folderID,
	); err != nil {
		return nil, err
	}

	if _, err := s.guards.EnsureFolderExists(
		ctx,
		folderID,
	); err != nil {
		return nil, apperrors.NotFoundError("folder not found")
	}

	path := make([]database.Folder, 0)
	currentID := folderID

	for currentID.Valid {
		folder, err := s.repo.GetFolderByID(ctx, currentID)
		if err != nil {
			return nil, apperrors.InternalError("failed to get folder path", err)
		}

		path = append([]database.Folder{folder}, path...)
		currentID = folder.ParentFolderID
	}

	return path, nil
}
