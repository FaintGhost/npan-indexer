# Task 001: 后端 GetProgress 竞态测试 (Red)

**depends-on**: (none)

## Description

为 `SyncManager.GetProgress()` 添加单元测试，验证当同步 goroutine 已启动（`IsRunning()=true`）但 progress store 还未更新时，`GetProgress()` 应该返回正确的 "running" 状态。

测试覆盖三个竞态场景：
1. `IsRunning()=true` 且 progress store 为空（nil）→ 应返回 status="running"
2. `IsRunning()=true` 且 progress store 里有旧数据（status="done"）→ 应返回 status="running"
3. `IsRunning()=true` 且 progress store 里有旧数据（status="interrupted"）→ 应返回 status="running"

## Execution Context

**Task Number**: 1 of 5
**Phase**: Core Features
**Prerequisites**: None

## BDD Scenario Reference

**Scenario**: GetProgress returns running when goroutine is active

```gherkin
Scenario: GetProgress 返回 running 当 goroutine 活跃但 store 为空
  Given SyncManager 的 running 标记为 true
  And progress store 不包含任何数据
  When 调用 GetProgress()
  Then 返回非 nil 的 SyncProgressState
  And status 为 "running"

Scenario: GetProgress 返回 running 当 goroutine 活跃但 store 有旧完成数据
  Given SyncManager 的 running 标记为 true
  And progress store 包含 status="done" 的旧数据
  When 调用 GetProgress()
  Then 返回 SyncProgressState 且 status 为 "running"
  And lastError 被清空

Scenario: GetProgress 返回 running 当 goroutine 活跃但 store 有中断数据
  Given SyncManager 的 running 标记为 true
  And progress store 包含 status="interrupted" 的旧数据
  When 调用 GetProgress()
  Then 返回 SyncProgressState 且 status 为 "running"
  And lastError 被清空
```

## Files to Modify/Create

- Create: `internal/service/sync_manager_progress_test.go`

## Steps

### Step 1: 创建测试文件

在 `internal/service/sync_manager_progress_test.go` 中创建测试，使用与现有测试（如 `sync_manager_mode_test.go`）一致的模式：
- 使用 `storage.NewJSONProgressStore` 和临时文件创建 progress store
- 直接构造 `SyncManager` 并设置内部 `running` 字段为 true（通过调用 `Start` 配合 mock API 或直接设置字段）
- 因为 `running` 是私有字段，需要通过 `Start()` 方法（传入不会实际运行的 mock API）来使 `IsRunning()` 返回 true

**注意**: `SyncManager.running` 是私有字段。测试在同一个 `service` 包内，可以直接赋值 `m.running = true` 并加锁。或者使用一个辅助方式来触发 running 状态。参考 `sync_manager_incremental_test.go` 的模式。

### Step 2: 验证测试失败

运行测试，确认三个场景都失败（当前 `GetProgress()` 不处理这些竞态场景）：
- 场景 1: progress 为 nil 时直接返回 nil，不会返回 running 状态
- 场景 2/3: progress 有旧 status 时直接返回旧状态，不会纠正为 running

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run TestGetProgress -v
```

## Success Criteria

- 三个测试用例全部编写完成
- 三个测试用例全部 FAIL（Red），因为当前 `GetProgress()` 不处理竞态场景
- 测试失败原因是断言失败（不是编译错误或 panic）
