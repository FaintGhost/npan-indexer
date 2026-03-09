# Task 002: [IMPL] SyncManager 状态抽象与游标语义 (GREEN)

**depends-on**: task-002-sync-manager-state-test, task-001-sqlite-state-store-impl

## Description

将 `SyncManager` 从 JSON store 硬编码切换为可注入的 progress/sync-state/checkpoint 抽象，并保持 full/incremental 游标、重启恢复与现有进度语义不变。

## Execution Context

**Task Number**: 002 of 009
**Phase**: Core Features
**Prerequisites**: `task-002-sync-manager-state-test` 已完成并处于 Red；`task-001-sqlite-state-store-impl` 已提供可用 SQLite state store

## BDD Scenario

```gherkin
Scenario: 全量同步成功后进度与增量游标写入 SQLite
  Given SyncManager 使用 SQLite progress store、sync state store 与 checkpoint store factory
  And 一次全量同步成功完成
  When 管理端或 CLI 读取同步状态
  Then 应能从 SQLite 读取到 status=done 的 SyncProgressState
  And 应能从 SQLite 读取到 LastSyncTime 大于 0 的 SyncState
  And 进度中的根目录统计、verification 与 completedRoots 应保持与迁移前一致

Scenario: 全量同步失败时不会错误推进增量游标
  Given SyncManager 使用 SQLite 状态存储
  And 一次全量同步在 crawl 过程中失败
  When 读取 SQLite 中的 sync state
  Then LastSyncTime 不应被写成新的成功时间点
  And SQLite 中的 SyncProgressState 应反映 error 状态与失败原因

Scenario: 进程重启后 running 状态会从 SQLite 恢复为 interrupted
  Given SQLite 中持久化了一份 status=running 的 SyncProgressState
  And 当前进程内没有活跃同步 goroutine
  When 管理端调用 GetSyncProgress
  Then 返回状态应被修正为 interrupted
  And LastError 应提示进程重启导致同步中断
  And 修正后的状态应回写到 SQLite
```

**Spec Source**: `../2026-03-08-sync-state-sqlite-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go`
- Modify: `internal/service/sync_manager_progress_test.go`
- Modify: `internal/service/sync_manager_incremental_test.go`
- Modify: `internal/service/sync_manager_routing_test.go`
- Modify: `cmd/server/main.go`

## Steps

### Step 1: Verify Scenario

- 确认本任务只改服务层抽象与 server 侧基础 wiring，不提前处理 CLI 单独命令行为。

### Step 2: Implement Logic (Green)

- 为 `SyncManager` 引入 progress/sync-state/checkpoint 抽象接口。
- 移除运行时对 `NewJSONSyncStateStore` 与 `NewJSONCheckpointStore` 的直接依赖。
- 将 full success、full failure、incremental success/failure、restart interrupted 的现有行为接到注入的新 store 上。
- 更新 server 初始化路径，使默认运行时使用 SQLite store bundle。

### Step 3: Verify Pass

- 运行 `task-002` 的服务层测试并确认通过。
- 检查 full/incremental 游标语义与设计一致。

### Step 4: Refactor & Safety Check

- 清理 JSON store 类型耦合，避免 `SyncManagerArgs` 残留具体实现类型。
- 确认 `GetProgress()` 的 interrupted 回写路径使用新的 progress store。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/service -run 'CursorUpdate|Incremental|Interrupted|SQLite' -count=1
GOCACHE=/tmp/go-build go test ./cmd/server ./internal/service -count=1
```

## Success Criteria

- `SyncManager` 不再依赖具体 JSON store 类型。
- full/incremental 游标语义、重启恢复语义保持不变。
- server 运行路径默认接入 SQLite store。
- `task-002` 相关测试全部转绿。
