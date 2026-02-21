# Task 009: Test dynamic throttle in RequestLimiter

**depends-on**: (none)

## Description

为 RequestLimiter 的动态速率调整功能创建测试。验证当搜索活跃时同步速率降低，搜索空闲后恢复。

## Execution Context

**Task Number**: 009 of 012
**Phase**: Sync Dynamic Throttle
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 4.1 (搜索活跃时降低同步速率), Scenario 4.2 (搜索空闲后恢复)

## Files to Modify/Create

- Create: `internal/indexer/limiter_throttle_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD specs 中 Scenario 4.1 和 4.2 存在

### Step 2: Define test interface

- 测试需要一个 `ActivityChecker` 接口（`IsActive() bool`），创建 mock 实现来控制活跃状态

### Step 3: Implement Tests (Red)

- `TestRequestLimiter_ThrottlesWhenActive`: 创建 RequestLimiter，设置 mock ActivityChecker 返回 `true`（搜索活跃），调用 `AdjustRate(checker)`，验证 limiter 的速率降低（通过观察 Schedule 调用间隔变长来验证）
- `TestRequestLimiter_RestoresWhenInactive`: 先设置活跃导致降速，然后将 mock 改为 `false`（搜索空闲），调用 `AdjustRate(checker)`，验证速率恢复
- 测试可通过 `RequestLimiter` 新增的公开方法 `AdjustRate(ActivityChecker)` 来触发速率调整
- **Verification**: 测试应编译失败（Red），因为 `AdjustRate` 方法和 `ActivityChecker` 接口尚未定义

### Step 4: Verify Red

- 运行测试确认编译失败

## Verification Commands

```bash
go test ./internal/indexer/... -run "TestRequestLimiter_Throttle" -v
```

## Success Criteria

- 测试编译失败（Red），因为缺少 AdjustRate 方法
- 使用 mock ActivityChecker 隔离对 SearchActivityTracker 的依赖
