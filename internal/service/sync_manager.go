package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"npan/internal/indexer"
	"npan/internal/metrics"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/storage"
)

type SyncStartRequest struct {
	Mode                models.SyncMode `json:"mode"`
	RootFolderIDs       []int64         `json:"root_folder_ids"`
	IncludeDepartments  *bool           `json:"include_departments"`
	PreserveRootCatalog *bool           `json:"preserve_root_catalog"`
	DepartmentIDs       []int64         `json:"department_ids"`
	ResumeProgress      *bool           `json:"resume_progress"`
	ForceRebuild        *bool           `json:"force_rebuild"`
	RootWorkers         int             `json:"root_workers"`
	ProgressEvery       int             `json:"progress_every"`
	CheckpointTemplate  string          `json:"checkpoint_template"`
	WindowOverlapMS     int64           `json:"window_overlap_ms"`
	IncrementalQuery    string          `json:"incremental_query"`
}

type SyncManager struct {
	index            search.IndexOperator
	progressStore    storage.ProgressStore
	syncStateStore   storage.SyncStateStore
	checkpointStores storage.CheckpointStoreFactory
	meiliHost        string
	meiliIndex       string

	defaultCheckpointTemplate string
	defaultRootWorkers        int
	defaultProgressEvery      int
	retry                     models.RetryPolicyOptions
	maxConcurrent             int
	minTimeMS                 int
	activityChecker           indexer.ActivityChecker

	defaultIncrementalQuery string
	defaultWindowOverlapMS  int64
	metricsReporter         metrics.SyncReporter

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

type SyncManagerArgs struct {
	Index              search.IndexOperator
	ProgressStore      storage.ProgressStore
	SyncStateStore     storage.SyncStateStore
	CheckpointStores   storage.CheckpointStoreFactory
	MeiliHost          string
	MeiliIndex         string
	CheckpointTemplate string
	RootWorkers        int
	ProgressEvery      int
	Retry              models.RetryPolicyOptions
	MaxConcurrent      int
	MinTimeMS          int
	ActivityChecker    indexer.ActivityChecker
	IncrementalQuery   string
	WindowOverlapMS    int64
	MetricsReporter    metrics.SyncReporter
}

func NewSyncManager(args SyncManagerArgs) *SyncManager {
	return &SyncManager{
		index:                     args.Index,
		progressStore:             args.ProgressStore,
		syncStateStore:            args.SyncStateStore,
		checkpointStores:          args.CheckpointStores,
		meiliHost:                 args.MeiliHost,
		meiliIndex:                args.MeiliIndex,
		defaultCheckpointTemplate: args.CheckpointTemplate,
		defaultRootWorkers:        args.RootWorkers,
		defaultProgressEvery:      args.ProgressEvery,
		retry:                     args.Retry,
		maxConcurrent:             args.MaxConcurrent,
		minTimeMS:                 args.MinTimeMS,
		activityChecker:           args.ActivityChecker,
		defaultIncrementalQuery:   args.IncrementalQuery,
		defaultWindowOverlapMS:    args.WindowOverlapMS,
		metricsReporter:           args.MetricsReporter,
	}
}

func (m *SyncManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *SyncManager) effectiveSyncStateStore() storage.SyncStateStore {
	return m.syncStateStore
}

func (m *SyncManager) effectiveCheckpointStoreFactory() storage.CheckpointStoreFactory {
	if m.checkpointStores != nil {
		return m.checkpointStores
	}
	return storage.NewJSONCheckpointStoreFactory()
}

func (m *SyncManager) GetProgress() (*models.SyncProgressState, error) {
	progress, err := m.progressStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load progress: %w", err)
	}

	isRunning := m.IsRunning()

	if progress == nil {
		if isRunning {
			// Sync goroutine started but hasn't saved initial progress yet.
			// Return a minimal running state so the frontend sees "running".
			return &models.SyncProgressState{
				Status:         "running",
				StartedAt:      time.Now().UnixMilli(),
				UpdatedAt:      time.Now().UnixMilli(),
				Roots:          []int64{},
				CompletedRoots: []int64{},
				RootProgress:   map[string]*models.RootSyncProgress{},
			}, nil
		}
		return nil, nil
	}

	// If progress says "running" but no goroutine is active (e.g. after
	// container restart), mark as interrupted so the UI shows the real state.
	if progress.Status == "running" && !isRunning {
		progress.Status = "interrupted"
		progress.LastError = "进程重启，同步中断"
		progress.ActiveRoot = nil
		for _, root := range progress.RootProgress {
			if root == nil || root.Status != "running" {
				continue
			}
			root.Status = "interrupted"
			root.Error = "进程重启，同步中断"
			root.CurrentFolderID = nil
			root.CurrentPageID = nil
			root.CurrentPageCount = nil
			root.QueueLength = nil
			root.UpdatedAt = time.Now().UnixMilli()
		}
		progress.UpdatedAt = time.Now().UnixMilli()
		if err := m.progressStore.Save(progress); err != nil {
			slog.Warn("保存进度失败", "error", err)
		}
	}

	// Goroutine is running but hasn't overwritten old progress yet.
	// Override status in-memory only (don't write to store — goroutine will save real progress).
	if isRunning && progress.Status != "running" {
		progress.Status = "running"
		progress.LastError = ""
	}

	return progress, nil
}

func (m *SyncManager) Cancel() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running || m.cancel == nil {
		return false
	}
	m.cancel()
	return true
}

func (m *SyncManager) GetIndexDocumentCount(ctx context.Context) (int64, error) {
	if m.index == nil {
		return 0, fmt.Errorf("索引服务未初始化")
	}
	return m.index.DocumentCount(ctx)
}

func (m *SyncManager) Start(api npan.API, request SyncStartRequest) error {
	effectiveMode, err := resolveMode(request.Mode)
	if err != nil {
		return err
	}

	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("已有全量同步任务在运行")
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.running = true
	m.cancel = cancel
	m.mu.Unlock()

	if m.metricsReporter != nil {
		m.metricsReporter.ReportSyncStarted(effectiveMode)
	}

	go func() {
		defer func() {
			m.mu.Lock()
			m.running = false
			m.cancel = nil
			m.mu.Unlock()
		}()

		_ = m.run(ctx, api, request)
	}()

	return nil
}

func buildCheckpointFilePath(template string, rootID int64, multiRoots bool) string {
	if !multiRoots {
		return template
	}

	if len(template) > 5 && template[len(template)-5:] == ".json" {
		return template[:len(template)-5] + fmt.Sprintf(".%d.json", rootID)
	}
	return template + fmt.Sprintf(".%d.json", rootID)
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func removeInt64(items []int64, target int64) []int64 {
	result := make([]int64, 0, len(items))
	for _, item := range items {
		if item == target {
			continue
		}
		result = append(result, item)
	}
	return result
}

func uniqueSorted(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func shouldPreserveRootCatalog(request SyncStartRequest, forceRebuild bool) bool {
	if forceRebuild {
		return false
	}
	if request.PreserveRootCatalog != nil {
		return *request.PreserveRootCatalog
	}
	return len(request.RootFolderIDs) > 0
}

func cloneRootSyncProgress(src *models.RootSyncProgress) *models.RootSyncProgress {
	if src == nil {
		return nil
	}
	cloned := *src
	if src.EstimatedTotalDocs != nil {
		v := *src.EstimatedTotalDocs
		cloned.EstimatedTotalDocs = &v
	}
	if src.CurrentFolderID != nil {
		v := *src.CurrentFolderID
		cloned.CurrentFolderID = &v
	}
	if src.CurrentPageID != nil {
		v := *src.CurrentPageID
		cloned.CurrentPageID = &v
	}
	if src.CurrentPageCount != nil {
		v := *src.CurrentPageCount
		cloned.CurrentPageCount = &v
	}
	if src.QueueLength != nil {
		v := *src.QueueLength
		cloned.QueueLength = &v
	}
	return &cloned
}

func mergeHistoricalRootCatalog(progress *models.SyncProgressState, existing *models.SyncProgressState) {
	if progress == nil || existing == nil {
		return
	}

	if progress.RootNames == nil {
		progress.RootNames = map[int64]string{}
	}
	for id, name := range existing.RootNames {
		if _, exists := progress.RootNames[id]; !exists {
			progress.RootNames[id] = name
		}
	}

	if progress.RootProgress == nil {
		progress.RootProgress = map[string]*models.RootSyncProgress{}
	}
	for key, rp := range existing.RootProgress {
		if _, exists := progress.RootProgress[key]; exists {
			continue
		}
		progress.RootProgress[key] = cloneRootSyncProgress(rp)
	}
}

func syncCatalogFields(progress *models.SyncProgressState) {
	if progress == nil {
		return
	}

	if progress.RootProgress == nil {
		progress.RootProgress = map[string]*models.RootSyncProgress{}
	}
	progress.CatalogRootProgress = progress.RootProgress

	if progress.RootNames == nil {
		progress.RootNames = map[int64]string{}
	}
	progress.CatalogRootNames = make(map[int64]string, len(progress.RootNames))
	for id, name := range progress.RootNames {
		progress.CatalogRootNames[id] = name
	}

	roots := make([]int64, 0, len(progress.RootProgress))
	seen := map[int64]struct{}{}
	for key, rp := range progress.RootProgress {
		rootID := int64(0)
		if rp != nil && rp.RootFolderID > 0 {
			rootID = rp.RootFolderID
		}
		if rootID <= 0 {
			parsed, err := strconv.ParseInt(key, 10, 64)
			if err == nil {
				rootID = parsed
			}
		}
		if rootID <= 0 {
			continue
		}
		if _, exists := seen[rootID]; exists {
			continue
		}
		seen[rootID] = struct{}{}
		roots = append(roots, rootID)
	}
	progress.CatalogRoots = uniqueSorted(roots)
}

func resolveMode(mode models.SyncMode) (models.SyncMode, error) {
	switch mode {
	case "":
		return models.SyncModeFull, nil
	case models.SyncModeFull:
		return models.SyncModeFull, nil
	case models.SyncModeIncremental:
		return models.SyncModeIncremental, nil
	default:
		return "", fmt.Errorf("不支持的同步模式: %s（可选: full|incremental）", mode)
	}
}

func (m *SyncManager) discoverRootFolders(ctx context.Context, api npan.API, request SyncStartRequest) ([]int64, map[int64]int64, map[int64]string, error) {
	roots := append([]int64{}, request.RootFolderIDs...)
	rootEstimateMap := map[int64]int64{}
	rootNameMap := map[int64]string{}

	for _, rootID := range request.RootFolderIDs {
		// Synthetic root entry; don't call upstream folder info endpoint.
		if rootID == 0 {
			rootNameMap[rootID] = "全部文件"
			continue
		}

		folder, err := api.GetFolderInfo(ctx, rootID)
		if err != nil {
			slog.Warn("获取根目录信息失败，降级继续", "root_id", rootID, "error", err)
			continue
		}
		if folder.Name != "" {
			rootNameMap[rootID] = folder.Name
		}
		estimate := folder.ItemCount + 1
		if estimate > 0 {
			rootEstimateMap[rootID] = estimate
		}
	}

	includeDepartments := request.IncludeDepartments == nil || *request.IncludeDepartments
	if includeDepartments {
		departmentIDs := append([]int64{}, request.DepartmentIDs...)
		if len(departmentIDs) == 0 {
			deps, err := api.ListUserDepartments(ctx)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("list user departments: %w", err)
			}
			for _, dep := range deps {
				departmentIDs = append(departmentIDs, dep.ID)
			}
		}

		for _, departmentID := range departmentIDs {
			folders, err := api.ListDepartmentFolders(ctx, departmentID)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("list department folders (dept %d): %w", departmentID, err)
			}
			for _, folder := range folders {
				roots = append(roots, folder.ID)
				if folder.Name != "" {
					rootNameMap[folder.ID] = folder.Name
				}
				estimate := folder.ItemCount + 1
				if estimate <= 0 {
					continue
				}
				if existing, exists := rootEstimateMap[folder.ID]; !exists || estimate > existing {
					rootEstimateMap[folder.ID] = estimate
				}
			}
		}
	}

	return uniqueSorted(roots), rootEstimateMap, rootNameMap, nil
}

func createInitialProgress(args struct {
	Roots              []int64
	RootCheckpointMap  map[int64]string
	RootEstimateMap    map[int64]int64
	RootNameMap        map[int64]string
	StartedAt          int64
	MeiliHost          string
	MeiliIndex         string
	CheckpointTemplate string
}) *models.SyncProgressState {
	rootProgress := map[string]*models.RootSyncProgress{}
	for _, root := range args.Roots {
		rootProgress[fmt.Sprintf("%d", root)] = &models.RootSyncProgress{
			RootFolderID:   root,
			CheckpointFile: args.RootCheckpointMap[root],
			Status:         "pending",
			Stats: models.CrawlStats{
				StartedAt: args.StartedAt,
				EndedAt:   args.StartedAt,
			},
			UpdatedAt: args.StartedAt,
		}
		if estimate, exists := args.RootEstimateMap[root]; exists && estimate > 0 {
			estimateCopy := estimate
			rootProgress[fmt.Sprintf("%d", root)].EstimatedTotalDocs = &estimateCopy
		}
	}

	return &models.SyncProgressState{
		Status:             "running",
		StartedAt:          args.StartedAt,
		UpdatedAt:          args.StartedAt,
		MeiliHost:          args.MeiliHost,
		MeiliIndex:         args.MeiliIndex,
		CheckpointTemplate: args.CheckpointTemplate,
		Roots:              append([]int64{}, args.Roots...),
		RootNames:          args.RootNameMap,
		CompletedRoots:     []int64{},
		AggregateStats: models.CrawlStats{
			StartedAt: args.StartedAt,
			EndedAt:   args.StartedAt,
		},
		RootProgress: rootProgress,
	}
}

func updateAggregateFromRoots(progress *models.SyncProgressState) {
	aggregate := models.CrawlStats{
		StartedAt: progress.AggregateStats.StartedAt,
		EndedAt:   time.Now().UnixMilli(),
	}

	for _, rootID := range progress.Roots {
		root := progress.RootProgress[fmt.Sprintf("%d", rootID)]
		if root == nil {
			continue
		}
		aggregate.FoldersVisited += root.Stats.FoldersVisited
		aggregate.FilesIndexed += root.Stats.FilesIndexed
		aggregate.PagesFetched += root.Stats.PagesFetched
		aggregate.FailedRequests += root.Stats.FailedRequests
	}

	progress.AggregateStats = aggregate
	progress.UpdatedAt = time.Now().UnixMilli()
}

func restoreProgress(existing *models.SyncProgressState, roots []int64, rootCheckpointMap map[int64]string, rootEstimateMap map[int64]int64, rootNameMap map[int64]string) *models.SyncProgressState {
	now := time.Now().UnixMilli()
	restored := *existing
	restored.Status = "running"
	restored.UpdatedAt = now
	restored.ActiveRoot = nil
	restored.LastError = ""
	restored.Roots = append([]int64{}, roots...)
	restored.RootNames = rootNameMap
	restored.CompletedRoots = []int64{}
	if restored.RootProgress == nil {
		restored.RootProgress = map[string]*models.RootSyncProgress{}
	}

	for _, rootID := range roots {
		key := fmt.Sprintf("%d", rootID)
		rp, exists := restored.RootProgress[key]
		if !exists || rp == nil {
			restored.RootProgress[key] = &models.RootSyncProgress{
				RootFolderID:   rootID,
				CheckpointFile: rootCheckpointMap[rootID],
				Status:         "pending",
				Stats: models.CrawlStats{
					StartedAt: existing.StartedAt,
					EndedAt:   existing.StartedAt,
				},
				UpdatedAt: now,
			}
			if estimate, hasEstimate := rootEstimateMap[rootID]; hasEstimate && estimate > 0 {
				estimateCopy := estimate
				restored.RootProgress[key].EstimatedTotalDocs = &estimateCopy
			}
			continue
		}

		rp.CheckpointFile = rootCheckpointMap[rootID]
		rp.UpdatedAt = now
		if estimate, hasEstimate := rootEstimateMap[rootID]; hasEstimate && estimate > 0 {
			estimateCopy := estimate
			rp.EstimatedTotalDocs = &estimateCopy
		} else {
			rp.EstimatedTotalDocs = nil
		}
		if rp.Status == "done" || containsInt64(existing.CompletedRoots, rootID) {
			rp.Status = "done"
			restored.CompletedRoots = append(restored.CompletedRoots, rootID)
		} else {
			rp.Status = "pending"
			rp.Error = ""
		}
	}

	updateAggregateFromRoots(&restored)
	return &restored
}

func (m *SyncManager) runSingleRoot(ctx context.Context, api npan.API, progress *models.SyncProgressState, progressMu *sync.Mutex, rootID int64, checkpointFile string, progressEvery int, limiter *indexer.RequestLimiter, resetStats bool) error {
	key := fmt.Sprintf("%d", rootID)

	progressMu.Lock()
	rp := progress.RootProgress[key]
	if rp == nil {
		progressMu.Unlock()
		return nil
	}

	resumeBase := rp.Stats
	rp.Status = "running"
	rp.Error = ""
	now := time.Now().UnixMilli()
	if resetStats {
		resumeBase = models.CrawlStats{
			StartedAt: now,
			EndedAt:   now,
		}
		rp.Stats = resumeBase
		progress.CompletedRoots = removeInt64(progress.CompletedRoots, rootID)
	}
	rp.UpdatedAt = now
	progress.ActiveRoot = &rootID
	updateAggregateFromRoots(progress)
	if err := m.progressStore.Save(progress); err != nil {
		progressMu.Unlock()
		return err
	}
	progressMu.Unlock()

	checkpointStore := m.effectiveCheckpointStoreFactory().ForKey(checkpointFile)

	stats, err := indexer.RunFullCrawl(ctx, indexer.FullCrawlDeps{
		API:             api,
		IndexWriter:     &indexWriter{index: m.index},
		Limiter:         limiter,
		CheckpointStore: checkpointStore,
		RootFolderID:    rootID,
		Retry:           m.retry,
		OnProgress: func(event indexer.ProgressEvent) {
			if progressEvery > 1 && event.Stats.PagesFetched%int64(progressEvery) != 0 {
				return
			}

			progressMu.Lock()
			defer progressMu.Unlock()

			root := progress.RootProgress[key]
			if root == nil {
				return
			}

			root.Stats = models.CrawlStats{
				FoldersVisited: resumeBase.FoldersVisited + event.Stats.FoldersVisited,
				FilesIndexed:   resumeBase.FilesIndexed + event.Stats.FilesIndexed,
				PagesFetched:   resumeBase.PagesFetched + event.Stats.PagesFetched,
				FailedRequests: resumeBase.FailedRequests + event.Stats.FailedRequests,
				StartedAt:      resumeBase.StartedAt,
				EndedAt:        time.Now().UnixMilli(),
			}
			root.CurrentFolderID = &event.CurrentFolderID
			root.CurrentPageID = &event.CurrentPageID
			root.CurrentPageCount = &event.CurrentPageCount
			root.QueueLength = &event.QueueLength
			root.UpdatedAt = time.Now().UnixMilli()

			updateAggregateFromRoots(progress)
			_ = m.progressStore.Save(progress)
		},
	})

	progressMu.Lock()
	defer progressMu.Unlock()

	rp = progress.RootProgress[key]
	if rp == nil {
		return err
	}

	if err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			rp.Status = "cancelled"
			rp.Error = ctx.Err().Error()
			rp.UpdatedAt = time.Now().UnixMilli()
			updateAggregateFromRoots(progress)
			if err := m.progressStore.Save(progress); err != nil {
				slog.Warn("保存进度失败", "error", err)
			}
			return err
		}

		rp.Status = "error"
		rp.Error = err.Error()
		rp.UpdatedAt = time.Now().UnixMilli()
		progress.Status = "error"
		progress.LastError = err.Error()
		updateAggregateFromRoots(progress)
		if err := m.progressStore.Save(progress); err != nil {
			slog.Warn("保存进度失败", "error", err)
		}
		return err
	}

	rp.Status = "done"
	rp.Error = ""
	rp.Stats = models.CrawlStats{
		FoldersVisited: resumeBase.FoldersVisited + stats.FoldersVisited,
		FilesIndexed:   resumeBase.FilesIndexed + stats.FilesIndexed,
		PagesFetched:   resumeBase.PagesFetched + stats.PagesFetched,
		FailedRequests: resumeBase.FailedRequests + stats.FailedRequests,
		StartedAt:      resumeBase.StartedAt,
		EndedAt:        stats.EndedAt,
	}
	rp.CurrentFolderID = nil
	rp.CurrentPageID = nil
	rp.CurrentPageCount = nil
	zero := int64(0)
	rp.QueueLength = &zero
	rp.UpdatedAt = time.Now().UnixMilli()

	if !containsInt64(progress.CompletedRoots, rootID) {
		progress.CompletedRoots = append(progress.CompletedRoots, rootID)
	}

	updateAggregateFromRoots(progress)
	return m.progressStore.Save(progress)
}

func (m *SyncManager) runIncremental(ctx context.Context, api npan.API, progress *models.SyncProgressState, request SyncStartRequest, limiter *indexer.RequestLimiter) error {
	query := request.IncrementalQuery
	if query == "" {
		query = m.defaultIncrementalQuery
	}
	if query == "" {
		query = "* OR *"
	}

	overlapMS := request.WindowOverlapMS
	if overlapMS <= 0 {
		overlapMS = m.defaultWindowOverlapMS
	}

	if progress.IncrementalStats == nil {
		progress.IncrementalStats = &models.IncrementalSyncStats{}
	}

	cursorBefore := progress.IncrementalStats.CursorBefore
	since := cursorBefore
	if since > 0 && overlapMS > 0 {
		overlapSec := overlapMS / 1000
		if overlapSec > 0 {
			since -= overlapSec
			if since < 0 {
				since = 0
			}
		}
	}

	changes, err := indexer.FetchIncrementalChanges(ctx, indexer.IncrementalFetchOptions{
		Since: since,
		Until: 0,
		Retry: m.retry,
		Fetch: func(ctx context.Context, start *int64, end *int64, pageID int64) (map[string]any, error) {
			var result map[string]any
			schedErr := limiter.Schedule(ctx, func() error {
				var fetchErr error
				result, fetchErr = api.SearchUpdatedWindow(ctx, query, start, end, pageID)
				return fetchErr
			})
			return result, schedErr
		},
	})
	if err != nil {
		return fmt.Errorf("fetch incremental changes: %w", err)
	}

	var upserts []models.IndexDocument
	var deleteIDs []string
	for _, item := range changes {
		if item.Deleted {
			deleteIDs = append(deleteIDs, item.Doc.DocID)
		} else {
			upserts = append(upserts, item.Doc)
		}
	}

	progress.IncrementalStats.ChangesFetched = int64(len(changes))

	if len(upserts) > 0 {
		err := indexer.WithRetryVoid(ctx, func() error {
			return m.index.UpsertDocuments(ctx, upserts)
		}, m.retry)
		if err != nil {
			progress.IncrementalStats.SkippedUpserts += int64(len(upserts))
		} else {
			progress.IncrementalStats.Upserted += int64(len(upserts))
		}
	}

	if len(deleteIDs) > 0 {
		err := indexer.WithRetryVoid(ctx, func() error {
			return m.index.DeleteDocuments(ctx, deleteIDs)
		}, m.retry)
		if err != nil {
			progress.IncrementalStats.SkippedDeletes += int64(len(deleteIDs))
		} else {
			progress.IncrementalStats.Deleted += int64(len(deleteIDs))
		}
	}

	progress.IncrementalStats.CursorAfter = time.Now().UnixMilli()
	return nil
}

func (m *SyncManager) run(ctx context.Context, api npan.API, request SyncStartRequest) error {
	// Mode resolution
	effectiveMode, err := resolveMode(request.Mode)
	if err != nil {
		return err
	}

	syncStateStore := m.effectiveSyncStateStore()
	var syncState *models.SyncState
	if syncStateStore != nil {
		syncState, _ = syncStateStore.Load()
	}

	if effectiveMode == models.SyncModeIncremental {
		if syncState == nil || syncState.LastSyncTime <= 0 {
			return fmt.Errorf("增量同步需要先执行一次全量同步")
		}
		return m.runIncrementalPath(ctx, api, request, syncState, syncStateStore)
	}

	// Full crawl path
	forceRebuild := request.ForceRebuild != nil && *request.ForceRebuild
	if forceRebuild {
		slog.Info("强制重建索引：清空所有文档")
		if err := m.index.DeleteAllDocuments(ctx); err != nil {
			return fmt.Errorf("清空索引失败: %w", err)
		}
		slog.Info("强制重建索引：重新应用索引设置")
		if err := m.index.EnsureSettings(ctx); err != nil {
			return fmt.Errorf("重新应用索引设置失败: %w", err)
		}
	}

	fullStartTime := time.Now()
	roots, rootEstimateMap, rootNameMap, err := m.discoverRootFolders(ctx, api, request)
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return fmt.Errorf("未发现可遍历的根目录")
	}

	checkpointTemplate := request.CheckpointTemplate
	if checkpointTemplate == "" {
		checkpointTemplate = m.defaultCheckpointTemplate
	}

	rootCheckpointMap := map[int64]string{}
	for _, root := range roots {
		rootCheckpointMap[root] = buildCheckpointFilePath(checkpointTemplate, root, len(roots) > 1)
	}

	resume := true
	if request.ResumeProgress != nil {
		resume = *request.ResumeProgress
	}
	if forceRebuild {
		resume = false
	}

	// Force rebuild and explicit non-resume runs must start from a clean
	// crawl checkpoint, otherwise full crawl may resume from a stale queue.
	if forceRebuild || !resume {
		for _, rootID := range roots {
			checkpointStore := m.effectiveCheckpointStoreFactory().ForKey(rootCheckpointMap[rootID])
			if err := checkpointStore.Clear(); err != nil {
				return fmt.Errorf("clear checkpoint for root %d: %w", rootID, err)
			}
		}
	}

	existing, err := m.progressStore.Load()
	if err != nil {
		return err
	}

	var progress *models.SyncProgressState
	if resume && existing != nil {
		progress = restoreProgress(existing, roots, rootCheckpointMap, rootEstimateMap, rootNameMap)
	} else {
		startedAt := time.Now().UnixMilli()
		progress = createInitialProgress(struct {
			Roots              []int64
			RootCheckpointMap  map[int64]string
			RootEstimateMap    map[int64]int64
			RootNameMap        map[int64]string
			StartedAt          int64
			MeiliHost          string
			MeiliIndex         string
			CheckpointTemplate string
		}{
			Roots:              roots,
			RootCheckpointMap:  rootCheckpointMap,
			RootEstimateMap:    rootEstimateMap,
			RootNameMap:        rootNameMap,
			StartedAt:          startedAt,
			MeiliHost:          m.meiliHost,
			MeiliIndex:         m.meiliIndex,
			CheckpointTemplate: checkpointTemplate,
		})
	}
	if shouldPreserveRootCatalog(request, forceRebuild) {
		mergeHistoricalRootCatalog(progress, existing)
	}
	syncCatalogFields(progress)

	if err := m.progressStore.Save(progress); err != nil {
		return err
	}

	rootWorkers := request.RootWorkers
	if rootWorkers <= 0 {
		rootWorkers = m.defaultRootWorkers
	}
	if rootWorkers <= 0 {
		rootWorkers = 1
	}

	progressEvery := request.ProgressEvery
	if progressEvery <= 0 {
		progressEvery = m.defaultProgressEvery
	}
	if progressEvery <= 0 {
		progressEvery = 1
	}

	semaphore := make(chan struct{}, rootWorkers)
	progressMu := &sync.Mutex{}
	limiter := indexer.NewRequestLimiter(m.maxConcurrent, m.minTimeMS)
	if m.activityChecker != nil {
		limiter.SetActivityChecker(m.activityChecker)
	}
	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	var (
		wg         sync.WaitGroup
		firstErr   error
		firstErrMu sync.Mutex
	)

	setFirstErr := func(err error) {
		if err == nil {
			return
		}
		firstErrMu.Lock()
		defer firstErrMu.Unlock()
		if firstErr != nil {
			return
		}
		firstErr = err
		runCancel()
	}

	for _, rootID := range roots {
		status := "pending"
		if rp := progress.RootProgress[fmt.Sprintf("%d", rootID)]; rp != nil {
			status = rp.Status
		}
		if resume && status == "done" {
			continue
		}

		wg.Add(1)
		go func(currentRoot int64) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
			case <-runCtx.Done():
				return
			}
			defer func() { <-semaphore }()

			if err := m.runSingleRoot(runCtx, api, progress, progressMu, currentRoot, rootCheckpointMap[currentRoot], progressEvery, limiter, false); err != nil {
				if errors.Is(err, context.Canceled) && ctx.Err() != nil {
					return
				}
				setFirstErr(err)
			}
		}(rootID)
	}

	wg.Wait()

	progressMu.Lock()
	defer progressMu.Unlock()

	if firstErr != nil {
		progress.Status = "error"
		progress.LastError = firstErr.Error()
		progress.ActiveRoot = nil
		updateAggregateFromRoots(progress)
		if err := m.progressStore.Save(progress); err != nil {
			slog.Warn("保存进度失败", "error", err)
		}
		if m.metricsReporter != nil {
			m.metricsReporter.ReportSyncFinished(metrics.SyncEvent{
				Mode:     models.SyncModeFull,
				Status:   "error",
				Duration: time.Since(fullStartTime),
				Stats:    progress.AggregateStats,
			})
		}
		return firstErr
	}

	if ctx.Err() != nil {
		progress.Status = "cancelled"
		progress.LastError = ctx.Err().Error()
		progress.ActiveRoot = nil
		updateAggregateFromRoots(progress)
		if err := m.progressStore.Save(progress); err != nil {
			slog.Warn("保存进度失败", "error", err)
		}
		if m.metricsReporter != nil {
			m.metricsReporter.ReportSyncFinished(metrics.SyncEvent{
				Mode:     models.SyncModeFull,
				Status:   "cancelled",
				Duration: time.Since(fullStartTime),
				Stats:    progress.AggregateStats,
			})
		}
		return ctx.Err()
	}

	progress.Status = "done"
	progress.LastError = ""
	progress.ActiveRoot = nil
	progress.Mode = string(models.SyncModeFull)
	updateAggregateFromRoots(progress)

	// Write cursor for future incremental runs
	if syncStateStore != nil {
		_ = syncStateStore.Save(&models.SyncState{LastSyncTime: time.Now().UnixMilli()})
	}

	meiliCount, err := m.index.DocumentCount(ctx)
	if err == nil {
		progress.Verification = buildVerification(meiliCount, progress.AggregateStats)
		appendRootEstimateWarnings(progress.Verification, progress)
	}

	if m.metricsReporter != nil {
		m.metricsReporter.ReportSyncFinished(metrics.SyncEvent{
			Mode:     models.SyncModeFull,
			Status:   "done",
			Duration: time.Since(fullStartTime),
			Stats:    progress.AggregateStats,
		})
	}

	return m.progressStore.Save(progress)
}

func (m *SyncManager) runIncrementalPath(ctx context.Context, api npan.API, request SyncStartRequest, syncState *models.SyncState, syncStateStore storage.SyncStateStore) error {
	incrStartTime := time.Now()
	now := incrStartTime.UnixMilli()

	cursorBefore := int64(0)
	if syncState != nil && syncState.LastSyncTime > 0 {
		cursorBefore = syncState.LastSyncTime
		if cursorBefore >= 1_000_000_000_000 {
			cursorBefore = cursorBefore / 1000
		}
	}

	// Preserve roots/rootProgress from the previous full sync so that
	// the admin UI can still display per-root details after an incremental run.
	existing, _ := m.progressStore.Load()

	progress := &models.SyncProgressState{
		Status:           "running",
		Mode:             string(models.SyncModeIncremental),
		StartedAt:        now,
		UpdatedAt:        now,
		MeiliHost:        m.meiliHost,
		MeiliIndex:       m.meiliIndex,
		Roots:            []int64{},
		CompletedRoots:   []int64{},
		RootNames:        map[int64]string{},
		RootProgress:     map[string]*models.RootSyncProgress{},
		AggregateStats:   models.CrawlStats{StartedAt: now, EndedAt: now},
		IncrementalStats: &models.IncrementalSyncStats{CursorBefore: cursorBefore},
	}

	if existing != nil {
		progress.Roots = existing.Roots
		progress.CompletedRoots = existing.CompletedRoots
		progress.RootNames = existing.RootNames
		progress.RootProgress = existing.RootProgress
		progress.AggregateStats = existing.AggregateStats
		progress.CatalogRoots = existing.CatalogRoots
		progress.CatalogRootNames = existing.CatalogRootNames
		progress.CatalogRootProgress = existing.CatalogRootProgress
	}
	syncCatalogFields(progress)

	if err := m.progressStore.Save(progress); err != nil {
		return err
	}

	limiter := indexer.NewRequestLimiter(m.maxConcurrent, m.minTimeMS)
	if m.activityChecker != nil {
		limiter.SetActivityChecker(m.activityChecker)
	}

	err := m.runIncremental(ctx, api, progress, request, limiter)
	if err == nil {
		err = m.runIncrementalRepairs(ctx, api, progress, request, limiter)
	}

	if err != nil {
		if ctx.Err() != nil {
			progress.Status = "cancelled"
			progress.LastError = ctx.Err().Error()
		} else {
			progress.Status = "error"
			progress.LastError = err.Error()
		}
	} else {
		progress.Status = "done"

		if syncStateStore != nil {
			_ = syncStateStore.Save(&models.SyncState{
				LastSyncTime: progress.IncrementalStats.CursorAfter,
			})
		}

		meiliCount, verErr := m.index.DocumentCount(ctx)
		if verErr == nil {
			progress.Verification = buildVerification(meiliCount, progress.AggregateStats)
		}
	}

	progress.UpdatedAt = time.Now().UnixMilli()
	syncCatalogFields(progress)

	if m.metricsReporter != nil {
		m.metricsReporter.ReportSyncFinished(metrics.SyncEvent{
			Mode:      models.SyncModeIncremental,
			Status:    progress.Status,
			Duration:  time.Since(incrStartTime),
			Stats:     progress.AggregateStats,
			IncrStats: progress.IncrementalStats,
		})
	}

	return m.progressStore.Save(progress)
}

type indexWriter struct {
	index search.IndexOperator
}

func (w *indexWriter) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
	return w.index.UpsertDocuments(ctx, docs)
}

func buildVerification(meiliCount int64, stats models.CrawlStats) *models.SyncVerification {
	crawled := stats.FilesIndexed + stats.FoldersVisited
	discovered := stats.FilesDiscovered + stats.FoldersVisited

	v := &models.SyncVerification{
		MeiliDocCount:      meiliCount,
		CrawledDocCount:    crawled,
		DiscoveredDocCount: discovered,
		SkippedCount:       stats.SkippedFiles,
		Verified:           true,
	}

	if meiliCount < crawled {
		v.Warnings = append(v.Warnings,
			fmt.Sprintf("索引文档数(%d) < 爬取写入数(%d)", meiliCount, crawled))
	}
	if discovered > 0 && crawled < discovered {
		v.Warnings = append(v.Warnings,
			fmt.Sprintf("已索引(%d) < 已发现(%d), 跳过(%d)", crawled, discovered, stats.SkippedFiles))
	}

	return v
}

func appendRootEstimateWarnings(verification *models.SyncVerification, progress *models.SyncProgressState) {
	if verification == nil || progress == nil {
		return
	}

	const (
		minAbsGap   int64   = 20
		minGapRatio float64 = 0.05
	)

	for _, rootID := range progress.Roots {
		root := progress.RootProgress[fmt.Sprintf("%d", rootID)]
		if root == nil || root.EstimatedTotalDocs == nil || *root.EstimatedTotalDocs <= 0 {
			continue
		}

		estimated := *root.EstimatedTotalDocs
		actual := root.Stats.FilesIndexed + root.Stats.FoldersVisited
		gap := estimated - actual
		if gap <= 0 {
			continue
		}

		gapRatio := float64(gap) / float64(estimated)
		if gap <= minAbsGap && gapRatio <= minGapRatio {
			continue
		}

		name := progress.RootNames[rootID]
		if name == "" {
			name = fmt.Sprintf("%d", rootID)
		}
		verification.Warnings = append(verification.Warnings, fmt.Sprintf(
			"根目录 %s(%d) 估计文档数(%d) 与实际索引统计(%d) 差异较大，差值=%d",
			name, rootID, estimated, actual, gap,
		))
	}
}
