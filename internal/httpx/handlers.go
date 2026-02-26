package httpx

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	cfg                          config.Config
	queryService                 searchService
	syncManager                  *service.SyncManager
	apiFactory                   func(token string, authOptions npan.AuthResolverOptions) npan.API
	inspectRootsMaxConcurrency   int
	inspectRootsPerFolderTimeout time.Duration
}

func NewHandlers(cfg config.Config, queryService search.Searcher, syncManager *service.SyncManager) *Handlers {
	return &Handlers{
		cfg:                          cfg,
		queryService:                 queryService,
		syncManager:                  syncManager,
		inspectRootsMaxConcurrency:   cfg.InspectRootsMaxConcurrency,
		inspectRootsPerFolderTimeout: cfg.InspectRootsPerFolderTimeout,
	}
}

type authPayload struct {
	Token        string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	SubID        int64  `json:"sub_id"`
	SubType      string `json:"sub_type"`
	OAuthHost    string `json:"oauth_host"`
}

func normalizePositiveIDs(ids []int64) ([]int64, error) {
	normalized := make([]int64, 0, len(ids))
	seen := map[int64]struct{}{}
	for _, id := range ids {
		if id <= 0 {
			return nil, strconv.ErrSyntax
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, nil
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
	if h.apiFactory != nil {
		return h.apiFactory(token, authOptions)
	}
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
