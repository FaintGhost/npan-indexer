package httpx

import (
	"crypto/subtle"
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

type Handlers struct {
	cfg          config.Config
	queryService *search.QueryService
	syncManager  *service.SyncManager
}

func NewHandlers(cfg config.Config, queryService *search.QueryService, syncManager *service.SyncManager) *Handlers {
	return &Handlers{
		cfg:          cfg,
		queryService: queryService,
		syncManager:  syncManager,
	}
}

func writeError(c *echo.Context, status int, message string) error {
	return c.JSON(status, map[string]any{
		"error": message,
	})
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

func parseBearerHeader(header string) string {
	value := strings.TrimSpace(header)
	if len(value) < 7 {
		return ""
	}
	if strings.ToLower(value[:7]) != "bearer " {
		return ""
	}
	return strings.TrimSpace(value[7:])
}

func (h *Handlers) requireAPIAccess(c *echo.Context) bool {
	expected := strings.TrimSpace(h.cfg.AdminAPIKey)
	if expected == "" {
		return true
	}

	provided := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
	if provided == "" {
		provided = parseBearerHeader(c.Request().Header.Get("Authorization"))
	}

	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		_ = writeError(c, http.StatusUnauthorized, "未授权")
		return false
	}

	return true
}

type authPayload struct {
	Token        string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	SubID        int64  `json:"sub_id"`
	SubType      string `json:"sub_type"`
	OAuthHost    string `json:"oauth_host"`
}

func (h *Handlers) resolveAuthOptions(c *echo.Context, payload authPayload, allowConfigFallback bool) npan.AuthResolverOptions {
	tokenFromHeader := parseBearerHeader(c.Request().Header.Get("Authorization"))

	tokenCandidates := []string{
		payload.Token,
		tokenFromHeader,
		strings.TrimSpace(c.QueryParam("token")),
	}
	clientIDCandidates := []string{payload.ClientID}
	clientSecretCandidates := []string{payload.ClientSecret}
	subIDCandidates := []int64{payload.SubID}
	oauthHostCandidates := []string{payload.OAuthHost}

	if allowConfigFallback {
		tokenCandidates = append(tokenCandidates, h.cfg.Token)
		clientIDCandidates = append(clientIDCandidates, h.cfg.ClientID)
		clientSecretCandidates = append(clientSecretCandidates, h.cfg.ClientSecret)
		subIDCandidates = append(subIDCandidates, h.cfg.SubID)
		oauthHostCandidates = append(oauthHostCandidates, h.cfg.OAuthHost)
	}

	subType := npan.TokenSubjectType(payload.SubType)
	if subType == "" && allowConfigFallback {
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

func (h *Handlers) resolveToken(c *echo.Context, payload authPayload, allowConfigFallback bool) (string, npan.AuthResolverOptions, error) {
	authOptions := h.resolveAuthOptions(c, payload, allowConfigFallback)
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
	if !h.requireAPIAccess(c) {
		return nil
	}

	var payload authPayload
	if err := c.Bind(&payload); err != nil {
		return writeError(c, http.StatusBadRequest, "请求体格式错误")
	}

	authOptions := h.resolveAuthOptions(c, payload, h.cfg.AllowConfigAuthFallback)
	if authOptions.ClientID == "" || authOptions.ClientSecret == "" || authOptions.SubID <= 0 {
		return writeError(c, http.StatusBadRequest, "缺少认证参数: client_id/client_secret/sub_id")
	}

	token, err := npan.RequestAccessToken(c.Request().Context(), nil, npan.TokenRequestOptions{
		OAuthHost:    authOptions.OAuthHost,
		ClientID:     authOptions.ClientID,
		ClientSecret: authOptions.ClientSecret,
		SubID:        authOptions.SubID,
		SubType:      authOptions.SubType,
	})
	if err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, token)
}

func (h *Handlers) RemoteSearch(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	queryWords := strings.TrimSpace(c.QueryParam("query"))
	if queryWords == "" {
		queryWords = strings.TrimSpace(c.QueryParam("q"))
	}
	if queryWords == "" {
		return writeError(c, http.StatusBadRequest, "缺少 query 参数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{}, h.cfg.AllowConfigAuthFallback)
	if err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	pageID := int64(0)
	if raw := strings.TrimSpace(c.QueryParam("page_id")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed < 0 {
			return writeError(c, http.StatusBadRequest, "page_id 必须是 >= 0 的整数")
		}
		pageID = parsed
	}

	searchInFolder, err := parseInt64Pointer(c.QueryParam("search_in_folder"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "search_in_folder 必须是整数")
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
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) DownloadURL(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	fileIDRaw := strings.TrimSpace(c.QueryParam("file_id"))
	if fileIDRaw == "" {
		return writeError(c, http.StatusBadRequest, "缺少 file_id 参数")
	}

	fileID, err := strconv.ParseInt(fileIDRaw, 10, 64)
	if err != nil || fileID <= 0 {
		return writeError(c, http.StatusBadRequest, "file_id 必须是正整数")
	}

	validPeriod, err := parseInt64Pointer(c.QueryParam("valid_period"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "valid_period 必须是整数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{}, h.cfg.AllowConfigAuthFallback)
	if err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	api := h.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(c.Request().Context(), fileID, validPeriod)
	if err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"file_id":      fileID,
		"download_url": downloadURL,
	})
}

func (h *Handlers) LocalSearch(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		query = strings.TrimSpace(c.QueryParam("q"))
	}
	if query == "" {
		return writeError(c, http.StatusBadRequest, "缺少 query 参数")
	}

	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeError(c, http.StatusBadRequest, "page 必须是正整数")
		}
		page = parsed
	}

	pageSize := int64(20)
	if raw := strings.TrimSpace(c.QueryParam("page_size")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeError(c, http.StatusBadRequest, "page_size 必须是正整数")
		}
		pageSize = parsed
	}

	parentID, err := parseInt64Pointer(c.QueryParam("parent_id"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "parent_id 必须是整数")
	}

	updatedAfter, err := parseInt64Pointer(c.QueryParam("updated_after"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "updated_after 必须是整数")
	}

	updatedBefore, err := parseInt64Pointer(c.QueryParam("updated_before"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "updated_before 必须是整数")
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
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) DemoSearch(c *echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		query = strings.TrimSpace(c.QueryParam("q"))
	}
	if query == "" {
		return writeError(c, http.StatusBadRequest, "缺少 query 参数")
	}

	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeError(c, http.StatusBadRequest, "page 必须是正整数")
		}
		page = parsed
	}

	pageSize := int64(30)
	if raw := strings.TrimSpace(c.QueryParam("page_size")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return writeError(c, http.StatusBadRequest, "page_size 必须是正整数")
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
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) DemoDownloadURL(c *echo.Context) error {
	fileIDRaw := strings.TrimSpace(c.QueryParam("file_id"))
	if fileIDRaw == "" {
		return writeError(c, http.StatusBadRequest, "缺少 file_id 参数")
	}

	fileID, err := strconv.ParseInt(fileIDRaw, 10, 64)
	if err != nil || fileID <= 0 {
		return writeError(c, http.StatusBadRequest, "file_id 必须是正整数")
	}

	validPeriod, err := parseInt64Pointer(c.QueryParam("valid_period"))
	if err != nil {
		return writeError(c, http.StatusBadRequest, "valid_period 必须是整数")
	}

	token, authOptions, err := h.resolveToken(c, authPayload{}, true)
	if err != nil {
		return writeError(c, http.StatusServiceUnavailable, "下载服务暂不可用，请联系管理员检查服务端凭据")
	}

	api := h.newAPIClient(token, authOptions)
	downloadService := service.NewDownloadURLService(api)
	downloadURL, err := downloadService.GetDownloadURL(c.Request().Context(), fileID, validPeriod)
	if err != nil {
		return writeError(c, http.StatusBadGateway, "生成下载链接失败，请稍后重试")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"file_id":      fileID,
		"download_url": downloadURL,
	})
}

type syncStartPayload struct {
	authPayload
	RootFolderIDs      []int64 `json:"root_folder_ids"`
	IncludeDepartments *bool   `json:"include_departments"`
	DepartmentIDs      []int64 `json:"department_ids"`
	ResumeProgress     *bool   `json:"resume_progress"`
	RootWorkers        int     `json:"root_workers"`
	ProgressEvery      int     `json:"progress_every"`
	CheckpointTemplate string  `json:"checkpoint_template"`
}

func (h *Handlers) StartFullSync(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	var payload syncStartPayload
	if err := c.Bind(&payload); err != nil {
		return writeError(c, http.StatusBadRequest, "请求体格式错误")
	}

	token, authOptions, err := h.resolveToken(c, payload.authPayload, h.cfg.AllowConfigAuthFallback)
	if err != nil {
		return writeError(c, http.StatusBadRequest, err.Error())
	}

	api := h.newAPIClient(token, authOptions)
	err = h.syncManager.Start(api, service.SyncStartRequest{
		RootFolderIDs:      payload.RootFolderIDs,
		IncludeDepartments: payload.IncludeDepartments,
		DepartmentIDs:      payload.DepartmentIDs,
		ResumeProgress:     payload.ResumeProgress,
		RootWorkers:        payload.RootWorkers,
		ProgressEvery:      payload.ProgressEvery,
		CheckpointTemplate: payload.CheckpointTemplate,
	})
	if err != nil {
		return writeError(c, http.StatusConflict, err.Error())
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"message": "全量同步任务已启动",
	})
}

func (h *Handlers) GetFullSyncProgress(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	progress, err := h.syncManager.GetProgress()
	if err != nil {
		return writeError(c, http.StatusInternalServerError, err.Error())
	}
	if progress == nil {
		return writeError(c, http.StatusNotFound, "未找到同步进度")
	}

	return c.JSON(http.StatusOK, progress)
}

func (h *Handlers) CancelFullSync(c *echo.Context) error {
	if !h.requireAPIAccess(c) {
		return nil
	}

	if !h.syncManager.Cancel() {
		return writeError(c, http.StatusConflict, "当前没有运行中的同步任务")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "同步取消信号已发送",
	})
}
