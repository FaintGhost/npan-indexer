package httpx

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func resolveDemoHTMLPath() string {
	candidates := []string{
		filepath.Join("web", "demo", "index.html"),
		filepath.Join("..", "web", "demo", "index.html"),
		filepath.Join("..", "..", "web", "demo", "index.html"),
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return candidates[0]
	}
	candidates = append([]string{
		filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "web", "demo", "index.html")),
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

func NewServer(handlers *Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	e.GET("/healthz", handlers.Health)
	demoHTMLPath := resolveDemoHTMLPath()
	demoFSPath := strings.TrimPrefix(filepath.ToSlash(demoHTMLPath), "/")
	e.GET("/demo", func(c *echo.Context) error {
		return c.FileFS(demoFSPath, os.DirFS("/"))
	})
	e.GET("/demo/", func(c *echo.Context) error {
		return c.FileFS(demoFSPath, os.DirFS("/"))
	})

	api := e.Group("/api/v1")
	api.POST("/token", handlers.Token)
	api.GET("/npan/search", handlers.RemoteSearch)
	api.GET("/download-url", handlers.DownloadURL)
	api.GET("/search/local", handlers.LocalSearch)
	api.POST("/sync/full/start", handlers.StartFullSync)
	api.GET("/sync/full/progress", handlers.GetFullSyncProgress)
	api.POST("/sync/full/cancel", handlers.CancelFullSync)

	return e
}
