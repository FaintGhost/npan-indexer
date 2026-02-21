package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"npan/internal/config"
	"npan/internal/httpx"
	"npan/internal/logx"
	"npan/internal/metrics"
	"npan/internal/search"
	"npan/internal/service"
	"npan/internal/storage"
	"npan/web"

	"github.com/labstack/echo/v5"
)

func main() {
	logger := logx.NewLogger()
	slog.SetDefault(logger)

	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		slog.Error("配置验证失败", "error", err)
		os.Exit(1)
	}

	// Metrics infrastructure
	promReg := metrics.NewRegistry()
	syncMetrics := metrics.NewSyncMetrics(promReg)
	searchMetrics := metrics.NewSearchMetrics(promReg)

	meiliIndex := search.NewMeiliIndex(cfg.MeiliHost, cfg.MeiliAPIKey, cfg.MeiliIndex)
	if err := meiliIndex.EnsureSettings(context.Background()); err != nil {
		logger.Error("初始化 Meili 设置失败", "error", err)
		os.Exit(1)
	}

	instrMeili := metrics.NewInstrumentedMeiliIndex(meiliIndex, searchMetrics)
	queryService := search.NewQueryService(instrMeili)
	tracker := search.NewSearchActivityTracker(5)
	cachedService := search.NewCachedQueryService(queryService, 256, 30*time.Second, tracker)
	instrSearch := metrics.NewInstrumentedSearchService(cachedService, cachedService, searchMetrics)

	progressStore := storage.NewJSONProgressStore(cfg.ProgressFile)
	syncReporter := metrics.NewPrometheusSyncReporter(syncMetrics)
	syncManager := service.NewSyncManager(service.SyncManagerArgs{
		Index:              meiliIndex,
		ProgressStore:      progressStore,
		MeiliHost:          cfg.MeiliHost,
		MeiliIndex:         cfg.MeiliIndex,
		CheckpointTemplate: cfg.CheckpointTemplate,
		RootWorkers:        cfg.SyncRootWorkers,
		ProgressEvery:      cfg.SyncProgressEvery,
		Retry:              cfg.Retry,
		MaxConcurrent:      cfg.SyncMaxConcurrent,
		MinTimeMS:          cfg.SyncMinTimeMS,
		ActivityChecker:    tracker,
		SyncStateFile:      cfg.SyncStateFile,
		IncrementalQuery:   cfg.IncrementalQuery,
		WindowOverlapMS:    cfg.SyncWindowOverlapMS,
		MetricsReporter:    syncReporter,
	})

	handlers := httpx.NewHandlers(cfg, instrSearch, syncManager)
	distFS := echo.MustSubFS(web.DistFS, "dist")
	e := httpx.NewServer(handlers, cfg.AdminAPIKey, distFS, promReg)

	httpServer := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           e,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		ReadTimeout:       cfg.ServerReadTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
		IdleTimeout:       cfg.ServerIdleTimeout,
	}

	// Metrics server (independent port)
	var metricsServer *http.Server
	if cfg.MetricsAddr != "" {
		metricsServer = metrics.NewMetricsServer(cfg.MetricsAddr, promReg)
		go func() {
			logger.Info("指标服务启动", "addr", cfg.MetricsAddr)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("指标服务启动失败", "error", err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("服务启动", "addr", cfg.ServerAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("收到停机信号，开始优雅关闭...")

	// Graceful shutdown: main server first (15s), then metrics server (5s)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("主服务优雅关闭失败", "error", err)
	}

	if metricsServer != nil {
		metricsShutdownCtx, metricsCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer metricsCancel()
		if err := metricsServer.Shutdown(metricsShutdownCtx); err != nil {
			slog.Error("指标服务优雅关闭失败", "error", err)
		}
	}
}
