package search

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"

	"npan/internal/models"
)

type MeiliIndex struct {
	index meilisearch.IndexManager
}

func NewMeiliIndex(host string, apiKey string, indexName string) *MeiliIndex {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	return &MeiliIndex{index: client.Index(indexName)}
}

func NewMeiliIndexFromManager(index meilisearch.IndexManager) *MeiliIndex {
	return &MeiliIndex{index: index}
}

func (m *MeiliIndex) EnsureSettings() error {
	_, err := m.index.UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{"name", "path_text"},
		FilterableAttributes: []string{"type", "parent_id", "modified_at", "in_trash", "is_deleted"},
		SortableAttributes:   []string{"modified_at", "size", "created_at"},
	})
	return err
}

func (m *MeiliIndex) UpsertDocuments(docs []models.IndexDocument) error {
	if len(docs) == 0 {
		return nil
	}

	primaryKey := "doc_id"
	_, err := m.index.AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &primaryKey})
	return err
}

func (m *MeiliIndex) DeleteDocuments(docIDs []string) error {
	if len(docIDs) == 0 {
		return nil
	}

	_, err := m.index.DeleteDocuments(docIDs, nil)
	return err
}

func (m *MeiliIndex) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	filters := make([]string, 0, 8)

	if params.Type != "" && params.Type != "all" {
		filters = append(filters, fmt.Sprintf("type = %s", params.Type))
	}
	if params.ParentID != nil {
		filters = append(filters, fmt.Sprintf("parent_id = %d", *params.ParentID))
	}
	if params.UpdatedAfter != nil {
		filters = append(filters, fmt.Sprintf("modified_at >= %d", *params.UpdatedAfter))
	}
	if params.UpdatedBefore != nil {
		filters = append(filters, fmt.Sprintf("modified_at <= %d", *params.UpdatedBefore))
	}
	if !params.IncludeDeleted {
		filters = append(filters, "is_deleted = false")
		filters = append(filters, "in_trash = false")
	}

	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	request := &meilisearch.SearchRequest{
		Filter:      filters,
		Page:        page,
		HitsPerPage: pageSize,
		Sort:        []string{"modified_at:desc"},
	}

	response, err := m.index.Search(params.Query, request)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]models.IndexDocument, 0, len(response.Hits))
	if err := response.Hits.DecodeInto(&docs); err != nil {
		return nil, 0, err
	}

	total := response.TotalHits
	if total == 0 {
		total = response.EstimatedTotalHits
	}
	if total == 0 {
		total = int64(len(docs))
	}

	return docs, total, nil
}
