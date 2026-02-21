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

	meiliIndex := search.NewMeiliIndex(cfg.MeiliHost, cfg.MeiliAPIKey, cfg.MeiliIndex)
	if err := meiliIndex.EnsureSettings(context.Background()); err != nil {
		logger.Error("初始化 Meili 设置失败", "error", err)
		os.Exit(1)
	}

	queryService := search.NewQueryService(meiliIndex)
	tracker := search.NewSearchActivityTracker(5)
	cachedService := search.NewCachedQueryService(queryService, 256, 30*time.Second, tracker)
	progressStore := storage.NewJSONProgressStore(cfg.ProgressFile)
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
	})

	handlers := httpx.NewHandlers(cfg, cachedService, syncManager)
	distFS := echo.MustSubFS(web.DistFS, "dist")
	e := httpx.NewServer(handlers, cfg.AdminAPIKey, distFS)

	httpServer := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           e,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		ReadTimeout:       cfg.ServerReadTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
		IdleTimeout:       cfg.ServerIdleTimeout,
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("优雅关闭失败", "error", err)
	}
}
