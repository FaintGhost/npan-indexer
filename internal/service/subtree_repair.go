package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"npan/internal/indexer"
	"npan/internal/models"
	"npan/internal/npan"
	"npan/internal/search"
)

const subtreeRepairPageSize int64 = 1000

type subtreeRepairMode string

const (
	subtreeRepairModeBackfill subtreeRepairMode = "backfill"
	subtreeRepairModeRebuild  subtreeRepairMode = "rebuild"
)

type subtreeRepairTarget struct {
	folder       models.NpanFolder
	localDocs    int64
	expectedDocs int64
	mode         subtreeRepairMode
}

type indexedTreeSnapshot struct {
	rootID int64

	rootDocID        string
	folderDocIDs     map[int64]string
	childFolders     map[int64][]int64
	directFileDocIDs map[int64][]string

	subtreeFolderCounts map[int64]int64
	subtreeFileCounts   map[int64]int64
}

func buildIndexedTreeSnapshot(ctx context.Context, idx search.IndexOperator, rootID int64) (*indexedTreeSnapshot, error) {
	snapshot := &indexedTreeSnapshot{
		rootID:              rootID,
		folderDocIDs:        map[int64]string{},
		childFolders:        map[int64][]int64{},
		directFileDocIDs:    map[int64][]string{},
		subtreeFolderCounts: map[int64]int64{},
		subtreeFileCounts:   map[int64]int64{},
	}

	queue := []int64{rootID}
	enqueued := map[int64]struct{}{rootID: {}}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]

		children, err := searchIndexedChildren(ctx, idx, parentID)
		if err != nil {
			return nil, err
		}

		for _, doc := range children {
			if doc.Type == models.ItemTypeFolder {
				if parentID == rootID && doc.SourceID == rootID {
					snapshot.rootDocID = doc.DocID
					continue
				}
				snapshot.folderDocIDs[doc.SourceID] = doc.DocID
				snapshot.childFolders[parentID] = append(snapshot.childFolders[parentID], doc.SourceID)
				if _, ok := enqueued[doc.SourceID]; !ok {
					enqueued[doc.SourceID] = struct{}{}
					queue = append(queue, doc.SourceID)
				}
				continue
			}
			snapshot.directFileDocIDs[parentID] = append(snapshot.directFileDocIDs[parentID], doc.DocID)
		}
	}

	snapshot.computeSubtreeCounts(rootID)
	return snapshot, nil
}

func searchIndexedChildren(ctx context.Context, idx search.IndexOperator, parentID int64) ([]models.IndexDocument, error) {
	var result []models.IndexDocument
	var page int64 = 1
	for {
		currentParent := parentID
		items, total, err := idx.Search(models.LocalSearchParams{
			Query:    "*",
			Page:     page,
			PageSize: subtreeRepairPageSize,
			ParentID: &currentParent,
		})
		if err != nil {
			return nil, fmt.Errorf("search indexed children for folder %d page %d: %w", parentID, page, err)
		}
		result = append(result, items...)
		if int64(len(result)) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return result, nil
}

func (s *indexedTreeSnapshot) computeSubtreeCounts(folderID int64) (int64, int64) {
	if folderCount, ok := s.subtreeFolderCounts[folderID]; ok {
		return folderCount, s.subtreeFileCounts[folderID]
	}

	var folderCount int64
	if folderID == s.rootID {
		if s.rootDocID != "" {
			folderCount = 1
		}
	} else if s.folderDocIDs[folderID] != "" {
		folderCount = 1
	}

	fileCount := int64(len(s.directFileDocIDs[folderID]))
	for _, childID := range s.childFolders[folderID] {
		childFolderCount, childFileCount := s.computeSubtreeCounts(childID)
		folderCount += childFolderCount
		fileCount += childFileCount
	}

	s.subtreeFolderCounts[folderID] = folderCount
	s.subtreeFileCounts[folderID] = fileCount
	return folderCount, fileCount
}

func (s *indexedTreeSnapshot) subtreeDocCount(folderID int64) int64 {
	folderCount, fileCount := s.computeSubtreeCounts(folderID)
	return folderCount + fileCount
}

func (s *indexedTreeSnapshot) directChildCount(folderID int64) int64 {
	return int64(len(s.directFileDocIDs[folderID]) + len(s.childFolders[folderID]))
}

func (s *indexedTreeSnapshot) collectSubtreeDocIDs(folderID int64) []string {
	docIDs := make([]string, 0, 1+len(s.directFileDocIDs[folderID]))
	if folderID == s.rootID {
		if s.rootDocID != "" {
			docIDs = append(docIDs, s.rootDocID)
		}
	} else if docID := s.folderDocIDs[folderID]; docID != "" {
		docIDs = append(docIDs, docID)
	}
	docIDs = append(docIDs, s.directFileDocIDs[folderID]...)
	for _, childID := range s.childFolders[folderID] {
		docIDs = append(docIDs, s.collectSubtreeDocIDs(childID)...)
	}
	return docIDs
}

func (m *SyncManager) fetchFolderInfo(ctx context.Context, api npan.API, folderID int64, limiter *indexer.RequestLimiter) (models.NpanFolder, error) {
	var folder models.NpanFolder
	err := limiter.Schedule(ctx, func() error {
		var innerErr error
		folder, innerErr = api.GetFolderInfo(ctx, folderID)
		return innerErr
	})
	return folder, err
}

func (m *SyncManager) listAllFolderChildren(ctx context.Context, api npan.API, folderID int64, limiter *indexer.RequestLimiter) ([]models.NpanFolder, int64, error) {
	var (
		pageID       int64
		directCount  int64
		childFolders []models.NpanFolder
	)

	for {
		var page models.FolderChildrenPage
		err := limiter.Schedule(ctx, func() error {
			var innerErr error
			page, innerErr = api.ListFolderChildren(ctx, folderID, pageID)
			return innerErr
		})
		if err != nil {
			return nil, 0, err
		}

		directCount += int64(len(page.Files) + len(page.Folders))
		childFolders = append(childFolders, page.Folders...)

		pageCount := page.PageCount
		if pageCount <= 0 {
			pageCount = 1
		}
		pageID++
		if pageID >= pageCount {
			break
		}
	}

	return childFolders, directCount, nil
}

func classifySubtreeRepair(localDocs int64, expectedDocs int64) subtreeRepairMode {
	if localDocs < expectedDocs {
		return subtreeRepairModeBackfill
	}
	return subtreeRepairModeRebuild
}

func (m *SyncManager) collectRepairTargets(ctx context.Context, api npan.API, snapshot *indexedTreeSnapshot, folder models.NpanFolder, limiter *indexer.RequestLimiter) ([]subtreeRepairTarget, error) {
	childFolders, liveDirectCount, err := m.listAllFolderChildren(ctx, api, folder.ID, limiter)
	if err != nil {
		return nil, fmt.Errorf("list live children for folder %d: %w", folder.ID, err)
	}

	localDocs := snapshot.subtreeDocCount(folder.ID)
	expectedDocs := folder.ItemCount + 1

	if snapshot.directChildCount(folder.ID) != liveDirectCount {
		return []subtreeRepairTarget{{
			folder:       folder,
			localDocs:    localDocs,
			expectedDocs: expectedDocs,
			mode:         classifySubtreeRepair(localDocs, expectedDocs),
		}}, nil
	}

	var targets []subtreeRepairTarget
	for _, child := range childFolders {
		childTargets, err := m.collectRepairTargets(ctx, api, snapshot, child, limiter)
		if err != nil {
			return nil, err
		}
		targets = append(targets, childTargets...)
	}
	if len(targets) > 0 {
		return targets, nil
	}

	if localDocs != expectedDocs {
		return []subtreeRepairTarget{{
			folder:       folder,
			localDocs:    localDocs,
			expectedDocs: expectedDocs,
			mode:         classifySubtreeRepair(localDocs, expectedDocs),
		}}, nil
	}
	return nil, nil
}

func (m *SyncManager) deleteSubtreeDocuments(ctx context.Context, docIDs []string) error {
	if len(docIDs) == 0 {
		return nil
	}
	return indexer.WithRetryVoid(ctx, func() error {
		return m.index.DeleteDocuments(ctx, docIDs)
	}, m.retry)
}

func (m *SyncManager) rebuildNestedFolderSubtree(ctx context.Context, api npan.API, folder models.NpanFolder, limiter *indexer.RequestLimiter) error {
	rootDoc := search.MapFolderToIndexDoc(folder, fmt.Sprintf("folder/%d/%s", folder.ID, folder.Name))
	if err := indexer.WithRetryVoid(ctx, func() error {
		return m.index.UpsertDocuments(ctx, []models.IndexDocument{rootDoc})
	}, m.retry); err != nil {
		return fmt.Errorf("upsert subtree root folder %d: %w", folder.ID, err)
	}

	queue := []int64{folder.ID}
	for len(queue) > 0 {
		currentFolderID := queue[0]
		queue = queue[1:]

		var pageID int64
		for {
			var page models.FolderChildrenPage
			err := limiter.Schedule(ctx, func() error {
				var innerErr error
				page, innerErr = api.ListFolderChildren(ctx, currentFolderID, pageID)
				return innerErr
			})
			if err != nil {
				return fmt.Errorf("list children for repair folder %d page %d: %w", currentFolderID, pageID, err)
			}

			docs := make([]models.IndexDocument, 0, len(page.Folders)+len(page.Files))
			for _, childFolder := range page.Folders {
				queue = append(queue, childFolder.ID)
				docs = append(docs, search.MapFolderToIndexDoc(childFolder, fmt.Sprintf("folder/%d/%s", childFolder.ID, childFolder.Name)))
			}
			for _, file := range page.Files {
				docs = append(docs, search.MapFileToIndexDoc(file, fmt.Sprintf("file/%d/%s", file.ID, file.Name)))
			}

			if len(docs) > 0 {
				if err := indexer.WithRetryVoid(ctx, func() error {
					return m.index.UpsertDocuments(ctx, docs)
				}, m.retry); err != nil {
					return fmt.Errorf("upsert subtree docs for folder %d page %d: %w", currentFolderID, pageID, err)
				}
			}

			pageCount := page.PageCount
			if pageCount <= 0 {
				pageCount = 1
			}
			pageID++
			if pageID >= pageCount {
				break
			}
		}
	}

	return nil
}

func refreshRootProgressFromSnapshot(progress *models.SyncProgressState, rootID int64, expectedDocs int64, snapshot *indexedTreeSnapshot) {
	if progress == nil || snapshot == nil {
		return
	}
	root := progress.RootProgress[fmt.Sprintf("%d", rootID)]
	if root == nil {
		return
	}

	root.Stats.FoldersVisited = snapshot.subtreeFolderCounts[rootID]
	root.Stats.FilesIndexed = snapshot.subtreeFileCounts[rootID]
	root.Stats.EndedAt = time.Now().UnixMilli()
	root.UpdatedAt = time.Now().UnixMilli()
	if expectedDocs > 0 {
		expectedCopy := expectedDocs
		root.EstimatedTotalDocs = &expectedCopy
	}
	updateAggregateFromRoots(progress)
}

func (m *SyncManager) runIncrementalRepairs(ctx context.Context, api npan.API, progress *models.SyncProgressState, request SyncStartRequest, limiter *indexer.RequestLimiter) error {
	if progress == nil || progress.IncrementalStats == nil || progress.IncrementalStats.ChangesFetched > 0 {
		return nil
	}

	progressEvery := request.ProgressEvery
	if progressEvery <= 0 {
		progressEvery = m.defaultProgressEvery
	}
	if progressEvery <= 0 {
		progressEvery = 1
	}

	progress.Status = "running"
	progress.LastError = ""
	progress.UpdatedAt = time.Now().UnixMilli()
	if err := m.progressStore.Save(progress); err != nil {
		return err
	}

	progressMu := &sync.Mutex{}
	for _, rootID := range progress.Roots {
		rootInfo, err := m.fetchFolderInfo(ctx, api, rootID, limiter)
		if err != nil {
			slog.Warn("获取根目录详情失败，跳过目录级补偿", "root_id", rootID, "error", err)
			continue
		}

		snapshot, err := buildIndexedTreeSnapshot(ctx, m.index, rootID)
		if err != nil {
			return err
		}

		targets, err := m.collectRepairTargets(ctx, api, snapshot, rootInfo, limiter)
		if err != nil {
			return err
		}

		for _, target := range targets {
			if target.mode == subtreeRepairModeRebuild {
				docIDs := snapshot.collectSubtreeDocIDs(target.folder.ID)
				if err := m.deleteSubtreeDocuments(ctx, docIDs); err != nil {
					return fmt.Errorf("delete subtree docs for folder %d: %w", target.folder.ID, err)
				}
			}

			if target.folder.ID == rootID {
				root := progress.RootProgress[fmt.Sprintf("%d", rootID)]
				if root == nil {
					continue
				}
				checkpointFile := root.CheckpointFile
				if checkpointFile == "" {
					checkpointFile = buildCheckpointFilePath(m.defaultCheckpointTemplate, rootID, len(progress.Roots) > 1)
					root.CheckpointFile = checkpointFile
				}
				checkpointStore := m.effectiveCheckpointStoreFactory().ForKey(checkpointFile)
				if err := checkpointStore.Clear(); err != nil {
					return fmt.Errorf("clear repair checkpoint for root %d: %w", rootID, err)
				}
				if err := m.runSingleRoot(ctx, api, progress, progressMu, rootID, checkpointFile, progressEvery, limiter, true); err != nil {
					return err
				}
				continue
			}

			if err := m.rebuildNestedFolderSubtree(ctx, api, target.folder, limiter); err != nil {
				return err
			}
		}

		updatedSnapshot, err := buildIndexedTreeSnapshot(ctx, m.index, rootID)
		if err != nil {
			return err
		}
		refreshRootProgressFromSnapshot(progress, rootID, rootInfo.ItemCount+1, updatedSnapshot)
	}

	return nil
}
