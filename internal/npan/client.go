package npan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"npan/internal/models"
)

type HTTPClientOptions struct {
	BaseURL        string
	Token          string
	TokenRefresher func(ctx context.Context) (string, error)
	Client         *http.Client
}

type HTTPClient struct {
	baseURL        string
	token          string
	tokenRefresher func(ctx context.Context) (string, error)
	client         *http.Client
	mu             sync.RWMutex
	refreshMu      sync.Mutex
}

func NewHTTPClient(options HTTPClientOptions) *HTTPClient {
	httpClient := options.Client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &HTTPClient{
		baseURL:        strings.TrimRight(options.BaseURL, "/"),
		token:          strings.TrimSpace(options.Token),
		tokenRefresher: options.TokenRefresher,
		client:         httpClient,
	}
}

func (c *HTTPClient) getToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

func (c *HTTPClient) setToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = strings.TrimSpace(token)
}

func (c *HTTPClient) refreshToken(ctx context.Context, staleToken string) (string, error) {
	if c.tokenRefresher == nil {
		return "", fmt.Errorf("token 无法自动刷新")
	}

	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	currentToken := c.getToken()
	if currentToken != "" && currentToken != staleToken {
		return currentToken, nil
	}

	newToken, err := c.tokenRefresher(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(newToken) == "" {
		return "", fmt.Errorf("刷新 token 失败: 新 token 为空")
	}

	c.setToken(newToken)
	return newToken, nil
}

func readStatusError(resp *http.Response, fallback string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = fallback
	}
	return &StatusError{Status: resp.StatusCode, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, message)}
}

func toInt64(input any, fallback int64) int64 {
	switch value := input.(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func toBool(input any) bool {
	switch value := input.(type) {
	case bool:
		return value
	case string:
		lower := strings.ToLower(strings.TrimSpace(value))
		return lower == "true" || lower == "1"
	case float64:
		return value != 0
	case int:
		return value != 0
	}
	return false
}

func (c *HTTPClient) request(ctx context.Context, method string, path string, query url.Values, out any) error {
	fullURL := c.baseURL + path
	if query != nil {
		encoded := query.Encode()
		if encoded != "" {
			fullURL = fullURL + "?" + encoded
		}
	}

	attempt := 0
	for {
		token := c.getToken()

		req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 && c.tokenRefresher != nil {
			_ = resp.Body.Close()
			if _, refreshErr := c.refreshToken(ctx, token); refreshErr != nil {
				return fmt.Errorf("请求返回 401，且刷新 token 失败: %w", refreshErr)
			}
			attempt++
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return readStatusError(resp, "请求失败")
		}

		if out == nil {
			return nil
		}
		return json.NewDecoder(io.LimitReader(resp.Body, 10*1024*1024)).Decode(out)
	}
}

func mapFolder(row map[string]any) models.NpanFolder {
	parentID := int64(0)
	if parentRaw, ok := row["parent"].(map[string]any); ok {
		parentID = toInt64(parentRaw["id"], 0)
	}

	return models.NpanFolder{
		ID:         toInt64(row["id"], 0),
		Name:       fmt.Sprintf("%v", row["name"]),
		ParentID:   parentID,
		ItemCount:  toInt64(row["item_count"], 0),
		ModifiedAt: toInt64(row["modified_at"], 0),
		InTrash:    toBool(row["in_trash"]),
		IsDeleted:  toBool(row["is_deleted"]),
	}
}

func mapFile(row map[string]any) models.NpanFile {
	parentID := int64(0)
	if parentRaw, ok := row["parent"].(map[string]any); ok {
		parentID = toInt64(parentRaw["id"], 0)
	}

	return models.NpanFile{
		ID:         toInt64(row["id"], 0),
		Name:       fmt.Sprintf("%v", row["name"]),
		ParentID:   parentID,
		Size:       toInt64(row["size"], 0),
		ModifiedAt: toInt64(row["modified_at"], 0),
		CreatedAt:  toInt64(row["created_at"], 0),
		SHA1:       fmt.Sprintf("%v", row["sha1"]),
		InTrash:    toBool(row["in_trash"]),
		IsDeleted:  toBool(row["is_deleted"]),
	}
}

func (c *HTTPClient) ListFolderChildren(ctx context.Context, folderID int64, pageID int64) (models.FolderChildrenPage, error) {
	var body struct {
		Files        []map[string]any `json:"files"`
		Folders      []map[string]any `json:"folders"`
		PageID       int64            `json:"page_id"`
		PageCount    int64            `json:"page_count"`
		PageCapacity int64            `json:"page_capacity"`
		TotalCount   int64            `json:"total_count"`
	}

	query := url.Values{}
	query.Set("page_id", strconv.FormatInt(pageID, 10))
	err := c.request(ctx, http.MethodGet, "/api/v2/folder/"+strconv.FormatInt(folderID, 10)+"/children", query, &body)
	if err != nil {
		return models.FolderChildrenPage{}, err
	}

	files := make([]models.NpanFile, 0, len(body.Files))
	for _, row := range body.Files {
		files = append(files, mapFile(row))
	}

	folders := make([]models.NpanFolder, 0, len(body.Folders))
	for _, row := range body.Folders {
		folders = append(folders, mapFolder(row))
	}

	return models.FolderChildrenPage{
		Files:        files,
		Folders:      folders,
		PageID:       body.PageID,
		PageCount:    body.PageCount,
		PageCapacity: body.PageCapacity,
		TotalCount:   body.TotalCount,
	}, nil
}

func (c *HTTPClient) GetDownloadURL(ctx context.Context, fileID int64, validPeriod *int64) (models.DownloadURLResult, error) {
	query := url.Values{}
	if validPeriod != nil {
		query.Set("valid_period", strconv.FormatInt(*validPeriod, 10))
	}

	var body models.DownloadURLResult
	err := c.request(ctx, http.MethodGet, "/api/v2/file/"+strconv.FormatInt(fileID, 10)+"/download", query, &body)
	if err != nil {
		return models.DownloadURLResult{}, err
	}
	return body, nil
}

func (c *HTTPClient) SearchUpdatedWindow(ctx context.Context, queryWords string, start *int64, end *int64, pageID int64) (map[string]any, error) {
	trimmedQuery := strings.TrimSpace(queryWords)
	if trimmedQuery == "" {
		trimmedQuery = "* OR *"
	}

	query := url.Values{}
	query.Set("query_words", trimmedQuery)
	query.Set("type", "all")
	query.Set("page_id", strconv.FormatInt(pageID, 10))
	query.Set("query_filter", "all")

	rangeStart := ""
	rangeEnd := ""
	if start != nil {
		rangeStart = strconv.FormatInt(*start, 10)
	}
	if end != nil {
		rangeEnd = strconv.FormatInt(*end, 10)
	}
	query.Set("updated_time_range", rangeStart+","+rangeEnd)

	body := map[string]any{}
	err := c.request(ctx, http.MethodGet, "/api/v2/item/search", query, &body)
	return body, err
}

func (c *HTTPClient) ListUserDepartments(ctx context.Context) ([]models.NpanDepartment, error) {
	var body struct {
		Departments []map[string]any `json:"departments"`
	}

	if err := c.request(ctx, http.MethodGet, "/api/v2/user/departments", nil, &body); err != nil {
		return nil, err
	}

	departments := make([]models.NpanDepartment, 0, len(body.Departments))
	for _, dep := range body.Departments {
		departments = append(departments, models.NpanDepartment{
			ID:   toInt64(dep["id"], 0),
			Name: fmt.Sprintf("%v", dep["name"]),
		})
	}

	return departments, nil
}

func (c *HTTPClient) ListDepartmentFolders(ctx context.Context, departmentID int64) ([]models.NpanFolder, error) {
	var body struct {
		Folders []map[string]any `json:"folders"`
	}

	query := url.Values{}
	query.Set("department_id", strconv.FormatInt(departmentID, 10))

	if err := c.request(ctx, http.MethodGet, "/api/v2/folder/department_folders", query, &body); err != nil {
		return nil, err
	}

	result := make([]models.NpanFolder, 0, len(body.Folders))
	for _, folder := range body.Folders {
		result = append(result, models.NpanFolder{
			ID:        toInt64(folder["id"], 0),
			Name:      fmt.Sprintf("%v", folder["name"]),
			ItemCount: toInt64(folder["item_count"], 0),
		})
	}

	return result, nil
}

func (c *HTTPClient) SearchItems(ctx context.Context, params models.RemoteSearchParams) (models.RemoteSearchResponse, error) {
	query := url.Values{}
	query.Set("query_words", params.QueryWords)
	query.Set("type", params.Type)
	query.Set("page_id", strconv.FormatInt(params.PageID, 10))
	query.Set("query_filter", params.QueryFilter)

	if params.SearchInFolder != nil {
		query.Set("search_in_folder", strconv.FormatInt(*params.SearchInFolder, 10))
	}
	if strings.TrimSpace(params.UpdatedTimeRange) != "" {
		query.Set("updated_time_range", params.UpdatedTimeRange)
	}

	var body struct {
		Files        []map[string]any `json:"files"`
		Folders      []map[string]any `json:"folders"`
		TotalCount   int64            `json:"total_count"`
		PageID       int64            `json:"page_id"`
		PageCapacity int64            `json:"page_capacity"`
		PageCount    int64            `json:"page_count"`
	}

	if err := c.request(ctx, http.MethodGet, "/api/v2/item/search", query, &body); err != nil {
		return models.RemoteSearchResponse{}, err
	}

	files := make([]models.RemoteSearchItem, 0, len(body.Files))
	for _, item := range body.Files {
		files = append(files, models.RemoteSearchItem{
			ID:   toInt64(item["id"], 0),
			Name: fmt.Sprintf("%v", item["name"]),
			Type: "file",
		})
	}

	folders := make([]models.RemoteSearchItem, 0, len(body.Folders))
	for _, item := range body.Folders {
		folders = append(folders, models.RemoteSearchItem{
			ID:   toInt64(item["id"], 0),
			Name: fmt.Sprintf("%v", item["name"]),
			Type: "folder",
		})
	}

	return models.RemoteSearchResponse{
		Files:        files,
		Folders:      folders,
		TotalCount:   body.TotalCount,
		PageID:       body.PageID,
		PageCapacity: body.PageCapacity,
		PageCount:    body.PageCount,
	}, nil
}
