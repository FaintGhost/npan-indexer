package httpx

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewServer(handlers *Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	e.GET("/healthz", handlers.Health)

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
