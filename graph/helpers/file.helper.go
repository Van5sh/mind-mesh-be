package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

// FileToModel maps a database.File row to the GraphQL File model. The
// schema declares File.properties as non-null (FileProperties!), so a
// stub referencing the file itself is always attached even though the
// file_properties row isn't loaded here.
func FileToModel(file database.File) *model.File {
	var folder *model.Folder

	if file.FolderID.Valid {
		folder = &model.Folder{
			ID: file.FolderID.String(),
		}
	}

	return &model.File{
		ID:     file.ID.String(),
		Name:   file.Name,
		Size:   int(file.Size),
		Folder: folder,
		Properties: &model.FileProperties{
			File: &model.File{ID: file.ID.String()},
		},
		CreatedAt: file.CreatedAt.Time,
		UpdatedAt: file.UpdatedAt.Time,
	}
}

// FolderToModel maps a database.Folder row to the GraphQL Folder model.
// ParentFolder/ChildFolders/Files are resolved separately (see
// folder.resolvers.go-style methods on *folderResolver), since gqlgen
// is configured to call dedicated resolvers for those fields instead of
// reading them off this struct.
func FolderToModel(folder database.Folder) *model.Folder {
	return &model.Folder{
		ID:   folder.ID.String(),
		Name: folder.Name,
		Project: &model.Project{
			ID: folder.ProjectID.String(),
		},
		CreatedAt: folder.CreatedAt.Time,
		UpdatedAt: folder.UpdatedAt.Time,
	}
}

func FileShareToModel(share database.FileShare) *model.FileShare {
	return &model.FileShare{
		ID: share.ID.String(),
		File: &model.File{
			ID: share.FileID.String(),
			Properties: &model.FileProperties{
				File: &model.File{ID: share.FileID.String()},
			},
		},
		SharedBy:   &model.User{ID: share.SharedBy.String()},
		SharedWith: &model.User{ID: share.SharedWith.String()},
		Permission: model.FilePermission(share.Permission),
		CreatedAt:  share.CreatedAt.Time,
	}
}

func MapFilesToModel(
	files []database.File,
) []*model.File {

	result := make([]*model.File, 0, len(files))

	for _, file := range files {
		result = append(result, FileToModel(file))
	}

	return result
}
