# Task 003: 前端轮询鲁棒性测试 (Red)

**depends-on**: (none)

## Description

为 `useSyncProgress` hook 添加单元测试，验证 `startSync` 调用后轮询不会因为收到暂时的非 running 状态而过早停止。

当前问题：`startSync` 调用 `startPolling('running')` 启动轮询，但轮询回调在收到第一个 `status !== 'running'` 的响应时立即停止。这导致如果后端还未更新进度（返回旧数据），轮询永久停止。

修复策略：`startSync` 成功后立即进行乐观更新（设置 progress 为 running），并在轮询中添加宽限期机制——`startSync` 启动的轮询在前 N 次轮询内即使收到非 running 状态也不停止。

## Execution Context

**Task Number**: 3 of 5
**Phase**: Core Features
**Prerequisites**: None（与 Task 001/002 并行）

## BDD Scenario Reference

```gherkin
Scenario: startSync 后 UI 立即显示 running（乐观更新）
  Given useSyncProgress hook 已初始化
  And POST /api/v1/admin/sync 返回 202
  When 调用 startSync
  Then progress.status 立即变为 "running"（不等待 GET 结果）

Scenario: startSync 后轮询不会因旧数据停止
  Given startSync 已成功调用
  And 后端 GET /api/v1/admin/sync 返回旧数据 status="done"
  When 第一次轮询返回 status="done"
  Then 轮询不停止（宽限期内）
  And 继续下一次轮询

Scenario: 宽限期结束后轮询正常停止
  Given startSync 已成功调用
  And 宽限期已过
  When 轮询返回 status="done"
  Then 轮询正常停止
```

## Files to Modify/Create

- Modify: `web/src/hooks/use-sync-progress.test.ts`

## Steps

### Step 1: 添加测试用例

在现有测试文件中添加以下测试：

1. **startSync 乐观更新测试**: 验证 `startSync` 成功后，`progress` 立即被设置为 running 状态（通过检查 POST 成功后、GET 返回前的状态）

2. **轮询宽限期测试**: mock `fetch` 使 GET 返回旧数据（status="done"），验证 startSync 之后轮询在宽限期内不停止（使用 `vi.useFakeTimers()` 控制时间，检查多次 `POLL_INTERVAL` 后 fetch 是否继续被调用）

3. **宽限期结束后正常停止测试**: 验证宽限期过后（超过 N 次轮询），如果仍然收到非 running 状态，轮询正确停止

### Step 2: 验证测试失败

运行测试确认全部失败，因为当前实现没有乐观更新和宽限期机制。

## Verification Commands

```bash
cd /root/workspace/npan/web && npx vitest run src/hooks/use-sync-progress.test.ts
```

## Success Criteria

- 三个新测试用例编写完成
- 测试全部 FAIL（Red）
- 失败原因是断言失败（非编译/运行时错误）
