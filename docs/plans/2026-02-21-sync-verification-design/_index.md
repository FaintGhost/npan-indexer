# Sync Verification Design

## Context

npan 的全量同步将网盘 API 中的文件/文件夹爬取并索引到 MeiliSearch。当前缺乏验证机制来确认"所有文件都被成功索引"。

### 已识别的 4 个可靠性缺口

1. **`FilesIndexed` 计数不准确** — 在 `UpsertDocuments` 之前递增，失败时虚高
2. **无"已发现文件总数"** — API 返回的 `TotalCount` 未被使用，无参照基准
3. **Upsert 失败直接终止** — MeiliSearch 写入无重试，一次失败中止整个根目录
4. **无事后对账** — MeiliSearch `GetStats()` 仅用于 Ping，不做验证

## Requirements

- 同步完成后能从多个维度验证"是否索引了全部文件"
- 采用四层互相印证的策略，任何单层失效不影响其他层的验证能力
- 改动尽量限制在 `indexer` 和 `service` 包，不影响外部 API 接口契约
- 前端进度 UI 能展示新增的统计指标（已发现 vs 已索引、跳过数）
- 保持向后兼容，旧的进度 JSON 文件加载不会报错

## Rationale

四层验证体系的设计理念是 **defense in depth**：

| 层级 | 验证时机 | 验证内容 | 独立性 |
|------|---------|---------|--------|
| L1 精确计数 | 爬取中 | 移动 FilesIndexed 到 Upsert 后 | 修复已有指标 |
| L2 发现计数 | 爬取中 | 利用 TotalCount 记录应索引总数 | 新增 FilesDiscovered |
| L3 Upsert 容错 | 爬取中 | 重试 + 跳过失败 + 记录 SkippedFiles | 新增容错机制 |
| L4 事后对账 | 完成后 | 查 MeiliSearch 文档数 vs 爬取统计 | 外部验证 |

## Detailed Design

### Layer 1: Fix FilesIndexed Counter

**文件**: `internal/indexer/full_crawl.go`

将 `stats.FilesIndexed += int64(len(page.Files))` 从 L131（Upsert 前）移到 Upsert 成功后。同时 `stats.FoldersVisited` 不受影响（它在出队时递增，语义正确）。

```go
// Before (当前)
stats.FilesIndexed += int64(len(page.Files))  // L131
if len(docs) > 0 {
    if err := deps.IndexWriter.UpsertDocuments(ctx, docs); err != nil { ... }
}

// After (修改后)
filesInBatch := int64(len(page.Files))
if len(docs) > 0 {
    if err := ...; err != nil { ... }
}
stats.FilesIndexed += filesInBatch
```

### Layer 2: Track FilesDiscovered

**文件**: `internal/models/models.go`, `internal/indexer/full_crawl.go`

在 `CrawlStats` 新增 `FilesDiscovered int64`。

在 `RunFullCrawl` 中，对每个文件夹的**第一页**响应，读取 `page.TotalCount` 累加到 `stats.FilesDiscovered`。`TotalCount` 是网盘 API 返回的该文件夹子项总数（文件 + 子文件夹），但我们更精确的做法是：累加每页实际返回的 `len(page.Files) + len(page.Folders)` 作为 discovered（因为 `TotalCount` 含义可能包含回收站项目等）。

考虑到精确性，采用**每页累加方案**：

```go
stats.FilesDiscovered += int64(len(page.Files))
// FoldersDiscovered 用 FoldersVisited 等价，不单独新增
```

这样 `FilesDiscovered`（应索引数）vs `FilesIndexed`（已索引数）的差值就能反映丢失量。

### Layer 3: Upsert Retry + Skip on Failure

#### 3a. `WithRetryVoid` 函数

**文件**: `internal/indexer/retry.go`

新增无返回值版本的 retry 辅助函数：

```go
func WithRetryVoid(ctx context.Context, operation func() error, opts models.RetryPolicyOptions) error {
    _, err := WithRetry(ctx, func() (struct{}, error) {
        return struct{}{}, operation()
    }, opts)
    return err
}
```

#### 3b. 扩展 `isRetriable` 识别 MeiliSearch 错误

**文件**: `internal/indexer/retry.go`

当前 `isRetriable` 只识别 `*npan.StatusError`，对 `*meilisearch.Error` 是盲的。新增：

```go
var meiliErr *meilisearch.Error
if errors.As(err, &meiliErr) {
    switch meiliErr.ErrCode {
    case meilisearch.MeilisearchTimeoutError,
        meilisearch.MeilisearchCommunicationError:
        return true
    case meilisearch.MeilisearchApiError,
        meilisearch.MeilisearchApiErrorWithoutMessage:
        return meiliErr.StatusCode == 429 ||
            (meiliErr.StatusCode >= 500 && meiliErr.StatusCode <= 599)
    }
}
```

#### 3c. full_crawl.go 中 Upsert 包裹 retry + 失败降级

```go
if len(docs) > 0 {
    err := WithRetryVoid(ctx, func() error {
        return deps.IndexWriter.UpsertDocuments(ctx, docs)
    }, deps.Retry)
    if err != nil {
        stats.FailedRequests++
        stats.SkippedFiles += filesInBatch
        // 记录日志但不终止爬取
        if deps.OnProgress != nil { ... }
        // continue 到下一页，而非 return
    } else {
        stats.FilesIndexed += filesInBatch
    }
}
```

**`CrawlStats` 新增字段**: `SkippedFiles int64`

### Layer 4: Post-Sync Reconciliation

**文件**: `internal/search/meili_index.go`, `internal/service/sync_manager.go`, `internal/models/models.go`

#### 4a. MeiliIndex 新增 DocumentCount 方法

```go
func (m *MeiliIndex) DocumentCount(ctx context.Context) (int64, error) {
    stats, err := m.index.GetStatsWithContext(ctx)
    if err != nil {
        return 0, err
    }
    return stats.NumberOfDocuments, nil
}
```

#### 4b. SyncProgressState 新增对账字段

```go
type SyncVerification struct {
    MeiliDocCount       int64  `json:"meiliDocCount"`
    CrawledDocCount     int64  `json:"crawledDocCount"`     // FilesIndexed + FoldersVisited
    DiscoveredDocCount  int64  `json:"discoveredDocCount"`  // FilesDiscovered + FoldersVisited
    SkippedCount        int64  `json:"skippedCount"`
    Verified            bool   `json:"verified"`
    Warnings            []string `json:"warnings,omitempty"`
}
```

在 `SyncProgressState` 新增 `Verification *SyncVerification`。

#### 4c. 同步完成时执行对账

在 `sync_manager.go` 的 `run()` 函数末尾，当 `progress.Status = "done"` 时：

```go
meiliCount, err := m.index.DocumentCount(ctx)
if err == nil {
    crawled := progress.AggregateStats.FilesIndexed + progress.AggregateStats.FoldersVisited
    discovered := progress.AggregateStats.FilesDiscovered + progress.AggregateStats.FoldersVisited
    verification := &models.SyncVerification{
        MeiliDocCount:      meiliCount,
        CrawledDocCount:    crawled,
        DiscoveredDocCount: discovered,
        SkippedCount:       progress.AggregateStats.SkippedFiles,
        Verified:           true,
    }
    if meiliCount < crawled {
        verification.Warnings = append(verification.Warnings,
            fmt.Sprintf("MeiliSearch 文档数(%d) < 爬取写入数(%d)", meiliCount, crawled))
    }
    if discovered > 0 && crawled < discovered {
        verification.Warnings = append(verification.Warnings,
            fmt.Sprintf("已索引(%d) < 已发现(%d), 跳过(%d)", crawled, discovered, skipped))
    }
    progress.Verification = verification
}
```

### Frontend Changes

#### sync-schemas.ts

```typescript
// CrawlStatsSchema 新增
filesDiscovered: z.number(),
skippedFiles: z.number(),

// SyncProgressSchema 新增
verification: z.object({
  meiliDocCount: z.number(),
  crawledDocCount: z.number(),
  discoveredDocCount: z.number(),
  skippedCount: z.number(),
  verified: z.boolean(),
  warnings: z.array(z.string()).optional().default([]),
}).optional().nullable(),
```

#### sync-progress-display.tsx

- 统计卡片新增"已发现"和"已跳过"
- 同步完成时显示对账结果：OK / Warning 标识
- Warning 时显示黄色提示横幅

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
