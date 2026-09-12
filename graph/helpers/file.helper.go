package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

func MapFilesToModel(
	files []database.File,
) []*model.File {

	result := make([]*model.File, 0, len(files))

	for _, file := range files {

		var folder *model.Folder

		if file.FolderID.Valid {
			folder = &model.Folder{
				ID: file.FolderID.String(),
			}
		}

		result = append(result, &model.File{
			ID:     file.ID.String(),
			Name:   file.Name,
			Size:   int(file.Size),
			Folder: folder,
		})
	}

	return result
}
