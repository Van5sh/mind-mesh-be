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

func NewFileService(repo *repository.FileRepository) *FileService {
	return &FileService{
		repo:   repo,
		guards: guards.NewFileGuard(repo),
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
		return database.File{}, err
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
		return nil, err
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
		return err
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
			return "", err
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
		return database.Folder{}, err
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
		return err
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
			return "", err
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