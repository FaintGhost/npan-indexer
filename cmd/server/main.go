package main

import (
	"log"

	"npan/internal/config"
	"npan/internal/httpx"
	"npan/internal/search"
	"npan/internal/service"
	"npan/internal/storage"
)

func main() {
	cfg := config.Load()

	meiliIndex := search.NewMeiliIndex(cfg.MeiliHost, cfg.MeiliAPIKey, cfg.MeiliIndex)
	if err := meiliIndex.EnsureSettings(); err != nil {
		log.Fatalf("初始化 Meili 设置失败: %v", err)
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

	log.Printf("Echo 服务启动: %s", cfg.ServerAddr)
	if err := server.Start(cfg.ServerAddr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
