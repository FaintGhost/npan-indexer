# Task 008: Implement SearchActivityTracker

**depends-on**: task-007

## Description

创建 SearchActivityTracker，使用 `sync/atomic` 记录最近搜索活动的 Unix 时间戳。提供 RecordActivity() 和 IsActive() 方法。

## Execution Context

**Task Number**: 008 of 012
**Phase**: Activity Tracking
**Prerequisites**: Task 007 测试已创建

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 4.3 (搜索请求被正确追踪)

## Files to Modify/Create

- Create: `internal/search/activity_tracker.go`

## Steps

### Step 1: Create ActivityTracker

- 在 `activity_tracker.go` 中创建 `SearchActivityTracker` 结构体
- 字段：`lastActive atomic.Int64`（Unix 秒时间戳）、`windowSec int64`
- 构造函数 `NewSearchActivityTracker(windowSec int64) *SearchActivityTracker`

### Step 2: Implement RecordActivity

- `RecordActivity()` 方法：`t.lastActive.Store(time.Now().Unix())`

### Step 3: Implement IsActive

- `IsActive() bool` 方法：`return time.Now().Unix() - t.lastActive.Load() < t.windowSec`

### Step 4: Verify Green

- 运行 Task 007 创建的测试，验证全部通过
- **Verification**: `go test ./internal/search/... -run "TestSearchActivityTracker" -v`

## Verification Commands

```bash
go test ./internal/search/... -run "TestSearchActivityTracker" -v
```

## Success Criteria

- Task 007 的测试全部通过（Green）
- 实现仅使用标准库（sync/atomic, time）
