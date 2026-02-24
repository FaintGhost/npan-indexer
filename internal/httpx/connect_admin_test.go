package httpx

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
	"npan/internal/models"
	"npan/internal/npan"
)

type adminConnectTestAPI struct {
	folderInfo map[int64]models.NpanFolder
	folderErrs map[int64]error
}

func (a *adminConnectTestAPI) ListFolderChildren(context.Context, int64, int64) (models.FolderChildrenPage, error) {
	return models.FolderChildrenPage{}, errors.New("not implemented")
}

func (a *adminConnectTestAPI) GetFolderInfo(_ context.Context, folderID int64) (models.NpanFolder, error) {
	if err, ok := a.folderErrs[folderID]; ok {
		return models.NpanFolder{}, err
	}
	if folder, ok := a.folderInfo[folderID]; ok {
		return folder, nil
	}
	return models.NpanFolder{}, errors.New("not found")
}

func (a *adminConnectTestAPI) GetDownloadURL(context.Context, int64, *int64) (models.DownloadURLResult, error) {
	return models.DownloadURLResult{}, errors.New("not implemented")
}

func (a *adminConnectTestAPI) SearchUpdatedWindow(context.Context, string, *int64, *int64, int64) (map[string]any, error) {
	return nil, errors.New("not implemented")
}

func (a *adminConnectTestAPI) ListUserDepartments(context.Context) ([]models.NpanDepartment, error) {
	return nil, errors.New("not implemented")
}

func (a *adminConnectTestAPI) ListDepartmentFolders(context.Context, int64) ([]models.NpanFolder, error) {
	return nil, errors.New("not implemented")
}

func (a *adminConnectTestAPI) SearchItems(context.Context, models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	return models.RemoteSearchResponse{}, errors.New("not implemented")
}

func TestConnectAdminStartSync_ForceRebuildWithRoots_ReturnsInvalidArgument(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	forceRebuild := true
	req := connect.NewRequest(&npanv1.StartSyncRequest{
		RootFolderIds: []int64{1},
		ForceRebuild:  &forceRebuild,
	})
	req.Header().Set("X-API-Key", testAdminKey)
	_, err := client.StartSync(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %v", got)
	}
}

func TestConnectAdminInspectRoots_PartialSuccess(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	handlers.apiFactory = func(_ string, _ npan.AuthResolverOptions) npan.API {
		return &adminConnectTestAPI{
			folderInfo: map[int64]models.NpanFolder{
				1: {ID: 1, Name: "root-1", ItemCount: 10},
			},
			folderErrs: map[int64]error{
				2: errors.New("upstream error"),
			},
		}
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	req := connect.NewRequest(&npanv1.InspectRootsRequest{FolderIds: []int64{1, 2}})
	req.Header().Set("X-API-Key", testAdminKey)
	req.Header().Set("Authorization", "Bearer dummy-token")
	resp, err := client.InspectRoots(context.Background(), req)
	if err != nil {
		t.Fatalf("InspectRoots returned error: %v", err)
	}
	if len(resp.Msg.GetItems()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Msg.GetItems()))
	}
	if len(resp.Msg.GetErrors()) != 1 {
		t.Fatalf("expected 1 error, got %d", len(resp.Msg.GetErrors()))
	}
	if got := resp.Msg.GetItems()[0].GetEstimatedTotalDocs(); got != 11 {
		t.Fatalf("expected estimated_total_docs=11, got %d", got)
	}
}

func TestConnectAdminGetSyncProgress_NotFoundAndCancelConflict(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)

	progressReq := connect.NewRequest(&npanv1.GetSyncProgressRequest{})
	progressReq.Header().Set("X-API-Key", testAdminKey)
	_, err := client.GetSyncProgress(context.Background(), progressReq)
	if err == nil {
		t.Fatalf("expected GetSyncProgress error")
	}
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Fatalf("expected not_found, got %v", got)
	}

	cancelReq := connect.NewRequest(&npanv1.CancelSyncRequest{})
	cancelReq.Header().Set("X-API-Key", testAdminKey)
	_, err = client.CancelSync(context.Background(), cancelReq)
	if err == nil {
		t.Fatalf("expected CancelSync error")
	}
	if got := connect.CodeOf(err); got != connect.CodeAborted {
		t.Fatalf("expected aborted, got %v", got)
	}
}
