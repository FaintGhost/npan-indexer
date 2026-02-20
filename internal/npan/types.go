package npan

import (
	"context"

	"npan/internal/models"
)

type API interface {
	ListFolderChildren(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error)
	GetDownloadURL(ctx context.Context, fileID int64, validPeriod *int64) (models.DownloadURLResult, error)
	SearchUpdatedWindow(ctx context.Context, queryWords string, start *int64, end *int64, pageID int64) (map[string]any, error)
	ListUserDepartments(ctx context.Context) ([]models.NpanDepartment, error)
	ListDepartmentFolders(ctx context.Context, departmentID int64) ([]models.NpanFolder, error)
	SearchItems(ctx context.Context, params models.RemoteSearchParams) (models.RemoteSearchResponse, error)
}

type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string {
	return e.Message
}
