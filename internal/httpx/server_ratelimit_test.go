package httpx

import (
  "net/http"
  "net/http/httptest"
  "path/filepath"
  "sync"
  "testing"

  "npan/internal/config"
  "npan/internal/service"
  "npan/internal/storage"
)

// newTestHandlers 构造测试用的最小 Handlers 实例。
func newTestHandlers(t *testing.T) *Handlers {
  t.Helper()

  progressFile := filepath.Join(t.TempDir(), "progress.json")
  progressStore := storage.NewJSONProgressStore(progressFile)
  syncManager := service.NewSyncManager(service.SyncManagerArgs{
    ProgressStore: progressStore,
  })

  return &Handlers{
    cfg:          config.Config{AllowConfigAuthFallback: true},
    queryService: &mockSearchService{},
    syncManager:  syncManager,
  }
}

func TestNewServer_RateLimitOnSearchEndpoint(t *testing.T) {
  const adminAPIKey = "test-admin-key"
  handlers := newTestHandlers(t)
  e := NewServer(handlers, adminAPIKey)

  const totalRequests = 50
  type result struct {
    statusCode int
    retryAfter string
  }
  results := make([]result, totalRequests)

  var wg sync.WaitGroup
  wg.Add(totalRequests)

  for i := 0; i < totalRequests; i++ {
    go func(idx int) {
      defer wg.Done()

      req := httptest.NewRequest(http.MethodGet, "/api/v1/app/search?query=test", nil)
      req.RemoteAddr = "192.168.1.100:12345"
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)

      results[idx] = result{
        statusCode: rec.Code,
        retryAfter: rec.Header().Get("Retry-After"),
      }
    }(i)
  }

  wg.Wait()

  got429 := 0
  got200 := 0
  for _, r := range results {
    switch r.statusCode {
    case http.StatusTooManyRequests:
      got429++
    case http.StatusOK:
      got200++
    }
  }

  if got429 == 0 {
    t.Errorf("expected some requests to return 429, but all %d requests succeeded (got200=%d)", totalRequests, got200)
  }

  // 验证所有 429 响应都包含 Retry-After header
  for i, r := range results {
    if r.statusCode == http.StatusTooManyRequests && r.retryAfter == "" {
      t.Errorf("request %d returned 429 but missing Retry-After header", i)
    }
  }
}

func TestNewServer_RateLimitOnAdminEndpoint(t *testing.T) {
  const adminAPIKey = "test-admin-key"
  handlers := newTestHandlers(t)
  e := NewServer(handlers, adminAPIKey)

  const totalRequests = 50
  type result struct {
    statusCode int
    retryAfter string
  }
  results := make([]result, totalRequests)

  var wg sync.WaitGroup
  wg.Add(totalRequests)

  for i := 0; i < totalRequests; i++ {
    go func(idx int) {
      defer wg.Done()

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sync/full/progress", nil)
      req.Header.Set("X-API-Key", adminAPIKey)
      req.RemoteAddr = "10.0.0.1:54321"
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)

      results[idx] = result{
        statusCode: rec.Code,
        retryAfter: rec.Header().Get("Retry-After"),
      }
    }(i)
  }

  wg.Wait()

  got429 := 0
  gotAuth := 0
  gotOther := 0
  for _, r := range results {
    switch r.statusCode {
    case http.StatusTooManyRequests:
      got429++
    case http.StatusUnauthorized:
      gotAuth++
    default:
      gotOther++
    }
  }

  if got429 == 0 {
    t.Errorf("expected some admin requests to return 429, but none did (gotAuth=%d, gotOther=%d)", gotAuth, gotOther)
  }

  // 验证所有 429 响应都包含 Retry-After header
  for i, r := range results {
    if r.statusCode == http.StatusTooManyRequests && r.retryAfter == "" {
      t.Errorf("request %d returned 429 but missing Retry-After header", i)
    }
  }
}
