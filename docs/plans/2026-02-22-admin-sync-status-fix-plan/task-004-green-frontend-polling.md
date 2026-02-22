# Task 004: 前端轮询鲁棒性实现 (Green)

**depends-on**: task-003

## Description

修改 `useSyncProgress` hook，添加乐观更新和轮询宽限期机制，使 startSync 后 UI 能立即反映 running 状态且轮询不会过早停止。

## Execution Context

**Task Number**: 4 of 5
**Phase**: Core Features
**Prerequisites**: Task 003 测试已编写

## BDD Scenario Reference

与 Task 003 相同的三个场景。

## Files to Modify/Create

- Modify: `web/src/hooks/use-sync-progress.ts`

## Steps

### Step 1: 添加乐观更新

在 `startSync` 函数中，`apiPost` 成功返回后、调用 `fetchProgress` 之前（或替代 `fetchProgress`），立即执行 `setProgress` 将状态设为 running 的最小对象。这确保 UI 立即更新而不依赖 GET 响应。

具体做法：POST 成功后，调用 `setProgress(prev => ...)` 设置一个 running 状态的 progress 对象。如果之前有 progress（prev 非 null），保留其他字段但更新 status 为 "running" 并清空 lastError；如果之前没有 progress，创建一个最小的 running progress 对象。

仍然调用 `fetchProgress()` 来获取真实数据（可能已经是 running，也可能是旧数据），但乐观更新确保 UI 不会闪回旧状态。

### Step 2: 添加轮询宽限期

修改 `startPolling` 函数，接受一个可选的 `gracePollCount` 参数（默认为 0）。当 `gracePollCount > 0` 时，前 N 次轮询即使收到非 running 状态也不停止轮询。

在 `startSync` 中调用 `startPolling('running', 5)` 传入宽限次数（5 次 × 2 秒 = 10 秒宽限期）。

修改轮询回调逻辑：
- 添加一个 `let remainingGrace = gracePollCount` 计数器
- 每次轮询后：如果 `result.status !== 'running'` 且 `remainingGrace > 0`，递减计数器但不停止
- 如果 `result.status !== 'running'` 且 `remainingGrace <= 0`，停止轮询（现有行为）
- 如果 `result.status === 'running'`，继续轮询（现有行为）

### Step 3: 确保乐观更新不会被 fetchProgress 覆盖为旧数据

在 `fetchProgress` 中，如果当前已经是 running（通过 ref 追踪 startSync 已调用），且 GET 返回非 running，不要覆盖 progress state。或更简单地：让 polling 的 `fetchProgress` 正常更新 state——因为有了宽限期，即使状态暂时闪回旧数据，下一次 poll 会再次更新。实际上，配合后端修复（Task 002），GET 应该已经返回 running 状态，所以这个边界情况基本不会发生。

### Step 4: 运行测试验证

运行 Task 003 的测试和所有现有测试，确认通过。

## Verification Commands

```bash
# 运行 use-sync-progress 测试
cd /root/workspace/npan/web && npx vitest run src/hooks/use-sync-progress.test.ts

# 运行所有前端单元测试
cd /root/workspace/npan/web && npx vitest run
```

## Success Criteria

- Task 003 的三个测试全部通过 (Green)
- 所有现有前端单元测试通过（无回归）
- startSync 成功后 UI 立即显示 running（乐观更新）
- 轮询在宽限期内不会因非 running 响应停止
- 宽限期过后轮询恢复正常停止逻辑
