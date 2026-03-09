# Task 004: [IMPL] Admin/CLI SQLite 兼容性 (GREEN)

**depends-on**: task-004-admin-cli-sqlite-test, task-002-sync-manager-state-impl, task-003-checkpoint-lifecycle-impl

## Description

把 Admin Connect API 与 CLI 的运行路径完整切到 SQLite 状态库，保证 `GetSyncProgress` / `WatchSyncProgress` / `sync-progress` 在外部语义上与迁移前保持兼容。

## Execution Context

**Task Number**: 004 of 009
**Phase**: Integration
**Prerequisites**: `task-004-admin-cli-sqlite-test` 已处于 Red；服务层与 checkpoint 的 SQLite 路径已可用

## BDD Scenario

```gherkin
Scenario: GetSyncProgress 在 SQLite 后端下保持当前响应语义
  Given 后端已切换到 SQLite 状态存储
  When AdminService.GetSyncProgress 被调用
  Then 返回的状态字段、根目录进度、聚合统计与错误信息应与迁移前保持兼容
  And 前端不需要因为状态存储切换而修改协议或字段语义

Scenario: WatchSyncProgress 在 SQLite 后端下持续推送最新进度
  Given 后端已切换到 SQLite 状态存储
  And 同步任务正在运行并周期性写入进度
  When AdminService.WatchSyncProgress 建立流式订阅
  Then 订阅方应持续收到最新的 SyncProgressState
  And 最终应收到 done、error 或 cancelled 的终态

Scenario: CLI sync-progress 从 SQLite 读取状态而不是依赖 JSON 文件
  Given 运行环境已切换到 SQLite 状态库
  And 旧的 progress JSON 文件不存在或未更新
  When 用户执行 CLI sync-progress 命令
  Then 命令仍应返回当前同步进度
  And 数据来源应是 SQLite 中的 progress 记录
```

**Spec Source**: `../2026-03-08-sync-state-sqlite-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/httpx/connect_admin_test.go`
- Modify: `internal/httpx/server_ratelimit_test.go`
- Modify: `internal/cli/root.go`
- Modify: `internal/cli/root_progress_test.go`
- Create: `internal/cli/root_sync_sqlite_test.go`
- Modify: `cmd/server/main.go`

## Steps

### Step 1: Verify Scenario

- 确认本任务聚焦 API/CLI wiring，不额外扩展前端或文档。

### Step 2: Implement Logic (Green)

- 将 server 与 CLI 的 SyncManager 初始化统一切到 SQLite store bundle。
- 更新 CLI `sync-progress` 读取路径，使其从 progress store 抽象读取，而不是直接 new JSON store。
- 确保 Connect admin 的 `GetSyncProgress` / `WatchSyncProgress` 通过同一 `SyncManager` 读取 SQLite 状态。
- 保持返回字段与 streaming 行为不变。

### Step 3: Verify Pass

- 运行 `task-004` 的接口层与 CLI 测试并确认通过。
- 确认在 legacy progress JSON 缺失时，CLI 仍能正确输出 SQLite 中的进度。

### Step 4: Refactor & Safety Check

- 清理 server/CLI 中残留的 JSON store 运行时主路径。
- 确认 Admin/CLI 共享同一套状态初始化方式，避免再次分叉。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/httpx ./internal/cli -run 'SyncProgress|WatchSyncProgress|SQLite' -count=1
GOCACHE=/tmp/go-build go test ./cmd/server ./internal/httpx ./internal/cli -count=1
```

## Success Criteria

- Admin Connect API 与 CLI 都从 SQLite 读取状态。
- `GetSyncProgress` / `WatchSyncProgress` / `sync-progress` 语义保持兼容。
- `task-004` 相关测试全部转绿。
