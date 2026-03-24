package httpx

import (
	"context"
	"errors"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/meilisearch/meilisearch-go"
	"github.com/prometheus/client_golang/prometheus"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
	"npan/internal/config"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/service"
	"npan/internal/storage"
)

type adminConnectTestAPI struct {
	folderInfo        map[int64]models.NpanFolder
	folderErrs        map[int64]error
	getFolderInfoHook func(context.Context, int64)
}

type adminConnectStatsIndex struct {
	meilisearch.IndexManager
	stats *meilisearch.StatsIndex
	err   error
}

func (i *adminConnectStatsIndex) GetStatsWithContext(context.Context) (*meilisearch.StatsIndex, error) {
	if i.err != nil {
		return nil, i.err
	}
	if i.stats == nil {
		return &meilisearch.StatsIndex{}, nil
	}
	return i.stats, nil
}

func (a *adminConnectTestAPI) ListFolderChildren(context.Context, int64, int64) (models.FolderChildrenPage, error) {
	return models.FolderChildrenPage{}, errors.New("not implemented")
}

func (a *adminConnectTestAPI) GetFolderInfo(ctx context.Context, folderID int64) (models.NpanFolder, error) {
	if a.getFolderInfoHook != nil {
		a.getFolderInfoHook(ctx, folderID)
	}
	if err, ok := a.folderErrs[folderID]; ok {
		return models.NpanFolder{}, err
	}
	if folder, ok := a.folderInfo[folderID]; ok {
		return folder, nil
	}
	return models.NpanFolder{}, errors.New("not found")
}

func TestConnectAdminInspectRoots_ConcurrentAndOrdered(t *testing.T) {
	prevConcurrency := inspectRootsMaxConcurrency
	prevPerFolderTimeout := inspectRootsPerFolderTimeout
	inspectRootsMaxConcurrency = 2
	inspectRootsPerFolderTimeout = 0
	defer func() {
		inspectRootsMaxConcurrency = prevConcurrency
		inspectRootsPerFolderTimeout = prevPerFolderTimeout
	}()

	var inFlight int32
	var maxInFlight int32
	started := make(chan int64, 4)
	release := make(chan struct{})

	handlers := newTestHandlers(t)
	handlers.apiFactory = func(_ string, _ npan.AuthResolverOptions) npan.API {
		return &adminConnectTestAPI{
			folderInfo: map[int64]models.NpanFolder{
				1: {ID: 1, Name: "root-1", ItemCount: 10},
				2: {ID: 2, Name: "root-2", ItemCount: 20},
				3: {ID: 3, Name: "root-3", ItemCount: 30},
			},
			getFolderInfoHook: func(_ context.Context, folderID int64) {
				current := atomic.AddInt32(&inFlight, 1)
				for {
					observed := atomic.LoadInt32(&maxInFlight)
					if current <= observed || atomic.CompareAndSwapInt32(&maxInFlight, observed, current) {
						break
					}
				}
				started <- folderID
				if folderID == 1 || folderID == 2 {
					<-release
				}
				atomic.AddInt32(&inFlight, -1)
			},
		}
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	req := connect.NewRequest(&npanv1.InspectRootsRequest{FolderIds: []int64{1, 2, 3}})
	req.Header().Set("X-API-Key", testAdminKey)
	req.Header().Set("Authorization", "Bearer dummy-token")

	resultCh := make(chan struct {
		resp *connect.Response[npanv1.InspectRootsResponse]
		err  error
	}, 1)
	go func() {
		resp, err := client.InspectRoots(context.Background(), req)
		resultCh <- struct {
			resp *connect.Response[npanv1.InspectRootsResponse]
			err  error
		}{resp: resp, err: err}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected first folder request to start")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected second folder request to start")
	}

	if got := atomic.LoadInt32(&maxInFlight); got < 2 {
		t.Fatalf("expected concurrent GetFolderInfo, maxInFlight=%d", got)
	}
	close(release)

	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Fatalf("InspectRoots returned error: %v", result.err)
		}
		items := result.resp.Msg.GetItems()
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].GetFolderId() != 1 || items[1].GetFolderId() != 2 || items[2].GetFolderId() != 3 {
			t.Fatalf("expected ordered items [1,2,3], got [%d,%d,%d]",
				items[0].GetFolderId(), items[1].GetFolderId(), items[2].GetFolderId())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("InspectRoots did not finish in time")
	}
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

func (a *adminConnectTestAPI) SearchItems(_ context.Context, params models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	return models.RemoteSearchResponse{}, errors.New("not implemented")
}

func TestConnectAdminStartSync_ForceRebuildWithRoots_ReturnsInvalidArgument(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), prometheus.NewRegistry())
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
	if !strings.Contains(err.Error(), "force_rebuild") {
		t.Fatalf("expected business guard message to mention force_rebuild, got %q", err.Error())
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
	if got := resp.Msg.GetItems()[0].GetItemCount(); got != 10 {
		t.Fatalf("expected item_count=10, got %d", got)
	}
	if got := resp.Msg.GetItems()[0].GetEstimatedTotalDocs(); got != 11 {
		t.Fatalf("expected estimated_total_docs=11, got %d", got)
	}
}

func TestConnectAdminGetIndexStats_Success(t *testing.T) {
	t.Parallel()

	progressStore := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))
	stubIndex := search.NewMeiliIndexFromManager(&adminConnectStatsIndex{
		stats: &meilisearch.StatsIndex{NumberOfDocuments: 12},
	})
	syncManager := service.NewSyncManager(service.SyncManagerArgs{
		Index:            stubIndex,
		ProgressStore:    progressStore,
		CheckpointStores: storage.NewJSONCheckpointStoreFactory(),
	})

	handlers := &Handlers{
		cfg:          config.Config{AllowConfigAuthFallback: true},
		queryService: &mockSearchService{},
		syncManager:  syncManager,
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	req := connect.NewRequest(&npanv1.GetIndexStatsRequest{})
	req.Header().Set("X-API-Key", testAdminKey)
	resp, err := client.GetIndexStats(context.Background(), req)
	if err != nil {
		t.Fatalf("GetIndexStats returned error: %v", err)
	}
	if got := resp.Msg.GetDocumentCount(); got != 12 {
		t.Fatalf("expected document_count=12, got %d", got)
	}
}

func TestConnectAdminGetIndexStats_ZeroDocument(t *testing.T) {
	t.Parallel()

	progressStore := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))
	stubIndex := search.NewMeiliIndexFromManager(&adminConnectStatsIndex{
		stats: &meilisearch.StatsIndex{NumberOfDocuments: 0},
	})
	syncManager := service.NewSyncManager(service.SyncManagerArgs{
		Index:            stubIndex,
		ProgressStore:    progressStore,
		CheckpointStores: storage.NewJSONCheckpointStoreFactory(),
	})

	handlers := &Handlers{
		cfg:          config.Config{AllowConfigAuthFallback: true},
		queryService: &mockSearchService{},
		syncManager:  syncManager,
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	req := connect.NewRequest(&npanv1.GetIndexStatsRequest{})
	req.Header().Set("X-API-Key", testAdminKey)
	resp, err := client.GetIndexStats(context.Background(), req)
	if err != nil {
		t.Fatalf("GetIndexStats returned error: %v", err)
	}
	if got := resp.Msg.GetDocumentCount(); got != 0 {
		t.Fatalf("expected document_count=0, got %d", got)
	}
}

func TestConnectAdminGetIndexStats_InternalError(t *testing.T) {
	t.Parallel()

	progressStore := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))
	syncManager := service.NewSyncManager(service.SyncManagerArgs{
		ProgressStore: progressStore,
	})

	handlers := &Handlers{
		cfg:          config.Config{AllowConfigAuthFallback: true},
		queryService: &mockSearchService{},
		syncManager:  syncManager,
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	req := connect.NewRequest(&npanv1.GetIndexStatsRequest{})
	req.Header().Set("X-API-Key", testAdminKey)
	_, err := client.GetIndexStats(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Fatalf("expected internal, got %v", got)
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

func TestConnectAdminWatchSyncProgress_StreamsUntilTerminal(t *testing.T) {
	originalInterval := watchSyncProgressPollInterval
	watchSyncProgressPollInterval = 20 * time.Millisecond
	t.Cleanup(func() {
		watchSyncProgressPollInterval = originalInterval
	})

	progressStore := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))
	syncManager := service.NewSyncManager(service.SyncManagerArgs{
		ProgressStore: progressStore,
	})
	handlers := &Handlers{
		cfg:          config.Config{AllowConfigAuthFallback: true},
		queryService: &mockSearchService{},
		syncManager:  syncManager,
	}

	now := time.Now().UnixMilli()
	if err := progressStore.Save(&models.SyncProgressState{
		Status:         "idle",
		StartedAt:      now,
		UpdatedAt:      now,
		Roots:          []int64{},
		CompletedRoots: []int64{},
		RootProgress:   map[string]*models.RootSyncProgress{},
	}); err != nil {
		t.Fatalf("save running progress: %v", err)
	}

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAdminServiceClient(ts.Client(), ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := connect.NewRequest(&npanv1.WatchSyncProgressRequest{})
	req.Header().Set("X-API-Key", testAdminKey)
	stream, err := client.WatchSyncProgress(ctx, req)
	if err != nil {
		t.Fatalf("WatchSyncProgress returned error: %v", err)
	}

	if !stream.Receive() {
		t.Fatalf("expected first streamed message, err=%v", stream.Err())
	}
	if got := stream.Msg().GetState().GetStatus(); got != npanv1.SyncStatus_SYNC_STATUS_IDLE {
		t.Fatalf("expected idle status, got %v", got)
	}

	doneAt := time.Now().UnixMilli()
	if err := progressStore.Save(&models.SyncProgressState{
		Status:         "done",
		StartedAt:      now,
		UpdatedAt:      doneAt,
		Roots:          []int64{},
		CompletedRoots: []int64{},
		RootProgress:   map[string]*models.RootSyncProgress{},
	}); err != nil {
		t.Fatalf("save done progress: %v", err)
	}

	deadline := time.After(500 * time.Millisecond)
	sawDone := false
	for !sawDone {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for done status, last err=%v", stream.Err())
		default:
		}

		if !stream.Receive() {
			t.Fatalf("stream closed before done status, err=%v", stream.Err())
		}
		if stream.Msg().GetState().GetStatus() == npanv1.SyncStatus_SYNC_STATUS_DONE {
			sawDone = true
		}
	}

	if stream.Receive() {
		t.Fatalf("expected stream to close after terminal status")
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("unexpected stream error after terminal status: %v", err)
	}
}

func TestToProtoSyncProgressState_PopulatesTimestampSidecar(t *testing.T) {
	t.Parallel()

	startedAt := int64(1700000000123)
	updatedAt := int64(1700000005123)
	rootUpdatedAt := int64(1700000008123)
	statsStartedAt := int64(1700000001123)
	statsEndedAt := int64(1700000002123)

	state := &models.SyncProgressState{
		Status:    "running",
		StartedAt: startedAt,
		UpdatedAt: updatedAt,
		AggregateStats: models.CrawlStats{
			StartedAt: statsStartedAt,
			EndedAt:   statsEndedAt,
		},
		RootProgress: map[string]*models.RootSyncProgress{
			"1": {
				RootFolderID:    1,
				Status:          "running",
				CurrentFolderID: ptrInt64(42),
				CurrentPageID:   ptrInt64(3),
				CurrentPageCount: ptrInt64(9),
				QueueLength:     ptrInt64(7),
				Error:           "nested repair in progress",
				Stats: models.CrawlStats{
					StartedAt: statsStartedAt,
					EndedAt:   statsEndedAt,
				},
				UpdatedAt: rootUpdatedAt,
			},
		},
	}

	got := toProtoSyncProgressState(state)
	if got == nil {
		t.Fatalf("expected non-nil response")
	}
	if got.GetStartedAtTs() == nil {
		t.Fatalf("expected started_at_ts to be set")
	}
	if got.GetUpdatedAtTs() == nil {
		t.Fatalf("expected updated_at_ts to be set")
	}
	if got.GetAggregateStats() == nil || got.GetAggregateStats().GetStartedAtTs() == nil || got.GetAggregateStats().GetEndedAtTs() == nil {
		t.Fatalf("expected aggregate stats timestamp sidecars to be set")
	}
	root := got.GetRootProgress()["1"]
	if root == nil || root.GetUpdatedAtTs() == nil {
		t.Fatalf("expected root progress updated_at_ts to be set")
	}
	if root.GetCurrentFolderId() != 42 || root.GetCurrentPageId() != 3 || root.GetCurrentPageCount() != 9 || root.GetQueueLength() != 7 {
		t.Fatalf("expected root progress runtime fields to round-trip, got %+v", root)
	}
	if root.GetError() != "nested repair in progress" {
		t.Fatalf("expected root error to round-trip, got %q", root.GetError())
	}
	if got.GetStartedAtTs().AsTime().UnixMilli() != time.UnixMilli(startedAt).UnixMilli() {
		t.Fatalf("unexpected started_at_ts value")
	}
	if got.GetUpdatedAtTs().AsTime().UnixMilli() != time.UnixMilli(updatedAt).UnixMilli() {
		t.Fatalf("unexpected updated_at_ts value")
	}
}

func ptrInt64(v int64) *int64 { return &v }
