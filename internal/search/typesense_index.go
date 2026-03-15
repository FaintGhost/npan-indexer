package search

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"npan/internal/models"
)

type TypesenseIndex struct {
	host       string
	apiKey     string
	collection string
	client     *http.Client
}

func NewTypesenseIndex(host string, apiKey string, collection string) *TypesenseIndex {
	return &TypesenseIndex{
		host:       strings.TrimRight(strings.TrimSpace(host), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		collection: strings.TrimSpace(collection),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type typesenseCollectionSchema struct {
	Name                string                     `json:"name"`
	DefaultSortingField string                     `json:"default_sorting_field,omitempty"`
	TokenSeparators     []string                   `json:"token_separators,omitempty"`
	Fields              []typesenseCollectionField `json:"fields"`
}

type typesenseCollectionField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Facet    bool   `json:"facet,omitempty"`
	Optional bool   `json:"optional,omitempty"`
	Sort     bool   `json:"sort,omitempty"`
}

type typesenseCollectionInfo struct {
	Name            string                     `json:"name"`
	NumDocuments    int64                      `json:"num_documents"`
	TokenSeparators []string                   `json:"token_separators"`
	Fields          []typesenseCollectionField `json:"fields"`
}

type typesenseSearchResponse struct {
	Found int64                `json:"found"`
	Hits  []typesenseSearchHit `json:"hits"`
}

type typesenseSearchHit struct {
	Document   json.RawMessage            `json:"document"`
	Highlights []typesenseHighlight       `json:"highlights"`
	Highlight  map[string]json.RawMessage `json:"highlight"`
}

type typesenseHighlight struct {
	Field   string `json:"field"`
	Snippet string `json:"snippet"`
	Value   string `json:"value"`
}

func (t *TypesenseIndex) EnsureSettings(ctx context.Context) error {
	info, status, err := t.fetchCollection(ctx)
	if err != nil {
		if status != http.StatusNotFound {
			return err
		}
	}
	if status == http.StatusOK {
		return validateTypesenseCollection(info)
	}

	schema := typesenseCollectionSchema{
		Name:                t.collection,
		DefaultSortingField: "modified_at",
		TokenSeparators:     []string{"-", "_"},
		Fields: []typesenseCollectionField{
			{Name: "doc_id", Type: "string"},
			{Name: "source_id", Type: "int64", Sort: true},
			{Name: "type", Type: "string", Facet: true},
			{Name: "name", Type: "string"},
			{Name: "name_base", Type: "string"},
			{Name: "name_ext", Type: "string", Optional: true},
			{Name: "file_category", Type: "string", Facet: true, Optional: true},
			{Name: "path_text", Type: "string"},
			{Name: "parent_id", Type: "int64", Facet: true, Sort: true},
			{Name: "modified_at", Type: "int64", Facet: true, Sort: true},
			{Name: "created_at", Type: "int64", Sort: true},
			{Name: "size", Type: "int64", Sort: true},
			{Name: "sha1", Type: "string", Optional: true},
			{Name: "in_trash", Type: "bool", Facet: true},
			{Name: "is_deleted", Type: "bool", Facet: true},
		},
	}

	var created typesenseCollectionInfo
	return t.doJSON(ctx, http.MethodPost, "/collections", nil, schema, &created)
}

func (t *TypesenseIndex) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var body bytes.Buffer
	for _, doc := range docs {
		encoded, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		body.Write(encoded)
		body.WriteByte('\n')
	}

	query := url.Values{}
	query.Set("action", "upsert")
	respBody, err := t.do(ctx, http.MethodPost, fmt.Sprintf("/collections/%s/documents/import", url.PathEscape(t.collection)), query, "text/plain", &body)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(respBody))
	var failures []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var result struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return fmt.Errorf("解析 Typesense import 结果失败: %w", err)
		}
		if !result.Success {
			failures = append(failures, result.Error)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(failures) > 0 {
		return fmt.Errorf("typesense 导入失败: %s", strings.Join(failures, "; "))
	}
	return nil
}

func (t *TypesenseIndex) DeleteDocuments(ctx context.Context, docIDs []string) error {
	if len(docIDs) == 0 {
		return nil
	}

	parts := make([]string, 0, len(docIDs))
	for _, docID := range docIDs {
		parts = append(parts, fmt.Sprintf("doc_id:=%s", quoteTypesenseString(docID)))
	}
	query := url.Values{}
	query.Set("filter_by", strings.Join(parts, " || "))
	query.Set("batch_size", fmt.Sprintf("%d", len(docIDs)))
	_, err := t.do(ctx, http.MethodDelete, fmt.Sprintf("/collections/%s/documents", url.PathEscape(t.collection)), query, "", nil)
	return err
}

func (t *TypesenseIndex) DeleteAllDocuments(ctx context.Context) error {
	if _, err := t.do(ctx, http.MethodDelete, fmt.Sprintf("/collections/%s", url.PathEscape(t.collection)), nil, "", nil); err != nil && !isNotFoundError(err) {
		return err
	}
	return t.EnsureSettings(ctx)
}

func (t *TypesenseIndex) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	ctx := context.Background()
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	queryText := strings.TrimSpace(preprocessQuery(params.Query))
	if queryText == "" {
		queryText = "*"
	}

	search := func(dropTokens int64) (*typesenseSearchResponse, error) {
		query := url.Values{}
		query.Set("q", queryText)
		query.Set("query_by", "name_base,name_ext,name,path_text")
		query.Set("query_by_weights", "8,6,4,1")
		query.Set("page", fmt.Sprintf("%d", page))
		query.Set("per_page", fmt.Sprintf("%d", pageSize))
		query.Set("exhaustive_search", "true")
		query.Set("sort_by", "_text_match:desc,modified_at:desc")
		query.Set("highlight_fields", "name")
		query.Set("highlight_start_tag", "<mark>")
		query.Set("highlight_end_tag", "</mark>")
		query.Set("drop_tokens_threshold", fmt.Sprintf("%d", dropTokens))
		if filterBy := buildTypesenseFilter(params); filterBy != "" {
			query.Set("filter_by", filterBy)
		}

		var response typesenseSearchResponse
		err := t.doJSON(ctx, http.MethodGet, fmt.Sprintf("/collections/%s/documents/search", url.PathEscape(t.collection)), query, nil, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	response, err := search(0)
	if err != nil {
		return nil, 0, err
	}
	if response.Found == 0 && len(response.Hits) == 0 && strings.TrimSpace(params.Query) != "" {
		response, err = search(1)
		if err != nil {
			return nil, 0, err
		}
	}

	docs := make([]models.IndexDocument, 0, len(response.Hits))
	for _, hit := range response.Hits {
		var doc models.IndexDocument
		if err := json.Unmarshal(hit.Document, &doc); err != nil {
			return nil, 0, err
		}
		doc.HighlightedName = typesenseHighlightName(hit)
		docs = append(docs, doc)
	}

	return docs, response.Found, nil
}

func (t *TypesenseIndex) Ping() error {
	_, err := t.do(context.Background(), http.MethodGet, "/health", nil, "", nil)
	return err
}

func (t *TypesenseIndex) DocumentCount(ctx context.Context) (int64, error) {
	info, _, err := t.fetchCollection(ctx)
	if err != nil {
		if isNotFoundError(err) {
			return 0, nil
		}
		return 0, err
	}
	return info.NumDocuments, nil
}

func (t *TypesenseIndex) fetchCollection(ctx context.Context) (typesenseCollectionInfo, int, error) {
	respBody, status, err := t.doWithStatus(ctx, http.MethodGet, fmt.Sprintf("/collections/%s", url.PathEscape(t.collection)), nil, "", nil)
	if err != nil {
		return typesenseCollectionInfo{}, status, err
	}
	var info typesenseCollectionInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return typesenseCollectionInfo{}, status, err
	}
	return info, status, nil
}

func validateTypesenseCollection(info typesenseCollectionInfo) error {
	required := map[string]string{
		"doc_id":        "string",
		"source_id":     "int64",
		"type":          "string",
		"name":          "string",
		"name_base":     "string",
		"name_ext":      "string",
		"file_category": "string",
		"path_text":     "string",
		"parent_id":     "int64",
		"modified_at":   "int64",
		"created_at":    "int64",
		"size":          "int64",
		"sha1":          "string",
		"in_trash":      "bool",
		"is_deleted":    "bool",
	}
	seen := map[string]string{}
	for _, field := range info.Fields {
		seen[field.Name] = field.Type
	}
	for name, wantType := range required {
		if gotType, ok := seen[name]; !ok {
			return fmt.Errorf("typesense collection 缺少字段 %s", name)
		} else if gotType != wantType {
			return fmt.Errorf("typesense collection 字段 %s 类型错误: got=%s want=%s", name, gotType, wantType)
		}
	}
	if !containsTokenSeparator(info.TokenSeparators, "-") || !containsTokenSeparator(info.TokenSeparators, "_") {
		return fmt.Errorf("typesense collection 缺少必要 token_separators，需包含 '-' 和 '_'")
	}
	return nil
}

func buildTypesenseFilter(params models.LocalSearchParams) string {
	filters := make([]string, 0, 8)
	if params.Type != "" && params.Type != "all" {
		filters = append(filters, fmt.Sprintf("type:=%s", quoteTypesenseString(params.Type)))
	}
	if params.ParentID != nil {
		filters = append(filters, fmt.Sprintf("parent_id:=%d", *params.ParentID))
	}
	if params.UpdatedAfter != nil {
		filters = append(filters, fmt.Sprintf("modified_at:>=%d", *params.UpdatedAfter))
	}
	if params.UpdatedBefore != nil {
		filters = append(filters, fmt.Sprintf("modified_at:<=%d", *params.UpdatedBefore))
	}
	if !params.IncludeDeleted {
		filters = append(filters, "is_deleted:=false")
		filters = append(filters, "in_trash:=false")
	}
	return strings.Join(filters, " && ")
}

func typesenseHighlightName(hit typesenseSearchHit) string {
	for _, highlight := range hit.Highlights {
		if highlight.Field != "name" {
			continue
		}
		if highlight.Snippet != "" {
			return highlight.Snippet
		}
		if highlight.Value != "" {
			return highlight.Value
		}
	}
	if raw, ok := hit.Highlight["name"]; ok {
		var nested struct {
			Snippet string `json:"snippet"`
			Value   string `json:"value"`
		}
		if err := json.Unmarshal(raw, &nested); err == nil {
			if nested.Snippet != "" {
				return nested.Snippet
			}
			return nested.Value
		}
		var plain string
		if err := json.Unmarshal(raw, &plain); err == nil {
			return plain
		}
	}
	return ""
}

func quoteTypesenseString(value string) string {
	escaped := strings.ReplaceAll(value, "`", "\\`")
	return "`" + escaped + "`"
}

func containsTokenSeparator(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func (t *TypesenseIndex) doJSON(ctx context.Context, method string, path string, query url.Values, body any, out any) error {
	var reader io.Reader
	contentType := ""
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
		contentType = "application/json"
	}

	respBody, err := t.do(ctx, method, path, query, contentType, reader)
	if err != nil {
		return err
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

func (t *TypesenseIndex) do(ctx context.Context, method string, path string, query url.Values, contentType string, body io.Reader) ([]byte, error) {
	respBody, _, err := t.doWithStatus(ctx, method, path, query, contentType, body)
	return respBody, err
}

func (t *TypesenseIndex) doWithStatus(ctx context.Context, method string, path string, query url.Values, contentType string, body io.Reader) ([]byte, int, error) {
	endpoint := t.host + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, 0, err
	}
	if t.apiKey != "" {
		req.Header.Set("X-TYPESENSE-API-KEY", t.apiKey)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, &typesenseAPIError{
		statusCode: resp.StatusCode,
		body:       strings.TrimSpace(string(respBody)),
	}
}

type typesenseAPIError struct {
	statusCode int
	body       string
}

func (e *typesenseAPIError) Error() string {
	if e.body == "" {
		return fmt.Sprintf("typesense 请求失败: status=%d", e.statusCode)
	}
	return fmt.Sprintf("typesense 请求失败: status=%d body=%s", e.statusCode, e.body)
}

func isNotFoundError(err error) bool {
	var apiErr *typesenseAPIError
	return errors.As(err, &apiErr) && apiErr.statusCode == http.StatusNotFound
}
