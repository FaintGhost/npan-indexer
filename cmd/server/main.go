package main

import (
	"context"
	"log/slog"
	"os"

	"npan/internal/config"
	"npan/internal/httpx"
	"npan/internal/logx"
	"npan/internal/search"
	"npan/internal/service"
	"npan/internal/storage"
)

func main() {
	logger := logx.NewLogger()
	slog.SetDefault(logger)

	cfg := config.Load()

	meiliIndex := search.NewMeiliIndex(cfg.MeiliHost, cfg.MeiliAPIKey, cfg.MeiliIndex)
	if err := meiliIndex.EnsureSettings(context.Background()); err != nil {
		logger.Error("初始化 Meili 设置失败", "error", err)
		os.Exit(1)
	}

	queryService := search.NewQueryService(meiliIndex)
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
	})

	handlers := httpx.NewHandlers(cfg, queryService, syncManager)
	server := httpx.NewServer(handlers)

	logger.Info("Echo 服务启动", "addr", cfg.ServerAddr)
	if err := server.Start(cfg.ServerAddr); err != nil {
		logger.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}
