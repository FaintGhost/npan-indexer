# Task 004: [TEST] Admin/CLI SQLite 兼容性 (RED)

**depends-on**: (none)

## Description

先用失败测试锁定 Admin Connect API 与 CLI 在 SQLite 后端下保持兼容，确保迁移不会破坏 `GetSyncProgress` / `WatchSyncProgress` / `sync-progress` 的外部行为。

## Execution Context

**Task Number**: 004 of 009
**Phase**: Integration
**Prerequisites**: 已阅读 `internal/httpx/connect_admin.go`、`internal/httpx/connect_admin_test.go`、`internal/cli/root.go`、`internal/cli/root_progress_test.go`

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
- Modify: `internal/cli/root_progress_test.go`
- Create: `internal/cli/root_sync_sqlite_test.go`

## Steps

### Step 1: Verify Scenario

- 确认场景覆盖了 Connect 读进度、streaming watch 与 CLI 读进度三条外部观测路径。

### Step 2: Implement Test (Red)

- 为 `GetSyncProgress` / `WatchSyncProgress` 增加基于 SQLite store bundle 的失败测试。
- 为 CLI `sync-progress` 增加“无 progress JSON 但 SQLite 中有记录”的失败测试。
- 使用临时 SQLite 文件、stub sync manager 或测试替身；禁止依赖真实网络。

### Step 3: Verify Red Failure

- 运行目标接口层与 CLI 测试并确认失败。
- 失败必须指向“wiring 仍读 JSON / SQLite 未接入命令路径 / watch 读不到 SQLite 进度”。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/httpx ./internal/cli -run 'SyncProgress|WatchSyncProgress|SQLite' -count=1
```

## Success Criteria

- 接口层与 CLI 的 SQLite 兼容性测试稳定处于 Red。
- 失败直接指向 wiring 或读取路径缺失。
- 测试不依赖真实外部服务。
