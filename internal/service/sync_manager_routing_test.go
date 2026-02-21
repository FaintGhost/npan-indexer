package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/meilisearch/meilisearch-go"

	"npan/internal/models"
	"npan/internal/search"
	"npan/internal/storage"
)

// ---------------------------------------------------------------------------
// Mock npan.API for routing tests
// ---------------------------------------------------------------------------

type mockAPIForRouting struct {
	listFolderChildrenFn func(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error)
}

func (m *mockAPIForRouting) ListFolderChildren(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
	if m.listFolderChildrenFn != nil {
		return m.listFolderChildrenFn(ctx, folderID, pageID)
	}
	return models.FolderChildrenPage{PageCount: 1}, nil
}

func (m *mockAPIForRouting) GetDownloadURL(_ context.Context, _ int64, _ *int64) (models.DownloadURLResult, error) {
	return models.DownloadURLResult{}, nil
}

func (m *mockAPIForRouting) SearchUpdatedWindow(_ context.Context, _ string, _ *int64, _ *int64, _ int64) (map[string]any, error) {
	return nil, nil
}

func (m *mockAPIForRouting) ListUserDepartments(_ context.Context) ([]models.NpanDepartment, error) {
	return nil, nil
}

func (m *mockAPIForRouting) ListDepartmentFolders(_ context.Context, _ int64) ([]models.NpanFolder, error) {
	return nil, nil
}

func (m *mockAPIForRouting) SearchItems(_ context.Context, _ models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	return models.RemoteSearchResponse{}, nil
}

// ---------------------------------------------------------------------------
// Stub meilisearch.IndexManager for routing tests
//
// Functional methods:
//   - AddDocumentsWithContext  (UpsertDocuments path)
//   - WaitForTaskWithContext   (waitTask path)
//   - GetStatsWithContext      (DocumentCount path)
//
// All other methods panic to detect unexpected calls.
// ---------------------------------------------------------------------------

type routingStubIndex struct {
	docCount int64
}

// ---------- functional stubs ----------

func (s *routingStubIndex) AddDocumentsWithContext(_ context.Context, _ any, _ *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	return &meilisearch.TaskInfo{TaskUID: 1}, nil
}

func (s *routingStubIndex) WaitForTaskWithContext(_ context.Context, _ int64, _ time.Duration) (*meilisearch.Task, error) {
	return &meilisearch.Task{Status: meilisearch.TaskStatusSucceeded}, nil
}

func (s *routingStubIndex) GetStatsWithContext(_ context.Context) (*meilisearch.StatsIndex, error) {
	return &meilisearch.StatsIndex{NumberOfDocuments: s.docCount}, nil
}

// ---------- remaining IndexManager stubs (panic on call) ----------

func (s *routingStubIndex) FetchInfo() (*meilisearch.IndexResult, error) {
	panic("unexpected call: FetchInfo")
}
func (s *routingStubIndex) FetchInfoWithContext(context.Context) (*meilisearch.IndexResult, error) {
	panic("unexpected call: FetchInfoWithContext")
}
func (s *routingStubIndex) FetchPrimaryKey() (*string, error) {
	panic("unexpected call: FetchPrimaryKey")
}
func (s *routingStubIndex) FetchPrimaryKeyWithContext(context.Context) (*string, error) {
	panic("unexpected call: FetchPrimaryKeyWithContext")
}
func (s *routingStubIndex) GetStats() (*meilisearch.StatsIndex, error) {
	panic("unexpected call: GetStats")
}
func (s *routingStubIndex) GetTask(int64) (*meilisearch.Task, error) {
	panic("unexpected call: GetTask")
}
func (s *routingStubIndex) GetTaskWithContext(context.Context, int64) (*meilisearch.Task, error) {
	panic("unexpected call: GetTaskWithContext")
}
func (s *routingStubIndex) GetTasks(*meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
	panic("unexpected call: GetTasks")
}
func (s *routingStubIndex) GetTasksWithContext(context.Context, *meilisearch.TasksQuery) (*meilisearch.TaskResult, error) {
	panic("unexpected call: GetTasksWithContext")
}
func (s *routingStubIndex) WaitForTask(int64, time.Duration) (*meilisearch.Task, error) {
	panic("unexpected call: WaitForTask")
}
func (s *routingStubIndex) GetDocument(string, *meilisearch.DocumentQuery, any) error {
	panic("unexpected call: GetDocument")
}
func (s *routingStubIndex) GetDocumentWithContext(context.Context, string, *meilisearch.DocumentQuery, any) error {
	panic("unexpected call: GetDocumentWithContext")
}
func (s *routingStubIndex) GetDocuments(*meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
	panic("unexpected call: GetDocuments")
}
func (s *routingStubIndex) GetDocumentsWithContext(context.Context, *meilisearch.DocumentsQuery, *meilisearch.DocumentsResult) error {
	panic("unexpected call: GetDocumentsWithContext")
}
func (s *routingStubIndex) AddDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocuments")
}
func (s *routingStubIndex) AddDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsInBatches")
}
func (s *routingStubIndex) AddDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsInBatchesWithContext")
}
func (s *routingStubIndex) AddDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsv")
}
func (s *routingStubIndex) AddDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvWithContext")
}
func (s *routingStubIndex) AddDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvInBatches")
}
func (s *routingStubIndex) AddDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvInBatchesWithContext")
}
func (s *routingStubIndex) AddDocumentsCsvFromReaderInBatches(io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvFromReaderInBatches")
}
func (s *routingStubIndex) AddDocumentsCsvFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvFromReaderInBatchesWithContext")
}
func (s *routingStubIndex) AddDocumentsCsvFromReader(io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvFromReader")
}
func (s *routingStubIndex) AddDocumentsCsvFromReaderWithContext(context.Context, io.Reader, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsCsvFromReaderWithContext")
}
func (s *routingStubIndex) AddDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjson")
}
func (s *routingStubIndex) AddDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonWithContext")
}
func (s *routingStubIndex) AddDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonInBatches")
}
func (s *routingStubIndex) AddDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonInBatchesWithContext")
}
func (s *routingStubIndex) AddDocumentsNdjsonFromReader(io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonFromReader")
}
func (s *routingStubIndex) AddDocumentsNdjsonFromReaderWithContext(context.Context, io.Reader, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonFromReaderWithContext")
}
func (s *routingStubIndex) AddDocumentsNdjsonFromReaderInBatches(io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonFromReaderInBatches")
}
func (s *routingStubIndex) AddDocumentsNdjsonFromReaderInBatchesWithContext(context.Context, io.Reader, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: AddDocumentsNdjsonFromReaderInBatchesWithContext")
}
func (s *routingStubIndex) UpdateDocuments(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocuments")
}
func (s *routingStubIndex) UpdateDocumentsWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsWithContext")
}
func (s *routingStubIndex) UpdateDocumentsInBatches(any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsInBatches")
}
func (s *routingStubIndex) UpdateDocumentsInBatchesWithContext(context.Context, any, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsInBatchesWithContext")
}
func (s *routingStubIndex) UpdateDocumentsCsv([]byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsCsv")
}
func (s *routingStubIndex) UpdateDocumentsCsvWithContext(context.Context, []byte, *meilisearch.CsvDocumentsQuery) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsCsvWithContext")
}
func (s *routingStubIndex) UpdateDocumentsCsvInBatches([]byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsCsvInBatches")
}
func (s *routingStubIndex) UpdateDocumentsCsvInBatchesWithContext(context.Context, []byte, int, *meilisearch.CsvDocumentsQuery) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsCsvInBatchesWithContext")
}
func (s *routingStubIndex) UpdateDocumentsNdjson([]byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsNdjson")
}
func (s *routingStubIndex) UpdateDocumentsNdjsonWithContext(context.Context, []byte, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsNdjsonWithContext")
}
func (s *routingStubIndex) UpdateDocumentsNdjsonInBatches([]byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsNdjsonInBatches")
}
func (s *routingStubIndex) UpdateDocumentsNdjsonInBatchesWithContext(context.Context, []byte, int, *meilisearch.DocumentOptions) ([]meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsNdjsonInBatchesWithContext")
}
func (s *routingStubIndex) UpdateDocumentsByFunction(*meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsByFunction")
}
func (s *routingStubIndex) UpdateDocumentsByFunctionWithContext(context.Context, *meilisearch.UpdateDocumentByFunctionRequest) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDocumentsByFunctionWithContext")
}
func (s *routingStubIndex) DeleteDocument(string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocument")
}
func (s *routingStubIndex) DeleteDocumentWithContext(context.Context, string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocumentWithContext")
}
func (s *routingStubIndex) DeleteDocuments([]string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocuments")
}
func (s *routingStubIndex) DeleteDocumentsWithContext(context.Context, []string, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocumentsWithContext")
}
func (s *routingStubIndex) DeleteDocumentsByFilter(any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocumentsByFilter")
}
func (s *routingStubIndex) DeleteDocumentsByFilterWithContext(context.Context, any, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteDocumentsByFilterWithContext")
}
func (s *routingStubIndex) DeleteAllDocuments(*meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteAllDocuments")
}
func (s *routingStubIndex) DeleteAllDocumentsWithContext(context.Context, *meilisearch.DocumentOptions) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: DeleteAllDocumentsWithContext")
}
func (s *routingStubIndex) Search(string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	panic("unexpected call: Search")
}
func (s *routingStubIndex) SearchWithContext(context.Context, string, *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	panic("unexpected call: SearchWithContext")
}
func (s *routingStubIndex) SearchRaw(string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
	panic("unexpected call: SearchRaw")
}
func (s *routingStubIndex) SearchRawWithContext(context.Context, string, *meilisearch.SearchRequest) (*json.RawMessage, error) {
	panic("unexpected call: SearchRawWithContext")
}
func (s *routingStubIndex) FacetSearch(*meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
	panic("unexpected call: FacetSearch")
}
func (s *routingStubIndex) FacetSearchWithContext(context.Context, *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
	panic("unexpected call: FacetSearchWithContext")
}
func (s *routingStubIndex) SearchSimilarDocuments(*meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
	panic("unexpected call: SearchSimilarDocuments")
}
func (s *routingStubIndex) SearchSimilarDocumentsWithContext(context.Context, *meilisearch.SimilarDocumentQuery, *meilisearch.SimilarDocumentResult) error {
	panic("unexpected call: SearchSimilarDocumentsWithContext")
}
func (s *routingStubIndex) GetSettings() (*meilisearch.Settings, error) {
	panic("unexpected call: GetSettings")
}
func (s *routingStubIndex) GetSettingsWithContext(context.Context) (*meilisearch.Settings, error) {
	panic("unexpected call: GetSettingsWithContext")
}
func (s *routingStubIndex) GetRankingRules() (*[]string, error) {
	panic("unexpected call: GetRankingRules")
}
func (s *routingStubIndex) GetRankingRulesWithContext(context.Context) (*[]string, error) {
	panic("unexpected call: GetRankingRulesWithContext")
}
func (s *routingStubIndex) GetDistinctAttribute() (*string, error) {
	panic("unexpected call: GetDistinctAttribute")
}
func (s *routingStubIndex) GetDistinctAttributeWithContext(context.Context) (*string, error) {
	panic("unexpected call: GetDistinctAttributeWithContext")
}
func (s *routingStubIndex) GetSearchableAttributes() (*[]string, error) {
	panic("unexpected call: GetSearchableAttributes")
}
func (s *routingStubIndex) GetSearchableAttributesWithContext(context.Context) (*[]string, error) {
	panic("unexpected call: GetSearchableAttributesWithContext")
}
func (s *routingStubIndex) GetDisplayedAttributes() (*[]string, error) {
	panic("unexpected call: GetDisplayedAttributes")
}
func (s *routingStubIndex) GetDisplayedAttributesWithContext(context.Context) (*[]string, error) {
	panic("unexpected call: GetDisplayedAttributesWithContext")
}
func (s *routingStubIndex) GetStopWords() (*[]string, error) {
	panic("unexpected call: GetStopWords")
}
func (s *routingStubIndex) GetStopWordsWithContext(context.Context) (*[]string, error) {
	panic("unexpected call: GetStopWordsWithContext")
}
func (s *routingStubIndex) GetSynonyms() (*map[string][]string, error) {
	panic("unexpected call: GetSynonyms")
}
func (s *routingStubIndex) GetSynonymsWithContext(context.Context) (*map[string][]string, error) {
	panic("unexpected call: GetSynonymsWithContext")
}
func (s *routingStubIndex) GetFilterableAttributes() (*[]any, error) {
	panic("unexpected call: GetFilterableAttributes")
}
func (s *routingStubIndex) GetFilterableAttributesWithContext(context.Context) (*[]any, error) {
	panic("unexpected call: GetFilterableAttributesWithContext")
}
func (s *routingStubIndex) GetSortableAttributes() (*[]string, error) {
	panic("unexpected call: GetSortableAttributes")
}
func (s *routingStubIndex) GetSortableAttributesWithContext(context.Context) (*[]string, error) {
	panic("unexpected call: GetSortableAttributesWithContext")
}
func (s *routingStubIndex) GetTypoTolerance() (*meilisearch.TypoTolerance, error) {
	panic("unexpected call: GetTypoTolerance")
}
func (s *routingStubIndex) GetTypoToleranceWithContext(context.Context) (*meilisearch.TypoTolerance, error) {
	panic("unexpected call: GetTypoToleranceWithContext")
}
func (s *routingStubIndex) GetPagination() (*meilisearch.Pagination, error) {
	panic("unexpected call: GetPagination")
}
func (s *routingStubIndex) GetPaginationWithContext(context.Context) (*meilisearch.Pagination, error) {
	panic("unexpected call: GetPaginationWithContext")
}
func (s *routingStubIndex) GetFaceting() (*meilisearch.Faceting, error) {
	panic("unexpected call: GetFaceting")
}
func (s *routingStubIndex) GetFacetingWithContext(context.Context) (*meilisearch.Faceting, error) {
	panic("unexpected call: GetFacetingWithContext")
}
func (s *routingStubIndex) GetEmbedders() (map[string]meilisearch.Embedder, error) {
	panic("unexpected call: GetEmbedders")
}
func (s *routingStubIndex) GetEmbeddersWithContext(context.Context) (map[string]meilisearch.Embedder, error) {
	panic("unexpected call: GetEmbeddersWithContext")
}
func (s *routingStubIndex) GetSearchCutoffMs() (int64, error) {
	panic("unexpected call: GetSearchCutoffMs")
}
func (s *routingStubIndex) GetSearchCutoffMsWithContext(context.Context) (int64, error) {
	panic("unexpected call: GetSearchCutoffMsWithContext")
}
func (s *routingStubIndex) GetSeparatorTokens() ([]string, error) {
	panic("unexpected call: GetSeparatorTokens")
}
func (s *routingStubIndex) GetSeparatorTokensWithContext(context.Context) ([]string, error) {
	panic("unexpected call: GetSeparatorTokensWithContext")
}
func (s *routingStubIndex) GetNonSeparatorTokens() ([]string, error) {
	panic("unexpected call: GetNonSeparatorTokens")
}
func (s *routingStubIndex) GetNonSeparatorTokensWithContext(context.Context) ([]string, error) {
	panic("unexpected call: GetNonSeparatorTokensWithContext")
}
func (s *routingStubIndex) GetDictionary() ([]string, error) {
	panic("unexpected call: GetDictionary")
}
func (s *routingStubIndex) GetDictionaryWithContext(context.Context) ([]string, error) {
	panic("unexpected call: GetDictionaryWithContext")
}
func (s *routingStubIndex) GetProximityPrecision() (meilisearch.ProximityPrecisionType, error) {
	panic("unexpected call: GetProximityPrecision")
}
func (s *routingStubIndex) GetProximityPrecisionWithContext(context.Context) (meilisearch.ProximityPrecisionType, error) {
	panic("unexpected call: GetProximityPrecisionWithContext")
}
func (s *routingStubIndex) GetLocalizedAttributes() ([]*meilisearch.LocalizedAttributes, error) {
	panic("unexpected call: GetLocalizedAttributes")
}
func (s *routingStubIndex) GetLocalizedAttributesWithContext(context.Context) ([]*meilisearch.LocalizedAttributes, error) {
	panic("unexpected call: GetLocalizedAttributesWithContext")
}
func (s *routingStubIndex) GetPrefixSearch() (*string, error) {
	panic("unexpected call: GetPrefixSearch")
}
func (s *routingStubIndex) GetPrefixSearchWithContext(context.Context) (*string, error) {
	panic("unexpected call: GetPrefixSearchWithContext")
}
func (s *routingStubIndex) GetFacetSearch() (bool, error) {
	panic("unexpected call: GetFacetSearch")
}
func (s *routingStubIndex) GetFacetSearchWithContext(context.Context) (bool, error) {
	panic("unexpected call: GetFacetSearchWithContext")
}
func (s *routingStubIndex) UpdateSettings(*meilisearch.Settings) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSettings")
}
func (s *routingStubIndex) UpdateSettingsWithContext(context.Context, *meilisearch.Settings) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSettingsWithContext")
}
func (s *routingStubIndex) ResetSettings() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSettings")
}
func (s *routingStubIndex) ResetSettingsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSettingsWithContext")
}
func (s *routingStubIndex) UpdateRankingRules(*[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateRankingRules")
}
func (s *routingStubIndex) UpdateRankingRulesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateRankingRulesWithContext")
}
func (s *routingStubIndex) ResetRankingRules() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetRankingRules")
}
func (s *routingStubIndex) ResetRankingRulesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetRankingRulesWithContext")
}
func (s *routingStubIndex) UpdateDistinctAttribute(string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDistinctAttribute")
}
func (s *routingStubIndex) UpdateDistinctAttributeWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDistinctAttributeWithContext")
}
func (s *routingStubIndex) ResetDistinctAttribute() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDistinctAttribute")
}
func (s *routingStubIndex) ResetDistinctAttributeWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDistinctAttributeWithContext")
}
func (s *routingStubIndex) UpdateSearchableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSearchableAttributes")
}
func (s *routingStubIndex) UpdateSearchableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSearchableAttributesWithContext")
}
func (s *routingStubIndex) ResetSearchableAttributes() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSearchableAttributes")
}
func (s *routingStubIndex) ResetSearchableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSearchableAttributesWithContext")
}
func (s *routingStubIndex) UpdateDisplayedAttributes(*[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDisplayedAttributes")
}
func (s *routingStubIndex) UpdateDisplayedAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDisplayedAttributesWithContext")
}
func (s *routingStubIndex) ResetDisplayedAttributes() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDisplayedAttributes")
}
func (s *routingStubIndex) ResetDisplayedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDisplayedAttributesWithContext")
}
func (s *routingStubIndex) UpdateStopWords(*[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateStopWords")
}
func (s *routingStubIndex) UpdateStopWordsWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateStopWordsWithContext")
}
func (s *routingStubIndex) ResetStopWords() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetStopWords")
}
func (s *routingStubIndex) ResetStopWordsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetStopWordsWithContext")
}
func (s *routingStubIndex) UpdateSynonyms(*map[string][]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSynonyms")
}
func (s *routingStubIndex) UpdateSynonymsWithContext(context.Context, *map[string][]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSynonymsWithContext")
}
func (s *routingStubIndex) ResetSynonyms() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSynonyms")
}
func (s *routingStubIndex) ResetSynonymsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSynonymsWithContext")
}
func (s *routingStubIndex) UpdateFilterableAttributes(*[]any) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFilterableAttributes")
}
func (s *routingStubIndex) UpdateFilterableAttributesWithContext(context.Context, *[]any) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFilterableAttributesWithContext")
}
func (s *routingStubIndex) ResetFilterableAttributes() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFilterableAttributes")
}
func (s *routingStubIndex) ResetFilterableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFilterableAttributesWithContext")
}
func (s *routingStubIndex) UpdateSortableAttributes(*[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSortableAttributes")
}
func (s *routingStubIndex) UpdateSortableAttributesWithContext(context.Context, *[]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSortableAttributesWithContext")
}
func (s *routingStubIndex) ResetSortableAttributes() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSortableAttributes")
}
func (s *routingStubIndex) ResetSortableAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSortableAttributesWithContext")
}
func (s *routingStubIndex) UpdateTypoTolerance(*meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateTypoTolerance")
}
func (s *routingStubIndex) UpdateTypoToleranceWithContext(context.Context, *meilisearch.TypoTolerance) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateTypoToleranceWithContext")
}
func (s *routingStubIndex) ResetTypoTolerance() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetTypoTolerance")
}
func (s *routingStubIndex) ResetTypoToleranceWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetTypoToleranceWithContext")
}
func (s *routingStubIndex) UpdatePagination(*meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdatePagination")
}
func (s *routingStubIndex) UpdatePaginationWithContext(context.Context, *meilisearch.Pagination) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdatePaginationWithContext")
}
func (s *routingStubIndex) ResetPagination() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetPagination")
}
func (s *routingStubIndex) ResetPaginationWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetPaginationWithContext")
}
func (s *routingStubIndex) UpdateFaceting(*meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFaceting")
}
func (s *routingStubIndex) UpdateFacetingWithContext(context.Context, *meilisearch.Faceting) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFacetingWithContext")
}
func (s *routingStubIndex) ResetFaceting() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFaceting")
}
func (s *routingStubIndex) ResetFacetingWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFacetingWithContext")
}
func (s *routingStubIndex) UpdateEmbedders(map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateEmbedders")
}
func (s *routingStubIndex) UpdateEmbeddersWithContext(context.Context, map[string]meilisearch.Embedder) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateEmbeddersWithContext")
}
func (s *routingStubIndex) ResetEmbedders() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetEmbedders")
}
func (s *routingStubIndex) ResetEmbeddersWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetEmbeddersWithContext")
}
func (s *routingStubIndex) UpdateSearchCutoffMs(int64) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSearchCutoffMs")
}
func (s *routingStubIndex) UpdateSearchCutoffMsWithContext(context.Context, int64) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSearchCutoffMsWithContext")
}
func (s *routingStubIndex) ResetSearchCutoffMs() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSearchCutoffMs")
}
func (s *routingStubIndex) ResetSearchCutoffMsWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSearchCutoffMsWithContext")
}
func (s *routingStubIndex) UpdateSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSeparatorTokens")
}
func (s *routingStubIndex) UpdateSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateSeparatorTokensWithContext")
}
func (s *routingStubIndex) ResetSeparatorTokens() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSeparatorTokens")
}
func (s *routingStubIndex) ResetSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetSeparatorTokensWithContext")
}
func (s *routingStubIndex) UpdateNonSeparatorTokens([]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateNonSeparatorTokens")
}
func (s *routingStubIndex) UpdateNonSeparatorTokensWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateNonSeparatorTokensWithContext")
}
func (s *routingStubIndex) ResetNonSeparatorTokens() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetNonSeparatorTokens")
}
func (s *routingStubIndex) ResetNonSeparatorTokensWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetNonSeparatorTokensWithContext")
}
func (s *routingStubIndex) UpdateDictionary([]string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDictionary")
}
func (s *routingStubIndex) UpdateDictionaryWithContext(context.Context, []string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateDictionaryWithContext")
}
func (s *routingStubIndex) ResetDictionary() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDictionary")
}
func (s *routingStubIndex) ResetDictionaryWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetDictionaryWithContext")
}
func (s *routingStubIndex) UpdateProximityPrecision(meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateProximityPrecision")
}
func (s *routingStubIndex) UpdateProximityPrecisionWithContext(context.Context, meilisearch.ProximityPrecisionType) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateProximityPrecisionWithContext")
}
func (s *routingStubIndex) ResetProximityPrecision() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetProximityPrecision")
}
func (s *routingStubIndex) ResetProximityPrecisionWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetProximityPrecisionWithContext")
}
func (s *routingStubIndex) UpdateLocalizedAttributes([]*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateLocalizedAttributes")
}
func (s *routingStubIndex) UpdateLocalizedAttributesWithContext(context.Context, []*meilisearch.LocalizedAttributes) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateLocalizedAttributesWithContext")
}
func (s *routingStubIndex) ResetLocalizedAttributes() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetLocalizedAttributes")
}
func (s *routingStubIndex) ResetLocalizedAttributesWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetLocalizedAttributesWithContext")
}
func (s *routingStubIndex) UpdatePrefixSearch(string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdatePrefixSearch")
}
func (s *routingStubIndex) UpdatePrefixSearchWithContext(context.Context, string) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdatePrefixSearchWithContext")
}
func (s *routingStubIndex) ResetPrefixSearch() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetPrefixSearch")
}
func (s *routingStubIndex) ResetPrefixSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetPrefixSearchWithContext")
}
func (s *routingStubIndex) UpdateFacetSearch(bool) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFacetSearch")
}
func (s *routingStubIndex) UpdateFacetSearchWithContext(context.Context, bool) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateFacetSearchWithContext")
}
func (s *routingStubIndex) ResetFacetSearch() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFacetSearch")
}
func (s *routingStubIndex) ResetFacetSearchWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: ResetFacetSearchWithContext")
}
func (s *routingStubIndex) UpdateIndex(*meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateIndex")
}
func (s *routingStubIndex) UpdateIndexWithContext(context.Context, *meilisearch.UpdateIndexRequestParams) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: UpdateIndexWithContext")
}
func (s *routingStubIndex) Delete(string) (bool, error) {
	panic("unexpected call: Delete")
}
func (s *routingStubIndex) DeleteWithContext(context.Context, string) (bool, error) {
	panic("unexpected call: DeleteWithContext")
}
func (s *routingStubIndex) Compact() (*meilisearch.TaskInfo, error) {
	panic("unexpected call: Compact")
}
func (s *routingStubIndex) CompactWithContext(context.Context) (*meilisearch.TaskInfo, error) {
	panic("unexpected call: CompactWithContext")
}
func (s *routingStubIndex) GetIndexReader() meilisearch.IndexReader         { return s }
func (s *routingStubIndex) GetTaskReader() meilisearch.TaskReader           { return s }
func (s *routingStubIndex) GetDocumentManager() meilisearch.DocumentManager { return s }
func (s *routingStubIndex) GetDocumentReader() meilisearch.DocumentReader   { return s }
func (s *routingStubIndex) GetSettingsManager() meilisearch.SettingsManager { return s }
func (s *routingStubIndex) GetSettingsReader() meilisearch.SettingsReader   { return s }
func (s *routingStubIndex) GetSearch() meilisearch.SearchReader             { return s }

// ---------------------------------------------------------------------------
// Helper: build a SyncManager wired to temp-dir-based stores
// ---------------------------------------------------------------------------

func newRoutingTestSyncManager(t *testing.T, stubIndex *routingStubIndex) *SyncManager {
	t.Helper()

	tmpDir := t.TempDir()
	progressFile := filepath.Join(tmpDir, "progress.json")
	syncStateFile := filepath.Join(tmpDir, "sync_state.json")
	checkpointFile := filepath.Join(tmpDir, "checkpoint.json")

	meiliIdx := search.NewMeiliIndexFromManager(stubIndex)
	progressStore := storage.NewJSONProgressStore(progressFile)

	return NewSyncManager(SyncManagerArgs{
		Index:              meiliIdx,
		ProgressStore:      progressStore,
		MeiliHost:          "http://127.0.0.1:7700",
		MeiliIndex:         "test_items",
		CheckpointTemplate: checkpointFile,
		RootWorkers:        1,
		ProgressEvery:      1,
		Retry:              models.RetryPolicyOptions{},
		MaxConcurrent:      10,
		MinTimeMS:          0,
		SyncStateFile:      syncStateFile,
	})
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestCursorUpdate_FullCrawlSuccess verifies that after a successful full
// crawl, run() writes the sync state file with LastSyncTime > 0.
//
// RED phase: This test is expected to FAIL until the cursor-write logic
// is implemented in run().
func TestCursorUpdate_FullCrawlSuccess(t *testing.T) {
	t.Parallel()

	const rootID int64 = 100

	api := &mockAPIForRouting{
		listFolderChildrenFn: func(_ context.Context, folderID int64, _ int64) (models.FolderChildrenPage, error) {
			if folderID == rootID {
				return models.FolderChildrenPage{
					Files: []models.NpanFile{
						{ID: 1, Name: "test.pdf", ParentID: rootID},
					},
					PageCount: 1,
				}, nil
			}
			return models.FolderChildrenPage{PageCount: 1}, nil
		},
	}

	stubIdx := &routingStubIndex{docCount: 2} // root folder doc + 1 file
	mgr := newRoutingTestSyncManager(t, stubIdx)

	beforeRun := time.Now().UnixMilli()

	noResume := false
	err := mgr.run(context.Background(), api, SyncStartRequest{
		Mode:               models.SyncModeFull,
		RootFolderIDs:      []int64{rootID},
		IncludeDepartments: &noResume, // false: skip department discovery
		ResumeProgress:     &noResume,
	})
	if err != nil {
		t.Fatalf("run() returned unexpected error: %v", err)
	}

	// Verify progress file shows status "done".
	progress, err := mgr.GetProgress()
	if err != nil {
		t.Fatalf("failed to load progress: %v", err)
	}
	if progress.Status != "done" {
		t.Fatalf("expected progress status 'done', got %q", progress.Status)
	}

	// Core assertion: sync state file must be written with LastSyncTime > 0.
	syncStateStore := storage.NewJSONSyncStateStore(mgr.syncStateFile)
	state, err := syncStateStore.Load()
	if err != nil {
		t.Fatalf("failed to load sync state: %v", err)
	}
	if state == nil {
		t.Fatal("sync state file was not written after successful full crawl")
	}
	if state.LastSyncTime <= 0 {
		t.Fatalf("expected LastSyncTime > 0, got %d", state.LastSyncTime)
	}
	if state.LastSyncTime < beforeRun {
		t.Fatalf("LastSyncTime (%d) is before run start (%d)", state.LastSyncTime, beforeRun)
	}

	t.Logf("OK: LastSyncTime = %d (run started at %d)", state.LastSyncTime, beforeRun)
}

// TestCursorUpdate_FullCrawlFailure verifies that after a failed full crawl,
// run() does NOT write a sync state (or leaves LastSyncTime at 0).
//
// RED phase: This test is expected to FAIL until the cursor-write logic
// is implemented in run().
func TestCursorUpdate_FullCrawlFailure(t *testing.T) {
	t.Parallel()

	const rootID int64 = 200
	crawlErr := errors.New("simulated API failure")

	api := &mockAPIForRouting{
		listFolderChildrenFn: func(_ context.Context, _ int64, _ int64) (models.FolderChildrenPage, error) {
			return models.FolderChildrenPage{}, crawlErr
		},
	}

	stubIdx := &routingStubIndex{docCount: 0}
	mgr := newRoutingTestSyncManager(t, stubIdx)

	noResume := false
	err := mgr.run(context.Background(), api, SyncStartRequest{
		Mode:               models.SyncModeFull,
		RootFolderIDs:      []int64{rootID},
		IncludeDepartments: &noResume,
		ResumeProgress:     &noResume,
	})
	if err == nil {
		t.Fatal("expected run() to return an error for a failed crawl, got nil")
	}

	// Core assertion: sync state file must NOT exist or LastSyncTime must be 0.
	syncStateStore := storage.NewJSONSyncStateStore(mgr.syncStateFile)
	state, err := syncStateStore.Load()
	if err != nil {
		t.Fatalf("failed to load sync state: %v", err)
	}
	if state != nil && state.LastSyncTime > 0 {
		t.Fatalf("sync state should not be written after a failed crawl, but got LastSyncTime = %d", state.LastSyncTime)
	}

	if state == nil {
		t.Log("OK: sync state file does not exist (expected after failed crawl)")
	} else {
		t.Logf("OK: sync state exists but LastSyncTime = %d (expected 0)", state.LastSyncTime)
	}
}

// Compile-time interface check.
var _ meilisearch.IndexManager = (*routingStubIndex)(nil)
