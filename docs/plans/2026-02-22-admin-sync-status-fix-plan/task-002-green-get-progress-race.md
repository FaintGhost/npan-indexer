# Task 002: 后端 GetProgress 竞态修复 (Green)

**depends-on**: task-001

## Description

修改 `SyncManager.GetProgress()` 方法，确保当 `IsRunning()=true` 时，返回的状态始终反映真实的运行状态。

## Execution Context

**Task Number**: 2 of 5
**Phase**: Core Features
**Prerequisites**: Task 001 测试已编写

## BDD Scenario Reference

与 Task 001 相同的三个场景。

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go` (GetProgress 方法，约 line 100-117)

## Steps

### Step 1: 修改 GetProgress 方法

在 `GetProgress()` 中添加两个逻辑分支：

1. **progress 为 nil 且 IsRunning() 为 true**: 返回一个最小的 "running" SyncProgressState（包含 status="running"、当前时间戳的 startedAt 和 updatedAt，以及必要的零值字段如空的 roots/completedRoots/rootProgress/aggregateStats）

2. **progress 非 nil 且 status 不是 "running" 但 IsRunning() 为 true**: 将 status 覆盖为 "running" 并清空 lastError（注意：只修改返回值的内存副本，不要写回 progress store，因为 goroutine 最终会写入正确的进度）

注意保持现有的 "interrupted" 检测逻辑不变（`progress.Status == "running" && !m.IsRunning()` → "interrupted"）。新逻辑应在 interrupted 检测之后执行。

### Step 2: 验证测试通过

运行 Task 001 的三个测试，确认全部通过。

### Step 3: 运行全部现有测试

确保没有引入回归。

## Verification Commands

```bash
# 运行 GetProgress 竞态测试
cd /root/workspace/npan && go test ./internal/service/ -run TestGetProgress -v

# 运行整个 service 包的测试
cd /root/workspace/npan && go test ./internal/service/ -v
```

## Success Criteria

- Task 001 的三个测试全部通过 (Green)
- 所有现有 service 包测试通过（无回归）
- `GetProgress()` 在以下情况下行为正确：
  - `IsRunning()=true`, store=nil → 返回 minimal running state
  - `IsRunning()=true`, store=done/interrupted/error → 返回 running（内存修改，不写 store）
  - `IsRunning()=false`, store=running → 返回 interrupted（现有逻辑保持不变）
  - `IsRunning()=false`, store=done → 返回 done（现有逻辑保持不变）
