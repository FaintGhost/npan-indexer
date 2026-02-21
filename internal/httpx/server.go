package httpx

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
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

// statusCapture wraps http.ResponseWriter to capture the response status code.
type statusCapture struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusCapture) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapture) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

// prometheusMiddleware registers HTTP request metrics with the given prometheus.Registerer
// and returns an Echo v5 middleware that records per-route request count and duration.
// Routes /healthz and /readyz are excluded from metrics.
func prometheusMiddleware(reg prometheus.Registerer) echo.MiddlewareFunc {
	labelNames := []string{"code", "method", "url"}

	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "npan",
		Name:      "requests_total",
		Help:      "How many HTTP requests processed, partitioned by status code, method, and route.",
	}, labelNames)
	reg.MustRegister(requestCount)

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "npan",
		Name:      "request_duration_seconds",
		Help:      "The HTTP request latencies in seconds.",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	}, labelNames)
	reg.MustRegister(requestDuration)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			p := c.Path()
			if p == "/healthz" || p == "/readyz" {
				return next(c)
			}

			capture := &statusCapture{ResponseWriter: c.Response(), status: http.StatusOK}
			c.SetResponse(capture)

			start := time.Now()
			err := next(c)
			elapsed := time.Since(start).Seconds()

			routePath := c.Path()
			if routePath == "/*" || routePath == "" {
				routePath = "/spa"
			}

			requestCount.WithLabelValues(strconv.Itoa(capture.status), c.Request().Method, routePath).Inc()
			requestDuration.WithLabelValues(strconv.Itoa(capture.status), c.Request().Method, routePath).Observe(elapsed)

			return err
		}
	}
}

func NewServer(handlers *Handlers, adminAPIKey string, distFS fs.FS, promReg prometheus.Registerer) *echo.Echo {
	e := echo.New()
	e.Logger = slog.Default()

	if promReg != nil {
		e.Use(prometheusMiddleware(promReg))
	}

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

	// Admin (requires API key, uses server-configured credentials for upstream API)
	admin := e.Group("/api/v1/admin", APIKeyAuth(adminAPIKey), ConfigFallbackAuth(), RateLimitMiddleware(5, 10))
	admin.POST("/sync", handlers.StartFullSync)
	admin.GET("/sync", handlers.GetFullSyncProgress)
	admin.DELETE("/sync", handlers.CancelFullSync)
	// Legacy routes for backward compatibility
	admin.POST("/sync/full", handlers.StartFullSync)
	admin.POST("/sync/start", handlers.StartFullSync)
	admin.GET("/sync/full/progress", handlers.GetFullSyncProgress)
	admin.GET("/sync/progress", handlers.GetFullSyncProgress)
	admin.POST("/sync/full/cancel", handlers.CancelFullSync)
	admin.POST("/sync/cancel", handlers.CancelFullSync)

	return e
}
