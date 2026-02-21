package indexer

import (
  "context"
  "errors"
  "fmt"
  "syscall"
  "testing"

  "npan/internal/models"
  "npan/internal/npan"
)

// mockCrawlAPI implements npan.API for crawl tests.
// Only ListFolderChildren has real logic; all other methods panic.
type mockCrawlAPI struct {
  // pages maps folderID -> list of pages (indexed by pageID).
  pages map[int64][]models.FolderChildrenPage
}

var _ npan.API = (*mockCrawlAPI)(nil)

func (m *mockCrawlAPI) ListFolderChildren(_ context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
  pages, ok := m.pages[folderID]
  if !ok || int(pageID) >= len(pages) {
    return models.FolderChildrenPage{PageCount: 1}, nil
  }
  return pages[pageID], nil
}

func (m *mockCrawlAPI) GetDownloadURL(_ context.Context, _ int64, _ *int64) (models.DownloadURLResult, error) {
  panic("not implemented")
}

func (m *mockCrawlAPI) SearchUpdatedWindow(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
  panic("not implemented")
}

func (m *mockCrawlAPI) ListUserDepartments(_ context.Context) ([]models.NpanDepartment, error) {
  panic("not implemented")
}

func (m *mockCrawlAPI) ListDepartmentFolders(_ context.Context, _ int64) ([]models.NpanFolder, error) {
  panic("not implemented")
}

func (m *mockCrawlAPI) SearchItems(_ context.Context, _ models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
  panic("not implemented")
}

// mockCrawlIndexWriter implements IndexWriter.
// It tracks call count and can be configured to fail on specific call numbers.
type mockCrawlIndexWriter struct {
  callCount  int
  failOnCall int   // 1-based; 0 means never fail
  failErr    error
}

func (w *mockCrawlIndexWriter) UpsertDocuments(_ context.Context, _ []models.IndexDocument) error {
  w.callCount++
  if w.failOnCall > 0 && w.callCount == w.failOnCall {
    return w.failErr
  }
  return nil
}

// memCheckpointStore is a simple in-memory CheckpointStore.
type memCheckpointStore struct {
  data *models.CrawlCheckpoint
}

func (s *memCheckpointStore) Load() (*models.CrawlCheckpoint, error) {
  return s.data, nil
}

func (s *memCheckpointStore) Save(checkpoint *models.CrawlCheckpoint) error {
  s.data = checkpoint
  return nil
}

func (s *memCheckpointStore) Clear() error {
  s.data = nil
  return nil
}

// makeFiles returns n NpanFile items with sequential IDs starting at startID.
func makeFiles(startID int64, n int) []models.NpanFile {
  files := make([]models.NpanFile, n)
  for i := range files {
    files[i] = models.NpanFile{
      ID:   startID + int64(i),
      Name: fmt.Sprintf("file-%d", startID+int64(i)),
    }
  }
  return files
}

// TestRunFullCrawl_FilesIndexedAfterUpsert verifies that FilesIndexed is
// incremented only when upsert succeeds (1 folder, 1 page, 10 files).
func TestRunFullCrawl_FilesIndexedAfterUpsert(t *testing.T) {
  t.Parallel()

  api := &mockCrawlAPI{
    pages: map[int64][]models.FolderChildrenPage{
      1: {
        {Files: makeFiles(1, 10), PageCount: 1},
      },
    },
  }
  writer := &mockCrawlIndexWriter{}
  store := &memCheckpointStore{}

  deps := FullCrawlDeps{
    API:             api,
    IndexWriter:     writer,
    Limiter:         NewRequestLimiter(10, 0),
    CheckpointStore: store,
    RootFolderID:    1,
    Retry:           models.RetryPolicyOptions{},
  }

  stats, err := RunFullCrawl(context.Background(), deps)
  if err != nil {
    t.Fatalf("expected no error, got: %v", err)
  }
  if stats.FilesIndexed != 10 {
    t.Errorf("FilesIndexed = %d, want 10", stats.FilesIndexed)
  }
  if stats.FilesDiscovered != 10 {
    t.Errorf("FilesDiscovered = %d, want 10", stats.FilesDiscovered)
  }
}

// TestRunFullCrawl_FilesDiscoveredPerPage verifies FilesDiscovered and
// FilesIndexed accumulate across multiple pages (1 folder, 2 pages, 5 files
// each).
func TestRunFullCrawl_FilesDiscoveredPerPage(t *testing.T) {
  t.Parallel()

  api := &mockCrawlAPI{
    pages: map[int64][]models.FolderChildrenPage{
      1: {
        {Files: makeFiles(1, 5), PageCount: 2},
        {Files: makeFiles(6, 5), PageCount: 2},
      },
    },
  }
  writer := &mockCrawlIndexWriter{}
  store := &memCheckpointStore{}

  deps := FullCrawlDeps{
    API:             api,
    IndexWriter:     writer,
    Limiter:         NewRequestLimiter(10, 0),
    CheckpointStore: store,
    RootFolderID:    1,
    Retry:           models.RetryPolicyOptions{},
  }

  stats, err := RunFullCrawl(context.Background(), deps)
  if err != nil {
    t.Fatalf("expected no error, got: %v", err)
  }
  if stats.FilesDiscovered != 10 {
    t.Errorf("FilesDiscovered = %d, want 10", stats.FilesDiscovered)
  }
  if stats.FilesIndexed != 10 {
    t.Errorf("FilesIndexed = %d, want 10", stats.FilesIndexed)
  }
}

// TestRunFullCrawl_SkippedFilesOnUpsertFailure verifies that a non-retriable
// upsert failure causes the page to be skipped (SkippedFiles++) rather than
// terminating the entire crawl (1 folder, 2 pages, 5 files each; first upsert
// fails).
func TestRunFullCrawl_SkippedFilesOnUpsertFailure(t *testing.T) {
  t.Parallel()

  api := &mockCrawlAPI{
    pages: map[int64][]models.FolderChildrenPage{
      1: {
        {Files: makeFiles(1, 5), PageCount: 2},
        {Files: makeFiles(6, 5), PageCount: 2},
      },
    },
  }
  // The first UpsertDocuments call handles the root folder doc + page-0 files.
  // We want that call to fail so the 5 files on page 0 are skipped.
  writer := &mockCrawlIndexWriter{
    failOnCall: 1,
    failErr:    errors.New("bad request"),
  }
  store := &memCheckpointStore{}

  deps := FullCrawlDeps{
    API:             api,
    IndexWriter:     writer,
    Limiter:         NewRequestLimiter(10, 0),
    CheckpointStore: store,
    RootFolderID:    1,
    Retry:           models.RetryPolicyOptions{},
  }

  stats, err := RunFullCrawl(context.Background(), deps)
  // The crawl must complete successfully — not terminate on the first failure.
  if err != nil {
    t.Fatalf("expected crawl to complete (err == nil), got: %v", err)
  }
  if stats.FilesDiscovered != 10 {
    t.Errorf("FilesDiscovered = %d, want 10", stats.FilesDiscovered)
  }
  if stats.FilesIndexed != 5 {
    t.Errorf("FilesIndexed = %d, want 5 (second page only)", stats.FilesIndexed)
  }
  if stats.SkippedFiles != 5 {
    t.Errorf("SkippedFiles = %d, want 5", stats.SkippedFiles)
  }
  if stats.FailedRequests != 1 {
    t.Errorf("FailedRequests = %d, want 1", stats.FailedRequests)
  }
}

// TestRunFullCrawl_UpsertRetrySuccess verifies that a retriable upsert error
// (syscall.ECONNRESET) is retried and, upon success, the files are counted as
// indexed (1 folder, 1 page, 10 files; first attempt fails, retry succeeds).
func TestRunFullCrawl_UpsertRetrySuccess(t *testing.T) {
  t.Parallel()

  api := &mockCrawlAPI{
    pages: map[int64][]models.FolderChildrenPage{
      1: {
        {Files: makeFiles(1, 10), PageCount: 1},
      },
    },
  }
  // First UpsertDocuments call fails with a retriable network error.
  writer := &mockCrawlIndexWriter{
    failOnCall: 1,
    failErr:    syscall.ECONNRESET,
  }
  store := &memCheckpointStore{}

  deps := FullCrawlDeps{
    API:             api,
    IndexWriter:     writer,
    Limiter:         NewRequestLimiter(10, 0),
    CheckpointStore: store,
    RootFolderID:    1,
    Retry: models.RetryPolicyOptions{
      MaxRetries:  2,
      BaseDelayMS: 1,
      MaxDelayMS:  1,
    },
  }

  stats, err := RunFullCrawl(context.Background(), deps)
  if err != nil {
    t.Fatalf("expected no error after retry, got: %v", err)
  }
  if stats.FilesIndexed != 10 {
    t.Errorf("FilesIndexed = %d, want 10", stats.FilesIndexed)
  }
  if stats.SkippedFiles != 0 {
    t.Errorf("SkippedFiles = %d, want 0", stats.SkippedFiles)
  }
}
