package httpx

import (
  "errors"
  "fmt"
  "log/slog"
  "net/http"

  "github.com/labstack/echo/v5"
)

// ErrorResponse 是统一的错误响应结构。
type ErrorResponse struct {
  Code      string `json:"code"`
  Message   string `json:"message"`
  RequestID string `json:"request_id,omitempty"`
}

// 标准错误码常量。
const (
  ErrCodeUnauthorized  = "UNAUTHORIZED"
  ErrCodeBadRequest    = "BAD_REQUEST"
  ErrCodeNotFound      = "NOT_FOUND"
  ErrCodeConflict      = "CONFLICT"
  ErrCodeRateLimited   = "RATE_LIMITED"
  ErrCodeInternalError = "INTERNAL_ERROR"
)

// writeErrorResponse 向客户端返回结构化 JSON 错误响应。
// 会自动从请求头提取 X-Request-Id 并包含在响应中。
func writeErrorResponse(c *echo.Context, status int, code string, message string) error {
  requestID := c.Request().Header.Get(echo.HeaderXRequestID)
  resp := ErrorResponse{
    Code:      code,
    Message:   message,
    RequestID: requestID,
  }
  return c.JSON(status, resp)
}

// httpStatusToErrCode 将 HTTP 状态码映射到错误码常量。
func httpStatusToErrCode(status int) string {
  switch status {
  case http.StatusBadRequest:
    return ErrCodeBadRequest
  case http.StatusUnauthorized:
    return ErrCodeUnauthorized
  case http.StatusNotFound:
    return ErrCodeNotFound
  case http.StatusConflict:
    return ErrCodeConflict
  case http.StatusTooManyRequests:
    return ErrCodeRateLimited
  default:
    return ErrCodeInternalError
  }
}

// customHTTPErrorHandler 是 Echo 的全局错误处理器。
// 将 *echo.HTTPError 及普通 error 统一转换为 ErrorResponse JSON 响应。
func customHTTPErrorHandler(c *echo.Context, err error) {
  var status int
  var code string
  var message string

  var he *echo.HTTPError
  if errors.As(err, &he) {
    status = he.Code
    code = httpStatusToErrCode(status)
    message = fmt.Sprintf("%v", he.Message)
    if status >= http.StatusInternalServerError {
      slog.Error("http server error", "status", status, "err", err)
    }
  } else {
    status = http.StatusInternalServerError
    code = ErrCodeInternalError
    message = "服务器内部错误"
    slog.Error("unhandled error", "err", err)
  }

  //nolint:errcheck
  _ = writeErrorResponse(c, status, code, message)
}
