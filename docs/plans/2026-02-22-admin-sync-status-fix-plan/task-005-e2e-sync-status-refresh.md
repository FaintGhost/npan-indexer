# Task 005: E2E 同步状态自动刷新测试

**depends-on**: task-002, task-004

## Description

添加 Playwright E2E 测试，验证在 /admin 页面点击"启动同步"后，UI 自动更新显示 running 状态，无需手动刷新。

## Execution Context

**Task Number**: 5 of 5
**Phase**: Testing
**Prerequisites**: 后端和前端修复均已完成

## BDD Scenario Reference

```gherkin
Scenario: 启动同步后 UI 自动显示运行状态
  Given 用户已认证并进入 /admin 页面
  When 用户点击"启动同步"按钮
  Then 页面自动显示"同步进行中"文字（无需手动刷新）
  And 按钮文字变为"同步进行中"

Scenario: 启动同步后显示运行中状态标签
  Given 用户已认证并进入 /admin 页面
  When 用户点击"启动同步"按钮
  Then 页面显示"运行中"状态标签（蓝色 badge）
```

## Files to Modify/Create

- Modify: `web/e2e/tests/admin.spec.ts` (添加测试用例到 'Admin 同步控制' describe)
- Modify: `web/e2e/pages/admin-page.ts` (如需添加新的 locator)

## Steps

### Step 1: 确认 admin-page.ts 有必要的 locator

检查 `AdminPage` 类是否有获取同步状态 badge（"运行中"标签）的 locator。如果没有，添加一个 `syncStatusBadge` locator，选择 SyncProgressDisplay 中状态标签的元素。

### Step 2: 添加 E2E 测试

在 `admin.spec.ts` 的 'Admin 同步控制' describe 中添加测试：

**测试: 启动同步后 UI 自动显示 running 状态**

1. 使用 `authenticatedPage` fixture（已认证）
2. 导航到 /admin 页面
3. 等待页面加载完成（heading "同步管理" 可见）
4. 点击"启动同步"按钮
5. 等待 POST 请求的响应
6. 断言：按钮文字变为"同步进行中"（使用 `toBeVisible` 或 `toHaveText`，带合理 timeout）
7. 断言：不需要手动 `page.reload()` 就能看到状态变化

注意：
- E2E 测试运行在真实后端环境中，同步可能很快完成（如果没有配置有效的 NPA token）。因此测试应该：
  - 使用 `waitForResponse` 等待 POST 响应
  - 在 POST 成功后立即检查 UI 状态（不加过长等待）
  - 使用合理的 timeout（如 5 秒）等待状态变化
- 如果同步因缺少 token 而立即失败，测试可以 skip（参考现有 "取消同步" 测试的模式）

### Step 3: 运行 E2E 测试

确保新测试和所有现有 E2E 测试通过。

## Verification Commands

```bash
# 运行 admin E2E 测试
cd /root/workspace/npan/web && npx playwright test e2e/tests/admin.spec.ts

# 运行所有 E2E 测试
cd /root/workspace/npan/web && npx playwright test
```

## Success Criteria

- 新 E2E 测试通过，验证启动同步后 UI 自动更新
- 所有现有 E2E 测试继续通过（无回归）
- 测试不依赖 `page.reload()` 来验证状态更新
