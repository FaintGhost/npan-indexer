package httpx

import (
  "encoding/json"
  "errors"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"

  "github.com/labstack/echo/v5"
)

// newEchoTestContext 创建测试用 echo.Context，同时返回 ResponseRecorder 以读取响应体
func newEchoTestContext(method, target string) (*echo.Context, *httptest.ResponseRecorder) {
  e := echo.New()
  req := httptest.NewRequest(method, target, nil)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)
  return c, rec
}

func TestWriteErrorResponse_ReturnsJSON(t *testing.T) {
  c, rec := newEchoTestContext(http.MethodGet, "/test")

  err := writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "缺少参数")
  if err != nil {
    t.Fatalf("writeErrorResponse returned unexpected error: %v", err)
  }

  if rec.Code != http.StatusBadRequest {
    t.Fatalf("expected status 400, got %d", rec.Code)
  }

  contentType := rec.Header().Get("Content-Type")
  if !strings.Contains(contentType, "application/json") {
    t.Fatalf("expected Content-Type to contain application/json, got %q", contentType)
  }

  var resp ErrorResponse
  if jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, rec.Body.String())
  }
  if resp.Code != ErrCodeBadRequest {
    t.Fatalf("expected code=%q, got %q", ErrCodeBadRequest, resp.Code)
  }
  if resp.Message != "缺少参数" {
    t.Fatalf("expected message=%q, got %q", "缺少参数", resp.Message)
  }
}

func TestWriteErrorResponse_IncludesRequestID(t *testing.T) {
  c, rec := newEchoTestContext(http.MethodGet, "/test")
  c.Request().Header.Set(echo.HeaderXRequestID, "test-req-123")

  err := writeErrorResponse(c, http.StatusUnauthorized, ErrCodeUnauthorized, "未授权")
  if err != nil {
    t.Fatalf("writeErrorResponse returned unexpected error: %v", err)
  }

  var resp ErrorResponse
  if jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, rec.Body.String())
  }
  if resp.RequestID != "test-req-123" {
    t.Fatalf("expected request_id=%q, got %q", "test-req-123", resp.RequestID)
  }
}

func TestCustomHTTPErrorHandler_EchoError(t *testing.T) {
  c, rec := newEchoTestContext(http.MethodGet, "/test")

  echoErr := &echo.HTTPError{Code: http.StatusNotFound, Message: "not found"}
  customHTTPErrorHandler(c, echoErr)

  if rec.Code != http.StatusNotFound {
    t.Fatalf("expected status 404, got %d", rec.Code)
  }

  var resp ErrorResponse
  if jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, rec.Body.String())
  }
  if resp.Code != ErrCodeNotFound {
    t.Fatalf("expected code=%q, got %q", ErrCodeNotFound, resp.Code)
  }
  if resp.Message == "" {
    t.Fatal("expected non-empty message")
  }
}

func TestCustomHTTPErrorHandler_GenericError(t *testing.T) {
  c, rec := newEchoTestContext(http.MethodGet, "/test")

  genericErr := errors.New("some internal error")
  customHTTPErrorHandler(c, genericErr)

  if rec.Code != http.StatusInternalServerError {
    t.Fatalf("expected status 500, got %d", rec.Code)
  }

  var resp ErrorResponse
  if jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp); jsonErr != nil {
    t.Fatalf("response body is not valid JSON: %v, body=%q", jsonErr, rec.Body.String())
  }
  if resp.Code != ErrCodeInternalError {
    t.Fatalf("expected code=%q, got %q", ErrCodeInternalError, resp.Code)
  }
  if resp.Message != "服务器内部错误" {
    t.Fatalf("expected message=%q, got %q", "服务器内部错误", resp.Message)
  }
}

func TestErrorResponse_NoInternalInfo(t *testing.T) {
  c, rec := newEchoTestContext(http.MethodGet, "/test")

  err := writeErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "服务器内部错误")
  if err != nil {
    t.Fatalf("writeErrorResponse returned unexpected error: %v", err)
  }

  body := rec.Body.String()
  for _, forbidden := range []string{"stack", "trace", "debug"} {
    if strings.Contains(body, forbidden) {
      t.Fatalf("response body should not contain %q, body=%q", forbidden, body)
    }
  }
}
