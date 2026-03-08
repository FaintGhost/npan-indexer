# Task 002: [TEST] SyncManager 状态抽象与游标语义 (RED)

**depends-on**: (none)

## Description

先用失败测试锁定 `SyncManager` 已不再硬编码 JSON store，并且在 SQLite 后端下继续保留 full/incremental 游标与重启恢复语义。该任务只负责编写和验证失败测试，不修改运行时代码。

## Execution Context

**Task Number**: 002 of 009
**Phase**: Core Features
**Prerequisites**: 已阅读 `internal/service/sync_manager.go`、`internal/service/sync_manager_progress_test.go`、`internal/service/sync_manager_incremental_test.go`

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

- Modify: `internal/service/sync_manager_progress_test.go`
- Modify: `internal/service/sync_manager_incremental_test.go`
- Modify: `internal/service/sync_manager_routing_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD 文档中覆盖了 full success、full failure、restart interrupted 三类核心状态语义。

### Step 2: Implement Test (Red)

- 将现有 `SyncManager` 测试扩展为基于 SQLite store bundle 的新失败用例，覆盖：
  - full success 后 progress 与 cursor 都写入 SQLite
  - full failure 时 cursor 不推进
  - incremental 成功/失败时继续沿用原有游标语义
  - `GetProgress()` 在“DB 是 running、内存未运行”场景下会把状态修正为 interrupted 并持久化
- 使用 stub API、stub index、临时 SQLite 文件隔离外部依赖。

### Step 3: Verify Red Failure

- 运行目标服务层测试并确认失败。
- 失败原因必须指向“SyncManager 仍硬编码 JSON / 尚未支持 store 抽象 / SQLite 路径未接通”，而不是网络或真实索引依赖。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/service -run 'CursorUpdate|Incremental|Interrupted|SQLite' -count=1
```

## Success Criteria

- 新增服务层测试稳定处于 Red。
- 失败直接指向 SyncManager 的状态抽象或 SQLite 接线缺失。
- 测试使用 test doubles，不依赖真实外部服务。
