# Task 007: 边界场景 E2E 测试

**depends-on**: Task 001

## Objective

编写边界场景和异常情况的 E2E 测试，覆盖 BDD specs Feature 4 的所有场景。

## BDD Scenarios Covered

- Scenario: 搜索特殊字符
- Scenario: 非常长的搜索查询
- Scenario: 快速连续搜索（防抖竞态）
- Scenario: 网络错误时显示错误状态
- Scenario: 搜索框纯空格不触发搜索
- Scenario: 浏览器后退/前进导航
- Scenario: Admin 认证过期（401 清除）

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/e2e/tests/search.spec.ts` | 修改：添加边界场景 describe block |
| `web/e2e/tests/admin.spec.ts` | 修改：添加认证过期测试 |

## Steps

### 1. 编写「搜索特殊字符」测试（search.spec.ts）

- 搜索 "C++ & .NET"
- 使用 `page.waitForResponse()` 捕获请求
- 断言请求 URL 中 query 参数正确 URL 编码
- 断言无前端 JavaScript 错误（监听 `page.on('pageerror')`）

### 2. 编写「非常长的搜索查询」测试（search.spec.ts）

- 生成 200 字符的搜索词（如重复字符串）
- 填入搜索框
- 断言搜索正常执行，不截断不报错
- 断言 API 请求发出且包含完整查询

### 3. 编写「快速连续搜索（防抖竞态）」测试（search.spec.ts）

- 快速依次输入 "a"、"ab"、"abc"
- 等待最终搜索完成
- 断言页面显示的是 query="abc" 的结果
- 断言前面的请求被取消或结果被丢弃（通过最终 DOM 状态验证）

### 4. 编写「网络错误时显示错误状态」测试（search.spec.ts）

- 使用 `page.route()` mock 搜索 API 返回网络错误（`route.abort()`）
- 输入搜索词
- 断言页面显示错误状态（`.border-rose-200` 或类似错误样式）
- 断言错误文本可见

### 5. 编写「搜索框纯空格不触发搜索」测试（search.spec.ts）

- 监听网络请求，记录所有搜索 API 调用
- 在搜索框输入 "   "（纯空格）
- 等待一段时间（超过防抖延迟）
- 断言没有发出搜索 API 请求
- 断言保持初始状态

### 6. 编写「浏览器后退/前进导航」测试（search.spec.ts）

- 在搜索页面搜索 "test" 并等待结果
- 导航到 `/admin/`
- 点击浏览器后退（`page.goBack()`）
- 断言返回到搜索页面 `/`

### 7. 编写「Admin 认证过期」测试（admin.spec.ts）

- 使用 `addInitScript` 注入 API Key
- 访问 `/admin/` 并确认已认证
- 通过 `page.evaluate()` 清除 localStorage 中的 key
- 刷新页面
- 断言显示 API Key 对话框

## Verification

```bash
cd web && bunx playwright test --list
# 应列出所有测试，边界场景部分约 7 个用例
```
