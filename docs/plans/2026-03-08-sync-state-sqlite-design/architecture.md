# Architecture: 同步状态 SQLite 迁移

## Current State

### 当前状态写入点

1. `internal/service/sync_manager.go`
  - `GetProgress()` 直接读取 `JSONProgressStore`
  - `run()` / `runSingleRoot()` 在全量路径中反复保存 progress
  - `runIncrementalPath()` 读取/写入 `JSONSyncStateStore`
  - `runSingleRoot()` 直接 `NewJSONCheckpointStore(checkpointFile)`

2. `internal/storage/json_store.go`
  - 三类 store 彼此独立
  - 每次都是整对象 JSON 序列化后原子覆写文件

3. `internal/cli/root.go`
  - `sync` 命令把 `cfg.ProgressFile` / `cfg.SyncStateFile` 作为 JSON 文件路径注入
  - `sync-progress` 直接读取 `NewJSONProgressStore(progressFile)`

### 当前问题

- 状态散落在多份 JSON 文件里，没有统一状态源。
- 一次同步生命周期中的 progress、cursor、checkpoint 不能共享事务边界。
- `SyncManager`、CLI、测试都直接耦合 JSON store，替换实现成本高。

## Target State

### 目标组件图

```text
cmd/server.main / internal/cli.root
  -> storage.NewSQLiteStateStores(...)
    -> SQLiteProgressStore
    -> SQLiteSyncStateStore
    -> SQLiteCheckpointStoreFactory
  -> service.NewSyncManager(...)
    -> ProgressStore interface
    -> SyncStateStore interface
    -> CheckpointStoreFactory interface
      -> indexer.RunFullCrawl(..., CheckpointStore)

legacy JSON files
  -> 仅作为 lazy import source
```

### 关键原则

1. **单一状态源**：SQLite 成为运行时主状态源。
2. **接口先行**：服务层不直接依赖 JSON/SQLite 具体类型。
3. **模型不重做**：仍沿用现有 `models.SyncProgressState` / `models.SyncState` / `models.CrawlCheckpoint`。
4. **外部行为不变**：Connect、CLI、前端、测试观测到的状态语义保持一致。

## Data Model

### SQLite 文件

- 默认状态库：`NPA_STATE_DB_FILE`
- 推荐默认值：`./data/state/sync-state.sqlite`

### 表结构

```sql
CREATE TABLE IF NOT EXISTS state_entries (
  namespace TEXT NOT NULL,
  key TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  updated_at_ms INTEGER NOT NULL,
  PRIMARY KEY(namespace, key)
);
```

### namespace / key 约定

| namespace | key | payload |
|---|---|---|
| `progress` | `default` | `models.SyncProgressState` |
| `sync_state` | `default` | `models.SyncState` |
| `checkpoint` | `<logical checkpoint key>` | `models.CrawlCheckpoint` |

### 为什么不拆表

当前业务的主要访问模式是：

- 整体读取进度对象
- 整体保存进度对象
- 按根目录 key 读取/写入 checkpoint
- 读取/写入单个增量 cursor

这更像 **KV + JSON document**，而不是复杂关系查询。此时：

- 拆表不会明显改善当前读写路径。
- 反而会增加 `RootProgress`、`CatalogRootProgress`、`Verification` 等嵌套结构的维护成本。

因此本轮选“SQLite 事务 + JSON payload”而不是“完全关系化建模”。

## Store Abstractions

### 新接口

```text
ProgressStore
  Load() (*models.SyncProgressState, error)
  Save(*models.SyncProgressState) error

SyncStateStore
  Load() (*models.SyncState, error)
  Save(*models.SyncState) error

CheckpointStore
  Load() (*models.CrawlCheckpoint, error)
  Save(*models.CrawlCheckpoint) error
  Clear() error

CheckpointStoreFactory
  ForKey(key string) CheckpointStore
```

### 预期实现

- `JSONProgressStore` / `JSONSyncStateStore` / `JSONCheckpointStore`
  - 作为 legacy / 测试 / fallback 实现继续保留
- `SQLiteProgressStore`
- `SQLiteSyncStateStore`
- `SQLiteCheckpointStoreFactory`
- `SQLiteCheckpointStore`

## Wiring Changes

### `internal/service/sync_manager.go`

需要改动点：

1. `SyncManager` 字段
  - `progressStore *storage.JSONProgressStore` -> `progressStore storage.ProgressStore`
  - `syncStateFile string` -> `syncStateStore storage.SyncStateStore`
  - 增加 `checkpointStores storage.CheckpointStoreFactory`

2. `SyncManagerArgs`
  - 去掉 `SyncStateFile string`
  - 增加 `SyncStateStore storage.SyncStateStore`
  - 增加 `CheckpointStores storage.CheckpointStoreFactory`

3. `runSingleRoot()`
  - 不再 `storage.NewJSONCheckpointStore(checkpointFile)`
  - 改为 `m.checkpointStores.ForKey(checkpointFile)`

4. `run()` / `runIncrementalPath()`
  - 不再在运行时直接 new `JSONSyncStateStore`
  - 统一走注入的 `m.syncStateStore`

### `cmd/server/main.go` / `internal/cli/root.go`

- 基于统一 config 创建 SQLite stores。
- 把 progress / sync_state / checkpoint factory 注入 `SyncManager`。
- `sync-progress` 命令直接使用同一个 progress store。

### `internal/config/config.go`

新增：

- `StateDBFile string`

保留：

- `ProgressFile`
- `SyncStateFile`
- `CheckpointTemplate`

用途从“主存储路径”调整为“legacy import source / 逻辑模板”。

## Migration Flow

### Progress / SyncState

```text
SQLiteProgressStore.Load()
  -> 先查 SQLite
  -> 若不存在，检查 legacy JSON progress 文件
  -> 若存在且可解析：写入 SQLite，返回结果
  -> 若不存在：返回 nil
```

`SyncStateStore.Load()` 同理。

### Checkpoint

```text
SQLiteCheckpointStore.Load(key)
  -> 先查 SQLite checkpoint/key
  -> 若不存在，检查 legacy checkpoint JSON 文件(key 视为逻辑路径)
  -> 若存在且可解析：写入 SQLite checkpoint/key，返回结果
  -> 若不存在：返回 nil
```

### 非破坏性要求

- 本轮不自动删除旧 JSON。
- 本轮不强制一次性迁移全部 checkpoint。
- 迁移失败时应保留原始错误上下文，便于排查损坏 JSON。

## Reliability Strategy

### SQLite 连接策略

- 单库共享连接池
- 通过 `database/sql` 管理连接
- 初始化时设置：
  - `journal_mode=WAL`
  - `synchronous=FULL`
  - `busy_timeout=5000`

### 写入策略

- `Save` / `Clear` 使用事务或单条原子 UPSERT/DELETE。
- 每次写入同时更新 `updated_at_ms`。
- checkpoint clear 必须是真正删除对应 row，避免 resume 误恢复旧值。

### 重启恢复

`GetProgress()` 现有语义保留：

- 如果 SQLite 中是 `running`，但进程内没有 goroutine 运行
- 则改写为 `interrupted`
- 并回写 SQLite

这意味着前端无需感知底层存储切换。

## Files Expected to Change

### 核心实现

- `internal/storage/json_store.go`
- `internal/storage/json_store_test.go`
- `internal/service/sync_manager.go`
- `internal/config/config.go`
- `internal/config/validate.go`
- `cmd/server/main.go`
- `internal/cli/root.go`

### 服务/接口测试

- `internal/service/sync_manager_progress_test.go`
- `internal/service/sync_manager_incremental_test.go`
- `internal/service/sync_manager_routing_test.go`
- `internal/httpx/connect_admin_test.go`
- `internal/httpx/server_ratelimit_test.go`

### 文档 / 运维

- `README.md`
- `docs/runbooks/index-sync-operations.md`

## Verification Layers

### Unit

- store CRUD 与 lazy import
- SyncManager 行为回归
- CLI 读状态回归

### Integration

- Admin Connect `GetSyncProgress` / `WatchSyncProgress`
- full -> cursor write
- incremental -> cursor advance

### End-to-End

- `GOCACHE=/tmp/go-build go test ./...`
- `cd web && bun vitest run`
- Docker smoke
- Playwright E2E
