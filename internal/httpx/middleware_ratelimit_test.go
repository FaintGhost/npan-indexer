package httpx

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"
  "time"

  "github.com/labstack/echo/v5"
)

// newRateLimitedEcho 构建挂载了速率限制中间件的 Echo 实例，并注册一个简单的 GET /test 路由。
func newRateLimitedEcho(rps float64, burst int) *echo.Echo {
  e := echo.New()
  e.Use(RateLimitMiddleware(rps, burst))
  e.GET("/test", func(c *echo.Context) error {
    return c.String(http.StatusOK, "ok")
  })
  return e
}

// doRequest 向 /test 发送一个带有指定 RemoteAddr 的 GET 请求，并返回响应。
func doRequest(e *echo.Echo, remoteAddr string) *httptest.ResponseRecorder {
  req := httptest.NewRequest(http.MethodGet, "/test", nil)
  req.RemoteAddr = remoteAddr
  rec := httptest.NewRecorder()
  e.ServeHTTP(rec, req)
  return rec
}

// --- TestRateLimit_ExceedsLimit_Returns429 ---

func TestRateLimit_ExceedsLimit_Returns429(t *testing.T) {
  // rps=2, burst=2：允许最多 burst 个令牌，超出后应返回 429
  e := newRateLimitedEcho(2, 2)
  ip := "192.168.1.1:12345"

  var lastCode int
  got429 := false
  for i := 0; i < 10; i++ {
    rec := doRequest(e, ip)
    lastCode = rec.Code
    if rec.Code == http.StatusTooManyRequests {
      got429 = true

      // 验证响应体包含 RATE_LIMITED code
      var body map[string]any
      if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
        t.Fatalf("429 response body is not valid JSON: %v", err)
      }
      if code, ok := body["code"]; !ok || code != "RATE_LIMITED" {
        t.Errorf(`expected "code":"RATE_LIMITED" in 429 body, got: %v`, body)
      }
      break
    }
  }

  if !got429 {
    t.Fatalf("expected a 429 response after exceeding burst, last status was %d", lastCode)
  }
}

// --- TestRateLimit_DifferentIPs_Independent ---

func TestRateLimit_DifferentIPs_Independent(t *testing.T) {
  // burst=3，每个 IP 独立计数，各自发 3 次都应成功
  e := newRateLimitedEcho(10, 3)

  ips := []string{
    "10.0.0.1:1111",
    "10.0.0.2:2222",
    "10.0.0.3:3333",
  }

  for _, ip := range ips {
    for i := 0; i < 3; i++ {
      rec := doRequest(e, ip)
      if rec.Code != http.StatusOK {
        t.Errorf("IP %s request %d: expected 200 got %d", ip, i+1, rec.Code)
      }
    }
  }
}

// --- TestRateLimit_RecoveryAfterWindow ---

func TestRateLimit_RecoveryAfterWindow(t *testing.T) {
  // rps=5, burst=2：先打满，等令牌恢复后再请求应通过
  e := newRateLimitedEcho(5, 2)
  ip := "172.16.0.1:9999"

  // 耗尽 burst
  for i := 0; i < 2; i++ {
    doRequest(e, ip)
  }

  // 此时再发应 429
  rec := doRequest(e, ip)
  if rec.Code != http.StatusTooManyRequests {
    t.Fatalf("expected 429 after burst exhausted, got %d", rec.Code)
  }

  // 等待令牌桶恢复（1 个令牌 = 1/rps = 200ms，等 300ms 确保至少 1 个令牌）
  time.Sleep(300 * time.Millisecond)

  rec = doRequest(e, ip)
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200 after recovery window, got %d", rec.Code)
  }
}

// --- TestRateLimit_ResponseContainsRetryAfter ---

func TestRateLimit_ResponseContainsRetryAfter(t *testing.T) {
  e := newRateLimitedEcho(1, 1)
  ip := "203.0.113.5:8080"

  // 第一次请求消耗令牌
  doRequest(e, ip)

  // 第二次请求触发限流
  rec := doRequest(e, ip)
  if rec.Code != http.StatusTooManyRequests {
    t.Fatalf("expected 429, got %d", rec.Code)
  }

  retryAfter := rec.Header().Get("Retry-After")
  if retryAfter == "" {
    t.Error("expected Retry-After header in 429 response, but it was missing")
  }
}
