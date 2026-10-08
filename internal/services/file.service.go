package services

import (
	"context"
	"errors"
	"fmt"
	"io"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/services/aws"
	"example/hello/internal/utils"
	"example/hello/internal/validators"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type FileService struct {
	repo   *repository.FileRepository
	guards *guards.FileGuard
	s3     *aws.S3Service
	sqs    *aws.SQSService
}

func NewFileService(
	repo *repository.FileRepository,
	guard *guards.FileGuard,
	s3Service *aws.S3Service,
	sqsService *aws.SQSService,
) *FileService {
	return &FileService{
		repo:   repo,
		guards: guard,
		s3:     s3Service,
		sqs:    sqsService,
	}
}

// supportedProcessingContentTypes must stay in sync with the content
// types ai/extraction/factory.py knows how to extract text from. Files
// of other types would just be queued, downloaded, and marked FAILED
// by the worker, so reject them up front instead.
var supportedProcessingContentTypes = map[string]bool{
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain": true,
}

func (s *FileService) CreateFile(
	ctx context.Context,
	params database.CreateFileParams,
	body io.Reader,
	contentType string,
) (database.File, error) {
	if err := validators.ValidateFileName(params.Name); err != nil {
		return database.File{}, err
	}
	// ProjectID is optional: unset means a personal file, uploaded to the
	// user's own drive rather than any project. A personal file can sit in
	// a personal folder owned by the same uploader - the destination-folder
	// check below (folder.ProjectID != params.ProjectID, plus an owner
	// check for the personal case) covers both that and the project case.
	if params.ProjectID.Valid {
		if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
			return database.File{}, err
		}
	}
	if err := validators.ValidateUUID("uploaded by", params.UploadedBy); err != nil {
		return database.File{}, err
	}
	if !supportedProcessingContentTypes[contentType] {
		return database.File{}, apperrors.Validation(
			fmt.Sprintf("unsupported file type: %s (supported: pdf, docx, plain text)", contentType),
		)
	}
	if params.FolderID.Valid {
		if err := validators.ValidateUUID(
			"folder id",
			params.FolderID,
		); err != nil {
			return database.File{}, err
		}

		folder, err := s.guards.EnsureFolderExists(
			ctx,
			params.FolderID,
		)
		if err != nil {
			return database.File{},
				apperrors.NotFoundError("destination folder not found")
		}

		if folder.ProjectID != params.ProjectID {
			return database.File{},
				apperrors.Validation("destination folder does not belong to the given project")
		}
		if !params.ProjectID.Valid && folder.OwnerID != params.UploadedBy {
			return database.File{},
				apperrors.Validation("destination folder belongs to another user's personal drive")
		}
	}

	name, err := s.resolveFileName(
		ctx,
		params.ProjectID,
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

	// file_properties doesn't exist yet at this point (created further
	// below), so fileS3Key necessarily falls back to file.Name here - which
	// is correct, since file.Name is still the original, never-renamed
	// value.
	s3Key := s.fileS3Key(ctx, file)
	if err := s.s3.S3Upload(
		ctx,
		s3Key,
		body,
		contentType,
	); err != nil {
		// Important: don't leave a file looking successfully uploaded.
		_ = s.repo.DeleteFile(ctx, file.ID)
		return database.File{},
			apperrors.InternalError("failed to upload file to storage", err)
	}

	// Create the AI-metadata row up front (status PENDING) so the worker's
	// status updates - which only UPDATE, never INSERT - have a row to hit.
	if _, err := s.repo.CreateFileAIMetadata(ctx, database.CreateFileAIMetadataParams{
		FileID:          file.ID,
		EmbeddingSynced: pgtype.Bool{Bool: false, Valid: true},
	}); err != nil {
		_ = s.repo.DeleteFile(ctx, file.ID)
		return database.File{},
			apperrors.InternalError("failed to initialize file processing status", err)
	}

	// Create the properties row up front too - trashFile/restoreFile write
	// to it (file_properties.deleted_at), and File.properties is non-null,
	// so every file needs one from the moment it exists.
	if _, err := s.repo.CreateFileProperties(ctx, database.CreateFilePropertiesParams{
		FileID:       file.ID,
		OriginalName: pgtype.Text{String: file.Name, Valid: true},
		IsIndexed:    false,
	}); err != nil {
		_ = s.repo.DeleteFile(ctx, file.ID)
		return database.File{},
			apperrors.InternalError("failed to initialize file properties", err)
	}

	// Personal files (no project) aren't queued for AI processing: RAG/chat
	// grounding is project-scoped, so there's nothing for the worker to
	// associate the embeddings with. Its file_ai_metadata row (created
	// above) is left at PENDING - there's no misleading "FAILED" state, it
	// simply never gets processed. Revisit if a project-less chat/RAG
	// feature is ever added.
	if !file.ProjectID.Valid {
		return file, nil
	}

	// Queue processing job. Field names must match ai/worker/main.py's
	// ProcessingJob model exactly (see FileProcessingJob for details).
	job := aws.FileProcessingJob{
		JobID:       uuid.NewString(),
		FileID:      file.ID.String(),
		ProjectID:   file.ProjectID.String(),
		StorageKey:  s3Key,
		ContentType: contentType,
	}

	if err := s.sqs.SendFileProcessingJob(ctx, job); err != nil {
		_, _ = s.repo.FailFileProcessing(ctx, database.FailFileProcessingParams{
			FileID:       file.ID,
			ErrorMessage: pgtype.Text{String: "failed to queue for processing", Valid: true},
		})
		return database.File{},
			apperrors.InternalError("failed to queue file processing", err)
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

// GetFilesByProjectIDs fetches the files of many projects in one query and
// returns them grouped by project ID (every requested ID is present in the
// map, with a nil slice when that project has no files). It exists for the
// Project.files dataloader - see graph/loaders.
func (s *FileService) GetFilesByProjectIDs(
	ctx context.Context,
	projectIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.File, error) {
	for _, id := range projectIDs {
		if err := validators.ValidateUUID("project id", id); err != nil {
			return nil, err
		}
	}

	files, err := s.repo.GetFilesByProjectIDs(ctx, projectIDs)
	if err != nil {
		return nil, apperrors.InternalError("failed to get files by projects", err)
	}

	grouped := make(map[pgtype.UUID][]database.File, len(projectIDs))
	for _, id := range projectIDs {
		grouped[id] = nil
	}
	for _, f := range files {
		grouped[f.ProjectID] = append(grouped[f.ProjectID], f)
	}

	return grouped, nil
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

	file, err := s.guards.EnsureFileExists(ctx, id)
	if err != nil {
		return apperrors.NotFoundError("file not found")
	}

	if err := s.s3.S3Delete(ctx, s.fileS3Key(ctx, file)); err != nil {
		return apperrors.InternalError("failed to delete file from storage", err)
	}

	if err := s.repo.DeleteFile(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete file", err)
	}

	return nil
}

// fileS3Key reconstructs the S3 key CreateFile originally uploaded this
// file under. It can't just use file.Name - RenameFile only updates the
// files row, it never touches S3, so a renamed file's key would otherwise
// go stale. file_properties.original_name is frozen at upload time (never
// updated by RenameFile) and is exactly the name CreateFile built the key
// from, so it survives renames correctly. Falls back to the current name
// only if that row is somehow missing (pre-dates the backfill migration or
// was never created) - best effort rather than blocking the delete.
func (s *FileService) fileS3Key(ctx context.Context, file database.File) string {
	name := file.Name
	if props, err := s.repo.GetFileProperties(ctx, file.ID); err == nil && props.OriginalName.Valid {
		name = props.OriginalName.String
	}

	if file.ProjectID.Valid {
		return fmt.Sprintf("projects/%s/files/%s/%s", file.ProjectID, file.ID, name)
	}
	return fmt.Sprintf("personal/%s/files/%s/%s", file.UploadedBy, file.ID, name)
}

// TrashFile soft-deletes a file (file_properties.deleted_at) - unlike
// DeleteFile, this is reversible via RestoreFile. The files row itself is
// untouched, so the pre-fetched row is still accurate to return.
func (s *FileService) TrashFile(
	ctx context.Context,
	id pgtype.UUID,
) (database.File, error) {
	if err := validators.ValidateUUID("file id", id); err != nil {
		return database.File{}, err
	}

	file, err := s.guards.EnsureFileExists(ctx, id)
	if err != nil {
		return database.File{}, apperrors.NotFoundError("file not found")
	}

	if err := s.repo.SoftDeleteFile(ctx, id); err != nil {
		return database.File{}, apperrors.InternalError("failed to trash file", err)
	}

	return file, nil
}

// RestoreFile undoes TrashFile.
func (s *FileService) RestoreFile(
	ctx context.Context,
	id pgtype.UUID,
) (database.File, error) {
	if err := validators.ValidateUUID("file id", id); err != nil {
		return database.File{}, err
	}

	file, err := s.guards.EnsureFileExists(ctx, id)
	if err != nil {
		return database.File{}, apperrors.NotFoundError("file not found")
	}

	if err := s.repo.RestoreFile(ctx, id); err != nil {
		return database.File{}, apperrors.InternalError("failed to restore file", err)
	}

	return file, nil
}

// GetTrashedFiles lists a project's soft-deleted files.
func (s *FileService) GetTrashedFiles(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetDeletedFiles(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get trashed files", err)
	}

	return files, nil
}

// GetTrashedPersonalFiles lists a user's soft-deleted personal (project-less)
// files.
func (s *FileService) GetTrashedPersonalFiles(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetTrashedPersonalFiles(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get trashed personal files", err)
	}

	return files, nil
}

func (s *FileService) resolveFileName(
	ctx context.Context,
	projectID pgtype.UUID,
	folderID pgtype.UUID,
	name string,
) (string, error) {
	originalName := name
	candidate := name

	for i := 1; ; i++ {
		exists, err := s.repo.FileNameExistsInFolder(
			ctx,
			database.FileNameExistsInFolderParams{
				ProjectID: projectID,
				FolderID:  folderID,
				Name:      candidate,
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
	if err := validators.ValidateFolderName(
		params.Name,
	); err != nil {
		return database.Folder{}, err
	}

	// ProjectID is optional: unset means a personal folder, in the owning
	// user's own drive rather than any project. A personal folder needs an
	// owner (that's what its authorization and listing queries key off);
	// a project folder gets one too, for free, as useful metadata.
	if params.ProjectID.Valid {
		if err := validators.ValidateUUID(
			"project id",
			params.ProjectID,
		); err != nil {
			return database.Folder{}, err
		}
	}
	if err := validators.ValidateUUID("owner id", params.OwnerID); err != nil {
		return database.Folder{}, err
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

		// A folder's space (project, or personal-to-an-owner) must match
		// its parent's - a personal folder can't sit inside a project
		// folder or another user's personal folder, and vice versa.
		if parent.ProjectID.Valid != params.ProjectID.Valid ||
			(parent.ProjectID.Valid && parent.ProjectID.Bytes != params.ProjectID.Bytes) {
			return database.Folder{},
				apperrors.Validation(
					"parent folder does not belong to the given project",
				)
		}
		if !params.ProjectID.Valid && parent.OwnerID != params.OwnerID {
			return database.Folder{},
				apperrors.Validation(
					"parent folder belongs to another user's personal drive",
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

// TrashFolder soft-deletes a folder (folders.deleted_at) - unlike
// DeleteFolder, this is reversible via RestoreFolder. It does not cascade to
// the folder's contents; a file/subfolder inside a trashed folder does not
// itself become trashed.
func (s *FileService) TrashFolder(
	ctx context.Context,
	id pgtype.UUID,
) (database.Folder, error) {
	if err := validators.ValidateUUID("folder id", id); err != nil {
		return database.Folder{}, err
	}

	folder, err := s.guards.EnsureFolderExists(ctx, id)
	if err != nil {
		return database.Folder{}, apperrors.NotFoundError("folder not found")
	}

	if err := s.repo.SoftDeleteFolder(ctx, id); err != nil {
		return database.Folder{}, apperrors.InternalError("failed to trash folder", err)
	}

	return folder, nil
}

// RestoreFolder undoes TrashFolder.
func (s *FileService) RestoreFolder(
	ctx context.Context,
	id pgtype.UUID,
) (database.Folder, error) {
	if err := validators.ValidateUUID("folder id", id); err != nil {
		return database.Folder{}, err
	}

	folder, err := s.guards.EnsureFolderExists(ctx, id)
	if err != nil {
		return database.Folder{}, apperrors.NotFoundError("folder not found")
	}

	if err := s.repo.RestoreFolder(ctx, id); err != nil {
		return database.Folder{}, apperrors.InternalError("failed to restore folder", err)
	}

	return folder, nil
}

// GetTrashedFolders lists a project's soft-deleted folders.
func (s *FileService) GetTrashedFolders(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetTrashedFolders(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get trashed folders", err)
	}

	return folders, nil
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

	existing, err := s.guards.EnsureFileExists(
		ctx,
		params.ID,
	)
	if err != nil {
		return database.File{},
			apperrors.NotFoundError("file not found")
	}

	// A zero-value FolderID means the file is being moved to the
	// project root, which is valid and has no folder to check.
	if params.FolderID.Valid {
		folder, err := s.guards.EnsureFolderExists(
			ctx,
			params.FolderID,
		)
		if err != nil {
			return database.File{},
				apperrors.NotFoundError("destination folder not found")
		}

		if folder.ProjectID != existing.ProjectID {
			return database.File{},
				apperrors.Validation("destination folder does not belong to the file's project")
		}
		if !existing.ProjectID.Valid && folder.OwnerID != existing.UploadedBy {
			return database.File{},
				apperrors.Validation("destination folder belongs to another user's personal drive")
		}
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
		file.ProjectID,
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

	existing, err := s.guards.EnsureFolderExists(
		ctx,
		params.ID,
	)
	if err != nil {
		return database.Folder{},
			apperrors.NotFoundError("folder not found")
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

		// MoveFolder only ever changes parent_folder_id, never which
		// project/owner a folder belongs to - the new parent must be in
		// that same space.
		if parent.ProjectID.Valid != existing.ProjectID.Valid ||
			(parent.ProjectID.Valid && parent.ProjectID.Bytes != existing.ProjectID.Bytes) {
			return database.Folder{},
				apperrors.Validation("destination folder belongs to another project")
		}
		if !existing.ProjectID.Valid && parent.OwnerID != existing.OwnerID {
			return database.Folder{},
				apperrors.Validation("destination folder belongs to another user's personal drive")
		}
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

// GetPersonalRootFolders returns a user's top-level personal (project-less)
// folders - the personal-drive counterpart to GetRootFolders.
func (s *FileService) GetPersonalRootFolders(
	ctx context.Context,
	ownerID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID("owner id", ownerID); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetStandaloneRootFolders(ctx, ownerID)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get personal root folders", err)
	}

	return folders, nil
}

// GetPersonalFolders returns every personal folder a user owns, flat - the
// personal-drive counterpart to GetFoldersByProjectID.
func (s *FileService) GetPersonalFolders(
	ctx context.Context,
	ownerID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID("owner id", ownerID); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetPersonalFolders(ctx, ownerID)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get personal folders", err)
	}

	return folders, nil
}

// GetTrashedPersonalFolders lists a user's soft-deleted personal folders.
func (s *FileService) GetTrashedPersonalFolders(
	ctx context.Context,
	ownerID pgtype.UUID,
) ([]database.Folder, error) {
	if err := validators.ValidateUUID("owner id", ownerID); err != nil {
		return nil, err
	}

	folders, err := s.repo.GetTrashedPersonalFolders(ctx, ownerID)
	if err != nil {
		return nil,
			apperrors.InternalError("failed to get trashed personal folders", err)
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

// GetFileStorage returns the storage record for a file. Storage is
// optional (a file's bytes may not have finished uploading yet), so a
// missing row is reported as apperrors.NotFound rather than Internal.
func (s *FileService) GetFileStorage(
	ctx context.Context,
	fileID pgtype.UUID,
) (database.FileStorage, error) {
	if err := validators.ValidateUUID("file id", fileID); err != nil {
		return database.FileStorage{}, err
	}

	storage, err := s.repo.GetFileStorage(ctx, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileStorage{}, apperrors.NotFoundError("file storage not found")
		}
		return database.FileStorage{}, apperrors.InternalError("failed to fetch file storage", err)
	}

	return storage, nil
}

// GetFileDownloadURL returns a presigned, time-limited URL for retrieving
// an uploaded file's bytes directly from S3.
func (s *FileService) GetFileDownloadURL(
	ctx context.Context,
	objectKey string,
) (string, error) {
	url, err := s.s3.S3GetUrl(ctx, objectKey)
	if err != nil {
		return "", apperrors.InternalError("failed to generate file download URL", err)
	}
	return url, nil
}

// GetFileProperties returns the properties record for a file.
func (s *FileService) GetFileProperties(
	ctx context.Context,
	fileID pgtype.UUID,
) (database.FileProperty, error) {
	if err := validators.ValidateUUID("file id", fileID); err != nil {
		return database.FileProperty{}, err
	}

	properties, err := s.repo.GetFileProperties(ctx, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileProperty{}, apperrors.NotFoundError("file properties not found")
		}
		return database.FileProperty{}, apperrors.InternalError("failed to fetch file properties", err)
	}

	return properties, nil
}

// GetFilePropertiesByFileIDs fetches the properties row of many files in one
// query. It exists for the File.properties dataloader.
func (s *FileService) GetFilePropertiesByFileIDs(
	ctx context.Context,
	fileIDs []pgtype.UUID,
) ([]database.FileProperty, error) {
	return fetchRows(ctx, fileIDs, "file properties", s.repo.GetFilePropertiesByFileIDs)
}

// GetFileAIMetadata returns the AI metadata record for a file. AI
// metadata is optional until the AI worker has processed the file.
func (s *FileService) GetFileAIMetadata(
	ctx context.Context,
	fileID pgtype.UUID,
) (database.FileAiMetadatum, error) {
	if err := validators.ValidateUUID("file id", fileID); err != nil {
		return database.FileAiMetadatum{}, err
	}

	metadata, err := s.repo.GetFileAIMetadata(ctx, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileAiMetadatum{}, apperrors.NotFoundError("file AI metadata not found")
		}
		return database.FileAiMetadatum{}, apperrors.InternalError("failed to fetch file AI metadata", err)
	}

	return metadata, nil
}

// GetFileSharesByFileID returns every share record for a file.
func (s *FileService) GetFileSharesByFileID(
	ctx context.Context,
	fileID pgtype.UUID,
) ([]database.FileShare, error) {
	if err := validators.ValidateUUID("file id", fileID); err != nil {
		return nil, err
	}

	shares, err := s.repo.GetFileSharesByFileID(ctx, fileID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch file shares", err)
	}

	return shares, nil
}

// GetFileSharesBySharedWith returns every share extended to a user.
func (s *FileService) GetFileSharesBySharedWith(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.FileShare, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	shares, err := s.repo.GetFileSharesBySharedWith(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch shared files", err)
	}

	return shares, nil
}

// GetFoldersByProjectIDs fetches the folders of many projects in one query,
// grouped by project ID. It exists for the Project.folders dataloader.
func (s *FileService) GetFoldersByProjectIDs(
	ctx context.Context,
	projectIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.Folder, error) {
	return fetchGrouped(ctx, projectIDs, "folders by projects",
		s.repo.GetFoldersByProjectIDs,
		func(f database.Folder) pgtype.UUID { return f.ProjectID })
}

// GetFilesByFolderIDs fetches the files inside many folders in one query,
// grouped by folder ID. It exists for the Folder.files dataloader.
func (s *FileService) GetFilesByFolderIDs(
	ctx context.Context,
	folderIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.File, error) {
	return fetchGrouped(ctx, folderIDs, "files by folders",
		s.repo.GetFilesByFolderIDs,
		func(f database.File) pgtype.UUID { return f.FolderID })
}

// GetFoldersByParentFolderIDs fetches the child folders of many folders in
// one query, grouped by parent folder ID. It exists for the
// Folder.childFolders dataloader.
func (s *FileService) GetFoldersByParentFolderIDs(
	ctx context.Context,
	parentIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.Folder, error) {
	return fetchGrouped(ctx, parentIDs, "child folders by parent folders",
		s.repo.GetFoldersByParentFolderIDs,
		func(f database.Folder) pgtype.UUID { return f.ParentFolderID })
}

// GetFoldersByIDs fetches many folders in one query (IDs with no folder are
// absent). It exists for the Folder.parentFolder dataloader.
func (s *FileService) GetFoldersByIDs(
	ctx context.Context,
	ids []pgtype.UUID,
) ([]database.Folder, error) {
	return fetchRows(ctx, ids, "folders", s.repo.GetFoldersByIDs)
}

// GetFileAIMetadataByFileIDs fetches the AI metadata of many files in one
// query (files with none are absent). It exists for the File.aiMetadata
// dataloader.
func (s *FileService) GetFileAIMetadataByFileIDs(
	ctx context.Context,
	fileIDs []pgtype.UUID,
) ([]database.FileAiMetadatum, error) {
	return fetchRows(ctx, fileIDs, "file AI metadata", s.repo.GetFileAIMetadataByFileIDs)
}

// GetFileStoragesByFileIDs fetches the storage rows of many files in one
// query (files with none are absent). It exists for the File.storage
// dataloader.
func (s *FileService) GetFileStoragesByFileIDs(
	ctx context.Context,
	fileIDs []pgtype.UUID,
) ([]database.FileStorage, error) {
	return fetchRows(ctx, fileIDs, "file storage", s.repo.GetFileStoragesByFileIDs)
}

// GetFileSharesByFileIDs fetches the shares of many files in one query,
// grouped by file ID. It exists for the File.shares dataloader.
func (s *FileService) GetFileSharesByFileIDs(
	ctx context.Context,
	fileIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.FileShare, error) {
	return fetchGrouped(ctx, fileIDs, "file shares by files",
		s.repo.GetFileSharesByFileIDs,
		func(s database.FileShare) pgtype.UUID { return s.FileID })
}

// GetFileReferencesByMessage gets the files referenced by one message. Only
// used as the ChatMessage.referencedFiles fallback when no dataloader is on
// the context - see GetFileReferencesByMessageIDs for the batched path.
func (s *FileService) GetFileReferencesByMessage(
	ctx context.Context,
	messageID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("message id", messageID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetMessageFileReferences(ctx, messageID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch referenced files", err)
	}
	return files, nil
}

// CreateMessageFileReference attaches a file reference (a #filename token
// resolved to an ID by the client, see CreateChatMessageInput) to a chat
// message. The caller validates that the file is actually visible in the
// message's chat before calling this - this just writes the row.
func (s *FileService) CreateMessageFileReference(
	ctx context.Context,
	params database.CreateMessageFileReferenceParams,
) (database.MessageFileReference, error) {
	if err := validators.ValidateUUID("message id", params.MessageID); err != nil {
		return database.MessageFileReference{}, err
	}
	if err := validators.ValidateUUID("file id", params.FileID); err != nil {
		return database.MessageFileReference{}, err
	}

	reference, err := s.repo.CreateMessageFileReference(ctx, params)
	if err != nil {
		return database.MessageFileReference{}, apperrors.InternalError(
			"failed to create message file reference",
			err,
		)
	}
	return reference, nil
}

// GetFileReferencesByMessageIDs fetches the files referenced by many
// messages in one query, grouped by message ID. It exists for the
// ChatMessage.referencedFiles dataloader.
func (s *FileService) GetFileReferencesByMessageIDs(
	ctx context.Context,
	messageIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.File, error) {
	grouped, err := fetchGrouped(ctx, messageIDs, "file references by messages",
		s.repo.GetFileReferencesForMessages,
		func(r database.GetFileReferencesForMessagesRow) pgtype.UUID { return r.MessageID })
	if err != nil {
		return nil, err
	}

	result := make(map[pgtype.UUID][]database.File, len(grouped))
	for messageID, rows := range grouped {
		files := make([]database.File, 0, len(rows))
		for _, row := range rows {
			files = append(files, database.File{
				ID:         row.ID,
				FolderID:   row.FolderID,
				ProjectID:  row.ProjectID,
				UploadedBy: row.UploadedBy,
				Name:       row.Name,
				Size:       row.Size,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			})
		}
		result[messageID] = files
	}
	return result, nil
}

// GetFileSharesBySharedWithIDs fetches the shares made with many users in one
// query, grouped by recipient. It exists for the User.sharedFiles dataloader.
func (s *FileService) GetFileSharesBySharedWithIDs(
	ctx context.Context,
	userIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.FileShare, error) {
	return fetchGrouped(ctx, userIDs, "file shares by recipients",
		s.repo.GetFileSharesBySharedWithIDs,
		func(s database.FileShare) pgtype.UUID { return s.SharedWith })
}

// GetPersonalFiles returns a user's root-level files that have no project -
// their personal drive space.
func (s *FileService) GetPersonalFiles(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetPersonalFiles(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get personal files", err)
	}
	return files, nil
}

// GetFileShareByUser returns the share row granting userID access to
// fileID, or a NotFound error if the file hasn't been shared with them.
// Used to authorize access to a personal (project-less) file that isn't
// theirs - the same FileShare mechanism project files could already use.
func (s *FileService) GetFileShareByUser(
	ctx context.Context,
	fileID pgtype.UUID,
	userID pgtype.UUID,
) (database.FileShare, error) {
	share, err := s.repo.GetFileShareByFileAndSharedWith(ctx, database.GetFileShareByFileAndSharedWithParams{
		FileID:     fileID,
		SharedWith: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileShare{}, apperrors.NotFoundError("file not shared with this user")
		}
		return database.FileShare{}, apperrors.InternalError("failed to check file share", err)
	}
	return share, nil
}

// CreateFileShare grants a user access to a file. Re-sharing with someone
// who already has a share updates their permission instead of hitting the
// file_id+shared_with unique constraint - a share is "this user's access
// level to this file", not a log of grants.
func (s *FileService) CreateFileShare(
	ctx context.Context,
	params database.FileShareParams,
) (database.FileShare, error) {
	if err := validators.ValidateUUID("file id", params.FileID); err != nil {
		return database.FileShare{}, err
	}
	if err := validators.ValidateUUID("shared by", params.SharedBy); err != nil {
		return database.FileShare{}, err
	}
	if err := validators.ValidateUUID("shared with", params.SharedWith); err != nil {
		return database.FileShare{}, err
	}
	if params.SharedBy == params.SharedWith {
		return database.FileShare{}, apperrors.Validation("cannot share a file with yourself")
	}

	if _, err := s.guards.EnsureFileExists(ctx, params.FileID); err != nil {
		return database.FileShare{}, apperrors.NotFoundError("file not found")
	}

	existing, err := s.repo.GetFileShareByFileAndSharedWith(ctx, database.GetFileShareByFileAndSharedWithParams{
		FileID:     params.FileID,
		SharedWith: params.SharedWith,
	})
	if err == nil {
		updated, err := s.repo.UpdateFileSharePermission(ctx, database.UpdateFileSharePermissionParams{
			ID:         existing.ID,
			Permission: params.Permission,
		})
		if err != nil {
			return database.FileShare{}, apperrors.InternalError("failed to update file share", err)
		}
		return updated, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return database.FileShare{}, apperrors.InternalError("failed to check existing file share", err)
	}

	share, err := s.repo.FileShare(ctx, params)
	if err != nil {
		return database.FileShare{}, apperrors.InternalError("failed to share file", err)
	}
	return share, nil
}

// GetFileShareByID returns a single share record by its own ID.
func (s *FileService) GetFileShareByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.FileShare, error) {
	if err := validators.ValidateUUID("file share id", id); err != nil {
		return database.FileShare{}, err
	}

	share, err := s.repo.GetFileShareByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileShare{}, apperrors.NotFoundError("file share not found")
		}
		return database.FileShare{}, apperrors.InternalError("failed to fetch file share", err)
	}
	return share, nil
}

// GetFavoriteFilesByProject lists a user's starred files within a project.
func (s *FileService) GetFavoriteFilesByProject(
	ctx context.Context,
	userID pgtype.UUID,
	projectID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetFavoriteFilesByProject(ctx, database.GetFavoriteFilesByProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, apperrors.InternalError("failed to get favorite files", err)
	}
	return files, nil
}

// GetFavoritePersonalFiles lists a user's starred personal (project-less)
// files - the personal-drive counterpart to GetFavoriteFilesByProject.
func (s *FileService) GetFavoritePersonalFiles(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.File, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	files, err := s.repo.GetFavoritePersonalFiles(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get favorite personal files", err)
	}
	return files, nil
}

// UpdateFileSharePermission changes an existing share's permission level.
func (s *FileService) UpdateFileSharePermission(
	ctx context.Context,
	params database.UpdateFileSharePermissionParams,
) (database.FileShare, error) {
	if err := validators.ValidateUUID("file share id", params.ID); err != nil {
		return database.FileShare{}, err
	}

	if _, err := s.repo.GetFileShareByID(ctx, params.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.FileShare{}, apperrors.NotFoundError("file share not found")
		}
		return database.FileShare{}, apperrors.InternalError("failed to fetch file share", err)
	}

	share, err := s.repo.UpdateFileSharePermission(ctx, params)
	if err != nil {
		return database.FileShare{}, apperrors.InternalError("failed to update file share", err)
	}
	return share, nil
}

// DeleteFileShare revokes a share, removing the recipient's access.
func (s *FileService) DeleteFileShare(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("file share id", id); err != nil {
		return err
	}

	if _, err := s.repo.GetFileShareByID(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.NotFoundError("file share not found")
		}
		return apperrors.InternalError("failed to fetch file share", err)
	}

	if err := s.repo.DeleteFileShare(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete file share", err)
	}
	return nil
}
