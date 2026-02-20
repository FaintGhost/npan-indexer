package indexer

import (
	"context"
	"time"

	"npan/internal/models"
)

func normalizeUnixSeconds(ts int64) int64 {
	if ts <= 0 {
		return 0
	}
	// 兼容历史毫秒游标，统一转为秒级语义。
	if ts >= 1_000_000_000_000 {
		return ts / 1000
	}
	return ts
}

type IncrementalInputItem struct {
	Doc     models.IndexDocument
	Deleted bool
}

type SyncStateStore interface {
	Load() (*models.SyncState, error)
	Save(state *models.SyncState) error
}

type IncrementalDeps struct {
	FetchChanges    func(ctx context.Context, since int64) ([]IncrementalInputItem, error)
	StateStore      SyncStateStore
	UpsertDocuments func(ctx context.Context, docs []models.IndexDocument) error
	DeleteDocuments func(ctx context.Context, docIDs []string) error
	NowProvider     func() int64
}

func RunIncrementalSync(ctx context.Context, deps IncrementalDeps) error {
	nowProvider := deps.NowProvider
	if nowProvider == nil {
		nowProvider = func() int64 { return time.Now().UnixMilli() }
	}

	state, err := deps.StateStore.Load()
	if err != nil {
		return err
	}

	since := int64(0)
	if state != nil {
		since = normalizeUnixSeconds(state.LastSyncTime)
	}

	changes, err := deps.FetchChanges(ctx, since)
	if err != nil {
		return err
	}

	upserts := make([]models.IndexDocument, 0, len(changes))
	deletes := make([]string, 0, len(changes))
	for _, item := range changes {
		if item.Deleted {
			deletes = append(deletes, item.Doc.DocID)
		} else {
			upserts = append(upserts, item.Doc)
		}
	}

	if len(upserts) > 0 {
		if err := deps.UpsertDocuments(ctx, upserts); err != nil {
			return err
		}
	}
	if len(deletes) > 0 {
		if err := deps.DeleteDocuments(ctx, deletes); err != nil {
			return err
		}
	}

	return deps.StateStore.Save(&models.SyncState{LastSyncTime: normalizeUnixSeconds(nowProvider())})
}
