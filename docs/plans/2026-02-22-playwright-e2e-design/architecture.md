# Architecture

## Docker Compose CI 架构

### 服务拓扑

```
docker-compose.ci.yml
├── meilisearch-ci    (getmeili/meilisearch:v1.35.1)
│   └── healthcheck: curl/wget http://127.0.0.1:7700/health
├── npan-ci           (build from Dockerfile)
│   ├── depends_on: meilisearch (healthy)
│   └── healthcheck: wget http://127.0.0.1:1323/healthz
└── playwright-ci     (mcr.microsoft.com/playwright:v1.52.0-noble)
    ├── depends_on: npan (healthy)
    ├── profiles: [e2e]         # 默认不启动
    ├── ipc: host               # 防止 Chromium OOM
    ├── init: true              # PID 1 信号处理
    └── volumes: ./web:/web     # 挂载测试代码
```

### 执行流程

```
make smoke-test                     make e2e-test
┌───────────────────┐               ┌───────────────────────────────┐
│ 1. build npan     │               │ 1. build npan                 │
│ 2. start meili    │               │ 2. start meili + npan         │
│ 3. start npan     │               │ 3. run smoke tests            │
│ 4. run smoke.sh   │               │ 4. docker compose run         │
│ 5. cleanup        │               │    --profile e2e playwright   │
└───────────────────┘               │ 5. cleanup                    │
                                    └───────────────────────────────┘
```

### Playwright 容器配置

```yaml
playwright:
  image: mcr.microsoft.com/playwright:v1.52.0-noble
  container_name: playwright-ci
  depends_on:
    npan:
      condition: service_healthy
  ipc: host
  init: true
  working_dir: /web
  volumes:
    - ./web:/web
  environment:
    BASE_URL: http://npan:1323
    MEILI_HOST: http://meilisearch:7700
    MEILI_API_KEY: ci-test-meili-key-5678
    MEILI_INDEX: npan_items
    E2E_ADMIN_API_KEY: ci-test-admin-api-key-1234
    CI: "true"
  command: sh -c "npm install --frozen-lockfile 2>/dev/null; npx playwright test"
  profiles:
    - e2e
```

**注意**：容器内使用 `npx` 而非 `bun`，因为 Playwright 官方镜像基于 Ubuntu Noble，预装 Node.js 但不含 bun。`npm install` 在有 `package-lock.json` 或 `bun.lock` 时可能需要适配。替代方案是在 Makefile 中先 `bun install`，然后挂载 `node_modules`。

### 替代方案：宿主机安装 + 容器运行

```makefile
e2e-test:
  @cleanup() { docker compose -f docker-compose.ci.yml down --volumes; }; \
  trap cleanup EXIT; \
  docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120 && \
  BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh && \
  cd web && bun install && BASE_URL=http://localhost:11323 bun run e2e
```

此方案在 GitHub Actions 中需要额外步骤 `bunx playwright install --with-deps chromium`。

## 项目结构

```
web/
├── e2e/
│   ├── fixtures/
│   │   ├── auth.ts          # addInitScript 注入 localStorage
│   │   └── seed.ts          # Meilisearch API 播种 + 等待索引
│   ├── pages/
│   │   ├── search-page.ts   # getByPlaceholder, getByRole 定位
│   │   └── admin-page.ts    # API Key Dialog, 同步控制定位
│   └── tests/
│       ├── search.spec.ts   # 搜索 + 下载 + 边界场景
│       └── admin.spec.ts    # 认证 + 同步生命周期
├── playwright.config.ts
├── package.json             # +@playwright/test, +e2e script
└── .gitignore               # +playwright-report/, +test-results/
```

### Page Object Model

两个页面足够简单，使用轻量 POM：

```typescript
// search-page.ts
export class SearchPage {
  readonly searchInput: Locator     // getByPlaceholder(/输入文件名/)
  readonly searchButton: Locator    // getByRole('button', { name: '搜索' })
  readonly clearButton: Locator     // getByRole('button') in input container
  readonly resultArticles: Locator  // page.locator('article')
  readonly sentinel: Locator        // 无限滚动哨兵

  async search(query: string) { ... }
  async searchImmediate(query: string) { ... }
  async waitForResults() { ... }
  async scrollToLoadMore() { ... }
}
```

```typescript
// admin-page.ts
export class AdminPage {
  readonly apiKeyInput: Locator     // input[type="password"]
  readonly submitButton: Locator    // getByRole('button', { name: /确认/ })
  readonly startSyncButton: Locator
  readonly cancelSyncButton: Locator
  readonly modeButtons: { auto, full, incremental }

  async authenticate(key: string) { ... }
  async selectMode(mode: string) { ... }
  async startSync() { ... }
  async cancelSync() { ... }
}
```

## Playwright 配置

```typescript
// playwright.config.ts
export default defineConfig({
  testDir: './e2e/tests',
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  timeout: 30_000,
  expect: { timeout: 8_000 },

  reporter: process.env.CI
    ? [['list'], ['html', { open: 'never' }]]
    : [['html', { open: 'on-failure' }]],

  use: {
    baseURL: process.env.BASE_URL ?? 'http://localhost:5173',
    headless: !!process.env.CI,
    viewport: { width: 1280, height: 720 },
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
    video: process.env.CI ? 'retain-on-failure' : 'off',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],

  webServer: process.env.CI ? undefined : {
    command: 'bun run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
  },
})
```

## GitHub Actions

```yaml
e2e-test:
  needs: [smoke-test]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: docker/setup-buildx-action@v3

    - name: Start app services
      run: docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120

    - name: Run E2E tests
      run: |
        docker compose -f docker-compose.ci.yml \
          run --rm \
          -e BASE_URL=http://npan:1323 \
          -e MEILI_HOST=http://meilisearch:7700 \
          -e MEILI_API_KEY=ci-test-meili-key-5678 \
          -e MEILI_INDEX=npan_items \
          -e E2E_ADMIN_API_KEY=ci-test-admin-api-key-1234 \
          -e CI=true \
          playwright

    - name: Upload artifacts
      if: ${{ !cancelled() }}
      uses: actions/upload-artifact@v4
      with:
        name: playwright-report-${{ github.run_id }}
        path: |
          web/playwright-report/
          web/test-results/
        retention-days: 7

    - name: Export logs on failure
      if: failure()
      run: docker compose -f docker-compose.ci.yml logs

    - name: Cleanup
      if: always()
      run: docker compose -f docker-compose.ci.yml down --volumes
```

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `web/package.json` | 修改：添加 `@playwright/test`、`e2e` 脚本 |
| `web/playwright.config.ts` | 新建 |
| `web/e2e/fixtures/auth.ts` | 新建 |
| `web/e2e/fixtures/seed.ts` | 新建 |
| `web/e2e/pages/search-page.ts` | 新建 |
| `web/e2e/pages/admin-page.ts` | 新建 |
| `web/e2e/tests/search.spec.ts` | 新建 |
| `web/e2e/tests/admin.spec.ts` | 新建 |
| `docker-compose.ci.yml` | 修改：添加 playwright service |
| `Makefile` | 修改：添加 e2e-test target |
| `.github/workflows/ci.yml` | 修改：添加 e2e-test job |
| `.gitignore` | 修改：添加 playwright 产物 |
