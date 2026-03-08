# Task 005: [TEST] routing 与 refinement 分类筛选 (RED)

**depends-on**: task-003-file-category-index-impl.md, task-004-instantsearch-results-impl.md

## Description

为 routing 与 `file_category` refinement 补充失败测试，锁定 query、page、分类筛选的 URL 恢复、浏览器前进后退恢复，以及非法筛选值回退默认分类等行为。该任务不实现生产代码。

## Execution Context

**Task Number**: 009 of 013
**Phase**: Search Migration
**Prerequisites**: `task-003-file-category-index-impl.md` 与 `task-004-instantsearch-results-impl.md` 已完成，且搜索结果已切换为 InstantSearch 状态模型

## BDD Scenario

```gherkin
Scenario: 文件分类筛选使用 file_category refinement
  Given 索引文档包含 file_category 字段并配置为 filterable
  When 用户选择 "文档" 分类筛选
  Then 搜索请求应携带对应 refinement
  And 结果总数应与筛选后的命中数一致
  And 页面不应再使用本地 items.filter 进行分类裁剪

Scenario: 非法 URL 筛选值应回退默认分类
  Given 用户直接访问带有非法筛选参数的搜索 URL
  When 搜索页初始化 routing 状态
  Then 当前 refinement 应回退到默认分类
  And 页面不应抛出异常

Scenario: query、page 和分类筛选可从 URL 恢复
  Given 用户已经在搜索页产生 query、page 和 file_category 状态
  When 用户刷新页面或通过分享链接重新打开
  Then 搜索页应从 URL 恢复相同的 InstantSearch 状态
  And 用户无需再次手动输入

Scenario: 浏览器前进后退可恢复搜索视图
  Given 用户依次切换了不同 query 或分类筛选
  When 用户使用浏览器后退或前进
  Then 搜索页应恢复到对应的 InstantSearch routing 状态
  And 结果列表应与 URL 保持一致
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`
- Create: `web/src/components/search-filters.test.tsx`
- Create: `web/src/lib/instantsearch-routing.test.ts`

## Steps

### Step 1: Verify Scenario

- 明确本任务覆盖状态所有权与 URL 同步，不再接受旧 `URLSearchParams + history.replaceState` 手工逻辑作为主路径。
- 确认 `query`、`page` 与 `file_category` 是需要恢复的核心公开搜索状态。

### Step 2: Implement Test (Red)

- 在组件测试中增加 URL 初始化恢复、非法分类回退默认值、切换 refinement 后 URL 同步的失败断言。
- 在 E2E 中增加浏览器后退/前进恢复搜索视图的断言，覆盖 query 与分类变化。
- 增加负向断言，锁定页面不应继续依赖本地 `items.filter(...)` 做分类裁剪。

### Step 3: Verify Red Failure

- 运行目标 Vitest 与相关 E2E 用例并确认失败。
- 失败原因应体现 routing 尚未交给 InstantSearch 或 refinement 尚未绑定 URL，而不是测试等待条件错误。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/search-filters.test.tsx src/lib/instantsearch-routing.test.ts
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright bunx playwright test web/e2e/tests/search.spec.ts --grep "筛选|URL|后退|前进"
```

## Success Criteria

- 新增 routing/refinement 用例稳定失败（Red）。
- 失败明确指向 URL 恢复、前进后退恢复或 refinement 绑定缺失。
- 覆盖 `query`、`page`、`file_category` 三类状态。
