package search

import (
  "context"
  "encoding/json"
  "io"
  "slices"
  "testing"
  "time"

  "github.com/meilisearch/meilisearch-go"
)

// settingsCaptureIndex is a mock IndexManager that captures the Settings
// passed to UpdateSettingsWithContext. All other methods panic or return
// zero values because they are irrelevant to the settings tests.
type settingsCaptureIndex struct {
  captured *meilisearch.Settings
}

// ---------- methods under test ----------

func (m *settingsCaptureIndex) UpdateSettingsWithContext(_ context.Context, s *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  m.captured = s
  return &meilisearch.TaskInfo{TaskUID: 1}, nil
}

func (m *settingsCaptureIndex) WaitForTaskWithContext(_ context.Context, _ int64, _ time.Duration) (*meilisearch.Task, error) {
  return &meilisearch.Task{Status: meilisearch.TaskStatusSucceeded}, nil
}

// ---------- helpers ----------

func callEnsureSettings(t *testing.T) *meilisearch.Settings {
  t.Helper()
  mock := &settingsCaptureIndex{}
  idx := NewMeiliIndexFromManager(mock)
  if err := idx.EnsureSettings(context.Background()); err != nil {
    t.Fatalf("EnsureSettings returned error: %v", err)
  }
  if mock.captured == nil {
    t.Fatal("EnsureSettings did not call UpdateSettingsWithContext")
  }
  return mock.captured
}

func containsString(slice []string, target string) bool {
  return slices.Contains(slice, target)
}

// ---------- tests ----------

func TestEnsureSettings_TypoTolerance(t *testing.T) {
  s := callEnsureSettings(t)

  if s.TypoTolerance == nil {
    t.Fatal("Settings.TypoTolerance is nil")
  }
  if !s.TypoTolerance.Enabled {
    t.Error("TypoTolerance.Enabled should be true")
  }

  if s.TypoTolerance.MinWordSizeForTypos.OneTypo != 5 {
    t.Errorf("MinWordSizeForTypos.OneTypo = %d, want 5", s.TypoTolerance.MinWordSizeForTypos.OneTypo)
  }
  if s.TypoTolerance.MinWordSizeForTypos.TwoTypos != 9 {
    t.Errorf("MinWordSizeForTypos.TwoTypos = %d, want 9", s.TypoTolerance.MinWordSizeForTypos.TwoTypos)
  }

  if !containsString(s.TypoTolerance.DisableOnAttributes, "path_text") {
    t.Error("TypoTolerance.DisableOnAttributes should contain \"path_text\"")
  }

  requiredExtensions := []string{
    "pdf", "docx", "xlsx", "pptx", "jpg", "png",
    "mp4", "zip", "rar", "exe", "apk", "bin", "iso",
  }
  for _, ext := range requiredExtensions {
    if !containsString(s.TypoTolerance.DisableOnWords, ext) {
      t.Errorf("TypoTolerance.DisableOnWords should contain %q", ext)
    }
  }
}

func TestEnsureSettings_StopWords(t *testing.T) {
  s := callEnsureSettings(t)

  if len(s.StopWords) == 0 {
    t.Fatal("Settings.StopWords is empty")
  }

  requiredWords := []string{"的", "了", "在", "是", "和"}
  for _, w := range requiredWords {
    if !containsString(s.StopWords, w) {
      t.Errorf("StopWords should contain %q", w)
    }
  }
}

func TestEnsureSettings_DisplayedAttributes(t *testing.T) {
  s := callEnsureSettings(t)

  if len(s.DisplayedAttributes) == 0 {
    t.Fatal("Settings.DisplayedAttributes is empty")
  }

  mustInclude := []string{"doc_id", "name", "type"}
  for _, attr := range mustInclude {
    if !containsString(s.DisplayedAttributes, attr) {
      t.Errorf("DisplayedAttributes should contain %q", attr)
    }
  }

  mustExclude := []string{"sha1", "in_trash", "is_deleted"}
  for _, attr := range mustExclude {
    if containsString(s.DisplayedAttributes, attr) {
      t.Errorf("DisplayedAttributes should NOT contain %q", attr)
    }
  }
}

func TestEnsureSettings_ProximityPrecision(t *testing.T) {
  s := callEnsureSettings(t)

  if s.ProximityPrecision != meilisearch.ByAttribute {
    t.Errorf("ProximityPrecision = %q, want %q", s.ProximityPrecision, meilisearch.ByAttribute)
  }
}

// ---------- remaining IndexManager stubs (unused, panic on call) ----------

func (m *settingsCaptureIndex) UpdateSettings(s *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSettings")
}
func (m *settingsCaptureIndex) WaitForTask(_ int64, _ time.Duration) (*meilisearch.Task, error) {
  panic("unexpected call: WaitForTask")
}

// IndexReader
func (m *settingsCaptureIndex) FetchInfo() (*meilisearch.IndexResult, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) FetchInfoWithContext(context.Context) (*meilisearch.IndexResult, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) FetchPrimaryKey() (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) FetchPrimaryKeyWithContext(context.Context) (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetStats() (*meilisearch.StatsIndex, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetStatsWithContext(context.Context) (*meilisearch.StatsIndex, error) {
  panic("unexpected call")
}

// TaskReader (remaining)
func (m *settingsCaptureIndex) GetTask(int64) (*meilisearch.Task, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetTaskWithContext(context.Context, int64) (*meilisearch.Task, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetTasks(*meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetTasksWithContext(context.Context, *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  panic("unexpected call")
}

// DocumentReader
func (m *settingsCaptureIndex) GetDocument(string, *meilisearch.DocumentQuery, any) error {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDocumentWithContext(context.Context, string, *meilisearch.DocumentQuery, any) error {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDocuments(*meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDocumentsWithContext(context.Context, *meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
  panic("unexpected call")
}

// DocumentManager
func (m *settingsCaptureIndex) AddDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvFromReaderInBatches(io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvFromReader(io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsCsvFromReaderWithContext(context.Context, io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonFromReader(io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonFromReaderWithContext(context.Context, io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonFromReaderInBatches(io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) AddDocumentsNdjsonFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsByFunction(*meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDocumentsByFunctionWithContext(context.Context, *meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocument(string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocumentWithContext(context.Context, string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocuments([]string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocumentsWithContext(context.Context, []string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocumentsByFilter(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteDocumentsByFilterWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteAllDocuments(*meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteAllDocumentsWithContext(context.Context, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}

// SearchReader
func (m *settingsCaptureIndex) Search(string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) SearchWithContext(context.Context, string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) SearchRaw(string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) SearchRawWithContext(context.Context, string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) FacetSearch(*meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) FacetSearchWithContext(context.Context, *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) SearchSimilarDocuments(*meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) SearchSimilarDocumentsWithContext(context.Context, *meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
  panic("unexpected call")
}

// SettingsReader
func (m *settingsCaptureIndex) GetSettings() (*meilisearch.Settings, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSettingsWithContext(context.Context) (*meilisearch.Settings, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetRankingRules() (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetRankingRulesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDistinctAttribute() (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDistinctAttributeWithContext(context.Context) (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSearchableAttributes() (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSearchableAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDisplayedAttributes() (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDisplayedAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetStopWords() (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetStopWordsWithContext(context.Context) (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSynonyms() (*map[string][]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSynonymsWithContext(context.Context) (*map[string][]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFilterableAttributes() (*[]any, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFilterableAttributesWithContext(context.Context) (*[]any, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSortableAttributes() (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSortableAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetTypoTolerance() (*meilisearch.TypoTolerance, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetTypoToleranceWithContext(context.Context) (*meilisearch.TypoTolerance, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetPagination() (*meilisearch.Pagination, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetPaginationWithContext(context.Context) (*meilisearch.Pagination, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFaceting() (*meilisearch.Faceting, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFacetingWithContext(context.Context) (*meilisearch.Faceting, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetEmbedders() (map[string]meilisearch.Embedder, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetEmbeddersWithContext(context.Context) (map[string]meilisearch.Embedder, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSearchCutoffMs() (int64, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSearchCutoffMsWithContext(context.Context) (int64, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSeparatorTokens() ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetSeparatorTokensWithContext(context.Context) ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetNonSeparatorTokens() ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetNonSeparatorTokensWithContext(context.Context) ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDictionary() ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetDictionaryWithContext(context.Context) ([]string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetProximityPrecision() (meilisearch.ProximityPrecisionType, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetProximityPrecisionWithContext(context.Context) (meilisearch.ProximityPrecisionType, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetLocalizedAttributes() ([]*meilisearch.LocalizedAttributes, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetLocalizedAttributesWithContext(context.Context) ([]*meilisearch.LocalizedAttributes, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetPrefixSearch() (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetPrefixSearchWithContext(context.Context) (*string, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFacetSearch() (bool, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) GetFacetSearchWithContext(context.Context) (bool, error) {
  panic("unexpected call")
}

// SettingsManager (remaining)
func (m *settingsCaptureIndex) ResetSettings() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSettingsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateRankingRules(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateRankingRulesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetRankingRules() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetRankingRulesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDistinctAttribute(string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDistinctAttributeWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDistinctAttribute() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDistinctAttributeWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSearchableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSearchableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSearchableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSearchableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDisplayedAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDisplayedAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDisplayedAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDisplayedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateStopWords(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateStopWordsWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetStopWords() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetStopWordsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSynonyms(*map[string][]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSynonymsWithContext(context.Context, *map[string][]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSynonyms() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSynonymsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFilterableAttributes(*[]any) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFilterableAttributesWithContext(context.Context, *[]any) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFilterableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFilterableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSortableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSortableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSortableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSortableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateTypoTolerance(*meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateTypoToleranceWithContext(context.Context, *meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetTypoTolerance() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetTypoToleranceWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdatePagination(*meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdatePaginationWithContext(context.Context, *meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetPagination() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetPaginationWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFaceting(*meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFacetingWithContext(context.Context, *meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFaceting() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFacetingWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateEmbedders(map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateEmbeddersWithContext(context.Context, map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetEmbedders() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetEmbeddersWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSearchCutoffMs(int64) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSearchCutoffMsWithContext(context.Context, int64) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSearchCutoffMs() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSearchCutoffMsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSeparatorTokens() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateNonSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateNonSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetNonSeparatorTokens() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetNonSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDictionary([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateDictionaryWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDictionary() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetDictionaryWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateProximityPrecision(meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateProximityPrecisionWithContext(context.Context, meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetProximityPrecision() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetProximityPrecisionWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateLocalizedAttributes([]*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateLocalizedAttributesWithContext(context.Context, []*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetLocalizedAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetLocalizedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdatePrefixSearch(string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdatePrefixSearchWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetPrefixSearch() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetPrefixSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFacetSearch(bool) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateFacetSearchWithContext(context.Context, bool) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFacetSearch() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) ResetFacetSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}

// IndexManager direct methods
func (m *settingsCaptureIndex) UpdateIndex(*meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) UpdateIndexWithContext(context.Context, *meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) Delete(string) (bool, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) DeleteWithContext(context.Context, string) (bool, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) Compact() (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}
func (m *settingsCaptureIndex) CompactWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call")
}

// IndexManager getter methods
func (m *settingsCaptureIndex) GetIndexReader() meilisearch.IndexReader       { return m }
func (m *settingsCaptureIndex) GetTaskReader() meilisearch.TaskReader         { return m }
func (m *settingsCaptureIndex) GetDocumentManager() meilisearch.DocumentManager { return m }
func (m *settingsCaptureIndex) GetDocumentReader() meilisearch.DocumentReader { return m }
func (m *settingsCaptureIndex) GetSettingsManager() meilisearch.SettingsManager { return m }
func (m *settingsCaptureIndex) GetSettingsReader() meilisearch.SettingsReader { return m }
func (m *settingsCaptureIndex) GetSearch() meilisearch.SearchReader           { return m }
