# 同步状态 SQLite 迁移设计

## Context

当前同步状态持久化分散在多份 JSON 文件中：

- `internal/storage/json_store.go` 负责 `SyncProgressState`、`SyncState`、`CrawlCheckpoint` 的文件读写。
- `internal/service/sync_manager.go` 在运行时直接硬编码 `JSONProgressStore`、`JSONSyncStateStore`、`JSONCheckpointStore`。
- `cmd/server/main.go` 与 `internal/cli/root.go` 也都直接实例化 JSON store。

这套实现虽然已经做了 `rename` 原子写，但仍有几个现实问题：

1. **状态分散**：全量进度、增量游标、每个根目录 checkpoint 分散在多份文件中，恢复与排查成本高。
2. **一致性不足**：一次同步中的多个状态更新无法放进同一个事务边界。
3. **恢复不稳**：频繁进度写入、重启后状态恢复、checkpoint 清理与 resume 语义都依赖文件系统行为。
4. **入口分叉**：Server、CLI、测试代码都直接绑死 JSON store，后续切换存储实现需要大面积改动。

本次目标不是重做同步模型，而是**把现有状态模型切到 SQLite，优先解决可靠性问题，同时保持现有外部行为不变**。

## Requirements

### Must

- 全量同步进度 `SyncProgressState` 持久化切到 SQLite。
- 增量游标 `SyncState.LastSyncTime` 持久化切到 SQLite。
- 每个根目录的 `CrawlCheckpoint` 持久化切到 SQLite。
- `AdminService.GetSyncProgress` / `WatchSyncProgress` 行为保持不变。
- CLI `sync` / `sync-progress` 读取同一份 SQLite 状态。
- 保留当前 `force_rebuild`、`resume_progress`、重启后 `running -> interrupted` 等语义。
- 支持从现有 JSON 文件**非破坏式迁移**到 SQLite。
- 构建链路必须继续兼容 `CGO_ENABLED=0`。

### Should

- 尽量少改 `models.SyncProgressState` / `models.SyncState` / `models.CrawlCheckpoint` 的结构。
- 尽量少改 Connect DTO、前端 schema、CLI 输出。
- 尽量把 SQLite 作为**单一状态源**，而不是继续维护多份文件。
- 迁移过程默认不删除旧 JSON 文件，降低回滚成本。

### Won't

- 本轮不重做 Admin UI。
- 本轮不引入多实例分布式锁或跨节点协调。
- 本轮不把同步状态改造成复杂关系型查询模型。
- 本轮不顺手改搜索、索引、鉴权等无关逻辑。

## Option Analysis

### Option A（推荐）：单 SQLite DB + namespace/key + JSON payload

使用一份共享 SQLite 文件作为状态库，把三类状态都写入同一张逻辑表：

- `progress/default` -> `SyncProgressState`
- `sync_state/default` -> `SyncState`
- `checkpoint/<logical-key>` -> `CrawlCheckpoint`

其中 payload 仍然存 JSON 文本，SQLite 负责事务、一致性和可靠落盘。

**优点**：

- 单一状态源，最符合“切换到 SQLite”的目标。
- 最小化模型改动，`models.*` 基本不用重做。
- 对 `GetProgress()`、CLI JSON 输出、Connect DTO 映射影响最小。
- checkpoint 可以继续沿用当前“逻辑文件路径”作为 key，不必改进度模型里 `CheckpointFile` 的表现。

**代价**：

- 需要把 `SyncManager` 从“直接 new JSON store”改为“注入接口/工厂”。
- 需要在 server/cli/config/test 中增加 SQLite wiring。

### Option B：每类状态各自一个 SQLite 文件

把 `progress`、`sync_state`、`checkpoint` 分别保存在不同 SQLite 文件，接口尽量贴近现有 JSON 文件路径。

**优点**：

- 改动局部较小。
- 部分调用点可以少改参数。

**缺点**：

- 仍然是多状态源，运维与恢复体验提升有限。
- “切到 SQLite”只解决了单文件可靠性，没解决状态分散问题。
- checkpoint 仍可能演化成每 root 一份 SQLite 文件，收益不高。

### Option C：继续保留 JSON，只加强写盘逻辑

比如继续优化 `fsync`、写频率、临时文件策略。

**不推荐**。

用户的问题是“当前任务同步不是很可靠”，而且代码已经做了原子 rename，继续在 JSON 文件层打补丁，收益有限、回报递减。

## Rationale

选择 Option A：

- 这是**最小模型改动**与**最大可靠性收益**的平衡点。
- 当前读写模式本来就是“整对象 Load/Save”，并不需要关系型拆表；直接存 JSON payload 更简单。
- 单 DB 能让 progress / cursor / checkpoint 落在同一实现内，后续更容易加迁移、观测和维护工具。
- 结合 `modernc.org/sqlite` 可以保持 `CGO_ENABLED=0` 构建链路不变。

## Detailed Design

### 1. 新增存储抽象，去掉 `SyncManager` 对 JSON store 的硬编码

新增通用接口：

- `ProgressStore`
  - `Load() (*models.SyncProgressState, error)`
  - `Save(state *models.SyncProgressState) error`
- `SyncStateStore`
  - `Load() (*models.SyncState, error)`
  - `Save(state *models.SyncState) error`
- `CheckpointStore`
  - `Load() (*models.CrawlCheckpoint, error)`
  - `Save(state *models.CrawlCheckpoint) error`
  - `Clear() error`
- `CheckpointStoreFactory`
  - `ForKey(key string) CheckpointStore`

`SyncManagerArgs` 改为注入：

- `ProgressStore`
- `SyncStateStore`
- `CheckpointStoreFactory`

这样 `internal/service/sync_manager.go` 不再自己 `new JSON store`，而是由上层 wiring 决定具体实现。

### 2. 使用单 SQLite 文件承载三类状态

新增 SQLite 状态库，比如默认路径：

- `./data/state/sync-state.sqlite`

SQLite 内部使用统一表：

```sql
CREATE TABLE IF NOT EXISTS state_entries (
  namespace TEXT NOT NULL,
  key TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  updated_at_ms INTEGER NOT NULL,
  PRIMARY KEY(namespace, key)
);
```

约定：

- `namespace = 'progress'`, `key = 'default'`
- `namespace = 'sync_state'`, `key = 'default'`
- `namespace = 'checkpoint'`, `key = <logical checkpoint key>`

这样：

- 不需要重构现有嵌套结构体。
- `SyncProgressState.RootProgress` / `CatalogRootProgress` / `IncrementalStats` 等复杂结构都可原样存 JSON。
- checkpoint 的 key 可以直接复用当前 `buildCheckpointFilePath(...)` 生成的逻辑路径字符串。

### 3. SQLite 驱动与运行参数

驱动选择：`modernc.org/sqlite`。

原因：

- 当前 `Dockerfile` 构建使用 `CGO_ENABLED=0`。
- `github.com/mattn/go-sqlite3` 依赖 CGO，不适合当前构建链路。
- `modernc.org/sqlite` 可通过 `database/sql` 使用，便于接入标准库测试与连接管理。

建议的连接初始化：

- `PRAGMA journal_mode = WAL`
- `PRAGMA synchronous = FULL`
- `PRAGMA busy_timeout = 5000`

目标是优先保证状态可靠性；如果后续发现进度频繁写入带来明显性能压力，再单独评估调优。

### 4. 迁移策略：按 namespace/key 惰性导入，且默认非破坏

本轮采用**惰性迁移 + 非破坏式保留**：

- 当 SQLite 中某个 namespace/key 不存在时：
  - 若对应 legacy JSON 文件存在，则读取 JSON。
  - 成功后写入 SQLite。
  - 返回导入后的对象。
- 一旦 SQLite 中已有记录，后续只读 SQLite，不再依赖 JSON。

具体映射：

- `progress/default` <- `cfg.ProgressFile`
- `sync_state/default` <- `cfg.SyncStateFile`
- `checkpoint/<key>` <- legacy checkpoint JSON 文件（由逻辑 key 对应）

这样做的好处：

- 首次切换不会丢历史状态。
- 不需要在升级脚本里强制搬迁所有 checkpoint 文件。
- 出问题时仍能人工对照旧 JSON 文件。

本轮**不自动删除、不自动 rename 旧 JSON 文件**，避免给线上恢复增加不可逆动作。

### 5. 配置与 wiring 变更

新增统一状态库配置，例如：

- `NPA_STATE_DB_FILE=./data/state/sync-state.sqlite`

保留当前配置项：

- `NPA_PROGRESS_FILE`
- `NPA_SYNC_STATE_FILE`
- `NPA_CHECKPOINT_FILE`

但它们的职责收敛为：

- 兼容旧 JSON 路径
- 提供 lazy import 的来源
- 保留 `CheckpointTemplate` 在进度模型中的逻辑展示语义

Server 与 CLI 统一改为：

1. 基于 `NPA_STATE_DB_FILE` 创建 SQLite store bundle。
2. 把 `ProgressStore` / `SyncStateStore` / `CheckpointStoreFactory` 注入 `SyncManager`。
3. `sync-progress` 命令改为从 SQLite progress store 读取，而不是直接读 JSON 文件。

### 6. 行为兼容边界

以下行为必须保持不变：

- `GetProgress()`：运行中返回 running；重启后把持久化的 running 改写为 interrupted。
- `runSingleRoot()`：按当前 `progressEvery` 节奏写进度。
- `force_rebuild` / `resume=false`：启动前清 checkpoint。
- `resume=true`：恢复旧 checkpoint。
- full 成功后写增量游标。
- incremental 失败/取消时不推进游标。
- `WatchSyncProgress` 首帧与终态推送行为不变。

其中 `RootSyncProgress.CheckpointFile` 与 `SyncProgressState.CheckpointTemplate` 继续保留现有字符串字段，用于兼容现有输出与测试，不要求把它们变成数据库内部 ID。

### 7. 测试策略

#### 存储层

- 新增 SQLite store 单测：
  - schema 初始化
  - progress/sync_state/checkpoint 的 save/load/clear
  - legacy JSON lazy import
  - 并发保存后的 payload 完整性
- 使用临时 SQLite 文件，不依赖真实外部服务。

#### 服务层

- 更新 `SyncManager` 相关测试，改为临时 SQLite DB：
  - 全量成功后写 cursor
  - 全量失败不写 cursor
  - `force_rebuild` / `resume=false` 清 checkpoint
  - incremental 成功/失败游标行为
  - `running -> interrupted` 重启恢复语义

#### 接口/CLI 层

- 更新 `connect_admin_test.go`，确保 `GetSyncProgress` / `WatchSyncProgress` 在 SQLite 后端下保持语义一致。
- 更新 `internal/cli/root.go` 相关测试，确保 `sync-progress` 改为读 SQLite。

## Success Criteria

- Server、CLI、测试代码不再直接依赖 `NewJSONProgressStore` / `NewJSONSyncStateStore` / `NewJSONCheckpointStore` 作为运行时主路径。
- SQLite 成为 progress / cursor / checkpoint 的主状态源。
- 已有 JSON 状态可被惰性导入，不阻塞首次升级。
- `go test ./...`、`cd web && bun vitest run`、Docker smoke、Playwright E2E 全部通过。
- 构建链路继续兼容 `CGO_ENABLED=0`。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为场景与测试策略
- [Architecture](./architecture.md) - 组件结构、数据流与改动范围
- [Best Practices](./best-practices.md) - 驱动选型、迁移策略与验证守则
