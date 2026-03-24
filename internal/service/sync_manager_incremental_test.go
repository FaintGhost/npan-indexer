package service

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
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
	listFolderChildrenFn  func(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error)
	getFolderInfoFn       func(ctx context.Context, folderID int64) (models.NpanFolder, error)
}

var _ npan.API = (*mockAPI)(nil)

func (m *mockAPI) ListFolderChildren(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
	if m.listFolderChildrenFn != nil {
		return m.listFolderChildrenFn(ctx, folderID, pageID)
	}
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

func (m *mockAPI) GetFolderInfo(ctx context.Context, folderID int64) (models.NpanFolder, error) {
	if m.getFolderInfoFn != nil {
		return m.getFolderInfoFn(ctx, folderID)
	}
	return models.NpanFolder{ID: folderID}, nil
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

type inMemoryIndexStub struct {
	docs      map[string]models.IndexDocument
	upserts   [][]models.IndexDocument
	deletes   [][]string
	deleteAll int
}

func newInMemoryIndexStub(docs []models.IndexDocument) *inMemoryIndexStub {
	items := make(map[string]models.IndexDocument, len(docs))
	for _, doc := range docs {
		items[doc.DocID] = doc
	}
	return &inMemoryIndexStub{docs: items}
}

func (s *inMemoryIndexStub) EnsureSettings(context.Context) error { return nil }

func (s *inMemoryIndexStub) UpsertDocuments(_ context.Context, docs []models.IndexDocument) error {
	cloned := append([]models.IndexDocument(nil), docs...)
	s.upserts = append(s.upserts, cloned)
	for _, doc := range docs {
		s.docs[doc.DocID] = doc
	}
	return nil
}

func (s *inMemoryIndexStub) DeleteDocuments(_ context.Context, docIDs []string) error {
	cloned := append([]string(nil), docIDs...)
	s.deletes = append(s.deletes, cloned)
	for _, docID := range docIDs {
		delete(s.docs, docID)
	}
	return nil
}

func (s *inMemoryIndexStub) DeleteAllDocuments(_ context.Context) error {
	s.deleteAll++
	s.docs = map[string]models.IndexDocument{}
	return nil
}

func (s *inMemoryIndexStub) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	items := make([]models.IndexDocument, 0, len(s.docs))
	for _, doc := range s.docs {
		if params.ParentID != nil && doc.ParentID != *params.ParentID {
			continue
		}
		if params.Type != "" && params.Type != "all" && string(doc.Type) != params.Type {
			continue
		}
		if !params.IncludeDeleted && (doc.InTrash || doc.IsDeleted) {
			continue
		}
		items = append(items, doc)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].DocID == items[j].DocID {
			return items[i].SourceID < items[j].SourceID
		}
		return items[i].DocID < items[j].DocID
	})
	total := int64(len(items))
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	start := int((page - 1) * pageSize)
	if start >= len(items) {
		return []models.IndexDocument{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	return append([]models.IndexDocument(nil), items[start:end]...), total, nil
}

func (s *inMemoryIndexStub) Ping() error { return nil }

func (s *inMemoryIndexStub) DocumentCount(context.Context) (int64, error) {
	return int64(len(s.docs)), nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestSyncManager(t *testing.T, idx search.IndexOperator) (*SyncManager, string) {
	t.Helper()
	tmpDir := t.TempDir()

	progressFile := filepath.Join(tmpDir, "progress.json")
	syncStateFile := filepath.Join(tmpDir, "sync_state.json")

	mgr := NewSyncManager(SyncManagerArgs{
		Index:              idx,
		ProgressStore:      storage.NewJSONProgressStore(progressFile),
		SyncStateStore:     storage.NewJSONSyncStateStore(syncStateFile),
		CheckpointStores:   storage.NewJSONCheckpointStoreFactory(),
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

func TestRunIncremental_DefaultQueryAndOpenEndedWindow(t *testing.T) {
	t.Parallel()

	stub := &incrementalStubIndex{}
	meiliIdx := search.NewMeiliIndexFromManager(stub)
	mgr, _ := newTestSyncManager(t, meiliIdx)
	mgr.defaultIncrementalQuery = ""
	limiter := indexer.NewRequestLimiter(2, 0)

	var capturedQuery string
	var capturedStart *int64
	var capturedEnd *int64

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, queryWords string, start *int64, end *int64, _ int64) (map[string]any, error) {
			capturedQuery = queryWords
			capturedStart = start
			capturedEnd = end
			return makeOnePage(nil, nil), nil
		},
	}

	progress := newTestProgress(1700000000)
	request := SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "",
		WindowOverlapMS:  5000,
	}

	err := mgr.runIncremental(context.Background(), api, progress, request, limiter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if capturedQuery != "* OR *" {
		t.Fatalf("expected fallback query '* OR *', got %q", capturedQuery)
	}
	if capturedStart == nil || *capturedStart != 1699999995 {
		t.Fatalf("expected overlapped start 1699999995, got %v", capturedStart)
	}
	if capturedEnd != nil {
		t.Fatalf("expected nil end for open-ended incremental window, got %v", *capturedEnd)
	}
}

func seedRepairProgress(t *testing.T, mgr *SyncManager, checkpointFile string) {
	t.Helper()
	err := mgr.progressStore.Save(&models.SyncProgressState{
		Status:         "done",
		Mode:           string(models.SyncModeFull),
		StartedAt:      1700000000000,
		UpdatedAt:      1700000000000,
		Roots:          []int64{100},
		CompletedRoots: []int64{100},
		RootNames:      map[int64]string{100: "Repair Root"},
		RootProgress: map[string]*models.RootSyncProgress{
			"100": {
				RootFolderID:   100,
				CheckpointFile: checkpointFile,
				Status:         "done",
				Stats: models.CrawlStats{
					FoldersVisited: 1,
					FilesIndexed:   0,
					StartedAt:      1700000000000,
					EndedAt:        1700000000000,
				},
				UpdatedAt: 1700000000000,
			},
		},
		AggregateStats: models.CrawlStats{
			FoldersVisited: 1,
			FilesIndexed:   0,
			StartedAt:      1700000000000,
			EndedAt:        1700000000000,
		},
	})
	if err != nil {
		t.Fatalf("seed progress: %v", err)
	}
}

func TestRunIncrementalPath_BackfillsOnlyDriftedSubfolderWhenCountsDisagree(t *testing.T) {
	t.Parallel()

	index := newInMemoryIndexStub([]models.IndexDocument{
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 100, Name: "全部文件", ParentID: 100}, "全部文件"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 200, Name: "Sub A", ParentID: 100}, "folder/200/Sub A"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 300, Name: "Sub B", ParentID: 100}, "folder/300/Sub B"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 2, Name: "keep.txt", ParentID: 300}, "file/2/keep.txt"),
	})
	mgr, tmpDir := newTestSyncManager(t, index)
	seedRepairProgress(t, mgr, filepath.Join(tmpDir, "repair-checkpoint.json"))

	var inspectRootCalls int32
	var inspectSubACalls int32
	var inspectSubBCalls int32
	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(nil, nil), nil
		},
		getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
			if folderID == 100 {
				return models.NpanFolder{ID: 100, Name: "Repair Root", ItemCount: 4}, nil
			}
			return models.NpanFolder{ID: folderID}, nil
		},
		listFolderChildrenFn: func(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
			switch folderID {
			case 100:
				atomic.AddInt32(&inspectRootCalls, 1)
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 2,
					Folders: []models.NpanFolder{
						{ID: 200, Name: "Sub A", ParentID: 100, ItemCount: 1},
						{ID: 300, Name: "Sub B", ParentID: 100, ItemCount: 1},
					},
				}, nil
			case 200:
				atomic.AddInt32(&inspectSubACalls, 1)
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 1,
					Files: []models.NpanFile{
						{ID: 1, Name: "missing.txt", ParentID: 200},
					},
				}, nil
			case 300:
				atomic.AddInt32(&inspectSubBCalls, 1)
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 1,
					Files: []models.NpanFile{
						{ID: 2, Name: "keep.txt", ParentID: 300},
					},
				}, nil
			}
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	err := mgr.runIncrementalPath(context.Background(), api, SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}, &models.SyncState{LastSyncTime: 1700000000000}, mgr.effectiveSyncStateStore())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if atomic.LoadInt32(&inspectRootCalls) == 0 || atomic.LoadInt32(&inspectSubACalls) < 2 || atomic.LoadInt32(&inspectSubBCalls) == 0 {
		t.Fatalf("expected live inspection plus targeted subfolder rebuild, got root=%d subA=%d subB=%d",
			inspectRootCalls, inspectSubACalls, inspectSubBCalls)
	}

	if len(index.deletes) != 0 {
		t.Fatalf("expected additive backfill without targeted delete, got %#v", index.deletes)
	}
	if len(index.upserts) == 0 {
		t.Fatalf("expected subtree rebuild upserts")
	}

	progress, err := mgr.progressStore.Load()
	if err != nil {
		t.Fatalf("load progress: %v", err)
	}
	root := progress.RootProgress["100"]
	if root == nil {
		t.Fatalf("expected repaired root progress")
	}
	if root.Stats.FoldersVisited != 3 || root.Stats.FilesIndexed != 2 {
		t.Fatalf("expected refreshed stats folders=3 files=2, got %+v", root.Stats)
	}
	if progress.IncrementalStats == nil || progress.IncrementalStats.ChangesFetched != 0 {
		t.Fatalf("expected no incremental changes, got %#v", progress.IncrementalStats)
	}
	if got, err := index.DocumentCount(context.Background()); err != nil || got != 5 {
		t.Fatalf("expected repaired document count=5, got=%d err=%v", got, err)
	}
}

func TestRunIncrementalPath_RebuildsDriftedSubfolderWhenLocalCountIsTooLarge(t *testing.T) {
	t.Parallel()

	index := newInMemoryIndexStub([]models.IndexDocument{
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 100, Name: "全部文件", ParentID: 100}, "全部文件"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 200, Name: "Sub A", ParentID: 100}, "folder/200/Sub A"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 300, Name: "Sub B", ParentID: 100}, "folder/300/Sub B"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 1, Name: "keep.txt", ParentID: 200}, "file/1/keep.txt"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 9, Name: "stale.txt", ParentID: 200}, "file/9/stale.txt"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 2, Name: "keep.txt", ParentID: 300}, "file/2/keep.txt"),
	})
	mgr, tmpDir := newTestSyncManager(t, index)
	seedRepairProgress(t, mgr, filepath.Join(tmpDir, "repair-checkpoint.json"))

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(nil, nil), nil
		},
		getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
			if folderID == 100 {
				return models.NpanFolder{ID: 100, Name: "Repair Root", ItemCount: 4}, nil
			}
			return models.NpanFolder{ID: folderID}, nil
		},
		listFolderChildrenFn: func(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
			switch folderID {
			case 100:
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 2,
					Folders: []models.NpanFolder{
						{ID: 200, Name: "Sub A", ParentID: 100, ItemCount: 1},
						{ID: 300, Name: "Sub B", ParentID: 100, ItemCount: 1},
					},
				}, nil
			case 200:
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 1,
					Files: []models.NpanFile{
						{ID: 1, Name: "keep.txt", ParentID: 200},
					},
				}, nil
			case 300:
				return models.FolderChildrenPage{
					PageCount:  1,
					TotalCount: 1,
					Files: []models.NpanFile{
						{ID: 2, Name: "keep.txt", ParentID: 300},
					},
				}, nil
			}
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	err := mgr.runIncrementalPath(context.Background(), api, SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}, &models.SyncState{LastSyncTime: 1700000000000}, mgr.effectiveSyncStateStore())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(index.deletes) != 1 {
		t.Fatalf("expected one targeted delete batch, got %d", len(index.deletes))
	}
	if len(index.deletes[0]) != 3 {
		t.Fatalf("expected subtree delete to remove stale folder docs, got %#v", index.deletes[0])
	}
	if len(index.upserts) == 0 {
		t.Fatalf("expected subtree rebuild upserts")
	}

	if got, err := index.DocumentCount(context.Background()); err != nil || got != 5 {
		t.Fatalf("expected rebuilt document count=5, got=%d err=%v", got, err)
	}
}

func TestRunIncrementalPath_RetriesTransientRepairTimeout(t *testing.T) {
	t.Parallel()

	index := newInMemoryIndexStub([]models.IndexDocument{
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 100, Name: "全部文件", ParentID: 100}, "全部文件"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 200, Name: "Sub A", ParentID: 100}, "folder/200/Sub A"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 300, Name: "Sub B", ParentID: 100}, "folder/300/Sub B"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 2, Name: "keep.txt", ParentID: 300}, "file/2/keep.txt"),
	})
	mgr, tmpDir := newTestSyncManager(t, index)
	seedRepairProgress(t, mgr, filepath.Join(tmpDir, "repair-checkpoint.json"))

	var subATries int32
	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(nil, nil), nil
		},
		getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
			if folderID == 100 {
				return models.NpanFolder{ID: 100, Name: "Repair Root", ItemCount: 4}, nil
			}
			return models.NpanFolder{ID: folderID}, nil
		},
		listFolderChildrenFn: func(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
			switch folderID {
			case 100:
				return models.FolderChildrenPage{
					PageCount: 1,
					Folders: []models.NpanFolder{
						{ID: 200, Name: "Sub A", ParentID: 100, ItemCount: 1},
						{ID: 300, Name: "Sub B", ParentID: 100, ItemCount: 1},
					},
				}, nil
			case 200:
				if atomic.AddInt32(&subATries, 1) == 1 {
					return models.FolderChildrenPage{}, errors.New(`net/http: TLS handshake timeout`)
				}
				return models.FolderChildrenPage{
					PageCount: 1,
					Files: []models.NpanFile{
						{ID: 1, Name: "missing.txt", ParentID: 200},
					},
				}, nil
			case 300:
				return models.FolderChildrenPage{
					PageCount: 1,
					Files: []models.NpanFile{
						{ID: 2, Name: "keep.txt", ParentID: 300},
					},
				}, nil
			}
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	err := mgr.runIncrementalPath(context.Background(), api, SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}, &models.SyncState{LastSyncTime: 1700000000000}, mgr.effectiveSyncStateStore())
	if err != nil {
		t.Fatalf("expected no error after retry, got: %v", err)
	}
	if atomic.LoadInt32(&subATries) < 2 {
		t.Fatalf("expected transient subtree repair timeout to be retried, got %d tries", subATries)
	}
}

func TestRunIncrementalPath_DoesNotFailWholeSyncWhenRepairTimesOut(t *testing.T) {
	t.Parallel()

	index := newInMemoryIndexStub([]models.IndexDocument{
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 100, Name: "全部文件", ParentID: 100}, "全部文件"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 200, Name: "Sub A", ParentID: 100}, "folder/200/Sub A"),
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 300, Name: "Sub B", ParentID: 100}, "folder/300/Sub B"),
		search.MapFileToIndexDoc(models.NpanFile{ID: 2, Name: "keep.txt", ParentID: 300}, "file/2/keep.txt"),
	})
	mgr, tmpDir := newTestSyncManager(t, index)
	seedRepairProgress(t, mgr, filepath.Join(tmpDir, "repair-checkpoint.json"))

	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(nil, nil), nil
		},
		getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
			if folderID == 100 {
				return models.NpanFolder{ID: 100, Name: "Repair Root", ItemCount: 4}, nil
			}
			return models.NpanFolder{ID: folderID}, nil
		},
		listFolderChildrenFn: func(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
			switch folderID {
			case 100:
				return models.FolderChildrenPage{
					PageCount: 1,
					Folders: []models.NpanFolder{
						{ID: 200, Name: "Sub A", ParentID: 100, ItemCount: 1},
						{ID: 300, Name: "Sub B", ParentID: 100, ItemCount: 1},
					},
				}, nil
			case 200:
				return models.FolderChildrenPage{}, errors.New(`net/http: TLS handshake timeout`)
			case 300:
				return models.FolderChildrenPage{
					PageCount: 1,
					Files: []models.NpanFile{
						{ID: 2, Name: "keep.txt", ParentID: 300},
					},
				}, nil
			}
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	err := mgr.runIncrementalPath(context.Background(), api, SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}, &models.SyncState{LastSyncTime: 1700000000000}, mgr.effectiveSyncStateStore())
	if err != nil {
		t.Fatalf("expected incremental sync to tolerate repair timeout, got: %v", err)
	}

	progress, err := mgr.progressStore.Load()
	if err != nil {
		t.Fatalf("load progress: %v", err)
	}
	if progress == nil {
		t.Fatalf("expected progress")
	}
	if progress.Status != "done" {
		t.Fatalf("expected overall incremental sync to complete, got %q", progress.Status)
	}
	root := progress.RootProgress["100"]
	if root == nil {
		t.Fatalf("expected root progress")
	}
	if root.Status != "error" {
		t.Fatalf("expected root repair failure to be visible, got %q", root.Status)
	}
	if root.Error == "" {
		t.Fatalf("expected root repair error message")
	}
}

func TestRunIncrementalPath_SkipsRepairWhenChangesFetched(t *testing.T) {
	t.Parallel()

	index := newInMemoryIndexStub([]models.IndexDocument{
		search.MapFolderToIndexDoc(models.NpanFolder{ID: 100, Name: "全部文件", ParentID: 100}, "全部文件"),
	})
	mgr, tmpDir := newTestSyncManager(t, index)
	seedRepairProgress(t, mgr, filepath.Join(tmpDir, "repair-checkpoint.json"))

	var repairCalls int32
	api := &mockAPI{
		searchUpdatedWindowFn: func(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
			return makeOnePage(
				[]map[string]any{
					makeFileEntry(1, "changed-file", 100, 1, false, false),
				},
				nil,
			), nil
		},
		getFolderInfoFn: func(_ context.Context, folderID int64) (models.NpanFolder, error) {
			return models.NpanFolder{ID: folderID, Name: "Repair Root", ItemCount: 1}, nil
		},
		listFolderChildrenFn: func(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
			atomic.AddInt32(&repairCalls, 1)
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	err := mgr.runIncrementalPath(context.Background(), api, SyncStartRequest{
		Mode:             models.SyncModeIncremental,
		IncrementalQuery: "*",
		WindowOverlapMS:  5000,
	}, &models.SyncState{LastSyncTime: 1700000000000}, mgr.effectiveSyncStateStore())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if atomic.LoadInt32(&repairCalls) != 0 {
		t.Fatalf("expected repair crawl to be skipped when incremental changes were fetched")
	}
}
