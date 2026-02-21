package service

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/meilisearch/meilisearch-go"

	"npan/internal/indexer"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/storage"
)

// ---------------------------------------------------------------------------
// mock: npan.API
// ---------------------------------------------------------------------------

// mockAPI implements npan.API for testing runIncremental.
// Only SearchUpdatedWindow is exercised; the rest return zero values.
type mockAPI struct {
	searchUpdatedWindowFn func(ctx context.Context, queryWords string, start *int64, end *int64, pageID int64) (map[string]any, error)
}

var _ npan.API = (*mockAPI)(nil)

func (m *mockAPI) ListFolderChildren(_ context.Context, _ int64, _ int64) (models.FolderChildrenPage, error) {
	return models.FolderChildrenPage{}, nil
}

func (m *mockAPI) GetDownloadURL(_ context.Context, _ int64, _ *int64) (models.DownloadURLResult, error) {
	return models.DownloadURLResult{}, nil
}

func (m *mockAPI) SearchUpdatedWindow(ctx context.Context, queryWords string, start *int64, end *int64, pageID int64) (map[string]any, error) {
	if m.searchUpdatedWindowFn != nil {
		return m.searchUpdatedWindowFn(ctx, queryWords, start, end, pageID)
	}
	return map[string]any{"page_count": float64(1)}, nil
}

func (m *mockAPI) ListUserDepartments(_ context.Context) ([]models.NpanDepartment, error) {
	return nil, nil
}

func (m *mockAPI) ListDepartmentFolders(_ context.Context, _ int64) ([]models.NpanFolder, error) {
	return nil, nil
}

func (m *mockAPI) SearchItems(_ context.Context, _ models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	return models.RemoteSearchResponse{}, nil
}

// ---------------------------------------------------------------------------
// mock: meilisearch.IndexManager for incremental tests
//
// Embeds routingStubIndex (from sync_manager_routing_test.go) and overrides
// AddDocumentsWithContext and DeleteDocumentsWithContext with controllable
// function fields.
// ---------------------------------------------------------------------------

type incrementalStubIndex struct {
	routingStubIndex
	addDocsFn    func(ctx context.Context, docs any, opts *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error)
	deleteDocsFn func(ctx context.Context, ids []string, opts *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error)
}

func (s *incrementalStubIndex) AddDocumentsWithContext(ctx context.Context, docs any, opts *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	if s.addDocsFn != nil {
		return s.addDocsFn(ctx, docs, opts)
	}
	return s.routingStubIndex.AddDocumentsWithContext(ctx, docs, opts)
}

func (s *incrementalStubIndex) DeleteDocumentsWithContext(ctx context.Context, ids []string, opts *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	if s.deleteDocsFn != nil {
		return s.deleteDocsFn(ctx, ids, opts)
	}
	return &meilisearch.TaskInfo{TaskUID: 1}, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestSyncManager(t *testing.T, idx *search.MeiliIndex) (*SyncManager, string) {
	t.Helper()
	tmpDir := t.TempDir()

	progressFile := filepath.Join(tmpDir, "progress.json")
	syncStateFile := filepath.Join(tmpDir, "sync_state.json")

	mgr := NewSyncManager(SyncManagerArgs{
		Index:              idx,
		ProgressStore:      storage.NewJSONProgressStore(progressFile),
		MeiliHost:          "http://127.0.0.1:7700",
		MeiliIndex:         "test_items",
		CheckpointTemplate: filepath.Join(tmpDir, "checkpoint.json"),
		RootWorkers:        1,
		ProgressEvery:      1,
		Retry: models.RetryPolicyOptions{
			MaxRetries:  2,
			BaseDelayMS: 1,
			MaxDelayMS:  1,
			JitterMS:    0,
		},
		MaxConcurrent:    2,
		MinTimeMS:        0,
		SyncStateFile:    syncStateFile,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	})

	return mgr, tmpDir
}

func makeOnePage(files []map[string]any, folders []map[string]any) map[string]any {
	if files == nil {
		files = []map[string]any{}
	}
	if folders == nil {
		folders = []map[string]any{}
	}
	return map[string]any{
		"files":      files,
		"folders":    folders,
		"page_count": float64(1),
	}
}

func makeFileEntry(id int64, name string, parentID int64, size int64, inTrash bool, isDeleted bool) map[string]any {
	return map[string]any{
		"id":          float64(id),
		"name":        name,
		"parent_id":   float64(parentID),
		"size":        float64(size),
		"modified_at": float64(1700000000),
		"created_at":  float64(1700000000),
		"in_trash":    inTrash,
		"is_deleted":  isDeleted,
	}
}

func makeFolderEntry(id int64, name string, parentID int64, inTrash bool, isDeleted bool) map[string]any {
	return map[string]any{
		"id":          float64(id),
		"name":        name,
		"parent_id":   float64(parentID),
		"modified_at": float64(1700000000),
		"in_trash":    inTrash,
		"is_deleted":  isDeleted,
	}
}

func newTestProgress(cursorBefore int64) *models.SyncProgressState {
	return &models.SyncProgressState{
		Status:    "running",
		Mode:      string(models.SyncModeIncremental),
		StartedAt: 1700000000000,
		UpdatedAt: 1700000000000,
		IncrementalStats: &models.IncrementalSyncStats{
			CursorBefore: cursorBefore,
			CursorAfter:  0,
		},
	}
}

// ---------------------------------------------------------------------------
// Test 1: Happy path – fetches changes, upserts and deletes, updates stats
// ---------------------------------------------------------------------------

func TestRunIncremental_Success(t *testing.T) {
	t.Parallel()

	stub := &incrementalStubIndex{}
	meiliIdx := search.NewMeiliIndexFromManager(stub)
	mgr, _ := newTestSyncManager(t, meiliIdx)
	limiter := indexer.NewRequestLimiter(2, 0)

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(
				[]map[string]any{
					makeFileEntry(1, "test.pdf", 100, 1024, false, false),
					makeFileEntry(2, "removed.doc", 100, 512, true, false),
				},
				[]map[string]any{
					makeFolderEntry(10, "docs", 0, false, false),
				},
			), nil
		},
	}

	progress := newTestProgress(1699990000)
	request := SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}

	err := mgr.runIncremental(context.Background(), api, progress, request, limiter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stats := progress.IncrementalStats
	if stats == nil {
		t.Fatal("expected IncrementalStats to be set")
	}

	// 3 changes fetched total (2 files + 1 folder)
	if stats.ChangesFetched != 3 {
		t.Errorf("expected ChangesFetched=3, got %d", stats.ChangesFetched)
	}

	// file_1 and folder_10 are upserts (not in_trash, not is_deleted)
	if stats.Upserted != 2 {
		t.Errorf("expected Upserted=2, got %d", stats.Upserted)
	}

	// file_2 is a delete (in_trash=true)
	if stats.Deleted != 1 {
		t.Errorf("expected Deleted=1, got %d", stats.Deleted)
	}

	if stats.SkippedUpserts != 0 {
		t.Errorf("expected SkippedUpserts=0, got %d", stats.SkippedUpserts)
	}
	if stats.SkippedDeletes != 0 {
		t.Errorf("expected SkippedDeletes=0, got %d", stats.SkippedDeletes)
	}

	if stats.CursorAfter <= stats.CursorBefore {
		t.Errorf("expected CursorAfter > CursorBefore, got before=%d after=%d",
			stats.CursorBefore, stats.CursorAfter)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Upsert fails once then succeeds on retry
// ---------------------------------------------------------------------------

func TestRunIncremental_UpsertRetrySuccess(t *testing.T) {
	t.Parallel()

	var addAttempts int32
	stub := &incrementalStubIndex{
		addDocsFn: func(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
			if atomic.AddInt32(&addAttempts, 1) == 1 {
				return nil, errors.New("connection timeout")
			}
			return &meilisearch.TaskInfo{TaskUID: 1}, nil
		},
	}
	meiliIdx := search.NewMeiliIndexFromManager(stub)
	mgr, _ := newTestSyncManager(t, meiliIdx)
	limiter := indexer.NewRequestLimiter(2, 0)

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(
				[]map[string]any{
					makeFileEntry(1, "retry-file.pdf", 100, 2048, false, false),
				},
				nil,
			), nil
		},
	}

	progress := newTestProgress(1699990000)
	request := SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}

	err := mgr.runIncremental(context.Background(), api, progress, request, limiter)
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}

	stats := progress.IncrementalStats
	if stats == nil {
		t.Fatal("expected IncrementalStats to be set")
	}
	if stats.Upserted != 1 {
		t.Errorf("expected Upserted=1 after retry, got %d", stats.Upserted)
	}
	if stats.SkippedUpserts != 0 {
		t.Errorf("expected SkippedUpserts=0 after successful retry, got %d", stats.SkippedUpserts)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Upsert always fails, tracked as skipped
// ---------------------------------------------------------------------------

func TestRunIncremental_UpsertExhaustsRetries(t *testing.T) {
	t.Parallel()

	stub := &incrementalStubIndex{
		addDocsFn: func(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
			return nil, errors.New("permanent error")
		},
	}
	meiliIdx := search.NewMeiliIndexFromManager(stub)
	mgr, _ := newTestSyncManager(t, meiliIdx)
	limiter := indexer.NewRequestLimiter(2, 0)

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(
				[]map[string]any{
					makeFileEntry(1, "will-fail.pdf", 100, 512, false, false),
					makeFileEntry(2, "also-fail.pdf", 100, 256, false, false),
				},
				nil,
			), nil
		},
	}

	progress := newTestProgress(1699990000)
	request := SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}

	// runIncremental should NOT return a fatal error when upsert retries are
	// exhausted; instead it records them as SkippedUpserts and continues.
	err := mgr.runIncremental(context.Background(), api, progress, request, limiter)
	if err != nil {
		t.Fatalf("expected graceful handling of exhausted upsert retries, got: %v", err)
	}

	stats := progress.IncrementalStats
	if stats == nil {
		t.Fatal("expected IncrementalStats to be set")
	}
	if stats.ChangesFetched != 2 {
		t.Errorf("expected ChangesFetched=2, got %d", stats.ChangesFetched)
	}
	if stats.SkippedUpserts != 2 {
		t.Errorf("expected SkippedUpserts=2, got %d", stats.SkippedUpserts)
	}
	if stats.Upserted != 0 {
		t.Errorf("expected Upserted=0 when all retries fail, got %d", stats.Upserted)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Delete succeeds
// ---------------------------------------------------------------------------

func TestRunIncremental_DeleteRetrySuccess(t *testing.T) {
	t.Parallel()

	stub := &incrementalStubIndex{}
	meiliIdx := search.NewMeiliIndexFromManager(stub)
	mgr, _ := newTestSyncManager(t, meiliIdx)
	limiter := indexer.NewRequestLimiter(2, 0)

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(
				[]map[string]any{
					makeFileEntry(5, "trashed.pdf", 100, 512, true, false),
				},
				[]map[string]any{
					makeFolderEntry(20, "deleted-folder", 0, false, true),
				},
			), nil
		},
	}

	progress := newTestProgress(1699990000)
	request := SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}

	err := mgr.runIncremental(context.Background(), api, progress, request, limiter)
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}

	stats := progress.IncrementalStats
	if stats == nil {
		t.Fatal("expected IncrementalStats to be set")
	}

	// Both items are marked for deletion (in_trash or is_deleted)
	if stats.Deleted != 2 {
		t.Errorf("expected Deleted=2 after retry, got %d", stats.Deleted)
	}
	if stats.SkippedDeletes != 0 {
		t.Errorf("expected SkippedDeletes=0 after successful retry, got %d", stats.SkippedDeletes)
	}
	if stats.ChangesFetched != 2 {
		t.Errorf("expected ChangesFetched=2, got %d", stats.ChangesFetched)
	}
}
