package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"npan/internal/indexer"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
	"npan/internal/storage"
)

type SyncStartRequest struct {
	Mode               models.SyncMode `json:"mode"`
	RootFolderIDs      []int64         `json:"root_folder_ids"`
	IncludeDepartments *bool           `json:"include_departments"`
	DepartmentIDs      []int64         `json:"department_ids"`
	ResumeProgress     *bool           `json:"resume_progress"`
	RootWorkers        int             `json:"root_workers"`
	ProgressEvery      int             `json:"progress_every"`
	CheckpointTemplate string          `json:"checkpoint_template"`
	WindowOverlapMS    int64           `json:"window_overlap_ms"`
	IncrementalQuery   string          `json:"incremental_query"`
}

type SyncManager struct {
	index         *search.MeiliIndex
	progressStore *storage.JSONProgressStore
	meiliHost     string
	meiliIndex    string

	defaultCheckpointTemplate string
	defaultRootWorkers        int
	defaultProgressEvery      int
	retry                     models.RetryPolicyOptions
	maxConcurrent             int
	minTimeMS                 int
	activityChecker           indexer.ActivityChecker

	syncStateFile             string
	defaultIncrementalQuery   string
	defaultWindowOverlapMS    int64

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

type SyncManagerArgs struct {
	Index              *search.MeiliIndex
	ProgressStore      *storage.JSONProgressStore
	MeiliHost          string
	MeiliIndex         string
	CheckpointTemplate string
	RootWorkers        int
	ProgressEvery      int
	Retry              models.RetryPolicyOptions
	MaxConcurrent      int
	MinTimeMS          int
	ActivityChecker    indexer.ActivityChecker
	SyncStateFile      string
	IncrementalQuery   string
	WindowOverlapMS    int64
}

func NewSyncManager(args SyncManagerArgs) *SyncManager {
	return &SyncManager{
		index:                     args.Index,
		progressStore:             args.ProgressStore,
		meiliHost:                 args.MeiliHost,
		meiliIndex:                args.MeiliIndex,
		defaultCheckpointTemplate: args.CheckpointTemplate,
		defaultRootWorkers:        args.RootWorkers,
		defaultProgressEvery:      args.ProgressEvery,
		retry:                     args.Retry,
		maxConcurrent:             args.MaxConcurrent,
		minTimeMS:                 args.MinTimeMS,
		activityChecker:           args.ActivityChecker,
		syncStateFile:             args.SyncStateFile,
		defaultIncrementalQuery:   args.IncrementalQuery,
		defaultWindowOverlapMS:    args.WindowOverlapMS,
	}
}

func (m *SyncManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *SyncManager) GetProgress() (*models.SyncProgressState, error) {
	return m.progressStore.Load()
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

func (m *SyncManager) Start(api npan.API, request SyncStartRequest) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("已有全量同步任务在运行")
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.running = true
	m.cancel = cancel
	m.mu.Unlock()

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

func resolveMode(mode models.SyncMode, state *models.SyncState) models.SyncMode {
	switch mode {
	case models.SyncModeFull:
		return models.SyncModeFull
	case models.SyncModeIncremental:
		return models.SyncModeIncremental
	default:
		if state != nil && state.LastSyncTime > 0 {
			return models.SyncModeIncremental
		}
		return models.SyncModeFull
	}
}

func (m *SyncManager) discoverRootFolders(ctx context.Context, api npan.API, request SyncStartRequest) ([]int64, map[int64]int64, map[int64]string, error) {
	roots := append([]int64{}, request.RootFolderIDs...)
	rootEstimateMap := map[int64]int64{}
	rootNameMap := map[int64]string{}

	includeDepartments := request.IncludeDepartments == nil || *request.IncludeDepartments
	if includeDepartments {
		departmentIDs := append([]int64{}, request.DepartmentIDs...)
		if len(departmentIDs) == 0 {
			deps, err := api.ListUserDepartments(ctx)
			if err != nil {
				return nil, nil, nil, err
			}
			for _, dep := range deps {
				departmentIDs = append(departmentIDs, dep.ID)
			}
		}

		for _, departmentID := range departmentIDs {
			folders, err := api.ListDepartmentFolders(ctx, departmentID)
			if err != nil {
				return nil, nil, nil, err
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

func (m *SyncManager) runSingleRoot(ctx context.Context, api npan.API, progress *models.SyncProgressState, progressMu *sync.Mutex, rootID int64, checkpointFile string, progressEvery int, limiter *indexer.RequestLimiter) error {
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
	rp.UpdatedAt = now
	progress.ActiveRoot = &rootID
	updateAggregateFromRoots(progress)
	if err := m.progressStore.Save(progress); err != nil {
		progressMu.Unlock()
		return err
	}
	progressMu.Unlock()

	checkpointStore := storage.NewJSONCheckpointStore(checkpointFile)

	stats, err := indexer.RunFullCrawl(ctx, indexer.FullCrawlDeps{
		API:             api,
		IndexWriter:     &meiliIndexWriter{index: m.index},
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
			_ = m.progressStore.Save(progress)
			return err
		}

		rp.Status = "error"
		rp.Error = err.Error()
		rp.UpdatedAt = time.Now().UnixMilli()
		progress.Status = "error"
		progress.LastError = err.Error()
		updateAggregateFromRoots(progress)
		_ = m.progressStore.Save(progress)
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
		query = "*"
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
		return err
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
	syncStateStore := storage.NewJSONSyncStateStore(m.syncStateFile)
	syncState, _ := syncStateStore.Load()
	effectiveMode := resolveMode(request.Mode, syncState)

	if effectiveMode == models.SyncModeIncremental {
		return m.runIncrementalPath(ctx, api, request, syncState, syncStateStore)
	}

	// Full crawl path
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

			if err := m.runSingleRoot(runCtx, api, progress, progressMu, currentRoot, rootCheckpointMap[currentRoot], progressEvery, limiter); err != nil {
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
		_ = m.progressStore.Save(progress)
		return firstErr
	}

	if ctx.Err() != nil {
		progress.Status = "cancelled"
		progress.LastError = ctx.Err().Error()
		progress.ActiveRoot = nil
		updateAggregateFromRoots(progress)
		_ = m.progressStore.Save(progress)
		return ctx.Err()
	}

	progress.Status = "done"
	progress.LastError = ""
	progress.ActiveRoot = nil
	progress.Mode = string(models.SyncModeFull)
	updateAggregateFromRoots(progress)

	// Write cursor for future incremental runs
	if m.syncStateFile != "" {
		_ = syncStateStore.Save(&models.SyncState{LastSyncTime: time.Now().UnixMilli()})
	}

	meiliCount, err := m.index.DocumentCount(ctx)
	if err == nil {
		progress.Verification = buildVerification(meiliCount, progress.AggregateStats)
	}

	return m.progressStore.Save(progress)
}

func (m *SyncManager) runIncrementalPath(ctx context.Context, api npan.API, request SyncStartRequest, syncState *models.SyncState, syncStateStore *storage.JSONSyncStateStore) error {
	now := time.Now().UnixMilli()

	cursorBefore := int64(0)
	if syncState != nil && syncState.LastSyncTime > 0 {
		cursorBefore = syncState.LastSyncTime
		if cursorBefore >= 1_000_000_000_000 {
			cursorBefore = cursorBefore / 1000
		}
	}

	progress := &models.SyncProgressState{
		Status:             "running",
		Mode:               string(models.SyncModeIncremental),
		StartedAt:          now,
		UpdatedAt:          now,
		MeiliHost:          m.meiliHost,
		MeiliIndex:         m.meiliIndex,
		Roots:              []int64{},
		CompletedRoots:     []int64{},
		RootProgress:       map[string]*models.RootSyncProgress{},
		AggregateStats:     models.CrawlStats{StartedAt: now, EndedAt: now},
		IncrementalStats:   &models.IncrementalSyncStats{CursorBefore: cursorBefore},
	}

	if err := m.progressStore.Save(progress); err != nil {
		return err
	}

	limiter := indexer.NewRequestLimiter(m.maxConcurrent, m.minTimeMS)
	if m.activityChecker != nil {
		limiter.SetActivityChecker(m.activityChecker)
	}

	err := m.runIncremental(ctx, api, progress, request, limiter)

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

		if m.syncStateFile != "" && syncStateStore != nil {
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
	return m.progressStore.Save(progress)
}

type meiliIndexWriter struct {
	index *search.MeiliIndex
}

func (w *meiliIndexWriter) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
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
			fmt.Sprintf("MeiliSearch 文档数(%d) < 爬取写入数(%d)", meiliCount, crawled))
	}
	if discovered > 0 && crawled < discovered {
		v.Warnings = append(v.Warnings,
			fmt.Sprintf("已索引(%d) < 已发现(%d), 跳过(%d)", crawled, discovered, stats.SkippedFiles))
	}

	return v
}
