package httpx

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"npan/internal/config"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/service"
)

// searchService 定义 handler 层对搜索服务的依赖。
type searchService interface {
	Query(models.LocalSearchParams) (search.QueryResult, error)
	Ping() error
}

type Handlers struct {
	cfg          config.Config
	queryService searchService
	syncManager  *service.SyncManager
}

func NewHandlers(cfg config.Config, queryService search.Searcher, syncManager *service.SyncManager) *Handlers {
	return &Handlers{
		cfg:          cfg,
		queryService: queryService,
		syncManager:  syncManager,
	}
}

func parseInt64Pointer(raw string) (*int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseBool(raw string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}

type authPayload struct {
	Token        string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	SubID        int64  `json:"sub_id"`
	SubType      string `json:"sub_type"`
	OAuthHost    string `json:"oauth_host"`
}

// allowConfigFallback 从 echo 上下文中读取 allow_config_fallback 标记。
// 若 EmbeddedAuth 中间件已设置该值则使用；否则使用全局配置。
func (h *Handlers) allowConfigFallback(c *echo.Context) bool {
	if v, ok := c.Get("allow_config_fallback").(bool); ok {
		return v
	}
	return h.cfg.AllowConfigAuthFallback
}

func (h *Handlers) resolveAuthOptions(c *echo.Context, payload authPayload) npan.AuthResolverOptions {
	tokenFromHeader := parseBearerHeaderValue(c.Request().Header.Get("Authorization"))
	fallback := h.allowConfigFallback(c)

	tokenCandidates := []string{
		payload.Token,
		tokenFromHeader,
		strings.TrimSpace(c.QueryParam("token")),
	}
	clientIDCandidates := []string{payload.ClientID}
	clientSecretCandidates := []string{payload.ClientSecret}
	subIDCandidates := []int64{payload.SubID}
	oauthHostCandidates := []string{payload.OAuthHost}

	if fallback {
		tokenCandidates = append(tokenCandidates, h.cfg.Token)
		clientIDCandidates = append(clientIDCandidates, h.cfg.ClientID)
		clientSecretCandidates = append(clientSecretCandidates, h.cfg.ClientSecret)
		subIDCandidates = append(subIDCandidates, h.cfg.SubID)
		oauthHostCandidates = append(oauthHostCandidates, h.cfg.OAuthHost)
	}

	subType := npan.TokenSubjectType(payload.SubType)
	if subType == "" && fallback {
		subType = h.cfg.SubType
	}
	if subType == "" {
		subType = npan.TokenSubjectUser
	}

	oauthHost := firstNotEmpty(oauthHostCandidates...)
	if oauthHost == "" {
		oauthHost = npan.DefaultOAuthHost
	}

	return npan.AuthResolverOptions{
		Token:        firstNotEmpty(tokenCandidates...),
		ClientID:     firstNotEmpty(clientIDCandidates...),
		ClientSecret: firstNotEmpty(clientSecretCandidates...),
		SubID:        firstPositive(subIDCandidates...),
		SubType:      subType,
		OAuthHost:    oauthHost,
	}
}

func (h *Handlers) resolveToken(c *echo.Context, payload authPayload) (string, npan.AuthResolverOptions, error) {
	authOptions := h.resolveAuthOptions(c, payload)
	token, err := npan.ResolveBearerToken(c.Request().Context(), nil, authOptions)
	if err != nil {
		return "", authOptions, err
	}
	return token, authOptions, nil
}

func firstNotEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstPositive(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (h *Handlers) newAPIClient(token string, authOptions npan.AuthResolverOptions) npan.API {
	return npan.NewHTTPClient(npan.HTTPClientOptions{
		BaseURL:        h.cfg.BaseURL,
		Token:          token,
		TokenRefresher: npan.NewTokenRefresher(nil, authOptions),
	})
}

func (h *Handlers) Health(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":       "ok",
		"running_sync": h.syncManager.IsRunning(),
	})
}

func (h *Handlers) Token(c *echo.Context) error {
	var payload authPayload
	if err := c.Bind(&payload); err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "请求体格式错误")
	}

	authOptions := h.resolveAuthOptions(c, payload)
	if authOptions.ClientID == "" || authOptions.ClientSecret == "" || authOptions.SubID <= 0 {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少认证参数: client_id/client_secret/sub_id")
	}

	token, err := npan.RequestAccessToken(c.Request().Context(), nil, npan.TokenRequestOptions{
		OAuthHost:    authOptions.OAuthHost,
		ClientID:     authOptions.ClientID,
		ClientSecret: authOptions.ClientSecret,
		SubID:        authOptions.SubID,
		SubType:      authOptions.SubType,
	})
	if err != nil {
		slog.Error("获取 token 失败", "error", err, "handler", "Token")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "认证失败，请检查凭据")
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handlers) RemoteSearch(c *echo.Context) error {
	queryWords := strings.TrimSpace(c.QueryParam("query"))
	if queryWords == "" {
		queryWords = strings.TrimSpace(c.QueryParam("q"))
	}
	if queryWords == "" {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少 query 参数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{})
	if err != nil {
		slog.Error("解析 token 失败", "error", err, "handler", "RemoteSearch")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "搜索请求失败，请稍后重试")
	}

	pageID := int64(0)
	if raw := strings.TrimSpace(c.QueryParam("page_id")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed < 0 {
			return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "page_id 必须是 >= 0 的整数")
		}
		pageID = parsed
	}

	searchInFolder, err := parseInt64Pointer(c.QueryParam("search_in_folder"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "search_in_folder 必须是整数")
	}

	api := h.newAPIClient(token, authOptions)
	result, err := api.SearchItems(c.Request().Context(), models.RemoteSearchParams{
		QueryWords:       queryWords,
		Type:             firstNotEmpty(c.QueryParam("type"), "all"),
		PageID:           pageID,
		QueryFilter:      firstNotEmpty(c.QueryParam("query_filter"), "all"),
		SearchInFolder:   searchInFolder,
		UpdatedTimeRange: strings.TrimSpace(c.QueryParam("updated_time_range")),
	})
	if err != nil {
		slog.Error("远程搜索失败", "error", err, "handler", "RemoteSearch")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "搜索请求失败，请稍后重试")
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) DownloadURL(c *echo.Context) error {
	fileIDRaw := strings.TrimSpace(c.QueryParam("file_id"))
	if fileIDRaw == "" {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少 file_id 参数")
	}

	fileID, err := strconv.ParseInt(fileIDRaw, 10, 64)
	if err != nil || fileID <= 0 {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "file_id 必须是正整数")
	}

	validPeriod, err := parseInt64Pointer(c.QueryParam("valid_period"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "valid_period 必须是整数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{})
	if err != nil {
		slog.Error("解析 token 失败", "error", err, "handler", "DownloadURL")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "获取下载链接失败")
	}

	api := h.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(c.Request().Context(), fileID, validPeriod)
	if err != nil {
		slog.Error("获取下载链接失败", "error", err, "handler", "DownloadURL")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "获取下载链接失败")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"file_id":      fileID,
		"download_url": downloadURL,
	})
}

func (h *Handlers) LocalSearch(c *echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		query = strings.TrimSpace(c.QueryParam("q"))
	}
	if query == "" {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少 query 参数")
	}

	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "page 必须是正整数")
		}
		page = parsed
	}

	pageSize := int64(20)
	if raw := strings.TrimSpace(c.QueryParam("page_size")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "page_size 必须是正整数")
		}
		pageSize = parsed
	}

	parentID, err := parseInt64Pointer(c.QueryParam("parent_id"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "parent_id 必须是整数")
	}

	updatedAfter, err := parseInt64Pointer(c.QueryParam("updated_after"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "updated_after 必须是整数")
	}

	updatedBefore, err := parseInt64Pointer(c.QueryParam("updated_before"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "updated_before 必须是整数")
	}

	result, err := h.queryService.Query(models.LocalSearchParams{
		Query:          query,
		Type:           firstNotEmpty(c.QueryParam("type"), "all"),
		Page:           page,
		PageSize:       pageSize,
		ParentID:       parentID,
		UpdatedAfter:   updatedAfter,
		UpdatedBefore:  updatedBefore,
		IncludeDeleted: parseBool(c.QueryParam("include_deleted"), false),
	})
	if err != nil {
		slog.Error("本地搜索失败", "error", err, "handler", "LocalSearch")
		return writeErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "搜索服务暂不可用")
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) AppSearch(c *echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		query = strings.TrimSpace(c.QueryParam("q"))
	}
	if query == "" {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少 query 参数")
	}

	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "page 必须是正整数")
		}
		page = parsed
	}

	pageSize := int64(30)
	if raw := strings.TrimSpace(c.QueryParam("page_size")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "page_size 必须是正整数")
		}
		pageSize = parsed
	}

	result, err := h.queryService.Query(models.LocalSearchParams{
		Query:          query,
		Type:           string(models.ItemTypeFile),
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: false,
	})
	if err != nil {
		slog.Error("应用搜索失败", "error", err, "handler", "AppSearch")
		return writeErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "搜索服务暂不可用")
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) AppDownloadURL(c *echo.Context) error {
	fileIDRaw := strings.TrimSpace(c.QueryParam("file_id"))
	if fileIDRaw == "" {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少 file_id 参数")
	}

	fileID, err := strconv.ParseInt(fileIDRaw, 10, 64)
	if err != nil || fileID <= 0 {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "file_id 必须是正整数")
	}

	validPeriod, err := parseInt64Pointer(c.QueryParam("valid_period"))
	if err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "valid_period 必须是整数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{})
	if err != nil {
		return writeErrorResponse(c, http.StatusServiceUnavailable, ErrCodeInternalError, "下载服务暂不可用，请联系管理员检查服务端凭据")
	}

	api := h.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(c.Request().Context(), fileID, validPeriod)
	if err != nil {
		return writeErrorResponse(c, http.StatusBadGateway, ErrCodeInternalError, "生成下载链接失败，请稍后重试")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"file_id":      fileID,
		"download_url": downloadURL,
	})
}

type syncStartPayload struct {
	authPayload
	Mode               string  `json:"mode"`
	RootFolderIDs      []int64 `json:"root_folder_ids"`
	IncludeDepartments *bool   `json:"include_departments"`
	DepartmentIDs      []int64 `json:"department_ids"`
	ResumeProgress     *bool   `json:"resume_progress"`
	RootWorkers        int     `json:"root_workers"`
	ProgressEvery      int     `json:"progress_every"`
	CheckpointTemplate string  `json:"checkpoint_template"`
	WindowOverlapMS    int64   `json:"window_overlap_ms"`
	IncrementalQuery   string  `json:"incremental_query"`
}

func (h *Handlers) StartFullSync(c *echo.Context) error {
	var payload syncStartPayload
	if err := c.Bind(&payload); err != nil {
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "请求体格式错误")
	}

	token, authOptions, err := h.resolveToken(c, payload.authPayload)
	if err != nil {
		slog.Error("解析 token 失败", "error", err, "handler", "StartFullSync")
		return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "启动同步失败")
	}

	api := h.newAPIClient(token, authOptions)
	err = h.syncManager.Start(api, service.SyncStartRequest{
		Mode:               models.SyncMode(payload.Mode),
		RootFolderIDs:      payload.RootFolderIDs,
		IncludeDepartments: payload.IncludeDepartments,
		DepartmentIDs:      payload.DepartmentIDs,
		ResumeProgress:     payload.ResumeProgress,
		RootWorkers:        payload.RootWorkers,
		ProgressEvery:      payload.ProgressEvery,
		CheckpointTemplate: payload.CheckpointTemplate,
		WindowOverlapMS:    payload.WindowOverlapMS,
		IncrementalQuery:   payload.IncrementalQuery,
	})
	if err != nil {
		slog.Error("启动同步失败", "error", err, "handler", "StartFullSync")
		return writeErrorResponse(c, http.StatusConflict, ErrCodeConflict, "启动同步失败")
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"message": "同步任务已启动",
	})
}

func (h *Handlers) GetFullSyncProgress(c *echo.Context) error {
	progress, err := h.syncManager.GetProgress()
	if err != nil {
		slog.Error("获取同步进度失败", "error", err, "handler", "GetFullSyncProgress")
		return writeErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "无法读取同步进度")
	}
	if progress == nil {
		return writeErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "未找到同步进度")
	}

	return c.JSON(http.StatusOK, progress)
}

func (h *Handlers) CancelFullSync(c *echo.Context) error {
	if !h.syncManager.Cancel() {
		return writeErrorResponse(c, http.StatusConflict, ErrCodeConflict, "当前没有运行中的同步任务")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "同步取消信号已发送",
	})
}

// Readyz 就绪检查端点，检测 Meilisearch 连通性。
func (h *Handlers) Readyz(c *echo.Context) error {
	if err := h.queryService.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"meili":  "unreachable",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ready",
	})
}
