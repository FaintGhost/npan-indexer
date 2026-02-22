# Task 005: 下载流程 E2E 测试

**depends-on**: Task 001

## Objective

编写文件下载流程的 E2E 测试，覆盖 BDD specs Feature 2 的所有场景。需要 mock 下载 API（CI 中 NPA_TOKEN 为 dummy）。

## BDD Scenarios Covered

- Scenario: 下载按钮初始状态
- Scenario: 点击下载显示加载状态
- Scenario: 下载成功后显示成功状态
- Scenario: 下载失败显示重试状态
- Scenario: 缓存的下载不重复请求 API
- Scenario: 多个文件可同时下载

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/e2e/tests/search.spec.ts` | 修改：在搜索 spec 中添加下载测试 describe block |

## Steps

### 1. 在 search.spec.ts 中添加下载测试 describe

下载测试需要搜索结果作为前提，因此放在 search.spec.ts 中，与搜索测试共享 Meilisearch 播种数据。

### 2. Mock 下载 API

所有下载测试开始前，使用 `page.route()` mock `/api/v1/app/download-url**`：
- 成功响应：`{ file_id: 1001, download_url: 'https://example.com/fake-download.pdf' }`
- 部分测试需要 mock 失败响应（502）

### 3. 拦截 window.open

使用 `page.addInitScript()` 拦截 `window.open`：
- 记录所有调用到 `(window as any).__openCalls` 数组
- 测试后通过 `page.evaluate()` 读取调用记录

### 4. 编写「下载按钮初始状态」测试

- 搜索并等待结果
- 断言每个结果卡片都有 "下载" 按钮
- 断言按钮为 idle 状态（可点击）

### 5. 编写「下载成功后显示成功状态」测试

- Mock 成功响应
- 拦截 window.open
- 点击第一个下载按钮
- 断言按钮文本变为 "获取中"（loading 状态）
- 等待 API 响应
- 断言按钮文本变为 "成功"（绿色）
- 断言 `window.open` 被调用且参数包含 download_url
- 等待按钮恢复为 "下载"

### 6. 编写「下载失败显示重试状态」测试

- Mock 下载 API 返回 502 错误
- 点击下载按钮
- 断言按钮文本变为 "重试"（红色）
- 断言按钮仍可点击

### 7. 编写「缓存的下载不重复请求 API」测试

- Mock 成功响应，记录 API 请求次数
- 点击同一文件的下载按钮，等待完成
- 再次点击同一文件的下载按钮
- 断言 API 请求只发出 1 次
- 断言 `window.open` 被调用了 2 次

### 8. 编写「多个文件可同时下载」测试

- Mock 成功响应
- 快速点击两个不同文件的下载按钮
- 断言两个按钮都进入 loading 状态
- 断言发出了 2 个 API 请求

## Verification

```bash
cd web && bunx playwright test search.spec.ts --list
# 应列出搜索 + 下载测试，下载部分约 5 个用例
```
