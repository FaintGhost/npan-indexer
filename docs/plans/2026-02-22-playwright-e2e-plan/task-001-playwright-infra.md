# Task 001: Playwright 项目基础设施

**depends-on**: (none)

## Objective

安装 Playwright 依赖，创建配置文件、测试 fixtures（Meilisearch 播种 + Auth 注入）和 Page Object Models。

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/package.json` | 修改：添加 `@playwright/test` 依赖、`e2e`/`e2e:ui`/`e2e:debug` 脚本 |
| `web/playwright.config.ts` | 新建 |
| `web/e2e/fixtures/seed.ts` | 新建 |
| `web/e2e/fixtures/auth.ts` | 新建 |
| `web/e2e/pages/search-page.ts` | 新建 |
| `web/e2e/pages/admin-page.ts` | 新建 |
| `.gitignore` | 修改：添加 `web/playwright-report/`、`web/test-results/` |

## Steps

### 1. 安装 Playwright

- 运行 `cd web && bun add -D @playwright/test`
- 运行 `cd web && bunx playwright install chromium` 安装浏览器
- 在 `package.json` 添加 scripts:
  - `"e2e": "playwright test"`
  - `"e2e:ui": "playwright test --ui"`
  - `"e2e:debug": "playwright test --debug"`

### 2. 创建 playwright.config.ts

按设计文档 architecture.md 的配置创建，关键点：
- `testDir: './e2e/tests'`
- `workers: 1`, `fullyParallel: false`
- `retries: process.env.CI ? 1 : 0`
- `timeout: 30_000`, `expect.timeout: 8_000`
- `baseURL` 从 `process.env.BASE_URL` 读取，默认 `http://localhost:5173`
- CI 中 reporter 用 `list` + `html`，本地用 `html` (open: on-failure)
- `screenshot: 'only-on-failure'`, `trace: 'on-first-retry'`
- 仅 chromium project
- 本地开发时自动启动 `bun run dev` 作为 webServer

### 3. 创建 Meilisearch 播种 fixture (seed.ts)

- 从环境变量读取 `MEILI_HOST`、`MEILI_API_KEY`、`MEILI_INDEX`（带默认值）
- 定义 `TEST_DOCUMENTS` 数组：3 条具名文档 + 35 条批量文档（test-file-000~034）
- 实现 `seedMeilisearch()`: 创建索引 → 批量插入文档 → 轮询等待索引完成
- 实现 `clearMeilisearch()`: 删除所有文档并等待完成
- 实现 `waitForMeiliTask(taskUid)`: 轮询 `/tasks/{uid}` 直到 succeeded/failed

### 4. 创建 Auth fixture (auth.ts)

- 从 `process.env.E2E_ADMIN_API_KEY` 读取（默认 `ci-test-admin-api-key-1234`）
- 扩展 `base.test` 提供 `authenticatedPage` fixture：
  - 使用 `context.addInitScript()` 注入 `localStorage['npan_admin_api_key']`
  - 导航到 `/admin/`，等待 API Key Dialog 不出现
- 导出 `test`、`expect`、`ADMIN_API_KEY`

### 5. 创建 Search Page Object (search-page.ts)

- `searchInput`: `getByPlaceholder(/输入文件名/)`
- `searchButton`: `getByRole('button', { name: '搜索' })`
- `resultArticles`: `page.locator('article')`
- `sentinel`: 无限滚动哨兵 `page.locator('.h-2').last()`
- 方法: `goto()`, `search(query)` (waitForResponse), `searchImmediate(query)` (click button), `waitForResults()`, `getResultCount()`, `scrollToLoadMore()`

### 6. 创建 Admin Page Object (admin-page.ts)

- `apiKeyInput`: `page.locator('input[type="password"]')`
- `submitButton`: `getByRole('button', { name: /确认/ })`
- `startSyncButton`: `getByRole('button', { name: /启动同步/ })`
- `cancelSyncButton`: `getByRole('button', { name: '取消同步' })`
- `modeButtons`: 自适应/全量/增量
- 方法: `goto()`, `submitApiKey(key)`, `injectApiKey(key)`, `selectMode(mode)`, `waitForAuthComplete()`

### 7. 更新 .gitignore

添加：`web/playwright-report/`、`web/test-results/`

## Verification

```bash
cd web && bunx playwright test --list
# 应显示 0 个测试（还没有 spec 文件），但不应报配置错误
```
