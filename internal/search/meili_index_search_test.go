package search

import (
  "context"
  "encoding/json"
  "io"
  "testing"
  "time"

  "github.com/meilisearch/meilisearch-go"

  "npan/internal/models"
)

// searchCaptureIndex captures the query and SearchRequest passed to Search,
// and returns a pre-configured SearchResponse. All other IndexManager methods
// are stubbed with zero-value returns.
type searchCaptureIndex struct {
  capturedQuery   string
  capturedRequest *meilisearch.SearchRequest
  response        *meilisearch.SearchResponse
}

func (s *searchCaptureIndex) Search(query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  s.capturedQuery = query
  s.capturedRequest = request
  return s.response, nil
}

// --- Stubs for IndexManager ---

func (s *searchCaptureIndex) FetchInfo() (*meilisearch.IndexResult, error) {
  return nil, nil
}
func (s *searchCaptureIndex) FetchInfoWithContext(_ context.Context) (*meilisearch.IndexResult, error) {
  return nil, nil
}
func (s *searchCaptureIndex) FetchPrimaryKey() (*string, error)    { return nil, nil }
func (s *searchCaptureIndex) FetchPrimaryKeyWithContext(_ context.Context) (*string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetStats() (*meilisearch.StatsIndex, error) { return nil, nil }
func (s *searchCaptureIndex) GetStatsWithContext(_ context.Context) (*meilisearch.StatsIndex, error) {
  return nil, nil
}

func (s *searchCaptureIndex) GetTask(_ int64) (*meilisearch.Task, error)    { return nil, nil }
func (s *searchCaptureIndex) GetTaskWithContext(_ context.Context, _ int64) (*meilisearch.Task, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetTasks(_ *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetTasksWithContext(_ context.Context, _ *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  return nil, nil
}
func (s *searchCaptureIndex) WaitForTask(_ int64, _ time.Duration) (*meilisearch.Task, error) {
  return nil, nil
}
func (s *searchCaptureIndex) WaitForTaskWithContext(_ context.Context, _ int64, _ time.Duration) (*meilisearch.Task, error) {
  return nil, nil
}

func (s *searchCaptureIndex) AddDocuments(_ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsWithContext(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsInBatches(_ any, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsInBatchesWithContext(_ context.Context, _ any, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsv(_ []byte, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvWithContext(_ context.Context, _ []byte, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvInBatches(_ []byte, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvInBatchesWithContext(_ context.Context, _ []byte, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvFromReaderInBatches(_ io.Reader, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvFromReaderInBatchesWithContext(_ context.Context, _ io.Reader, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvFromReader(_ io.Reader, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsCsvFromReaderWithContext(_ context.Context, _ io.Reader, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjson(_ []byte, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonWithContext(_ context.Context, _ []byte, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonInBatches(_ []byte, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonInBatchesWithContext(_ context.Context, _ []byte, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonFromReader(_ io.Reader, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonFromReaderWithContext(_ context.Context, _ io.Reader, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonFromReaderInBatches(_ io.Reader, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) AddDocumentsNdjsonFromReaderInBatchesWithContext(_ context.Context, _ io.Reader, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocuments(_ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsWithContext(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsInBatches(_ any, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsInBatchesWithContext(_ context.Context, _ any, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsCsv(_ []byte, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsCsvWithContext(_ context.Context, _ []byte, _ *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsCsvInBatches(_ []byte, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsCsvInBatchesWithContext(_ context.Context, _ []byte, _ int, _ *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsNdjson(_ []byte, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsNdjsonWithContext(_ context.Context, _ []byte, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsNdjsonInBatches(_ []byte, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsNdjsonInBatchesWithContext(_ context.Context, _ []byte, _ int, _ *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsByFunction(_ *meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDocumentsByFunctionWithContext(_ context.Context, _ *meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocument(_ string, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocumentWithContext(_ context.Context, _ string, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocuments(_ []string, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocumentsWithContext(_ context.Context, _ []string, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocumentsByFilter(_ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteDocumentsByFilterWithContext(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteAllDocuments(_ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) DeleteAllDocumentsWithContext(_ context.Context, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetDocument(_ string, _ *meilisearch.DocumentQuery, _ any) error {
  return nil
}
func (s *searchCaptureIndex) GetDocumentWithContext(_ context.Context, _ string, _ *meilisearch.DocumentQuery, _ any) error {
  return nil
}
func (s *searchCaptureIndex) GetDocuments(_ *meilisearch.DocumentsQuery, _ *meilisearch.DocumentsResult) error {
  return nil
}
func (s *searchCaptureIndex) GetDocumentsWithContext(_ context.Context, _ *meilisearch.DocumentsQuery, _ *meilisearch.DocumentsResult) error {
  return nil
}

func (s *searchCaptureIndex) UpdateSettings(_ *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSettingsWithContext(_ context.Context, _ *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSettings() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetSettingsWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSettings() (*meilisearch.Settings, error) { return nil, nil }
func (s *searchCaptureIndex) GetSettingsWithContext(_ context.Context) (*meilisearch.Settings, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateRankingRules(_ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateRankingRulesWithContext(_ context.Context, _ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetRankingRules() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetRankingRulesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetRankingRules() (*[]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetRankingRulesWithContext(_ context.Context) (*[]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDistinctAttribute(_ string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDistinctAttributeWithContext(_ context.Context, _ string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetDistinctAttribute() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetDistinctAttributeWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetDistinctAttribute() (*string, error) { return nil, nil }
func (s *searchCaptureIndex) GetDistinctAttributeWithContext(_ context.Context) (*string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSearchableAttributes(_ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSearchableAttributesWithContext(_ context.Context, _ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSearchableAttributes() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSearchableAttributesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSearchableAttributes() (*[]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetSearchableAttributesWithContext(_ context.Context) (*[]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDisplayedAttributes(_ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDisplayedAttributesWithContext(_ context.Context, _ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetDisplayedAttributes() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetDisplayedAttributesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetDisplayedAttributes() (*[]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetDisplayedAttributesWithContext(_ context.Context) (*[]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateStopWords(_ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateStopWordsWithContext(_ context.Context, _ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetStopWords() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetStopWordsWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetStopWords() (*[]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetStopWordsWithContext(_ context.Context) (*[]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSynonyms(_ *map[string][]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSynonymsWithContext(_ context.Context, _ *map[string][]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSynonyms() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetSynonymsWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSynonyms() (*map[string][]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetSynonymsWithContext(_ context.Context) (*map[string][]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFilterableAttributes(_ *[]any) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFilterableAttributesWithContext(_ context.Context, _ *[]any) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetFilterableAttributes() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetFilterableAttributesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetFilterableAttributes() (*[]any, error) { return nil, nil }
func (s *searchCaptureIndex) GetFilterableAttributesWithContext(_ context.Context) (*[]any, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSortableAttributes(_ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSortableAttributesWithContext(_ context.Context, _ *[]string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSortableAttributes() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSortableAttributesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSortableAttributes() (*[]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetSortableAttributesWithContext(_ context.Context) (*[]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateTypoTolerance(_ *meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateTypoToleranceWithContext(_ context.Context, _ *meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetTypoTolerance() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetTypoToleranceWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetTypoTolerance() (*meilisearch.TypoTolerance, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetTypoToleranceWithContext(_ context.Context) (*meilisearch.TypoTolerance, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdatePagination(_ *meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdatePaginationWithContext(_ context.Context, _ *meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetPagination() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetPaginationWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetPagination() (*meilisearch.Pagination, error) { return nil, nil }
func (s *searchCaptureIndex) GetPaginationWithContext(_ context.Context) (*meilisearch.Pagination, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFaceting(_ *meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFacetingWithContext(_ context.Context, _ *meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetFaceting() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetFacetingWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetFaceting() (*meilisearch.Faceting, error) { return nil, nil }
func (s *searchCaptureIndex) GetFacetingWithContext(_ context.Context) (*meilisearch.Faceting, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateEmbedders(_ map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateEmbeddersWithContext(_ context.Context, _ map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetEmbedders() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetEmbeddersWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetEmbedders() (map[string]meilisearch.Embedder, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetEmbeddersWithContext(_ context.Context) (map[string]meilisearch.Embedder, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSearchCutoffMs(_ int64) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSearchCutoffMsWithContext(_ context.Context, _ int64) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSearchCutoffMs() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetSearchCutoffMsWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSearchCutoffMs() (int64, error) { return 0, nil }
func (s *searchCaptureIndex) GetSearchCutoffMsWithContext(_ context.Context) (int64, error) {
  return 0, nil
}
func (s *searchCaptureIndex) UpdateSeparatorTokens(_ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateSeparatorTokensWithContext(_ context.Context, _ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSeparatorTokens() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetSeparatorTokensWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetSeparatorTokens() ([]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetSeparatorTokensWithContext(_ context.Context) ([]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateNonSeparatorTokens(_ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateNonSeparatorTokensWithContext(_ context.Context, _ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetNonSeparatorTokens() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetNonSeparatorTokensWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetNonSeparatorTokens() ([]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetNonSeparatorTokensWithContext(_ context.Context) ([]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDictionary(_ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateDictionaryWithContext(_ context.Context, _ []string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetDictionary() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetDictionaryWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetDictionary() ([]string, error) { return nil, nil }
func (s *searchCaptureIndex) GetDictionaryWithContext(_ context.Context) ([]string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateProximityPrecision(_ meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateProximityPrecisionWithContext(_ context.Context, _ meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetProximityPrecision() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetProximityPrecisionWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetProximityPrecision() (meilisearch.ProximityPrecisionType, error) {
  return "", nil
}
func (s *searchCaptureIndex) GetProximityPrecisionWithContext(_ context.Context) (meilisearch.ProximityPrecisionType, error) {
  return "", nil
}
func (s *searchCaptureIndex) UpdateLocalizedAttributes(_ []*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateLocalizedAttributesWithContext(_ context.Context, _ []*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetLocalizedAttributes() (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetLocalizedAttributesWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetLocalizedAttributes() ([]*meilisearch.LocalizedAttributes, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetLocalizedAttributesWithContext(_ context.Context) ([]*meilisearch.LocalizedAttributes, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdatePrefixSearch(_ string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdatePrefixSearchWithContext(_ context.Context, _ string) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetPrefixSearch() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetPrefixSearchWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetPrefixSearch() (*string, error) { return nil, nil }
func (s *searchCaptureIndex) GetPrefixSearchWithContext(_ context.Context) (*string, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFacetSearch(_ bool) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateFacetSearchWithContext(_ context.Context, _ bool) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) ResetFacetSearch() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) ResetFacetSearchWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) GetFacetSearch() (bool, error) { return false, nil }
func (s *searchCaptureIndex) GetFacetSearchWithContext(_ context.Context) (bool, error) {
  return false, nil
}

func (s *searchCaptureIndex) SearchWithContext(_ context.Context, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  return s.Search(query, request)
}
func (s *searchCaptureIndex) SearchRaw(_ string, _ *meilisearch.SearchRequest) (*json.RawMessage, error) {
  return nil, nil
}
func (s *searchCaptureIndex) SearchRawWithContext(_ context.Context, _ string, _ *meilisearch.SearchRequest) (*json.RawMessage, error) {
  return nil, nil
}
func (s *searchCaptureIndex) FacetSearch(_ *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  return nil, nil
}
func (s *searchCaptureIndex) FacetSearchWithContext(_ context.Context, _ *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  return nil, nil
}
func (s *searchCaptureIndex) SearchSimilarDocuments(_ *meilisearch.SimilarDocumentQuery, _ *meilisearch.SimilarDocumentResult) error {
  return nil
}
func (s *searchCaptureIndex) SearchSimilarDocumentsWithContext(_ context.Context, _ *meilisearch.SimilarDocumentQuery, _ *meilisearch.SimilarDocumentResult) error {
  return nil
}

func (s *searchCaptureIndex) GetIndexReader() meilisearch.IndexReader       { return s }
func (s *searchCaptureIndex) GetTaskReader() meilisearch.TaskReader         { return s }
func (s *searchCaptureIndex) GetDocumentManager() meilisearch.DocumentManager { return s }
func (s *searchCaptureIndex) GetDocumentReader() meilisearch.DocumentReader { return s }
func (s *searchCaptureIndex) GetSettingsManager() meilisearch.SettingsManager { return s }
func (s *searchCaptureIndex) GetSettingsReader() meilisearch.SettingsReader { return s }
func (s *searchCaptureIndex) GetSearch() meilisearch.SearchReader           { return s }

func (s *searchCaptureIndex) UpdateIndex(_ *meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) UpdateIndexWithContext(_ context.Context, _ *meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  return nil, nil
}
func (s *searchCaptureIndex) Delete(_ string) (bool, error) { return false, nil }
func (s *searchCaptureIndex) DeleteWithContext(_ context.Context, _ string) (bool, error) {
  return false, nil
}
func (s *searchCaptureIndex) Compact() (*meilisearch.TaskInfo, error) { return nil, nil }
func (s *searchCaptureIndex) CompactWithContext(_ context.Context) (*meilisearch.TaskInfo, error) {
  return nil, nil
}

// --- Helpers ---

// mustRawMessage marshals v to json.RawMessage; panics on failure.
func mustRawMessage(v any) json.RawMessage {
  b, err := json.Marshal(v)
  if err != nil {
    panic(err)
  }
  return json.RawMessage(b)
}

// buildHitsWithFormatted creates Hits containing both regular fields and a
// _formatted sub-object. This simulates Meilisearch highlighting behavior.
func buildHitsWithFormatted() meilisearch.Hits {
  hit := meilisearch.Hit{
    "doc_id":      mustRawMessage("file-1"),
    "source_id":   mustRawMessage(100),
    "type":        mustRawMessage("file"),
    "name":        mustRawMessage("test file.pdf"),
    "path_text":   mustRawMessage("/docs/test file.pdf"),
    "parent_id":   mustRawMessage(10),
    "modified_at": mustRawMessage(1700000000),
    "created_at":  mustRawMessage(1699000000),
    "size":        mustRawMessage(2048),
    "_formatted": mustRawMessage(map[string]any{
      "doc_id":    "file-1",
      "name":      "test <mark>file</mark>.pdf",
      "path_text": "/docs/test <mark>file</mark>.pdf",
    }),
  }
  return meilisearch.Hits{hit}
}

// newSearchCaptureIndex creates a searchCaptureIndex with a pre-configured
// response containing highlighted hit data.
func newSearchCaptureIndex() *searchCaptureIndex {
  return &searchCaptureIndex{
    response: &meilisearch.SearchResponse{
      Hits:      buildHitsWithFormatted(),
      TotalHits: 1,
    },
  }
}

// --- Tests ---

func TestSearch_RequestIncludesHighlightParams(t *testing.T) {
  mock := newSearchCaptureIndex()
  idx := NewMeiliIndexFromManager(mock)

  _, _, err := idx.Search(models.LocalSearchParams{
    Query: "file",
  })
  if err != nil {
    t.Fatalf("Search returned error: %v", err)
  }

  req := mock.capturedRequest
  if req == nil {
    t.Fatal("expected SearchRequest to be captured, got nil")
  }

  // Verify AttributesToHighlight contains "name"
  if len(req.AttributesToHighlight) == 0 {
    t.Fatal("expected AttributesToHighlight to be set, got empty slice")
  }
  found := false
  for _, attr := range req.AttributesToHighlight {
    if attr == "name" {
      found = true
      break
    }
  }
  if !found {
    t.Errorf("expected AttributesToHighlight to contain \"name\", got %v", req.AttributesToHighlight)
  }

  // Verify HighlightPreTag
  if req.HighlightPreTag != "<mark>" {
    t.Errorf("expected HighlightPreTag = \"<mark>\", got %q", req.HighlightPreTag)
  }

  // Verify HighlightPostTag
  if req.HighlightPostTag != "</mark>" {
    t.Errorf("expected HighlightPostTag = \"</mark>\", got %q", req.HighlightPostTag)
  }
}

func TestSearch_RequestIncludesAttributesToRetrieve(t *testing.T) {
  mock := newSearchCaptureIndex()
  idx := NewMeiliIndexFromManager(mock)

  _, _, err := idx.Search(models.LocalSearchParams{
    Query: "file",
  })
  if err != nil {
    t.Fatalf("Search returned error: %v", err)
  }

  req := mock.capturedRequest
  if req == nil {
    t.Fatal("expected SearchRequest to be captured, got nil")
  }

  // Verify AttributesToRetrieve is set
  if len(req.AttributesToRetrieve) == 0 {
    t.Fatal("expected AttributesToRetrieve to be set, got empty slice")
  }

  // Verify excluded fields are not present
  excluded := []string{"sha1", "in_trash", "is_deleted"}
  for _, excl := range excluded {
    for _, attr := range req.AttributesToRetrieve {
      if attr == excl {
        t.Errorf("expected AttributesToRetrieve to NOT contain %q, but it does: %v", excl, req.AttributesToRetrieve)
      }
    }
  }
}

func TestSearch_ResponseIncludesHighlightedName(t *testing.T) {
  mock := newSearchCaptureIndex()
  idx := NewMeiliIndexFromManager(mock)

  docs, _, err := idx.Search(models.LocalSearchParams{
    Query: "file",
  })
  if err != nil {
    t.Fatalf("Search returned error: %v", err)
  }

  if len(docs) == 0 {
    t.Fatal("expected at least one document, got none")
  }

  expected := "test <mark>file</mark>.pdf"
  if docs[0].HighlightedName != expected {
    t.Errorf("expected HighlightedName = %q, got %q", expected, docs[0].HighlightedName)
  }
}

func TestReorderQuery(t *testing.T) {
  tests := []struct {
    input string
    want  string
  }{
    {"mx40 spec pdf", "pdf mx40 spec"},
    {"mx40 pdf spec", "pdf mx40 spec"},
    {"mx40 spec", "mx40 spec"},
    {"pdf", "pdf"},
    {"mx40 spec pdf docx", "pdf docx mx40 spec"},
    {"report", "report"},
    {"", ""},
    {"MX40 SPEC PDF", "PDF MX40 SPEC"},
  }
  for _, tt := range tests {
    got := reorderQuery(tt.input)
    if got != tt.want {
      t.Errorf("reorderQuery(%q) = %q, want %q", tt.input, got, tt.want)
    }
  }
}
