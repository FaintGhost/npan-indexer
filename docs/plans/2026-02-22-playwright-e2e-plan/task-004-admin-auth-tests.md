# Task 004: Admin 认证 E2E 测试

**depends-on**: Task 001

## Objective

编写管理后台认证流程的 E2E 测试，覆盖 BDD specs Feature 3 中的认证相关场景。

## BDD Scenarios Covered

- Scenario: 未认证时显示 API Key 对话框
- Scenario: 空 API Key 显示本地错误
- Scenario: 错误 API Key 显示服务端错误
- Scenario: 正确 API Key 进入管理界面
- Scenario: 刷新页面保持认证状态
- Scenario: 返回搜索链接

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/e2e/tests/admin.spec.ts` | 新建 |

## Steps

### 1. 创建 admin.spec.ts 基本结构

- 导入 `test`、`expect`、`ADMIN_API_KEY` 从 `../fixtures/auth`
- 导入 `AdminPage` 从 `../pages/admin-page`
- 不需要 Meilisearch 播种（认证测试不依赖搜索数据）

### 2. 编写「未认证时显示 API Key 对话框」测试

- 访问 `/admin/`（不注入 localStorage）
- 断言全屏 API Key 对话框可见
- 断言密码输入框可见
- 断言 "确认" 按钮可见

### 3. 编写「空 API Key 显示本地错误」测试

- 访问 `/admin/`
- 不输入任何内容，直接点击确认按钮
- 断言显示 "请输入 API Key" 错误提示

### 4. 编写「错误 API Key 显示服务端错误」测试

- 访问 `/admin/`
- 输入 "wrong-key-00000" 并点击确认
- 断言按钮显示 loading 状态（"验证中..."）
- 等待 API 返回 401
- 断言显示 "API Key 无效" 错误
- 断言对话框仍然可见

### 5. 编写「正确 API Key 进入管理界面」测试

- 访问 `/admin/`
- 输入 `ADMIN_API_KEY` 并点击确认
- 断言对话框消失
- 断言页面显示同步管理界面
- 断言 localStorage 中保存了 API Key（使用 `page.evaluate()`）

### 6. 编写「刷新页面保持认证状态」测试

- 使用 `authenticatedPage` fixture（或 `addInitScript` 注入 API Key）
- 访问 `/admin/`
- 断言不显示 API Key 对话框
- 断言直接显示同步管理界面
- 刷新页面（`page.reload()`）
- 断言仍然显示同步管理界面，不显示对话框

### 7. 编写「返回搜索链接」测试

- 使用已认证页面访问 `/admin/`
- 点击 "返回搜索" 链接
- 断言页面导航到 `/`

## Verification

```bash
cd web && bunx playwright test admin.spec.ts --list
# 应列出约 6 个认证测试用例
```
