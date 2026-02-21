package httpx

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/labstack/echo/v5"
)

// --- 辅助函数 ---

func newMiddlewareTestContext(method, target string, setupReq func(req *http.Request)) (*echo.Echo, *httptest.ResponseRecorder, *echo.Context) {
  e := echo.New()
  req := httptest.NewRequest(method, target, nil)
  if setupReq != nil {
    setupReq(req)
  }
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)
  return e, rec, c
}

func successHandler(c *echo.Context) error {
  return c.String(http.StatusOK, "ok")
}

// --- APIKeyAuth 测试 ---

func TestAPIKeyAuth_NoKey_Returns401(t *testing.T) {
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth("test-admin-key-32chars-minimum!!")
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if rec.Code != http.StatusUnauthorized {
    t.Fatalf("expected 401, got %d", rec.Code)
  }
  if nextCalled {
    t.Fatal("next handler should not be called")
  }

  // 验证响应体包含 UNAUTHORIZED code
  body := rec.Body.String()
  var resp map[string]any
  if jsonErr := json.Unmarshal([]byte(body), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, body)
  }
  if resp["code"] != "UNAUTHORIZED" {
    t.Fatalf("expected code=UNAUTHORIZED, got %v", resp["code"])
  }

  // 验证不包含内部信息
  if _, hasStack := resp["stack"]; hasStack {
    t.Fatal("response should not contain 'stack' field")
  }
  if _, hasConfig := resp["config"]; hasConfig {
    t.Fatal("response should not contain 'config' field")
  }
}

func TestAPIKeyAuth_WrongKey_Returns401(t *testing.T) {
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", func(req *http.Request) {
    req.Header.Set("X-API-Key", "wrong-key")
  })

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth("test-admin-key-32chars-minimum!!")
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if rec.Code != http.StatusUnauthorized {
    t.Fatalf("expected 401, got %d", rec.Code)
  }
  if nextCalled {
    t.Fatal("next handler should not be called")
  }
}

func TestAPIKeyAuth_ValidXAPIKey_Passes(t *testing.T) {
  const adminKey = "test-admin-key-32chars-minimum!!"

  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", func(req *http.Request) {
    req.Header.Set("X-API-Key", adminKey)
  })

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth(adminKey)
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if !nextCalled {
    t.Fatal("next handler should have been called")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
}

func TestAPIKeyAuth_ValidBearerToken_Passes(t *testing.T) {
  const adminKey = "test-admin-key-32chars-minimum!!"

  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", func(req *http.Request) {
    req.Header.Set("Authorization", "Bearer "+adminKey)
  })

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth(adminKey)
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if !nextCalled {
    t.Fatal("next handler should have been called")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
}

func TestAPIKeyAuth_EmptyBearerToken_Returns401(t *testing.T) {
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", func(req *http.Request) {
    req.Header.Set("Authorization", "Bearer ")
  })

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth("test-admin-key-32chars-minimum!!")
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if rec.Code != http.StatusUnauthorized {
    t.Fatalf("expected 401, got %d", rec.Code)
  }
  if nextCalled {
    t.Fatal("next handler should not be called")
  }
}

func TestAPIKeyAuth_ResponseFormat(t *testing.T) {
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  next := func(c *echo.Context) error {
    return c.String(http.StatusOK, "ok")
  }

  middleware := APIKeyAuth("test-admin-key-32chars-minimum!!")
  handler := middleware(next)
  err := handler(c)

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if rec.Code != http.StatusUnauthorized {
    t.Fatalf("expected 401, got %d", rec.Code)
  }

  body := rec.Body.String()
  var resp map[string]any
  if jsonErr := json.Unmarshal([]byte(body), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, body)
  }
  if _, hasCode := resp["code"]; !hasCode {
    t.Fatal("response JSON should contain 'code' field")
  }
  if _, hasMessage := resp["message"]; !hasMessage {
    t.Fatal("response JSON should contain 'message' field")
  }
}

// --- EmbeddedAuth 测试 ---

func TestEmbeddedAuth_SetsAuthMode(t *testing.T) {
  _, _, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  next := func(c *echo.Context) error {
    return c.String(http.StatusOK, "ok")
  }

  middleware := EmbeddedAuth()
  handler := middleware(next)
  if err := handler(c); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  val := c.Get("auth_mode")
  if val != "embedded" {
    t.Fatalf("expected auth_mode=embedded, got %v", val)
  }
}

func TestEmbeddedAuth_SetsConfigFallback(t *testing.T) {
  _, _, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  next := func(c *echo.Context) error {
    return c.String(http.StatusOK, "ok")
  }

  middleware := EmbeddedAuth()
  handler := middleware(next)
  if err := handler(c); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  val := c.Get("allow_config_fallback")
  if val != true {
    t.Fatalf("expected allow_config_fallback=true, got %v", val)
  }
}

func TestEmbeddedAuth_CallsNextHandler(t *testing.T) {
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := EmbeddedAuth()
  handler := middleware(next)
  if err := handler(c); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  if !nextCalled {
    t.Fatal("next handler should have been called")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
}

func TestEmbeddedAuth_NoAPIKeyRequired(t *testing.T) {
  // 请求无任何认证头，EmbeddedAuth 不做 key 校验，应正常通过
  _, rec, c := newMiddlewareTestContext(http.MethodGet, "/test", nil)

  nextCalled := false
  next := func(c *echo.Context) error {
    nextCalled = true
    return c.String(http.StatusOK, "ok")
  }

  middleware := EmbeddedAuth()
  handler := middleware(next)
  if err := handler(c); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  if !nextCalled {
    t.Fatal("next handler should have been called without any auth header")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
}
