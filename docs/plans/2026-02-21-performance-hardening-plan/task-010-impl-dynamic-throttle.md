# Task 010: Implement dynamic throttle

**depends-on**: task-009, task-008

## Description

在 RequestLimiter 中添加动态速率调整功能。新增 `ActivityChecker` 接口和 `AdjustRate` 方法，当搜索活跃时将速率降低 50%，空闲时恢复。

## Execution Context

**Task Number**: 010 of 012
**Phase**: Sync Dynamic Throttle
**Prerequisites**: Task 009 测试已创建, Task 008 SearchActivityTracker 已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 4.1 (搜索活跃时降低同步速率), Scenario 4.2 (搜索空闲后恢复)

## Files to Modify/Create

- Modify: `internal/indexer/limiter.go` — 添加 ActivityChecker 接口、AdjustRate 方法、保存 baseRate 字段

## Steps

### Step 1: Define ActivityChecker interface

- 在 `limiter.go` 中定义 `ActivityChecker` 接口，包含 `IsActive() bool` 方法
- SearchActivityTracker 已隐式实现此接口

### Step 2: Save base rate

- 在 `RequestLimiter` 结构体中添加 `baseRate rate.Limit` 字段
- 在 `NewRequestLimiter` 中保存计算出的 baseRate

### Step 3: Implement AdjustRate

- 添加 `AdjustRate(checker ActivityChecker)` 方法
- 若 `checker.IsActive()`：`l.limiter.SetLimit(l.baseRate / 2)`
- 否则：`l.limiter.SetLimit(l.baseRate)`
- 若 baseRate 为 `rate.Inf`，降速时使用一个合理的有限速率（如 `rate.Every(100ms)`）

### Step 4: Integrate into Schedule

- 在 `Schedule` 方法中，若 `l.checker` 非 nil，在每次调用前执行 `l.AdjustRate(l.checker)`
- 或者在 `RequestLimiter` 中保存 checker 引用，通过 `SetActivityChecker(checker ActivityChecker)` 方法设置

### Step 5: Verify Green

- 运行 Task 009 创建的测试，验证全部通过
- **Verification**: `go test ./internal/indexer/... -run "TestRequestLimiter_Throttle" -v`

## Verification Commands

```bash
go test ./internal/indexer/... -run "TestRequestLimiter_Throttle" -v
go test ./internal/indexer/... -v
```

## Success Criteria

- Task 009 的测试全部通过（Green）
- 原有 `TestRequestLimiter_*` 测试仍然通过（无回归）
- `SearchActivityTracker` 满足 `ActivityChecker` 接口
