# Playwright E2E Testing Design

## Context

npan 项目当前测试体系包含 Go 单元测试 (11 packages)、前端 Vitest 单元测试 (174 tests)、API 冒烟测试 (34 curl tests)。缺少浏览器级别的端到端测试，无法验证前后端联调的完整用户旅程。

## Requirements

1. 覆盖搜索页完整流程：输入 → 防抖/立即搜索 → 结果渲染 → 无限滚动 → 清空
2. 覆盖下载流程：点击下载 → 状态机 (idle → loading → success/error) → window.open
3. 覆盖管理后台全流程：API Key 认证 → 模式选择 → 启动/取消同步 → 进度展示
4. 覆盖边界场景：空结果、网络错误、特殊字符、快速连续搜索
5. Docker Compose CI 集成：Playwright 容器化，smoke-test 后自动运行
6. GitHub Actions 中失败时上传截图/trace 供调试

## Rationale

- **Playwright 而非 Cypress**：原生支持 Docker 官方镜像、`waitForResponse` API 更适合 SPA 防抖测试、bun 兼容性更好
- **独立容器而非嵌入 npan**：npan 基于 Alpine (~20MB)，Playwright 需要 ~1.5GB 浏览器依赖，职责分离
- **仅 Chromium**：CI 速度优先，覆盖主要用户，本地可选开 Firefox/WebKit
- **单 worker**：测试共享 Meilisearch 数据，避免并发竞争
- **`page.route()` mock 下载 API**：CI 中 `NPA_TOKEN=ci-dummy-token` 导致真实下载失败，mock 后可完整测试流程

## Detailed Design

### 项目结构

```
web/
├── e2e/
│   ├── fixtures/
│   │   ├── auth.ts          # localStorage API Key 注入
│   │   └── seed.ts          # Meilisearch 测试数据播种
│   ├── pages/
│   │   ├── search-page.ts   # 搜索页 Page Object
│   │   └── admin-page.ts    # 管理页 Page Object
│   └── tests/
│       ├── search.spec.ts   # 搜索 + 下载 E2E
│       └── admin.spec.ts    # 管理后台 E2E
├── playwright.config.ts     # Playwright 配置
└── package.json             # 添加 @playwright/test + e2e 脚本
```

### Docker Compose 架构

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│  meilisearch-ci │◄────│   npan-ci    │     │ playwright-ci   │
│  :7700          │     │  :1323       │◄────│ (Chromium)      │
│  (healthcheck)  │     │  (healthcheck)│     │ depends_on:     │
└─────────────────┘     └──────────────┘     │  npan: healthy  │
                                              └─────────────────┘
```

- `playwright` 服务使用 `profiles: [e2e]`，smoke-test 阶段不启动
- 容器内通过 Docker 网络 `http://npan:1323` 访问应用
- 必须设置 `ipc: host` 防止 Chromium OOM 崩溃

### 关键技术方案

| 场景 | 方案 |
|------|------|
| 280ms 防抖 | `waitForResponse('**/api/v1/app/search*')` 替代 `waitForTimeout` |
| 无限滚动 | `sentinel.scrollIntoViewIfNeeded()` + `expect(articles).toHaveCount(N)` |
| 下载拦截 | `page.route('**/download-url**')` mock 返回，验证按钮状态机 |
| Admin 认证 | `context.addInitScript()` 注入 localStorage，无需 reload |
| 数据播种 | `beforeAll` 中通过 Meilisearch API 插入 38 条测试文档 + 等待索引完成 |
| 竞态测试 | 快速连续输入 → 验证最终结果匹配最后一次查询 |

### CI Pipeline 变更

```
unit-test-go ─┐
unit-test-fe ─┼→ smoke-test → e2e-test
generate-chk ─┘
```

- `e2e-test` job 依赖 `smoke-test` 通过
- 失败时上传 `playwright-report/` 和 `test-results/` (retention: 7 days)
- 预拉取 Playwright 镜像加速

### 性能预期

| 指标 | 值 |
|------|-----|
| 测试用例数 | ~20 |
| 浏览器 | Chromium only |
| Workers | 1 |
| 预计 CI 耗时 | 60-90s |
| 重试 | CI 中 1 次 |
| 超时 | 测试 30s / 断言 8s |

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为场景和测试策略
- [Architecture](./architecture.md) - 系统架构和组件详情
- [Best Practices](./best-practices.md) - 安全、性能和代码质量指南
