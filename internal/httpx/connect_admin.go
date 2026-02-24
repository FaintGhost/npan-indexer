package httpx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	resp := &npanv1.InspectRootsResponse{
		Items: make([]*npanv1.InspectRootItem, 0, len(folderIDs)),
	}
	for _, folderID := range folderIDs {
		folder, infoErr := api.GetFolderInfo(ctx, folderID)
		if infoErr != nil {
			resp.Errors = append(resp.Errors, &npanv1.InspectRootError{
				FolderId: folderID,
				Message:  "获取目录信息失败",
			})
			continue
		}
		estimate := folder.ItemCount + 1
		if estimate < 0 {
			estimate = 0
		}
		resp.Items = append(resp.Items, &npanv1.InspectRootItem{
			FolderId:           folderID,
			Name:               folder.Name,
			ItemCount:          folder.ItemCount,
			EstimatedTotalDocs: estimate,
		})
	}

	return connect.NewResponse(resp), nil
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
	case npanv1.SyncMode_SYNC_MODE_AUTO:
		return models.SyncModeAuto
	case npanv1.SyncMode_SYNC_MODE_FULL:
		return models.SyncModeFull
	case npanv1.SyncMode_SYNC_MODE_INCREMENTAL:
		return models.SyncModeIncremental
	default:
		return ""
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
	case "auto":
		mode := npanv1.SyncMode_SYNC_MODE_AUTO
		return &mode
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
