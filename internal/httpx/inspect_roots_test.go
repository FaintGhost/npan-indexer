package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"npan/internal/config"
	"npan/internal/models"
	"npan/internal/npan"
)

type inspectRootsTestAPI struct {
	getFolderInfoFn func(ctx context.Context, folderID int64) (models.NpanFolder, error)
}

func (a *inspectRootsTestAPI) ListFolderChildren(context.Context, int64, int64) (models.FolderChildrenPage, error) {
	return models.FolderChildrenPage{}, nil
}
func (a *inspectRootsTestAPI) GetFolderInfo(ctx context.Context, folderID int64) (models.NpanFolder, error) {
	if a.getFolderInfoFn != nil {
		return a.getFolderInfoFn(ctx, folderID)
	}
	return models.NpanFolder{ID: folderID}, nil
}
func (a *inspectRootsTestAPI) GetDownloadURL(context.Context, int64, *int64) (models.DownloadURLResult, error) {
	return models.DownloadURLResult{}, nil
}
func (a *inspectRootsTestAPI) SearchUpdatedWindow(context.Context, string, *int64, *int64, int64) (map[string]any, error) {
	return nil, nil
}
func (a *inspectRootsTestAPI) ListUserDepartments(context.Context) ([]models.NpanDepartment, error) {
	return nil, nil
}
func (a *inspectRootsTestAPI) ListDepartmentFolders(context.Context, int64) ([]models.NpanFolder, error) {
	return nil, nil
}
func (a *inspectRootsTestAPI) SearchItems(context.Context, models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	return models.RemoteSearchResponse{}, nil
}

func TestInspectRoots_PartialSuccess(t *testing.T) {
	t.Parallel()

	handler := &Handlers{
		cfg: config.Config{
			Token:                   "server-token",
			AllowConfigAuthFallback: true,
		},
		apiFactory: func(_ string, _ npan.AuthResolverOptions) npan.API {
			return &inspectRootsTestAPI{
				getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
					if folderID == 9999 {
						return models.NpanFolder{}, errors.New("not found")
					}
					return models.NpanFolder{
						ID:        folderID,
						Name:      "PIXELHUE",
						ItemCount: 4151,
					}, nil
				},
			}
		},
	}

	e := echo.New()
	body := map[string]any{
		"folder_ids": []int64{1001, 9999},
	}
	rawBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roots/inspect", bytes.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.InspectRoots(c); err != nil {
		t.Fatalf("InspectRoots returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp inspectRootsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 success item, got %d", len(resp.Items))
	}
	if resp.Items[0].FolderID != 1001 {
		t.Fatalf("expected folder_id=1001, got %d", resp.Items[0].FolderID)
	}
	if resp.Items[0].EstimatedTotalDocs != 4152 {
		t.Fatalf("expected estimated_total_docs=4152, got %d", resp.Items[0].EstimatedTotalDocs)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("expected 1 error item, got %d", len(resp.Errors))
	}
	if resp.Errors[0].FolderID != 9999 {
		t.Fatalf("expected error folder_id=9999, got %d", resp.Errors[0].FolderID)
	}
}

func TestInspectRoots_RejectsInvalidFolderIDs(t *testing.T) {
	t.Parallel()

	handler := &Handlers{
		cfg: config.Config{
			Token:                   "server-token",
			AllowConfigAuthFallback: true,
		},
	}

	e := echo.New()
	body := map[string]any{
		"folder_ids": []int64{0, 1001},
	}
	rawBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roots/inspect", bytes.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.InspectRoots(c); err != nil {
		t.Fatalf("InspectRoots returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
