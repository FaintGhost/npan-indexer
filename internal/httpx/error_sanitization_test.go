package httpx

import (
  "encoding/json"
  "errors"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"

  "github.com/labstack/echo/v5"

  "npan/internal/config"
  "npan/internal/models"
  "npan/internal/search"
  "npan/internal/service"
)

// errorSearchService 是一个总是返回错误的 searchService 实现。
type errorSearchService struct {
  queryErr error
  pingErr  error
}

func (m *errorSearchService) Query(_ models.LocalSearchParams) (search.QueryResult, error) {
  return search.QueryResult{}, m.queryErr
}
func (m *errorSearchService) Ping() error { return m.pingErr }

func TestErrorSanitization_MeiliError_NoDetails(t *testing.T) {
  t.Parallel()

  mock := &errorSearchService{
    queryErr: errors.New("meilisearch: connection refused to http://meili:7700/indexes/npan_items/search"),
  }
  h := &Handlers{
    cfg:          config.Config{},
    queryService: mock,
    syncManager:  &service.SyncManager{},
  }

  e := echo.New()
  req := httptest.NewRequest(http.MethodGet, "/api/v1/search/local?q=test", nil)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)

  _ = h.LocalSearch(c)

  body := rec.Body.String()
  bodyLower := strings.ToLower(body)

  if strings.Contains(bodyLower, "meilisearch") {
    t.Errorf("response leaks meilisearch details: %s", body)
  }
  if strings.Contains(bodyLower, "meili:7700") {
    t.Errorf("response leaks internal host: %s", body)
  }
  if strings.Contains(bodyLower, "connection refused") {
    t.Errorf("response leaks connection error: %s", body)
  }

  var resp map[string]any
  if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
    t.Fatalf("response not valid JSON: %v", err)
  }
  if _, ok := resp["code"]; !ok {
    t.Error("response missing 'code' field")
  }
  if _, ok := resp["message"]; !ok {
    t.Error("response missing 'message' field")
  }
}

func TestErrorSanitization_TokenError_NoSecret(t *testing.T) {
  t.Parallel()

  h := &Handlers{
    cfg: config.Config{
      ClientSecret:            "super-secret-value-12345",
      AllowConfigAuthFallback: true,
    },
    syncManager: &service.SyncManager{},
  }

  e := echo.New()
  // Token endpoint requires client credentials — will fail to obtain token
  req := httptest.NewRequest(http.MethodPost, "/api/v1/token", strings.NewReader(`{}`))
  req.Header.Set("Content-Type", "application/json")
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)

  _ = h.Token(c)

  body := rec.Body.String()
  if strings.Contains(body, "super-secret-value-12345") {
    t.Errorf("response leaks client_secret: %s", body)
  }
}

func TestErrorSanitization_InternalError_NoStack(t *testing.T) {
  t.Parallel()

  e := echo.New()
  e.HTTPErrorHandler = customHTTPErrorHandler

  // Trigger a panic via a handler
  e.GET("/test-panic", func(c *echo.Context) error {
    panic("internal panic at handlers.go:42")
  })
  // Echo's Recover middleware catches panics — simulate a generic 500
  req := httptest.NewRequest(http.MethodGet, "/test-500", nil)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)

  // Use the custom error handler directly
  customHTTPErrorHandler(c, errors.New("runtime error: index out of range [5] with length 3"))

  body := rec.Body.String()
  bodyLower := strings.ToLower(body)

  if strings.Contains(bodyLower, "runtime error") {
    t.Errorf("response leaks runtime error: %s", body)
  }
  if strings.Contains(bodyLower, ".go:") {
    t.Errorf("response leaks file path: %s", body)
  }
  if strings.Contains(bodyLower, "goroutine") {
    t.Errorf("response leaks goroutine info: %s", body)
  }
  if strings.Contains(bodyLower, "panic") {
    t.Errorf("response leaks panic info: %s", body)
  }

  var resp map[string]any
  if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
    t.Fatalf("response not valid JSON: %v", err)
  }
  if _, ok := resp["code"]; !ok {
    t.Error("response missing 'code' field")
  }
}

func TestErrorSanitization_AllErrors_HaveUnifiedFormat(t *testing.T) {
  t.Parallel()

  mock := &errorSearchService{
    queryErr: errors.New("internal error"),
  }
  h := &Handlers{
    cfg:          config.Config{},
    queryService: mock,
    syncManager:  &service.SyncManager{},
  }

  tests := []struct {
    name   string
    method string
    path   string
  }{
    {"LocalSearch", http.MethodGet, "/api/v1/search/local?q=test"},
  }

  for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
      e := echo.New()
      req := httptest.NewRequest(tc.method, tc.path, nil)
      rec := httptest.NewRecorder()
      c := e.NewContext(req, rec)
      _ = h.LocalSearch(c)

      var resp map[string]any
      if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
        t.Fatalf("response not valid JSON: %v", err)
      }

      if _, ok := resp["code"]; !ok {
        t.Error("response missing 'code' field")
      }
      if _, ok := resp["message"]; !ok {
        t.Error("response missing 'message' field")
      }

      // Ensure no internal debug fields
      for _, key := range []string{"stack", "trace", "debug", "error"} {
        if _, ok := resp[key]; ok {
          t.Errorf("response should not contain '%s' field", key)
        }
      }
    })
  }
}
