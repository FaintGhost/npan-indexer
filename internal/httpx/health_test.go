package httpx

import (
  "encoding/json"
  "errors"
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/labstack/echo/v5"

  "npan/internal/models"
  "npan/internal/search"
  "npan/internal/service"
)

// mockSearchService 实现 searchService 接口，用于测试。
type mockSearchService struct {
  pingErr  error
  queryErr error
}

func (m *mockSearchService) Ping() error { return m.pingErr }
func (m *mockSearchService) Query(_ models.LocalSearchParams) (search.QueryResult, error) {
  return search.QueryResult{}, m.queryErr
}

func TestHealthz_AlwaysReturns200(t *testing.T) {
  e := echo.New()
  req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
  rec := httptest.NewRecorder()

  h := &Handlers{syncManager: &service.SyncManager{}}
  c := e.NewContext(req, rec)

  if err := h.Health(c); err != nil {
    t.Fatalf("Health() returned unexpected error: %v", err)
  }

  if rec.Code != http.StatusOK {
    t.Fatalf("expected status 200, got %d", rec.Code)
  }

  var body map[string]any
  if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    t.Fatalf("response body is not valid JSON: %v", err)
  }

  if status, ok := body["status"]; !ok || status != "ok" {
    t.Errorf(`expected "status":"ok" in response body, got: %v`, body)
  }
}

func TestReadyz_MeiliAvailable_Returns200(t *testing.T) {
  e := echo.New()
  req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)

  h := &Handlers{queryService: &mockSearchService{pingErr: nil}}

  if err := h.Readyz(c); err != nil {
    t.Fatalf("Readyz() returned unexpected error: %v", err)
  }

  if rec.Code != http.StatusOK {
    t.Fatalf("expected status 200, got %d", rec.Code)
  }

  var body map[string]any
  if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    t.Fatalf("response body is not valid JSON: %v", err)
  }

  if status, ok := body["status"]; !ok || status != "ready" {
    t.Errorf(`expected "status":"ready" in response body, got: %v`, body)
  }
}

func TestReadyz_MeiliUnavailable_Returns503(t *testing.T) {
  e := echo.New()
  req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)

  h := &Handlers{queryService: &mockSearchService{pingErr: errors.New("connection refused")}}

  if err := h.Readyz(c); err != nil {
    t.Fatalf("Readyz() returned unexpected error: %v", err)
  }

  if rec.Code != http.StatusServiceUnavailable {
    t.Fatalf("expected status 503, got %d", rec.Code)
  }

  var body map[string]any
  if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    t.Fatalf("response body is not valid JSON: %v", err)
  }

  if status, ok := body["status"]; !ok || status != "not_ready" {
    t.Errorf(`expected "status":"not_ready" in response body, got: %v`, body)
  }
}
