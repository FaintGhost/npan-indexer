package httpx

import (
	"io/fs"
	"log/slog"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// spaHandler serves the Vite build output with SPA fallback.
// - Requests for assets/* are served with immutable cache headers.
// - Any unknown path falls back to index.html (client-side routing).
func spaHandler(distFS fs.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		p := c.Param("*")
		if p == "" {
			p = "index.html"
		}
		p = path.Clean(p)

		// Try to open the file from embedded FS.
		f, err := distFS.Open(p)
		if err == nil {
			f.Close()
			// Set cache headers based on path.
			if strings.HasPrefix(p, "assets/") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			return c.FileFS(p, distFS)
		}

		// SPA fallback: serve index.html with no-cache.
		c.Response().Header().Set("Cache-Control", "no-cache")
		return c.FileFS("index.html", distFS)
	}
}

func NewServer(handlers *Handlers, adminAPIKey string, distFS fs.FS) *echo.Echo {
	e := echo.New()
	e.Logger = slog.Default()
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(RateLimitMiddleware(20, 40))

	// Public endpoints (no auth)
	e.GET("/healthz", handlers.Health)
	e.GET("/readyz", handlers.Readyz)

	// SPA frontend served from embedded Vite build output.
	// Specific routes (/api, /healthz, /readyz) take priority over the catch-all.
	spa := spaHandler(distFS)
	e.GET("/*", spa)

	// App API (embedded auth — config fallback always enabled)
	appAPI := e.Group("/api/v1/app", EmbeddedAuth())
	appAPI.GET("/search", handlers.AppSearch)
	appAPI.GET("/download-url", handlers.AppDownloadURL)

	// API (requires API key)
	api := e.Group("/api/v1", APIKeyAuth(adminAPIKey))
	api.POST("/token", handlers.Token)
	api.GET("/search/remote", handlers.RemoteSearch)
	api.GET("/search/local", handlers.LocalSearch)
	api.GET("/download-url", handlers.DownloadURL)

	// Admin (requires API key)
	admin := e.Group("/api/v1/admin", APIKeyAuth(adminAPIKey), RateLimitMiddleware(5, 10))
	admin.POST("/sync/full", handlers.StartFullSync)
	admin.GET("/sync/full/progress", handlers.GetFullSyncProgress)
	admin.POST("/sync/full/cancel", handlers.CancelFullSync)

	return e
}
