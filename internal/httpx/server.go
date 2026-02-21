package httpx

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func resolveAppHTMLPath() string {
	candidates := []string{
		filepath.Join("web", "app", "index.html"),
		filepath.Join("..", "web", "app", "index.html"),
		filepath.Join("..", "..", "web", "app", "index.html"),
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return candidates[0]
	}
	candidates = append([]string{
		filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "web", "app", "index.html")),
	}, candidates...)

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		absCandidate, absErr := filepath.Abs(candidate)
		if absErr == nil {
			return absCandidate
		}
		return candidate
	}

	return candidates[0]
}

func NewServer(handlers *Handlers, adminAPIKey string) *echo.Echo {
	e := echo.New()
	e.Logger = slog.Default()
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	// Public endpoints (no auth)
	e.GET("/healthz", handlers.Health)
	e.GET("/readyz", handlers.Readyz)
	appHTMLPath := resolveAppHTMLPath()
	appFSPath := strings.TrimPrefix(filepath.ToSlash(appHTMLPath), "/")
	e.GET("/app", func(c *echo.Context) error {
		return c.FileFS(appFSPath, os.DirFS("/"))
	})
	e.GET("/app/", func(c *echo.Context) error {
		return c.FileFS(appFSPath, os.DirFS("/"))
	})

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
	admin := e.Group("/api/v1/admin", APIKeyAuth(adminAPIKey))
	admin.POST("/sync/full", handlers.StartFullSync)
	admin.GET("/sync/full/progress", handlers.GetFullSyncProgress)
	admin.POST("/sync/full/cancel", handlers.CancelFullSync)

	return e
}
