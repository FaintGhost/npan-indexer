package httpx

import (
  "crypto/subtle"
  "net/http"
  "strings"

  "github.com/labstack/echo/v5"
)

// APIKeyAuth 验证 X-API-Key header 或 Authorization: Bearer token。
// 使用 constant-time 比较防止计时攻击。
func APIKeyAuth(adminKey string) echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      provided := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
      if provided == "" {
        provided = parseBearerHeaderValue(c.Request().Header.Get("Authorization"))
      }

      if subtle.ConstantTimeCompare([]byte(provided), []byte(adminKey)) != 1 {
        return writeErrorResponse(c, http.StatusUnauthorized, ErrCodeUnauthorized,
          "未授权：缺少或无效的 API Key")
      }
      return next(c)
    }
  }
}

// parseBearerHeaderValue 从 Authorization header 中提取 Bearer token。
func parseBearerHeaderValue(header string) string {
  value := strings.TrimSpace(header)
  if len(value) < 7 {
    return ""
  }
  if !strings.EqualFold(value[:7], "bearer ") {
    return ""
  }
  return strings.TrimSpace(value[7:])
}

// EmbeddedAuth 为内嵌前端请求自动注入服务端凭据标记。
// 应用于 /api/v1/app/* 路由组。
func EmbeddedAuth() echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      c.Set("auth_mode", "embedded")
      c.Set("allow_config_fallback", true)
      return next(c)
    }
  }
}

// ConfigFallbackAuth 允许使用服务端配置的凭据访问上游 API。
// 用于已通过 APIKeyAuth 认证的管理路由。
func ConfigFallbackAuth() echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      c.Set("allow_config_fallback", true)
      return next(c)
    }
  }
}
