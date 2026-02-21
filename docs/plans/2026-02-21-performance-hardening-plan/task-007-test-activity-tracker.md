# Task 007: Test SearchActivityTracker

**depends-on**: (none)

## Description

为 SearchActivityTracker 创建测试。该组件使用原子操作记录最近的搜索活动时间戳，提供 IsActive() 方法判断是否在活跃窗口内。

## Execution Context

**Task Number**: 007 of 012
**Phase**: Activity Tracking
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 4.3 (搜索请求被正确追踪)

## Files to Modify/Create

- Create: `internal/search/activity_tracker_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD specs 中 Scenario 4.3 存在

### Step 2: Implement Tests (Red)

- `TestSearchActivityTracker_RecordAndIsActive`: 创建 tracker（windowSec=1），调用 RecordActivity()，立即检查 IsActive() 返回 true
- `TestSearchActivityTracker_ExpiresAfterWindow`: 创建 tracker（windowSec=1），调用 RecordActivity()，等待 > 1 秒，检查 IsActive() 返回 false
- `TestSearchActivityTracker_InitiallyInactive`: 创建 tracker，不调用 RecordActivity()，检查 IsActive() 返回 false
- **Verification**: 测试应编译失败（Red），因为 SearchActivityTracker 尚未定义

### Step 3: Verify Red

- 运行测试确认编译失败

## Verification Commands

```bash
go test ./internal/search/... -run "TestSearchActivityTracker" -v
```

## Success Criteria

- 测试编译失败（Red），因为缺少 SearchActivityTracker 类型
- 测试逻辑正确映射 BDD Scenario 4.3
- 无外部依赖（纯内存操作 + time.Sleep）
