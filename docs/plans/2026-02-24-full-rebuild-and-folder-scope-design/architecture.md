# Architecture

## Existing Behavior Snapshot

- 全量爬取入口：`internal/service/sync_manager.go` `run()`
- 单 root 执行：`runSingleRoot()` -> `indexer.RunFullCrawl()`
- 断点机制：`RunFullCrawl()` 启动时先 `CheckpointStore.Load()`
- 问题点：`force_rebuild` 仅清 Meili，不清 checkpoint

## Target Architecture

### A. Checkpoint Reset Path

```text
POST /api/v1/admin/sync (mode=full, force_rebuild=true)
  -> SyncManager.run()
    -> discoverRootFolders()
    -> build rootCheckpointMap
    -> resetCheckpointFiles(rootCheckpointMap)   <-- NEW
    -> runSingleRoot(root...)
      -> RunFullCrawl()
        -> checkpoint.Load() == nil
        -> 从 root/page=0 开始
```

### B. Folder Scope Path

```text
Admin UI 输入 root_folder_ids
  -> startSync(body.root_folder_ids, body.include_departments=false)
    -> StartFullSync payload 透传
    -> discoverRootFolders() 仅按指定 root + (可选)部门
```

### C. Official Estimate Path

```text
discoverRootFolders()
  for explicit root id:
    -> npan.GetFolderInfo(rootID)
    -> rootNameMap[rootID] = name
    -> rootEstimateMap[rootID] = item_count + 1

progress root:
  estimatedTotalDocs = rootEstimateMap[rootID]
```

### D. Completion Warning Path

```text
all roots done
  -> for each root:
    actual = filesIndexed + foldersVisited
    estimated = estimatedTotalDocs
    if diff too large:
      append verification.warnings
```

## File-Level Changes

### Backend

- `internal/service/sync_manager.go`
  - 新增 checkpoint 重置流程（`force_rebuild`/`resume=false`）
  - 在 `discoverRootFolders()` 对显式 root 调用 `GetFolderInfo`
  - 同步完成后追加 root 级差异告警
- `internal/npan/types.go`
  - `API` 接口新增 `GetFolderInfo`
- `internal/npan/client.go`
  - 实现 `GetFolderInfo`（`GET /api/v2/folder/{id}/info`）
- `internal/models/models.go`
  - 若需要细化展示，扩展 verification warning 结构（可选，优先保持字符串告警）

### Frontend

- `web/src/components/admin-sync-page.tsx`
  - 新增目录 ID 输入控件
  - 启动同步时透传 `root_folder_ids`
  - 指定目录时显式发送 `include_departments=false`
- `web/src/hooks/use-sync-progress.ts`
  - `startSync` 扩展参数，支持 `includeDepartments`
- `web/src/components/sync-progress-display.tsx`
  - 显示 estimate 与 actual 差值提示

## Data Contract Changes

### Start Sync Request（已有字段复用）

- `root_folder_ids?: number[]`
- `include_departments?: boolean`

### Progress Response

- 不新增硬性字段，优先复用：
  - `rootProgress[*].estimatedTotalDocs`
  - `verification.warnings`

## Testing Strategy Hooks

- 后端：
  - `sync_manager_routing_test.go`：验证 `force_rebuild` 会忽略旧 checkpoint
  - `sync_manager_estimate_test.go`：显式 root 的 estimate/name 填充
  - `sync_manager_verification_test.go`：差异告警生成
- 前端：
  - `use-sync-progress.test.ts`：请求体包含 `root_folder_ids/include_departments`
  - `admin-page.test.tsx`：目录输入与启动行为
  - `sync-progress-display.test.tsx`：estimate/actual 告警展示
