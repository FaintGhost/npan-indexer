package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/meilisearch/meilisearch-go"

	"npan/internal/models"
)

type MeiliIndex struct {
	index meilisearch.IndexManager
}

const defaultTaskPollInterval = 100 * time.Millisecond

func NewMeiliIndex(host string, apiKey string, indexName string) *MeiliIndex {
	client := meilisearch.New(host,
		meilisearch.WithAPIKey(apiKey),
		meilisearch.WithCustomJsonMarshaler(sonic.Marshal),
		meilisearch.WithCustomJsonUnmarshaler(sonic.Unmarshal),
	)
	return &MeiliIndex{index: client.Index(indexName)}
}

func NewMeiliIndexFromManager(index meilisearch.IndexManager) *MeiliIndex {
	return &MeiliIndex{index: index}
}

func (m *MeiliIndex) waitTask(ctx context.Context, taskInfo *meilisearch.TaskInfo) error {
	if taskInfo == nil {
		return fmt.Errorf("meilisearch 未返回 task 信息")
	}

	task, err := m.index.WaitForTaskWithContext(ctx, taskInfo.TaskUID, defaultTaskPollInterval)
	if err != nil {
		return err
	}

	if task.Status == meilisearch.TaskStatusSucceeded {
		return nil
	}

	return fmt.Errorf(
		"meilisearch task 执行失败: task_uid=%d status=%s code=%s message=%s",
		taskInfo.TaskUID,
		task.Status,
		task.Error.Code,
		task.Error.Message,
	)
}

func (m *MeiliIndex) EnsureSettings(ctx context.Context) error {
	taskInfo, err := m.index.UpdateSettingsWithContext(ctx, &meilisearch.Settings{
		RankingRules:         []string{"words", "typo", "proximity", "attribute", "exactness", "modified_at:desc"},
		SearchableAttributes: []string{"name", "path_text"},
		FilterableAttributes: []string{"type", "parent_id", "modified_at", "in_trash", "is_deleted"},
		SortableAttributes:   []string{"modified_at", "size", "created_at"},
		DisplayedAttributes:  []string{"doc_id", "source_id", "type", "name", "path_text", "parent_id", "modified_at", "created_at", "size"},
		StopWords:            []string{"的", "了", "在", "是", "和", "就", "都", "而", "及", "与"},
		TypoTolerance: &meilisearch.TypoTolerance{
			Enabled: true,
			MinWordSizeForTypos: meilisearch.MinWordSizeForTypos{
				OneTypo:  5,
				TwoTypos: 9,
			},
			DisableOnAttributes: []string{"path_text"},
			DisableOnWords:      []string{"pdf", "docx", "xlsx", "pptx", "jpg", "png", "mp4", "zip", "rar", "exe", "apk", "bin", "iso"},
		},
		ProximityPrecision: meilisearch.ByAttribute,
	})
	if err != nil {
		return err
	}
	return m.waitTask(ctx, taskInfo)
}

func (m *MeiliIndex) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
	if len(docs) == 0 {
		return nil
	}

	primaryKey := "doc_id"
	taskInfo, err := m.index.AddDocumentsWithContext(ctx, docs, &meilisearch.DocumentOptions{PrimaryKey: &primaryKey})
	if err != nil {
		return err
	}
	return m.waitTask(ctx, taskInfo)
}

func (m *MeiliIndex) DeleteDocuments(ctx context.Context, docIDs []string) error {
	if len(docIDs) == 0 {
		return nil
	}

	taskInfo, err := m.index.DeleteDocumentsWithContext(ctx, docIDs, nil)
	if err != nil {
		return err
	}
	return m.waitTask(ctx, taskInfo)
}

// knownExtensions 是常见文件扩展名集合，用于查询预处理。
var knownExtensions = map[string]bool{
	"pdf": true, "docx": true, "xlsx": true, "pptx": true,
	"doc": true, "xls": true, "ppt": true,
	"jpg": true, "jpeg": true, "png": true, "gif": true, "bmp": true,
	"mp4": true, "avi": true, "mov": true, "mkv": true,
	"zip": true, "rar": true, "7z": true, "tar": true, "gz": true,
	"exe": true, "apk": true, "bin": true, "iso": true,
	"dwg": true, "dxf": true, "cad": true,
	"txt": true, "csv": true, "json": true, "xml": true,
}

// reorderQuery 将查询中的文件扩展名移到前面，确保实际搜索词
// 留在最后以获得 Meilisearch 的前缀匹配。
// 例如 "mx40 spec pdf" → "pdf mx40 spec"，使 "spec" 前缀匹配 "specifications"。
func reorderQuery(query string) string {
	words := strings.Fields(query)
	if len(words) <= 1 {
		return query
	}
	var ext, terms []string
	for _, w := range words {
		if knownExtensions[strings.ToLower(w)] {
			ext = append(ext, w)
		} else {
			terms = append(terms, w)
		}
	}
	if len(ext) == 0 || len(terms) == 0 {
		return query
	}
	return strings.Join(append(ext, terms...), " ")
}

func (m *MeiliIndex) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	filters := make([]string, 0, 8)

	if params.Type != "" && params.Type != "all" {
		filters = append(filters, fmt.Sprintf("type = '%s'", params.Type))
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
		AttributesToRetrieve: []string{
			"doc_id", "source_id", "type", "name", "path_text",
			"parent_id", "modified_at", "created_at", "size",
		},
		AttributesToHighlight: []string{"name"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
	}

	response, err := m.index.Search(reorderQuery(params.Query), request)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]models.IndexDocument, 0, len(response.Hits))
	if err := response.Hits.DecodeInto(&docs); err != nil {
		return nil, 0, err
	}

	for i, hit := range response.Hits {
		if i >= len(docs) {
			break
		}
		formatted, ok := hit["_formatted"]
		if !ok {
			continue
		}
		var formattedObj map[string]json.RawMessage
		if err := json.Unmarshal(formatted, &formattedObj); err != nil {
			continue
		}
		if nameRaw, ok := formattedObj["name"]; ok {
			var name string
			if err := json.Unmarshal(nameRaw, &name); err == nil {
				docs[i].HighlightedName = name
			}
		}
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

// Ping 检查 Meilisearch 索引连通性。
func (m *MeiliIndex) Ping() error {
	_, err := m.index.GetStats()
	return err
}
