# Task 007: 全链路验证与发布闸门

**depends-on**: task-002-search-bootstrap-fallback-impl.md, task-003-file-category-index-impl.md, task-004-instantsearch-results-impl.md, task-005-routing-refinement-impl.md, task-006-download-integration-impl.md

## Description

执行 React InstantSearch 直连方案的全链路验证、结果对比与灰度回滚闸门确认，确保新旧双栈都可控，且在验证不足时不会贸然默认开启。

## Execution Context

**Task Number**: 013 of 013
**Phase**: Verification
**Prerequisites**: 搜索配置引导、`file_category`、结果渲染、routing/refinement 与下载集成都已完成并通过对应 Green 任务

## BDD Scenario

```gherkin
Scenario: 页面启动时成功获取公开搜索配置
  Given 后端返回公开搜索配置 host、indexName、searchApiKey 和 instantsearchEnabled=true
  When 用户打开搜索页
  Then 前端应初始化 InstantSearch search client
  And 浏览器不应调用 AppService.AppSearch

Scenario: 公开搜索配置不可用时回退旧链路
  Given 后端返回 instantsearchEnabled=false 或公开搜索配置缺失
  When 用户打开搜索页
  Then 前端应回退到现有 AppSearch 链路
  And 页面仍可完成搜索与下载

Scenario: 文件分类筛选使用 file_category refinement
  Given 索引文档包含 file_category 字段并配置为 filterable
  When 用户选择 "文档" 分类筛选
  Then 搜索请求应携带对应 refinement
  And 结果总数应与筛选后的命中数一致
  And 页面不应再使用本地 items.filter 进行分类裁剪

Scenario: query、page 和分类筛选可从 URL 恢复
  Given 用户已经在搜索页产生 query、page 和 file_category 状态
  When 用户刷新页面或通过分享链接重新打开
  Then 搜索页应从 URL 恢复相同的 InstantSearch 状态
  And 用户无需再次手动输入

Scenario: 命中结果名称显示高亮
  Given Meilisearch 返回带有 _formatted.name 的 hits
  When 结果卡片渲染文件名称
  Then 页面应展示高亮后的名称
  And 未命中高亮的结果应展示原始名称

Scenario: 搜索结果下载仍通过 AppDownloadURL
  Given 用户在搜索结果中点击下载按钮
  When 前端发起下载动作
  Then 前端应调用 AppService.AppDownloadURL
  And 不应尝试从 Meilisearch 响应中直接获取下载地址

Scenario: 直连链路异常时可切回 AppSearch
  Given 预发或线上发现直连 Meilisearch 存在不可接受的问题
  When 运维关闭 instantsearchEnabled 开关
  Then 前端应切回 AppSearch 搜索链路
  And 下载与页面其他功能保持可用
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`
- Modify: `web/e2e/fixtures/seed.ts`
- Modify: `tasks/todo.md`（回填执行结果）

## Steps

### Step 1: Verify Functional Gates

- 汇总并执行 Go 单测、前端单测、搜索页 E2E、下载 E2E 与 fallback 场景验证。
- 确认新旧双栈在开关开启/关闭下都能完成搜索与下载。

### Step 2: Compare New vs Legacy Behavior

- 对比新旧链路的命中总数、前 10 条结果、高亮输出、空态与错误态。
- 记录可接受差异与不可接受差异，尤其关注未复刻 `preprocessQuery()` / `All -> Last` fallback 带来的召回变化。

### Step 3: Define Rollout and Rollback Gates

- 明确默认开启前必须满足的通过项、监控项、人工验收项与回滚触发条件。
- 保留 `instantsearchEnabled` 作为发布闸门，未达标前不得移除 `AppSearch`。

### Step 4: Final Verification Sign-off

- 执行 Docker 冒烟与 Playwright 长链路回归。
- 在 `tasks/todo.md` 回填验证结果、差异结论与灰度建议。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./... -count=1
cd web && bun vitest run
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
git diff --check
```

## Success Criteria

- Go / 前端 / smoke / E2E 全部通过。
- 已完成新旧链路结果对比，并有明确差异结论。
- 已定义 `instantsearchEnabled` 的灰度启用与回滚条件。
- 默认启用前不存在未解释的高风险结果偏差或下载回归。
