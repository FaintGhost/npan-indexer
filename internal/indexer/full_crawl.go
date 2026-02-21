package indexer

import (
	"context"
	"fmt"
	"time"

	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
)

type IndexWriter interface {
	UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error
}

type CheckpointStore interface {
	Load() (*models.CrawlCheckpoint, error)
	Save(checkpoint *models.CrawlCheckpoint) error
	Clear() error
}

type FullCrawlDeps struct {
	API             npan.API
	IndexWriter     IndexWriter
	Limiter         *RequestLimiter
	CheckpointStore CheckpointStore
	RootFolderID    int64
	Retry           models.RetryPolicyOptions
	OnProgress      func(event ProgressEvent)
}

type ProgressEvent struct {
	RootFolderID     int64
	CurrentFolderID  int64
	CurrentPageID    int64
	CurrentPageCount int64
	QueueLength      int64
	Stats            models.CrawlStats
}

func defaultCheckpoint(rootFolderID int64) *models.CrawlCheckpoint {
	return &models.CrawlCheckpoint{
		Queue: []int64{rootFolderID},
	}
}

func RunFullCrawl(ctx context.Context, deps FullCrawlDeps) (models.CrawlStats, error) {
	startedAt := time.Now().UnixMilli()
	stats := models.CrawlStats{
		FoldersVisited: 0,
		FilesIndexed:   0,
		PagesFetched:   0,
		FailedRequests: 0,
		StartedAt:      startedAt,
		EndedAt:        startedAt,
	}

	state, err := deps.CheckpointStore.Load()
	if err != nil {
		return stats, err
	}
	if state == nil {
		state = defaultCheckpoint(deps.RootFolderID)
	}

	queue := append([]int64{}, state.Queue...)

	for len(queue) > 0 {
		folderID := queue[0]
		queue = queue[1:]
		stats.FoldersVisited++

		pageID := int64(0)
		pageCount := int64(1)

		for pageID < pageCount {
			checkpoint := &models.CrawlCheckpoint{
				Queue: []int64{folderID},
			}
			checkpoint.Queue = append(checkpoint.Queue, queue...)
			checkpoint.CurrentFolderID = &folderID
			checkpoint.CurrentPageID = &pageID

			if err := deps.CheckpointStore.Save(checkpoint); err != nil {
				return stats, err
			}

			var page models.FolderChildrenPage
			err := deps.Limiter.Schedule(ctx, func() error {
				result, requestErr := WithRetry(ctx, func() (models.FolderChildrenPage, error) {
					return deps.API.ListFolderChildren(ctx, folderID, pageID)
				}, deps.Retry)
				if requestErr != nil {
					return requestErr
				}
				page = result
				return nil
			})
			if err != nil {
				stats.FailedRequests++
				return stats, err
			}

			stats.PagesFetched++
			pageCount = page.PageCount
			if pageCount <= 0 {
				pageCount = 1
			}

			for _, folder := range page.Folders {
				queue = append(queue, folder.ID)
			}

			docs := make([]models.IndexDocument, 0, len(page.Folders)+len(page.Files)+1)
			if folderID == deps.RootFolderID && pageID == 0 {
				docs = append(docs, search.MapFolderToIndexDoc(models.NpanFolder{
					ID:       deps.RootFolderID,
					Name:     "全部文件",
					ParentID: deps.RootFolderID,
				}, "全部文件"))
			}

			for _, folder := range page.Folders {
				docs = append(docs, search.MapFolderToIndexDoc(folder, fmt.Sprintf("folder/%d/%s", folder.ID, folder.Name)))
			}
			for _, file := range page.Files {
				docs = append(docs, search.MapFileToIndexDoc(file, fmt.Sprintf("file/%d/%s", file.ID, file.Name)))
			}

			stats.FilesDiscovered += int64(len(page.Files))
			filesInBatch := int64(len(page.Files))
			if len(docs) > 0 {
				err := WithRetryVoid(ctx, func() error {
					return deps.IndexWriter.UpsertDocuments(ctx, docs)
				}, deps.Retry)
				if err != nil {
					stats.FailedRequests++
					stats.SkippedFiles += filesInBatch
					if ctx.Err() != nil {
						return stats, ctx.Err()
					}
				} else {
					stats.FilesIndexed += filesInBatch
				}
			} else {
				stats.FilesIndexed += filesInBatch
			}

			if deps.OnProgress != nil {
				deps.OnProgress(ProgressEvent{
					RootFolderID:     deps.RootFolderID,
					CurrentFolderID:  folderID,
					CurrentPageID:    pageID,
					CurrentPageCount: pageCount,
					QueueLength:      int64(len(queue)),
					Stats:            stats,
				})
			}

			pageID++
		}

		if err := deps.CheckpointStore.Save(&models.CrawlCheckpoint{Queue: append([]int64{}, queue...)}); err != nil {
			return stats, err
		}
	}

	if err := deps.CheckpointStore.Clear(); err != nil {
		return stats, err
	}

	stats.EndedAt = time.Now().UnixMilli()
	return stats, nil
}
