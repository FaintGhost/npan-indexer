# Architecture

## Overview

四层验证体系嵌入在现有的同步架构中，不改变整体结构，只在关键路径上增加计数器和验证逻辑。

```
┌─────────────────────────────────────────────────────────┐
│                    SyncManager.run()                     │
│                                                         │
│  ┌─────────────┐    ┌──────────────┐    ┌────────────┐ │
│  │ discover     │───>│ create/restore│───>│ runSingle  │ │
│  │ RootFolders  │    │ Progress     │    │ Root (×N)  │ │
│  └─────────────┘    └──────────────┘    └─────┬──────┘ │
│                                               │         │
│                                    ┌──────────▼───────┐ │
│                                    │ RunFullCrawl     │ │
│                                    │                  │ │
│                                    │ L1: Fix counter  │ │
│                                    │ L2: Track discovered│
│                                    │ L3: Retry+skip   │ │
│                                    └──────────────────┘ │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │ L4: Post-sync reconciliation                     │   │
│  │  MeiliIndex.DocumentCount() vs AggregateStats    │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Changed Files

### Backend

| File | Change Type | Description |
|------|------------|-------------|
| `internal/models/models.go` | Modify | CrawlStats 新增 FilesDiscovered, SkippedFiles; 新增 SyncVerification struct; SyncProgressState 新增 Verification 字段 |
| `internal/indexer/retry.go` | Modify | 新增 WithRetryVoid; isRetriable 扩展识别 *meilisearch.Error |
| `internal/indexer/full_crawl.go` | Modify | FilesIndexed 移到 upsert 后; 新增 FilesDiscovered 累加; upsert 包裹 retry + 失败跳过 |
| `internal/search/meili_index.go` | Modify | 新增 DocumentCount 方法 |
| `internal/service/sync_manager.go` | Modify | 完成时执行对账逻辑, 填充 Verification |

### Frontend

| File | Change Type | Description |
|------|------------|-------------|
| `web/src/lib/sync-schemas.ts` | Modify | CrawlStatsSchema 新增字段; SyncProgressSchema 新增 verification |
| `web/src/components/sync-progress-display.tsx` | Modify | 显示已发现/已跳过统计; 显示对账结果 |

### Tests

| File | Change Type | Description |
|------|------------|-------------|
| `internal/indexer/retry_test.go` | Modify | WithRetryVoid 测试; MeiliSearch 错误分类测试 |
| `internal/indexer/full_crawl_test.go` | New/Modify | 计数器准确性测试; 跳过行为测试 |
| `internal/service/sync_manager_estimate_test.go` | Modify | 对账逻辑测试 |
| `web/src/components/sync-progress-display.test.tsx` | Modify | 新统计卡片和对账结果 UI 测试 |
| `web/src/lib/sync-schemas.test.ts` | Modify | Schema 新字段验证 |

## Data Flow

### Crawl-time (L1 + L2 + L3)

```
ListFolderChildren(folderID, pageID)
  │
  ├── stats.FilesDiscovered += len(page.Files)     // L2: 无论 upsert 成败
  │
  ├── WithRetryVoid(UpsertDocuments(docs))          // L3: 重试
  │     │
  │     ├── success → stats.FilesIndexed += len     // L1: 只在成功后
  │     │
  │     └── fail (retry exhausted)
  │           ├── stats.SkippedFiles += len          // L3: 记录跳过
  │           ├── stats.FailedRequests++
  │           └── continue (不终止)                  // L3: 降级继续
  │
  └── OnProgress callback → save to JSON
```

### Post-sync (L4)

```
all roots done
  │
  ├── MeiliIndex.DocumentCount(ctx)
  │     └── GET /indexes/{uid}/stats → NumberOfDocuments
  │
  ├── compute CrawledDocCount = FilesIndexed + FoldersVisited
  ├── compute DiscoveredDocCount = FilesDiscovered + FoldersVisited
  │
  ├── compare & generate warnings
  │
  └── progress.Verification = &SyncVerification{...}
      └── save to progress JSON
```

## Backward Compatibility

- `CrawlStats` 新字段（`FilesDiscovered`, `SkippedFiles`）使用 JSON zero-value（`0`），旧数据加载时自动为 0
- `SyncProgressState.Verification` 使用 `omitempty` + pointer，旧数据加载时为 nil
- 前端 Zod schema 使用 `.optional().default(0)` / `.optional().nullable()`，兼容旧数据
- MeiliSearch `GetStats` 错误不阻塞同步完成
