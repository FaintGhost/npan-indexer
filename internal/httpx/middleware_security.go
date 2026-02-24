package httpx

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func SecureHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			return next(c)
		}
	}
}

func CORSConfig(allowedOrigins []string) middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{
			"Authorization",
			"X-API-Key",
			"Content-Type",
			// Connect-RPC (connect-web / connect-es) browser requests
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			// Keep grpc-web migration path open.
			"Grpc-Timeout",
		},
		ExposeHeaders: []string{
			"Connect-Error-Reason",
			"Connect-Error-Details",
		},
		MaxAge: 3600,
	}
}

func ParseCORSOrigins(envValue string) []string {
	if envValue == "" {
		return []string{}
	}
	parts := strings.Split(envValue, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
