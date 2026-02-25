package httpx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"

	"google.golang.org/protobuf/types/known/timestamppb"
	npanv1 "npan/gen/go/npan/v1"
	"npan/internal/models"
	"npan/internal/service"
)

type adminConnectServer struct {
	handlers *Handlers
}

var watchSyncProgressPollInterval = 2 * time.Second
var inspectRootsMaxConcurrency = 6
var inspectRootsPerFolderTimeout = 10 * time.Second

func newAdminConnectServer(handlers *Handlers) *adminConnectServer {
	return &adminConnectServer{handlers: handlers}
}

func (s *adminConnectServer) StartSync(ctx context.Context, req *connect.Request[npanv1.StartSyncRequest]) (*connect.Response[npanv1.StartSyncResponse], error) {
	if s.handlers == nil || s.handlers.syncManager == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("同步服务未初始化"))
	}

	if force := req.Msg.ForceRebuild != nil && req.Msg.GetForceRebuild(); force && len(req.Msg.GetRootFolderIds()) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("force_rebuild 仅允许全量全库执行"))
	}

	token, authOptions, err := s.handlers.resolveTokenForConnect(ctx, req.Header(), authPayload{}, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("启动同步失败"))
	}
	api := s.handlers.newAPIClient(token, authOptions)

	checkpointTemplate := req.Msg.GetCheckpointTemplate()
	if checkpointTemplate != "" {
		if validateErr := validateCheckpointTemplate(checkpointTemplate); validateErr != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, validateErr)
		}
	}

	rootWorkers, err := optionalInt64ToInt(req.Msg.RootWorkers)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	progressEvery, err := optionalInt64ToInt(req.Msg.ProgressEvery)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	startErr := s.handlers.syncManager.Start(api, service.SyncStartRequest{
		Mode:                fromProtoSyncMode(req.Msg.Mode),
		RootFolderIDs:       req.Msg.GetRootFolderIds(),
		IncludeDepartments:  req.Msg.IncludeDepartments,
		PreserveRootCatalog: req.Msg.PreserveRootCatalog,
		DepartmentIDs:       req.Msg.GetDepartmentIds(),
		ResumeProgress:      req.Msg.ResumeProgress,
		ForceRebuild:        req.Msg.ForceRebuild,
		RootWorkers:         rootWorkers,
		ProgressEvery:       progressEvery,
		CheckpointTemplate:  checkpointTemplate,
		WindowOverlapMS:     req.Msg.GetWindowOverlapMs(),
		IncrementalQuery:    req.Msg.GetIncrementalQuery(),
	})
	if startErr != nil {
		return nil, connect.NewError(connect.CodeAborted, errors.New("启动同步失败"))
	}

	return connect.NewResponse(&npanv1.StartSyncResponse{
		Message: "同步任务已启动",
	}), nil
}

func (s *adminConnectServer) InspectRoots(ctx context.Context, req *connect.Request[npanv1.InspectRootsRequest]) (*connect.Response[npanv1.InspectRootsResponse], error) {
	if s.handlers == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("服务未初始化"))
	}

	if len(req.Msg.GetFolderIds()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("folder_ids 不能为空"))
	}
	folderIDs, err := normalizePositiveIDs(req.Msg.GetFolderIds())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("folder_ids 必须是正整数数组"))
	}

	token, authOptions, err := s.handlers.resolveTokenForConnect(ctx, req.Header(), authPayload{}, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("拉取目录详情失败"))
	}

	api := s.handlers.newAPIClient(token, authOptions)
	type inspectJob struct {
		index    int
		folderID int64
	}
	type inspectResult struct {
		index int
		item  *npanv1.InspectRootItem
		err   *npanv1.InspectRootError
	}

	workers := s.handlers.inspectRootsMaxConcurrency
	if workers <= 0 {
		workers = inspectRootsMaxConcurrency
	}
	if workers <= 0 {
		workers = 1
	}
	if workers > len(folderIDs) {
		workers = len(folderIDs)
	}

	jobs := make(chan inspectJob, len(folderIDs))
	results := make(chan inspectResult, len(folderIDs))

	worker := func() {
		for job := range jobs {
			folderCtx := ctx
			cancel := func() {}
			perFolderTimeout := s.handlers.inspectRootsPerFolderTimeout
			if perFolderTimeout <= 0 {
				perFolderTimeout = inspectRootsPerFolderTimeout
			}
			if perFolderTimeout > 0 {
				folderCtx, cancel = context.WithTimeout(ctx, perFolderTimeout)
			}

			folder, infoErr := api.GetFolderInfo(folderCtx, job.folderID)
			cancel()

			if infoErr != nil {
				results <- inspectResult{
					index: job.index,
					err: &npanv1.InspectRootError{
						FolderId: job.folderID,
						Message:  "获取目录信息失败",
					},
				}
				continue
			}

			estimate := folder.ItemCount + 1
			if estimate < 0 {
				estimate = 0
			}

			results <- inspectResult{
				index: job.index,
				item: &npanv1.InspectRootItem{
					FolderId:           job.folderID,
					Name:               folder.Name,
					ItemCount:          folder.ItemCount,
					EstimatedTotalDocs: estimate,
				},
			}
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker()
		}()
	}

	for index, folderID := range folderIDs {
		jobs <- inspectJob{index: index, folderID: folderID}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	orderedItems := make([]*npanv1.InspectRootItem, len(folderIDs))
	orderedErrors := make([]*npanv1.InspectRootError, len(folderIDs))
	for result := range results {
		if result.item != nil {
			orderedItems[result.index] = result.item
			continue
		}
		orderedErrors[result.index] = result.err
	}

	resp := &npanv1.InspectRootsResponse{
		Items: make([]*npanv1.InspectRootItem, 0, len(folderIDs)),
	}
	for i := range folderIDs {
		if orderedItems[i] != nil {
			resp.Items = append(resp.Items, orderedItems[i])
		}
		if orderedErrors[i] != nil {
			resp.Errors = append(resp.Errors, orderedErrors[i])
		}
	}

	return connect.NewResponse(resp), nil
}

func (s *adminConnectServer) GetIndexStats(ctx context.Context, _ *connect.Request[npanv1.GetIndexStatsRequest]) (*connect.Response[npanv1.GetIndexStatsResponse], error) {
	if s.handlers == nil || s.handlers.syncManager == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("同步服务未初始化"))
	}

	count, err := s.handlers.syncManager.GetIndexDocumentCount(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("无法读取索引状态"))
	}

	return connect.NewResponse(&npanv1.GetIndexStatsResponse{DocumentCount: count}), nil
}

func (s *adminConnectServer) GetSyncProgress(_ context.Context, _ *connect.Request[npanv1.GetSyncProgressRequest]) (*connect.Response[npanv1.GetSyncProgressResponse], error) {
	if s.handlers == nil || s.handlers.syncManager == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("同步服务未初始化"))
	}

	progress, err := s.handlers.syncManager.GetProgress()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("无法读取同步进度"))
	}
	if progress == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("未找到同步进度"))
	}

	return connect.NewResponse(&npanv1.GetSyncProgressResponse{
		State: toProtoSyncProgressState(progress),
	}), nil
}

func (s *adminConnectServer) WatchSyncProgress(
	ctx context.Context,
	_ *connect.Request[npanv1.WatchSyncProgressRequest],
	stream *connect.ServerStream[npanv1.WatchSyncProgressResponse],
) error {
	if s.handlers == nil || s.handlers.syncManager == nil {
		return connect.NewError(connect.CodeInternal, errors.New("同步服务未初始化"))
	}

	sendProgress := func() (bool, error) {
		progress, err := s.handlers.syncManager.GetProgress()
		if err != nil {
			return false, connect.NewError(connect.CodeInternal, errors.New("无法读取同步进度"))
		}
		if progress == nil {
			return false, nil
		}
		if err := stream.Send(&npanv1.WatchSyncProgressResponse{
			State: toProtoSyncProgressState(progress),
		}); err != nil {
			return false, err
		}
		return isSyncTerminalStatus(progress.Status), nil
	}

	done, err := sendProgress()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	if done {
		return nil
	}

	ticker := time.NewTicker(watchSyncProgressPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		case <-ticker.C:
			done, err := sendProgress()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
			if done {
				return nil
			}
		}
	}
}

func (s *adminConnectServer) CancelSync(_ context.Context, _ *connect.Request[npanv1.CancelSyncRequest]) (*connect.Response[npanv1.CancelSyncResponse], error) {
	if s.handlers == nil || s.handlers.syncManager == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("同步服务未初始化"))
	}

	if !s.handlers.syncManager.Cancel() {
		return nil, connect.NewError(connect.CodeAborted, errors.New("当前没有运行中的同步任务"))
	}

	return connect.NewResponse(&npanv1.CancelSyncResponse{
		Message: "同步取消信号已发送",
	}), nil
}

func optionalInt64ToInt(v *int64) (int, error) {
	if v == nil {
		return 0, nil
	}
	maxInt := int64(^uint(0) >> 1)
	minInt := -maxInt - 1
	if *v > maxInt || *v < minInt {
		return 0, fmt.Errorf("数值超出 int 范围: %d", *v)
	}
	return int(*v), nil
}

func fromProtoSyncMode(mode *npanv1.SyncMode) models.SyncMode {
	if mode == nil {
		return ""
	}
	switch *mode {
	case npanv1.SyncMode_SYNC_MODE_FULL:
		return models.SyncModeFull
	case npanv1.SyncMode_SYNC_MODE_INCREMENTAL:
		return models.SyncModeIncremental
	default:
		return models.SyncMode(fmt.Sprintf("invalid(%d)", int32(*mode)))
	}
}

func toProtoSyncProgressState(state *models.SyncProgressState) *npanv1.SyncProgressState {
	if state == nil {
		return nil
	}

	resp := &npanv1.SyncProgressState{
		Status:              toProtoSyncStatus(state.Status),
		StartedAt:           state.StartedAt,
		StartedAtTs:         millisToProtoTimestamp(state.StartedAt),
		UpdatedAt:           state.UpdatedAt,
		UpdatedAtTs:         millisToProtoTimestamp(state.UpdatedAt),
		Roots:               state.Roots,
		CompletedRoots:      state.CompletedRoots,
		ActiveRoot:          state.ActiveRoot,
		AggregateStats:      toProtoCrawlStats(state.AggregateStats),
		RootNames:           int64MapToStringKeyMap(state.RootNames),
		RootProgress:        toProtoRootProgressMap(state.RootProgress),
		CatalogRoots:        state.CatalogRoots,
		CatalogRootNames:    int64MapToStringKeyMap(state.CatalogRootNames),
		CatalogRootProgress: toProtoRootProgressMap(state.CatalogRootProgress),
	}

	if mode := toProtoSyncMode(state.Mode); mode != nil {
		resp.Mode = mode
	}
	if state.IncrementalStats != nil {
		resp.IncrementalStats = &npanv1.IncrementalSyncStats{
			ChangesFetched: state.IncrementalStats.ChangesFetched,
			Upserted:       state.IncrementalStats.Upserted,
			Deleted:        state.IncrementalStats.Deleted,
			SkippedUpserts: state.IncrementalStats.SkippedUpserts,
			SkippedDeletes: state.IncrementalStats.SkippedDeletes,
			CursorBefore:   state.IncrementalStats.CursorBefore,
			CursorAfter:    state.IncrementalStats.CursorAfter,
		}
	}
	if state.Verification != nil {
		resp.Verification = &npanv1.SyncVerification{
			MeiliDocCount:      state.Verification.MeiliDocCount,
			CrawledDocCount:    state.Verification.CrawledDocCount,
			DiscoveredDocCount: state.Verification.DiscoveredDocCount,
			SkippedCount:       state.Verification.SkippedCount,
			Verified:           state.Verification.Verified,
			Warnings:           state.Verification.Warnings,
		}
	}
	if lastError := toOptionalString(state.LastError); lastError != nil {
		resp.LastError = lastError
	}

	return resp
}

func toProtoSyncStatus(raw string) npanv1.SyncStatus {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "idle":
		return npanv1.SyncStatus_SYNC_STATUS_IDLE
	case "running":
		return npanv1.SyncStatus_SYNC_STATUS_RUNNING
	case "done":
		return npanv1.SyncStatus_SYNC_STATUS_DONE
	case "error":
		return npanv1.SyncStatus_SYNC_STATUS_ERROR
	case "cancelled":
		return npanv1.SyncStatus_SYNC_STATUS_CANCELLED
	case "interrupted":
		return npanv1.SyncStatus_SYNC_STATUS_INTERRUPTED
	default:
		return npanv1.SyncStatus_SYNC_STATUS_UNSPECIFIED
	}
}

func toProtoSyncMode(raw string) *npanv1.SyncMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "full":
		mode := npanv1.SyncMode_SYNC_MODE_FULL
		return &mode
	case "incremental":
		mode := npanv1.SyncMode_SYNC_MODE_INCREMENTAL
		return &mode
	default:
		return nil
	}
}

func isSyncTerminalStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "error", "cancelled", "interrupted":
		return true
	default:
		return false
	}
}

func toProtoCrawlStats(stats models.CrawlStats) *npanv1.CrawlStats {
	return &npanv1.CrawlStats{
		FoldersVisited:  stats.FoldersVisited,
		FilesIndexed:    stats.FilesIndexed,
		FilesDiscovered: stats.FilesDiscovered,
		SkippedFiles:    stats.SkippedFiles,
		PagesFetched:    stats.PagesFetched,
		FailedRequests:  stats.FailedRequests,
		StartedAt:       stats.StartedAt,
		StartedAtTs:     millisToProtoTimestamp(stats.StartedAt),
		EndedAt:         stats.EndedAt,
		EndedAtTs:       millisToProtoTimestamp(stats.EndedAt),
	}
}

func toProtoRootProgressMap(in map[string]*models.RootSyncProgress) map[string]*npanv1.RootSyncProgress {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]*npanv1.RootSyncProgress, len(in))
	for key, value := range in {
		if value == nil {
			continue
		}
		out[key] = &npanv1.RootSyncProgress{
			RootFolderId:       value.RootFolderID,
			Status:             value.Status,
			EstimatedTotalDocs: value.EstimatedTotalDocs,
			Stats:              toProtoCrawlStats(value.Stats),
			UpdatedAt:          value.UpdatedAt,
			UpdatedAtTs:        millisToProtoTimestamp(value.UpdatedAt),
		}
	}
	return out
}

func int64MapToStringKeyMap(in map[int64]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[strconv.FormatInt(key, 10)] = value
	}
	return out
}

func millisToProtoTimestamp(raw int64) *timestamppb.Timestamp {
	if raw <= 0 {
		return nil
	}
	return timestamppb.New(time.UnixMilli(raw))
}
