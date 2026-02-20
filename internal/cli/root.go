package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"npan/internal/config"
	"npan/internal/indexer"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/service"
	"npan/internal/storage"
)

type authOptions struct {
	token        string
	clientID     string
	clientSecret string
	subID        int64
	subType      string
	oauthHost    string
	baseURL      string
}

type syncProgressOutputMode string

const (
	syncProgressOutputHuman syncProgressOutputMode = "human"
	syncProgressOutputJSON  syncProgressOutputMode = "json"
)

type progressRenderSnapshot struct {
	updatedAtMillis int64
	filesIndexed    int64
	pagesFetched    int64
}

func printJSON(value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func resolveSyncProgressOutputMode(raw string) (syncProgressOutputMode, error) {
	mode := syncProgressOutputMode(strings.ToLower(strings.TrimSpace(raw)))
	switch mode {
	case "", syncProgressOutputHuman:
		return syncProgressOutputHuman, nil
	case syncProgressOutputJSON:
		return syncProgressOutputJSON, nil
	default:
		return "", fmt.Errorf("不支持的进度输出模式: %s（可选: human|json）", raw)
	}
}

func formatDurationCompact(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int64(d / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func renderSyncFullProgressHuman(progress *models.SyncProgressState, snapshot *progressRenderSnapshot) string {
	timestamp := time.Now().Format("15:04:05")
	if progress != nil && progress.UpdatedAt > 0 {
		timestamp = time.UnixMilli(progress.UpdatedAt).Format("15:04:05")
	}

	totalRoots := len(progress.Roots)
	completedRoots := len(progress.CompletedRoots)
	activeRoot := "-"
	if progress.ActiveRoot != nil {
		activeRoot = fmt.Sprintf("%d", *progress.ActiveRoot)
	}

	fileRateText := "-"
	pageRateText := "-"
	if snapshot != nil && snapshot.updatedAtMillis > 0 && progress.UpdatedAt > snapshot.updatedAtMillis {
		deltaMillis := progress.UpdatedAt - snapshot.updatedAtMillis
		if deltaMillis > 0 {
			fileDelta := progress.AggregateStats.FilesIndexed - snapshot.filesIndexed
			pageDelta := progress.AggregateStats.PagesFetched - snapshot.pagesFetched
			if fileDelta < 0 {
				fileDelta = 0
			}
			if pageDelta < 0 {
				pageDelta = 0
			}
			fileRateText = fmt.Sprintf("%d/s", fileDelta*1000/deltaMillis)
			pageRateText = fmt.Sprintf("%d/s", pageDelta*1000/deltaMillis)
		}
	}

	if snapshot != nil {
		snapshot.updatedAtMillis = progress.UpdatedAt
		snapshot.filesIndexed = progress.AggregateStats.FilesIndexed
		snapshot.pagesFetched = progress.AggregateStats.PagesFetched
	}

	elapsedMillis := progress.UpdatedAt - progress.StartedAt
	if elapsedMillis < 0 {
		elapsedMillis = 0
	}

	activeRootDetail := ""
	if progress.ActiveRoot != nil {
		if rp := progress.RootProgress[fmt.Sprintf("%d", *progress.ActiveRoot)]; rp != nil {
			currentFolder := "-"
			if rp.CurrentFolderID != nil {
				currentFolder = fmt.Sprintf("%d", *rp.CurrentFolderID)
			}
			currentPage := "-"
			if rp.CurrentPageID != nil && rp.CurrentPageCount != nil && *rp.CurrentPageCount > 0 {
				currentPage = fmt.Sprintf("%d/%d", *rp.CurrentPageID+1, *rp.CurrentPageCount)
			} else if rp.CurrentPageID != nil {
				currentPage = fmt.Sprintf("%d", *rp.CurrentPageID)
			}
			queue := "-"
			if rp.QueueLength != nil {
				queue = fmt.Sprintf("%d", *rp.QueueLength)
			}
			activeRootDetail = fmt.Sprintf(" root{folder=%s page=%s queue=%s}", currentFolder, currentPage, queue)
		}
	}

	return fmt.Sprintf(
		"[%s] status=%s roots=%d/%d active=%s files=%d pages=%d folders=%d failed=%d file_rate=%s page_rate=%s elapsed=%s%s",
		timestamp,
		progress.Status,
		completedRoots,
		totalRoots,
		activeRoot,
		progress.AggregateStats.FilesIndexed,
		progress.AggregateStats.PagesFetched,
		progress.AggregateStats.FoldersVisited,
		progress.AggregateStats.FailedRequests,
		fileRateText,
		pageRateText,
		formatDurationCompact(time.Duration(elapsedMillis)*time.Millisecond),
		activeRootDetail,
	)
}

func printSyncFullProgress(progress *models.SyncProgressState, mode syncProgressOutputMode, snapshot *progressRenderSnapshot) error {
	if mode == syncProgressOutputJSON {
		return printJSON(map[string]any{
			"status":      progress.Status,
			"active_root": progress.ActiveRoot,
			"completed":   len(progress.CompletedRoots),
			"total_roots": len(progress.Roots),
			"stats":       progress.AggregateStats,
			"updated_at":  progress.UpdatedAt,
		})
	}

	fmt.Println(renderSyncFullProgressHuman(progress, snapshot))
	return nil
}

func parseInt64CSV(raw string) ([]int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无法解析整数列表 %q: %w", raw, err)
		}
		result = append(result, parsed)
	}
	return result, nil
}

func resolveAuthOptions(cfg config.Config, options authOptions) npan.AuthResolverOptions {
	return npan.AuthResolverOptions{
		Token:        firstNotEmpty(options.token, cfg.Token),
		ClientID:     firstNotEmpty(options.clientID, cfg.ClientID),
		ClientSecret: firstNotEmpty(options.clientSecret, cfg.ClientSecret),
		SubID:        firstPositive(options.subID, cfg.SubID),
		SubType:      npan.TokenSubjectType(firstNotEmpty(options.subType, string(cfg.SubType))),
		OAuthHost:    firstNotEmpty(options.oauthHost, cfg.OAuthHost),
	}
}

func resolveToken(ctx context.Context, cfg config.Config, options authOptions) (string, npan.AuthResolverOptions, error) {
	authOptions := resolveAuthOptions(cfg, options)
	token, err := npan.ResolveBearerToken(ctx, nil, authOptions)
	if err != nil {
		return "", authOptions, err
	}
	return token, authOptions, nil
}

func newAPIClient(baseURL string, token string, authOptions npan.AuthResolverOptions) npan.API {
	return npan.NewHTTPClient(npan.HTTPClientOptions{
		BaseURL:        baseURL,
		Token:          token,
		TokenRefresher: npan.NewTokenRefresher(nil, authOptions),
	})
}

func firstNotEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstPositive(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func normalizeUnixSeconds(value int64) int64 {
	if value <= 0 {
		return 0
	}
	if value >= 1_000_000_000_000 {
		return value / 1000
	}
	return value
}

func addAuthFlags(command *cobra.Command, options *authOptions, cfg config.Config) {
	command.Flags().StringVar(&options.token, "token", cfg.Token, "用户 Bearer token")
	command.Flags().StringVar(&options.clientID, "client-id", cfg.ClientID, "开放平台 client_id")
	command.Flags().StringVar(&options.clientSecret, "client-secret", cfg.ClientSecret, "开放平台 client_secret")
	command.Flags().Int64Var(&options.subID, "sub-id", cfg.SubID, "用户 ID 或企业 ID")
	command.Flags().StringVar(&options.subType, "sub-type", string(cfg.SubType), "subject 类型: user|enterprise")
	command.Flags().StringVar(&options.oauthHost, "oauth-host", cfg.OAuthHost, "OAuth 地址")
	command.Flags().StringVar(&options.baseURL, "base-url", cfg.BaseURL, "OpenAPI 基地址")
}

func NewRootCommand(cfg config.Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "npan-cli",
		Short: "Npan Go CLI",
	}

	rootCmd.AddCommand(newTokenCommand(cfg))
	rootCmd.AddCommand(newSearchRemoteCommand(cfg))
	rootCmd.AddCommand(newSearchLocalCommand(cfg))
	rootCmd.AddCommand(newDownloadURLCommand(cfg))
	rootCmd.AddCommand(newSyncFullCommand(cfg))
	rootCmd.AddCommand(newSyncIncrementalCommand(cfg))
	rootCmd.AddCommand(newSyncProgressCommand(cfg))

	return rootCmd
}

func newTokenCommand(cfg config.Config) *cobra.Command {
	var options authOptions

	cmd := &cobra.Command{
		Use:   "token",
		Short: "获取访问 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			subType := npan.TokenSubjectType(firstNotEmpty(options.subType, string(cfg.SubType), "user"))
			token, err := npan.RequestAccessToken(cmd.Context(), nil, npan.TokenRequestOptions{
				OAuthHost:    firstNotEmpty(options.oauthHost, cfg.OAuthHost),
				ClientID:     firstNotEmpty(options.clientID, cfg.ClientID),
				ClientSecret: firstNotEmpty(options.clientSecret, cfg.ClientSecret),
				SubID:        firstPositive(options.subID, cfg.SubID),
				SubType:      subType,
			})
			if err != nil {
				return err
			}
			return printJSON(token)
		},
	}

	cmd.Flags().StringVar(&options.clientID, "client-id", cfg.ClientID, "开放平台 client_id")
	cmd.Flags().StringVar(&options.clientSecret, "client-secret", cfg.ClientSecret, "开放平台 client_secret")
	cmd.Flags().Int64Var(&options.subID, "sub-id", cfg.SubID, "用户 ID 或企业 ID")
	cmd.Flags().StringVar(&options.subType, "sub-type", string(cfg.SubType), "subject 类型: user|enterprise")
	cmd.Flags().StringVar(&options.oauthHost, "oauth-host", cfg.OAuthHost, "OAuth 地址")

	return cmd
}

func newSearchRemoteCommand(cfg config.Config) *cobra.Command {
	var options authOptions
	var query string
	var searchType string
	var pageID int64
	var queryFilter string
	var searchInFolder int64
	var hasSearchInFolder bool
	var updatedTimeRange string

	cmd := &cobra.Command{
		Use:   "search-remote",
		Short: "调用 Npan 平台接口搜索",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("--query 不能为空")
			}

			token, authOptions, err := resolveToken(cmd.Context(), cfg, options)
			if err != nil {
				return err
			}

			api := newAPIClient(firstNotEmpty(options.baseURL, cfg.BaseURL), token, authOptions)

			var searchInFolderPtr *int64
			if hasSearchInFolder {
				searchInFolderPtr = &searchInFolder
			}

			result, err := api.SearchItems(cmd.Context(), models.RemoteSearchParams{
				QueryWords:       query,
				Type:             searchType,
				PageID:           pageID,
				QueryFilter:      queryFilter,
				SearchInFolder:   searchInFolderPtr,
				UpdatedTimeRange: updatedTimeRange,
			})
			if err != nil {
				return err
			}

			return printJSON(result)
		},
	}

	addAuthFlags(cmd, &options, cfg)
	cmd.Flags().StringVar(&query, "query", "", "搜索关键词")
	cmd.Flags().StringVar(&searchType, "type", "all", "搜索类型: file|folder|all")
	cmd.Flags().Int64Var(&pageID, "page-id", 0, "页码，从 0 开始")
	cmd.Flags().StringVar(&queryFilter, "query-filter", "all", "过滤类型: file_name|content|creator|all")
	cmd.Flags().Int64Var(&searchInFolder, "search-in-folder", 0, "父目录 ID")
	cmd.Flags().BoolVar(&hasSearchInFolder, "with-search-in-folder", false, "是否启用 search-in-folder")
	cmd.Flags().StringVar(&updatedTimeRange, "updated-time-range", "", "更新时间范围: start,end")

	return cmd
}

func newSearchLocalCommand(cfg config.Config) *cobra.Command {
	var query string
	var indexName string
	var meiliHost string
	var meiliKey string
	var searchType string
	var page int64
	var pageSize int64
	var parentID int64
	var hasParentID bool
	var updatedAfter int64
	var hasUpdatedAfter bool
	var updatedBefore int64
	var hasUpdatedBefore bool
	var includeDeleted bool

	cmd := &cobra.Command{
		Use:   "search-local",
		Short: "搜索本地 Meilisearch 索引",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("--query 不能为空")
			}

			meiliIndex := search.NewMeiliIndex(meiliHost, meiliKey, indexName)
			queryService := search.NewQueryService(meiliIndex)

			var parentIDPtr *int64
			if hasParentID {
				parentIDPtr = &parentID
			}
			var updatedAfterPtr *int64
			if hasUpdatedAfter {
				updatedAfterPtr = &updatedAfter
			}
			var updatedBeforePtr *int64
			if hasUpdatedBefore {
				updatedBeforePtr = &updatedBefore
			}

			result, err := queryService.Query(models.LocalSearchParams{
				Query:          query,
				Type:           searchType,
				Page:           page,
				PageSize:       pageSize,
				ParentID:       parentIDPtr,
				UpdatedAfter:   updatedAfterPtr,
				UpdatedBefore:  updatedBeforePtr,
				IncludeDeleted: includeDeleted,
			})
			if err != nil {
				return err
			}

			return printJSON(result)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "搜索关键词")
	cmd.Flags().StringVar(&indexName, "index", cfg.MeiliIndex, "Meili 索引名")
	cmd.Flags().StringVar(&meiliHost, "meili-host", cfg.MeiliHost, "Meili 地址")
	cmd.Flags().StringVar(&meiliKey, "meili-key", cfg.MeiliAPIKey, "Meili API key")
	cmd.Flags().StringVar(&searchType, "type", "all", "搜索类型: file|folder|all")
	cmd.Flags().Int64Var(&page, "page", 1, "页码，从 1 开始")
	cmd.Flags().Int64Var(&pageSize, "page-size", 20, "每页数量")
	cmd.Flags().Int64Var(&parentID, "parent-id", 0, "父目录 ID")
	cmd.Flags().BoolVar(&hasParentID, "with-parent-id", false, "是否启用 parent-id")
	cmd.Flags().Int64Var(&updatedAfter, "updated-after", 0, "起始更新时间")
	cmd.Flags().BoolVar(&hasUpdatedAfter, "with-updated-after", false, "是否启用 updated-after")
	cmd.Flags().Int64Var(&updatedBefore, "updated-before", 0, "截止更新时间")
	cmd.Flags().BoolVar(&hasUpdatedBefore, "with-updated-before", false, "是否启用 updated-before")
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "是否包含删除/回收站")

	return cmd
}

func newDownloadURLCommand(cfg config.Config) *cobra.Command {
	var options authOptions
	var fileID int64
	var validPeriod int64
	var withValidPeriod bool

	cmd := &cobra.Command{
		Use:   "download-url",
		Short: "获取文件实时下载链接",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fileID <= 0 {
				return fmt.Errorf("--file-id 必须是正整数")
			}

			token, authOptions, err := resolveToken(cmd.Context(), cfg, options)
			if err != nil {
				return err
			}

			api := newAPIClient(firstNotEmpty(options.baseURL, cfg.BaseURL), token, authOptions)

			downloadService := service.NewDownloadURLService(api)
			var validPeriodPtr *int64
			if withValidPeriod {
				validPeriodPtr = &validPeriod
			}

			downloadURL, err := downloadService.GetDownloadURL(cmd.Context(), fileID, validPeriodPtr)
			if err != nil {
				return err
			}

			return printJSON(map[string]any{
				"file_id":      fileID,
				"download_url": downloadURL,
			})
		},
	}

	addAuthFlags(cmd, &options, cfg)
	cmd.Flags().Int64Var(&fileID, "file-id", 0, "文件 ID")
	cmd.Flags().Int64Var(&validPeriod, "valid-period", 0, "下载链接有效秒数")
	cmd.Flags().BoolVar(&withValidPeriod, "with-valid-period", false, "是否携带 valid-period")

	return cmd
}

func newSyncFullCommand(cfg config.Config) *cobra.Command {
	var options authOptions
	var rootFolderIDsRaw string
	var includeDepartments bool
	var departmentIDsRaw string
	var resumeProgress bool
	var rootWorkers int
	var progressEvery int
	var progressOutput string
	var checkpointTemplate string
	var meiliHost string
	var meiliKey string
	var meiliIndexName string

	cmd := &cobra.Command{
		Use:   "sync-full",
		Short: "执行全量同步到 Meilisearch",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, authOptions, err := resolveToken(cmd.Context(), cfg, options)
			if err != nil {
				return err
			}

			roots, err := parseInt64CSV(rootFolderIDsRaw)
			if err != nil {
				return err
			}
			if len(roots) == 0 {
				roots = append([]int64{}, cfg.DefaultRootFolderIDs...)
			}

			departmentIDs, err := parseInt64CSV(departmentIDsRaw)
			if err != nil {
				return err
			}
			if len(departmentIDs) == 0 {
				departmentIDs = append([]int64{}, cfg.DefaultDepartmentIDs...)
			}

			meiliIndex := search.NewMeiliIndex(meiliHost, meiliKey, meiliIndexName)
			if err := meiliIndex.EnsureSettings(cmd.Context()); err != nil {
				return err
			}

			syncManager := service.NewSyncManager(service.SyncManagerArgs{
				Index:              meiliIndex,
				ProgressStore:      storage.NewJSONProgressStore(cfg.ProgressFile),
				MeiliHost:          meiliHost,
				MeiliIndex:         meiliIndexName,
				CheckpointTemplate: checkpointTemplate,
				RootWorkers:        rootWorkers,
				ProgressEvery:      progressEvery,
				Retry:              cfg.Retry,
				MaxConcurrent:      cfg.SyncMaxConcurrent,
				MinTimeMS:          cfg.SyncMinTimeMS,
			})

			api := newAPIClient(firstNotEmpty(options.baseURL, cfg.BaseURL), token, authOptions)

			if err := syncManager.Start(api, service.SyncStartRequest{
				RootFolderIDs:      roots,
				IncludeDepartments: &includeDepartments,
				DepartmentIDs:      departmentIDs,
				ResumeProgress:     &resumeProgress,
				RootWorkers:        rootWorkers,
				ProgressEvery:      progressEvery,
				CheckpointTemplate: checkpointTemplate,
			}); err != nil {
				return err
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			outputMode, err := resolveSyncProgressOutputMode(progressOutput)
			if err != nil {
				return err
			}
			snapshot := &progressRenderSnapshot{}

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for syncManager.IsRunning() {
				select {
				case <-cmd.Context().Done():
					syncManager.Cancel()
					if !waitSyncManagerStopped(syncManager, 15*time.Second) {
						return fmt.Errorf("同步取消超时，任务仍在后台运行")
					}
					return cmd.Context().Err()
				case <-sigCh:
					syncManager.Cancel()
					if !waitSyncManagerStopped(syncManager, 15*time.Second) {
						return fmt.Errorf("收到中断信号，等待同步停止超时")
					}
					progress, _ := syncManager.GetProgress()
					if progress != nil {
						_ = printSyncFullProgress(progress, outputMode, snapshot)
					}
					return fmt.Errorf("收到中断信号，同步已取消")
				case <-ticker.C:
					progress, loadErr := syncManager.GetProgress()
					if loadErr != nil || progress == nil {
						continue
					}
					_ = printSyncFullProgress(progress, outputMode, snapshot)
				}
			}

			progress, err := syncManager.GetProgress()
			if err != nil {
				return err
			}
			if progress == nil {
				return fmt.Errorf("未找到同步进度")
			}
			if progress.Status == "error" {
				if progress.LastError != "" {
					return fmt.Errorf("同步失败: %s", progress.LastError)
				}
				return fmt.Errorf("同步失败")
			}

			return printJSON(progress)
		},
	}

	addAuthFlags(cmd, &options, cfg)
	cmd.Flags().StringVar(&rootFolderIDsRaw, "root-folder-ids", "", "根目录 ID 列表，逗号分隔")
	cmd.Flags().BoolVar(&includeDepartments, "include-departments", cfg.DefaultIncludeDepartments, "是否自动扫描部门根目录")
	cmd.Flags().StringVar(&departmentIDsRaw, "department-ids", "", "部门 ID 列表，逗号分隔")
	cmd.Flags().BoolVar(&resumeProgress, "resume-progress", true, "是否从现有进度恢复")
	cmd.Flags().IntVar(&rootWorkers, "root-workers", cfg.SyncRootWorkers, "根目录并发 worker 数")
	cmd.Flags().IntVar(&progressEvery, "progress-every", cfg.SyncProgressEvery, "每处理 N 页记录一次进度")
	cmd.Flags().StringVar(&progressOutput, "progress-output", "human", "进度输出模式: human|json")
	cmd.Flags().StringVar(&checkpointTemplate, "checkpoint-template", cfg.CheckpointTemplate, "checkpoint 文件模板")

	cmd.Flags().StringVar(&meiliHost, "meili-host", cfg.MeiliHost, "Meili 地址")
	cmd.Flags().StringVar(&meiliKey, "meili-key", cfg.MeiliAPIKey, "Meili API key")
	cmd.Flags().StringVar(&meiliIndexName, "meili-index", cfg.MeiliIndex, "Meili 索引名")

	return cmd
}

func waitSyncManagerStopped(syncManager *service.SyncManager, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for syncManager.IsRunning() {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
	return true
}

func newSyncIncrementalCommand(cfg config.Config) *cobra.Command {
	var options authOptions
	var meiliHost string
	var meiliKey string
	var meiliIndexName string
	var syncStateFile string
	var windowOverlapMS int64
	var incrementalQueryWords string

	cmd := &cobra.Command{
		Use:   "sync-incremental",
		Short: "执行增量同步到 Meilisearch",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, authOptions, err := resolveToken(cmd.Context(), cfg, options)
			if err != nil {
				return err
			}

			api := newAPIClient(firstNotEmpty(options.baseURL, cfg.BaseURL), token, authOptions)
			meiliIndex := search.NewMeiliIndex(meiliHost, meiliKey, meiliIndexName)
			if err := meiliIndex.EnsureSettings(cmd.Context()); err != nil {
				return err
			}

			stateStore := storage.NewJSONSyncStateStore(syncStateFile)
			beforeState, err := stateStore.Load()
			if err != nil {
				return err
			}

			if windowOverlapMS < 0 {
				windowOverlapMS = 0
			}

			windowEnd := time.Now().Unix()
			sinceUsed := int64(0)
			totalChanges := 0
			upsertCount := 0
			deleteCount := 0

			windowOverlapSeconds := windowOverlapMS / 1000
			if windowOverlapMS%1000 != 0 {
				windowOverlapSeconds++
			}

			err = indexer.RunIncrementalSync(cmd.Context(), indexer.IncrementalDeps{
				FetchChanges: func(ctx context.Context, since int64) ([]indexer.IncrementalInputItem, error) {
					sinceUsed = since
					querySince := since - windowOverlapSeconds
					if querySince < 0 {
						querySince = 0
					}

					changes, err := indexer.FetchIncrementalChanges(ctx, indexer.IncrementalFetchOptions{
						Since: querySince,
						Until: windowEnd,
						Retry: cfg.Retry,
						Fetch: func(ctx context.Context, start *int64, end *int64, pageID int64) (map[string]any, error) {
							return api.SearchUpdatedWindow(ctx, incrementalQueryWords, start, end, pageID)
						},
					})
					if err != nil {
						return nil, err
					}

					totalChanges = len(changes)
					upsertCount = 0
					deleteCount = 0
					for _, change := range changes {
						if change.Deleted {
							deleteCount++
						} else {
							upsertCount++
						}
					}

					return changes, nil
				},
				StateStore: stateStore,
				UpsertDocuments: func(ctx context.Context, docs []models.IndexDocument) error {
					return meiliIndex.UpsertDocuments(ctx, docs)
				},
				DeleteDocuments: func(ctx context.Context, docIDs []string) error {
					return meiliIndex.DeleteDocuments(ctx, docIDs)
				},
				NowProvider: func() int64 {
					return windowEnd
				},
			})
			if err != nil {
				return err
			}

			afterState, err := stateStore.Load()
			if err != nil {
				return err
			}

			beforeCursor := int64(0)
			if beforeState != nil {
				beforeCursor = normalizeUnixSeconds(beforeState.LastSyncTime)
			}
			afterCursor := int64(0)
			if afterState != nil {
				afterCursor = normalizeUnixSeconds(afterState.LastSyncTime)
			}

			return printJSON(map[string]any{
				"status":        "done",
				"cursor_before": beforeCursor,
				"cursor_after":  afterCursor,
				"since_used":    sinceUsed,
				"window_end":    windowEnd,
				"time_unit":     "seconds",
				"changes_total": totalChanges,
				"upserts":       upsertCount,
				"deletes":       deleteCount,
			})
		},
	}

	addAuthFlags(cmd, &options, cfg)
	cmd.Flags().StringVar(&meiliHost, "meili-host", cfg.MeiliHost, "Meili 地址")
	cmd.Flags().StringVar(&meiliKey, "meili-key", cfg.MeiliAPIKey, "Meili API key")
	cmd.Flags().StringVar(&meiliIndexName, "meili-index", cfg.MeiliIndex, "Meili 索引名")
	cmd.Flags().StringVar(&syncStateFile, "sync-state-file", cfg.SyncStateFile, "增量游标状态文件路径")
	cmd.Flags().Int64Var(&windowOverlapMS, "window-overlap-ms", 2000, "增量窗口回看毫秒数，防止边界漏数")
	cmd.Flags().StringVar(&incrementalQueryWords, "incremental-query-words", cfg.IncrementalQuery, "增量查询词（默认 * OR *，可覆盖）")

	return cmd
}

func newSyncProgressCommand(cfg config.Config) *cobra.Command {
	var progressFile string

	cmd := &cobra.Command{
		Use:   "sync-progress",
		Short: "查看全量同步进度",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := storage.NewJSONProgressStore(progressFile)
			progress, err := store.Load()
			if err != nil {
				return err
			}
			if progress == nil {
				return fmt.Errorf("未找到进度文件: %s", progressFile)
			}

			return printJSON(progress)
		},
	}

	cmd.Flags().StringVar(&progressFile, "progress-file", cfg.ProgressFile, "进度文件路径")
	return cmd
}
