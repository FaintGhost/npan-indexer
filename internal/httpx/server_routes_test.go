package httpx

import (
  "net/http"
  "net/http/httptest"
  "testing"
)

const testAdminKey = "test-admin-key-1234567890"

func TestRoutes_PublicEndpoints_NoAuthRequired(t *testing.T) {
  t.Parallel()

  e := NewServer(&Handlers{}, testAdminKey, testDistFS(), nil)
  endpoints := []struct {
    method string
    path   string
  }{
    {http.MethodGet, "/healthz"},
    {http.MethodGet, "/readyz"},
    {http.MethodGet, "/api/v1/app/search?q=test"},
    {http.MethodGet, "/api/v1/app/download-url?file_id=1"},
  }

  for _, ep := range endpoints {
    t.Run(ep.method+" "+ep.path, func(t *testing.T) {
      req := httptest.NewRequest(ep.method, ep.path, nil)
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)
      if rec.Code == http.StatusUnauthorized {
        t.Errorf("%s %s returned 401, expected no auth required", ep.method, ep.path)
      }
    })
  }
}

func TestRoutes_APIEndpoints_RequireAuth(t *testing.T) {
  t.Parallel()

  e := NewServer(&Handlers{}, testAdminKey, testDistFS(), nil)
  endpoints := []struct {
    method string
    path   string
  }{
    {http.MethodPost, "/api/v1/token"},
    {http.MethodGet, "/api/v1/search/remote?q=test"},
    {http.MethodGet, "/api/v1/search/local?q=test"},
    {http.MethodGet, "/api/v1/download-url?file_id=1"},
  }

  for _, ep := range endpoints {
    t.Run(ep.method+" "+ep.path, func(t *testing.T) {
      req := httptest.NewRequest(ep.method, ep.path, nil)
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)
      if rec.Code != http.StatusUnauthorized {
        t.Errorf("%s %s returned %d, expected 401", ep.method, ep.path, rec.Code)
      }
    })
  }
}

func TestRoutes_AdminEndpoints_RequireAuth(t *testing.T) {
  t.Parallel()

  e := NewServer(&Handlers{}, testAdminKey, testDistFS(), nil)
  endpoints := []struct {
    method string
    path   string
  }{
    {http.MethodPost, "/api/v1/admin/sync/full"},
    {http.MethodGet, "/api/v1/admin/sync/full/progress"},
    {http.MethodPost, "/api/v1/admin/sync/full/cancel"},
  }

  for _, ep := range endpoints {
    t.Run(ep.method+" "+ep.path, func(t *testing.T) {
      req := httptest.NewRequest(ep.method, ep.path, nil)
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)
      if rec.Code != http.StatusUnauthorized {
        t.Errorf("%s %s returned %d, expected 401", ep.method, ep.path, rec.Code)
      }
    })
  }
}

func TestRoutes_APIEndpoints_WithKey_Pass(t *testing.T) {
  t.Parallel()

  e := NewServer(&Handlers{}, testAdminKey, testDistFS(), nil)
  endpoints := []struct {
    method string
    path   string
  }{
    {http.MethodPost, "/api/v1/token"},
    {http.MethodGet, "/api/v1/search/remote?q=test"},
    {http.MethodGet, "/api/v1/search/local?q=test"},
    {http.MethodGet, "/api/v1/download-url?file_id=1"},
    {http.MethodPost, "/api/v1/admin/sync/full"},
    {http.MethodGet, "/api/v1/admin/sync/full/progress"},
    {http.MethodPost, "/api/v1/admin/sync/full/cancel"},
  }

  for _, ep := range endpoints {
    t.Run(ep.method+" "+ep.path, func(t *testing.T) {
      req := httptest.NewRequest(ep.method, ep.path, nil)
      req.Header.Set("X-API-Key", testAdminKey)
      rec := httptest.NewRecorder()
      e.ServeHTTP(rec, req)
      if rec.Code == http.StatusUnauthorized {
        t.Errorf("%s %s with valid key returned 401", ep.method, ep.path)
      }
    })
  }
}
