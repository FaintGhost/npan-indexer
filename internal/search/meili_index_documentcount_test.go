package search

import (
  "context"
  "encoding/json"
  "errors"
  "io"
  "testing"
  "time"

  "github.com/meilisearch/meilisearch-go"
)

// docCountStubIndex is a mock IndexManager whose only functional method is
// GetStatsWithContext. All other methods panic to catch unexpected calls.
type docCountStubIndex struct {
  stats *meilisearch.StatsIndex
  err   error
}

// ---------- method under test ----------

func (m *docCountStubIndex) GetStatsWithContext(_ context.Context) (*meilisearch.StatsIndex, error) {
  return m.stats, m.err
}

// ---------- remaining IndexManager stubs (panic on call) ----------

// IndexReader
func (m *docCountStubIndex) FetchInfo() (*meilisearch.IndexResult, error) {
  panic("unexpected call: FetchInfo")
}
func (m *docCountStubIndex) FetchInfoWithContext(context.Context) (*meilisearch.IndexResult, error) {
  panic("unexpected call: FetchInfoWithContext")
}
func (m *docCountStubIndex) FetchPrimaryKey() (*string, error) {
  panic("unexpected call: FetchPrimaryKey")
}
func (m *docCountStubIndex) FetchPrimaryKeyWithContext(context.Context) (*string, error) {
  panic("unexpected call: FetchPrimaryKeyWithContext")
}
func (m *docCountStubIndex) GetStats() (*meilisearch.StatsIndex, error) {
  panic("unexpected call: GetStats")
}

// TaskReader
func (m *docCountStubIndex) GetTask(int64) (*meilisearch.Task, error) {
  panic("unexpected call: GetTask")
}
func (m *docCountStubIndex) GetTaskWithContext(context.Context, int64) (*meilisearch.Task, error) {
  panic("unexpected call: GetTaskWithContext")
}
func (m *docCountStubIndex) GetTasks(*meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  panic("unexpected call: GetTasks")
}
func (m *docCountStubIndex) GetTasksWithContext(context.Context, *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
  panic("unexpected call: GetTasksWithContext")
}
func (m *docCountStubIndex) WaitForTask(int64, time.Duration) (*meilisearch.Task, error) {
  panic("unexpected call: WaitForTask")
}
func (m *docCountStubIndex) WaitForTaskWithContext(context.Context, int64, time.Duration) (*meilisearch.Task, error) {
  panic("unexpected call: WaitForTaskWithContext")
}

// DocumentReader
func (m *docCountStubIndex) GetDocument(string, *meilisearch.DocumentQuery, any) error {
  panic("unexpected call: GetDocument")
}
func (m *docCountStubIndex) GetDocumentWithContext(context.Context, string, *meilisearch.DocumentQuery, any) error {
  panic("unexpected call: GetDocumentWithContext")
}
func (m *docCountStubIndex) GetDocuments(*meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
  panic("unexpected call: GetDocuments")
}
func (m *docCountStubIndex) GetDocumentsWithContext(context.Context, *meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
  panic("unexpected call: GetDocumentsWithContext")
}

// DocumentManager
func (m *docCountStubIndex) AddDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocuments")
}
func (m *docCountStubIndex) AddDocumentsWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsWithContext")
}
func (m *docCountStubIndex) AddDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsInBatches")
}
func (m *docCountStubIndex) AddDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsInBatchesWithContext")
}
func (m *docCountStubIndex) AddDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsv")
}
func (m *docCountStubIndex) AddDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvWithContext")
}
func (m *docCountStubIndex) AddDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvInBatches")
}
func (m *docCountStubIndex) AddDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvInBatchesWithContext")
}
func (m *docCountStubIndex) AddDocumentsCsvFromReaderInBatches(io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvFromReaderInBatches")
}
func (m *docCountStubIndex) AddDocumentsCsvFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvFromReaderInBatchesWithContext")
}
func (m *docCountStubIndex) AddDocumentsCsvFromReader(io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvFromReader")
}
func (m *docCountStubIndex) AddDocumentsCsvFromReaderWithContext(context.Context, io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsCsvFromReaderWithContext")
}
func (m *docCountStubIndex) AddDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjson")
}
func (m *docCountStubIndex) AddDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonWithContext")
}
func (m *docCountStubIndex) AddDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonInBatches")
}
func (m *docCountStubIndex) AddDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonInBatchesWithContext")
}
func (m *docCountStubIndex) AddDocumentsNdjsonFromReader(io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonFromReader")
}
func (m *docCountStubIndex) AddDocumentsNdjsonFromReaderWithContext(context.Context, io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonFromReaderWithContext")
}
func (m *docCountStubIndex) AddDocumentsNdjsonFromReaderInBatches(io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonFromReaderInBatches")
}
func (m *docCountStubIndex) AddDocumentsNdjsonFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: AddDocumentsNdjsonFromReaderInBatchesWithContext")
}
func (m *docCountStubIndex) UpdateDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocuments")
}
func (m *docCountStubIndex) UpdateDocumentsWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsInBatches")
}
func (m *docCountStubIndex) UpdateDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsInBatchesWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsCsv")
}
func (m *docCountStubIndex) UpdateDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsCsvWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsCsvInBatches")
}
func (m *docCountStubIndex) UpdateDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsCsvInBatchesWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsNdjson")
}
func (m *docCountStubIndex) UpdateDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsNdjsonWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsNdjsonInBatches")
}
func (m *docCountStubIndex) UpdateDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsNdjsonInBatchesWithContext")
}
func (m *docCountStubIndex) UpdateDocumentsByFunction(*meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsByFunction")
}
func (m *docCountStubIndex) UpdateDocumentsByFunctionWithContext(context.Context, *meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDocumentsByFunctionWithContext")
}
func (m *docCountStubIndex) DeleteDocument(string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocument")
}
func (m *docCountStubIndex) DeleteDocumentWithContext(context.Context, string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocumentWithContext")
}
func (m *docCountStubIndex) DeleteDocuments([]string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocuments")
}
func (m *docCountStubIndex) DeleteDocumentsWithContext(context.Context, []string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocumentsWithContext")
}
func (m *docCountStubIndex) DeleteDocumentsByFilter(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocumentsByFilter")
}
func (m *docCountStubIndex) DeleteDocumentsByFilterWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteDocumentsByFilterWithContext")
}
func (m *docCountStubIndex) DeleteAllDocuments(*meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteAllDocuments")
}
func (m *docCountStubIndex) DeleteAllDocumentsWithContext(context.Context, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: DeleteAllDocumentsWithContext")
}

// SearchReader
func (m *docCountStubIndex) Search(string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  panic("unexpected call: Search")
}
func (m *docCountStubIndex) SearchWithContext(context.Context, string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
  panic("unexpected call: SearchWithContext")
}
func (m *docCountStubIndex) SearchRaw(string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
  panic("unexpected call: SearchRaw")
}
func (m *docCountStubIndex) SearchRawWithContext(context.Context, string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
  panic("unexpected call: SearchRawWithContext")
}
func (m *docCountStubIndex) FacetSearch(*meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  panic("unexpected call: FacetSearch")
}
func (m *docCountStubIndex) FacetSearchWithContext(context.Context, *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
  panic("unexpected call: FacetSearchWithContext")
}
func (m *docCountStubIndex) SearchSimilarDocuments(*meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
  panic("unexpected call: SearchSimilarDocuments")
}
func (m *docCountStubIndex) SearchSimilarDocumentsWithContext(context.Context, *meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
  panic("unexpected call: SearchSimilarDocumentsWithContext")
}

// SettingsReader
func (m *docCountStubIndex) GetSettings() (*meilisearch.Settings, error) {
  panic("unexpected call: GetSettings")
}
func (m *docCountStubIndex) GetSettingsWithContext(context.Context) (*meilisearch.Settings, error) {
  panic("unexpected call: GetSettingsWithContext")
}
func (m *docCountStubIndex) GetRankingRules() (*[]string, error) {
  panic("unexpected call: GetRankingRules")
}
func (m *docCountStubIndex) GetRankingRulesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call: GetRankingRulesWithContext")
}
func (m *docCountStubIndex) GetDistinctAttribute() (*string, error) {
  panic("unexpected call: GetDistinctAttribute")
}
func (m *docCountStubIndex) GetDistinctAttributeWithContext(context.Context) (*string, error) {
  panic("unexpected call: GetDistinctAttributeWithContext")
}
func (m *docCountStubIndex) GetSearchableAttributes() (*[]string, error) {
  panic("unexpected call: GetSearchableAttributes")
}
func (m *docCountStubIndex) GetSearchableAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call: GetSearchableAttributesWithContext")
}
func (m *docCountStubIndex) GetDisplayedAttributes() (*[]string, error) {
  panic("unexpected call: GetDisplayedAttributes")
}
func (m *docCountStubIndex) GetDisplayedAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call: GetDisplayedAttributesWithContext")
}
func (m *docCountStubIndex) GetStopWords() (*[]string, error) {
  panic("unexpected call: GetStopWords")
}
func (m *docCountStubIndex) GetStopWordsWithContext(context.Context) (*[]string, error) {
  panic("unexpected call: GetStopWordsWithContext")
}
func (m *docCountStubIndex) GetSynonyms() (*map[string][]string, error) {
  panic("unexpected call: GetSynonyms")
}
func (m *docCountStubIndex) GetSynonymsWithContext(context.Context) (*map[string][]string, error) {
  panic("unexpected call: GetSynonymsWithContext")
}
func (m *docCountStubIndex) GetFilterableAttributes() (*[]any, error) {
  panic("unexpected call: GetFilterableAttributes")
}
func (m *docCountStubIndex) GetFilterableAttributesWithContext(context.Context) (*[]any, error) {
  panic("unexpected call: GetFilterableAttributesWithContext")
}
func (m *docCountStubIndex) GetSortableAttributes() (*[]string, error) {
  panic("unexpected call: GetSortableAttributes")
}
func (m *docCountStubIndex) GetSortableAttributesWithContext(context.Context) (*[]string, error) {
  panic("unexpected call: GetSortableAttributesWithContext")
}
func (m *docCountStubIndex) GetTypoTolerance() (*meilisearch.TypoTolerance, error) {
  panic("unexpected call: GetTypoTolerance")
}
func (m *docCountStubIndex) GetTypoToleranceWithContext(context.Context) (*meilisearch.TypoTolerance, error) {
  panic("unexpected call: GetTypoToleranceWithContext")
}
func (m *docCountStubIndex) GetPagination() (*meilisearch.Pagination, error) {
  panic("unexpected call: GetPagination")
}
func (m *docCountStubIndex) GetPaginationWithContext(context.Context) (*meilisearch.Pagination, error) {
  panic("unexpected call: GetPaginationWithContext")
}
func (m *docCountStubIndex) GetFaceting() (*meilisearch.Faceting, error) {
  panic("unexpected call: GetFaceting")
}
func (m *docCountStubIndex) GetFacetingWithContext(context.Context) (*meilisearch.Faceting, error) {
  panic("unexpected call: GetFacetingWithContext")
}
func (m *docCountStubIndex) GetEmbedders() (map[string]meilisearch.Embedder, error) {
  panic("unexpected call: GetEmbedders")
}
func (m *docCountStubIndex) GetEmbeddersWithContext(context.Context) (map[string]meilisearch.Embedder, error) {
  panic("unexpected call: GetEmbeddersWithContext")
}
func (m *docCountStubIndex) GetSearchCutoffMs() (int64, error) {
  panic("unexpected call: GetSearchCutoffMs")
}
func (m *docCountStubIndex) GetSearchCutoffMsWithContext(context.Context) (int64, error) {
  panic("unexpected call: GetSearchCutoffMsWithContext")
}
func (m *docCountStubIndex) GetSeparatorTokens() ([]string, error) {
  panic("unexpected call: GetSeparatorTokens")
}
func (m *docCountStubIndex) GetSeparatorTokensWithContext(context.Context) ([]string, error) {
  panic("unexpected call: GetSeparatorTokensWithContext")
}
func (m *docCountStubIndex) GetNonSeparatorTokens() ([]string, error) {
  panic("unexpected call: GetNonSeparatorTokens")
}
func (m *docCountStubIndex) GetNonSeparatorTokensWithContext(context.Context) ([]string, error) {
  panic("unexpected call: GetNonSeparatorTokensWithContext")
}
func (m *docCountStubIndex) GetDictionary() ([]string, error) {
  panic("unexpected call: GetDictionary")
}
func (m *docCountStubIndex) GetDictionaryWithContext(context.Context) ([]string, error) {
  panic("unexpected call: GetDictionaryWithContext")
}
func (m *docCountStubIndex) GetProximityPrecision() (meilisearch.ProximityPrecisionType, error) {
  panic("unexpected call: GetProximityPrecision")
}
func (m *docCountStubIndex) GetProximityPrecisionWithContext(context.Context) (meilisearch.ProximityPrecisionType, error) {
  panic("unexpected call: GetProximityPrecisionWithContext")
}
func (m *docCountStubIndex) GetLocalizedAttributes() ([]*meilisearch.LocalizedAttributes, error) {
  panic("unexpected call: GetLocalizedAttributes")
}
func (m *docCountStubIndex) GetLocalizedAttributesWithContext(context.Context) ([]*meilisearch.LocalizedAttributes, error) {
  panic("unexpected call: GetLocalizedAttributesWithContext")
}
func (m *docCountStubIndex) GetPrefixSearch() (*string, error) {
  panic("unexpected call: GetPrefixSearch")
}
func (m *docCountStubIndex) GetPrefixSearchWithContext(context.Context) (*string, error) {
  panic("unexpected call: GetPrefixSearchWithContext")
}
func (m *docCountStubIndex) GetFacetSearch() (bool, error) {
  panic("unexpected call: GetFacetSearch")
}
func (m *docCountStubIndex) GetFacetSearchWithContext(context.Context) (bool, error) {
  panic("unexpected call: GetFacetSearchWithContext")
}

// SettingsManager
func (m *docCountStubIndex) UpdateSettings(*meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSettings")
}
func (m *docCountStubIndex) UpdateSettingsWithContext(context.Context, *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSettingsWithContext")
}
func (m *docCountStubIndex) ResetSettings() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSettings")
}
func (m *docCountStubIndex) ResetSettingsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSettingsWithContext")
}
func (m *docCountStubIndex) UpdateRankingRules(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateRankingRules")
}
func (m *docCountStubIndex) UpdateRankingRulesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateRankingRulesWithContext")
}
func (m *docCountStubIndex) ResetRankingRules() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetRankingRules")
}
func (m *docCountStubIndex) ResetRankingRulesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetRankingRulesWithContext")
}
func (m *docCountStubIndex) UpdateDistinctAttribute(string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDistinctAttribute")
}
func (m *docCountStubIndex) UpdateDistinctAttributeWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDistinctAttributeWithContext")
}
func (m *docCountStubIndex) ResetDistinctAttribute() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDistinctAttribute")
}
func (m *docCountStubIndex) ResetDistinctAttributeWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDistinctAttributeWithContext")
}
func (m *docCountStubIndex) UpdateSearchableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSearchableAttributes")
}
func (m *docCountStubIndex) UpdateSearchableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSearchableAttributesWithContext")
}
func (m *docCountStubIndex) ResetSearchableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSearchableAttributes")
}
func (m *docCountStubIndex) ResetSearchableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSearchableAttributesWithContext")
}
func (m *docCountStubIndex) UpdateDisplayedAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDisplayedAttributes")
}
func (m *docCountStubIndex) UpdateDisplayedAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDisplayedAttributesWithContext")
}
func (m *docCountStubIndex) ResetDisplayedAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDisplayedAttributes")
}
func (m *docCountStubIndex) ResetDisplayedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDisplayedAttributesWithContext")
}
func (m *docCountStubIndex) UpdateStopWords(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateStopWords")
}
func (m *docCountStubIndex) UpdateStopWordsWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateStopWordsWithContext")
}
func (m *docCountStubIndex) ResetStopWords() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetStopWords")
}
func (m *docCountStubIndex) ResetStopWordsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetStopWordsWithContext")
}
func (m *docCountStubIndex) UpdateSynonyms(*map[string][]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSynonyms")
}
func (m *docCountStubIndex) UpdateSynonymsWithContext(context.Context, *map[string][]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSynonymsWithContext")
}
func (m *docCountStubIndex) ResetSynonyms() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSynonyms")
}
func (m *docCountStubIndex) ResetSynonymsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSynonymsWithContext")
}
func (m *docCountStubIndex) UpdateFilterableAttributes(*[]any) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFilterableAttributes")
}
func (m *docCountStubIndex) UpdateFilterableAttributesWithContext(context.Context, *[]any) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFilterableAttributesWithContext")
}
func (m *docCountStubIndex) ResetFilterableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFilterableAttributes")
}
func (m *docCountStubIndex) ResetFilterableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFilterableAttributesWithContext")
}
func (m *docCountStubIndex) UpdateSortableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSortableAttributes")
}
func (m *docCountStubIndex) UpdateSortableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSortableAttributesWithContext")
}
func (m *docCountStubIndex) ResetSortableAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSortableAttributes")
}
func (m *docCountStubIndex) ResetSortableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSortableAttributesWithContext")
}
func (m *docCountStubIndex) UpdateTypoTolerance(*meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateTypoTolerance")
}
func (m *docCountStubIndex) UpdateTypoToleranceWithContext(context.Context, *meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateTypoToleranceWithContext")
}
func (m *docCountStubIndex) ResetTypoTolerance() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetTypoTolerance")
}
func (m *docCountStubIndex) ResetTypoToleranceWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetTypoToleranceWithContext")
}
func (m *docCountStubIndex) UpdatePagination(*meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdatePagination")
}
func (m *docCountStubIndex) UpdatePaginationWithContext(context.Context, *meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdatePaginationWithContext")
}
func (m *docCountStubIndex) ResetPagination() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetPagination")
}
func (m *docCountStubIndex) ResetPaginationWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetPaginationWithContext")
}
func (m *docCountStubIndex) UpdateFaceting(*meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFaceting")
}
func (m *docCountStubIndex) UpdateFacetingWithContext(context.Context, *meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFacetingWithContext")
}
func (m *docCountStubIndex) ResetFaceting() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFaceting")
}
func (m *docCountStubIndex) ResetFacetingWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFacetingWithContext")
}
func (m *docCountStubIndex) UpdateEmbedders(map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateEmbedders")
}
func (m *docCountStubIndex) UpdateEmbeddersWithContext(context.Context, map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateEmbeddersWithContext")
}
func (m *docCountStubIndex) ResetEmbedders() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetEmbedders")
}
func (m *docCountStubIndex) ResetEmbeddersWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetEmbeddersWithContext")
}
func (m *docCountStubIndex) UpdateSearchCutoffMs(int64) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSearchCutoffMs")
}
func (m *docCountStubIndex) UpdateSearchCutoffMsWithContext(context.Context, int64) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSearchCutoffMsWithContext")
}
func (m *docCountStubIndex) ResetSearchCutoffMs() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSearchCutoffMs")
}
func (m *docCountStubIndex) ResetSearchCutoffMsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSearchCutoffMsWithContext")
}
func (m *docCountStubIndex) UpdateSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSeparatorTokens")
}
func (m *docCountStubIndex) UpdateSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateSeparatorTokensWithContext")
}
func (m *docCountStubIndex) ResetSeparatorTokens() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSeparatorTokens")
}
func (m *docCountStubIndex) ResetSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetSeparatorTokensWithContext")
}
func (m *docCountStubIndex) UpdateNonSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateNonSeparatorTokens")
}
func (m *docCountStubIndex) UpdateNonSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateNonSeparatorTokensWithContext")
}
func (m *docCountStubIndex) ResetNonSeparatorTokens() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetNonSeparatorTokens")
}
func (m *docCountStubIndex) ResetNonSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetNonSeparatorTokensWithContext")
}
func (m *docCountStubIndex) UpdateDictionary([]string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDictionary")
}
func (m *docCountStubIndex) UpdateDictionaryWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateDictionaryWithContext")
}
func (m *docCountStubIndex) ResetDictionary() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDictionary")
}
func (m *docCountStubIndex) ResetDictionaryWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetDictionaryWithContext")
}
func (m *docCountStubIndex) UpdateProximityPrecision(meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateProximityPrecision")
}
func (m *docCountStubIndex) UpdateProximityPrecisionWithContext(context.Context, meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateProximityPrecisionWithContext")
}
func (m *docCountStubIndex) ResetProximityPrecision() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetProximityPrecision")
}
func (m *docCountStubIndex) ResetProximityPrecisionWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetProximityPrecisionWithContext")
}
func (m *docCountStubIndex) UpdateLocalizedAttributes([]*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateLocalizedAttributes")
}
func (m *docCountStubIndex) UpdateLocalizedAttributesWithContext(context.Context, []*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateLocalizedAttributesWithContext")
}
func (m *docCountStubIndex) ResetLocalizedAttributes() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetLocalizedAttributes")
}
func (m *docCountStubIndex) ResetLocalizedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetLocalizedAttributesWithContext")
}
func (m *docCountStubIndex) UpdatePrefixSearch(string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdatePrefixSearch")
}
func (m *docCountStubIndex) UpdatePrefixSearchWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdatePrefixSearchWithContext")
}
func (m *docCountStubIndex) ResetPrefixSearch() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetPrefixSearch")
}
func (m *docCountStubIndex) ResetPrefixSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetPrefixSearchWithContext")
}
func (m *docCountStubIndex) UpdateFacetSearch(bool) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFacetSearch")
}
func (m *docCountStubIndex) UpdateFacetSearchWithContext(context.Context, bool) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateFacetSearchWithContext")
}
func (m *docCountStubIndex) ResetFacetSearch() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFacetSearch")
}
func (m *docCountStubIndex) ResetFacetSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: ResetFacetSearchWithContext")
}

// IndexManager direct methods
func (m *docCountStubIndex) UpdateIndex(*meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateIndex")
}
func (m *docCountStubIndex) UpdateIndexWithContext(context.Context, *meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: UpdateIndexWithContext")
}
func (m *docCountStubIndex) Delete(string) (bool, error) {
  panic("unexpected call: Delete")
}
func (m *docCountStubIndex) DeleteWithContext(context.Context, string) (bool, error) {
  panic("unexpected call: DeleteWithContext")
}
func (m *docCountStubIndex) Compact() (*meilisearch.TaskInfo, error) {
  panic("unexpected call: Compact")
}
func (m *docCountStubIndex) CompactWithContext(context.Context) (*meilisearch.TaskInfo, error) {
  panic("unexpected call: CompactWithContext")
}

// IndexManager getter methods
func (m *docCountStubIndex) GetIndexReader() meilisearch.IndexReader         { return m }
func (m *docCountStubIndex) GetTaskReader() meilisearch.TaskReader           { return m }
func (m *docCountStubIndex) GetDocumentManager() meilisearch.DocumentManager { return m }
func (m *docCountStubIndex) GetDocumentReader() meilisearch.DocumentReader   { return m }
func (m *docCountStubIndex) GetSettingsManager() meilisearch.SettingsManager { return m }
func (m *docCountStubIndex) GetSettingsReader() meilisearch.SettingsReader   { return m }
func (m *docCountStubIndex) GetSearch() meilisearch.SearchReader             { return m }

// ---------- tests ----------

func TestMeiliIndex_DocumentCount(t *testing.T) {
  t.Parallel()

  mock := &docCountStubIndex{
    stats: &meilisearch.StatsIndex{NumberOfDocuments: 42},
  }
  idx := NewMeiliIndexFromManager(mock)

  count, err := idx.DocumentCount(context.Background())
  if err != nil {
    t.Fatalf("DocumentCount returned unexpected error: %v", err)
  }
  if count != 42 {
    t.Errorf("DocumentCount = %d, want 42", count)
  }
}

func TestMeiliIndex_DocumentCount_Error(t *testing.T) {
  t.Parallel()

  sentinel := errors.New("stats unavailable")
  mock := &docCountStubIndex{
    err: sentinel,
  }
  idx := NewMeiliIndexFromManager(mock)

  count, err := idx.DocumentCount(context.Background())
  if count != 0 {
    t.Errorf("DocumentCount = %d on error, want 0", count)
  }
  if !errors.Is(err, sentinel) {
    t.Errorf("DocumentCount error = %v, want %v", err, sentinel)
  }
}
