# Task 005: [IMPL] routing 与 refinement 分类筛选 (GREEN)

**depends-on**: task-005-routing-refinement-test.md

## Description

实现 InstantSearch routing 与 `file_category` refinement UI，移除旧本地扩展名过滤拥有权，统一由 URL 与 InstantSearch state 驱动搜索状态。

## Execution Context

**Task Number**: 010 of 013
**Phase**: Search Migration
**Prerequisites**: `task-005-routing-refinement-test.md` 已完成且处于 Red

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

- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/lib/file-category.ts`
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`
- Create: `web/src/components/search-filters.tsx`
- Create: `web/src/components/search-filters.test.tsx`
- Create: `web/src/lib/instantsearch-routing.ts`
- Create: `web/src/lib/instantsearch-routing.test.ts`

## Steps

### Step 1: Centralize Routing State

- 让 query、page 与 refinement 统一交给 InstantSearch routing 管理。
- 清理旧 `query/activeQuery/activeFilter` 与手工 `replaceState/popstate` 主路径，避免双向同步竞态。

### Step 2: Implement Refinement UI

- 用基于 hooks 的分类筛选组件承载 `file_category` refinement。
- 保持现有“全部/文档/图片/视频/压缩包/其他”交互语义与可访问性结构。

### Step 3: Normalize URL Recovery Rules

- 为非法 URL 分类值定义回退默认分类规则。
- 确保刷新、分享链接、浏览器后退前进都能恢复一致搜索视图。

### Step 4: Verify Green

- 运行 task-005 新增测试并确认通过。
- 回归搜索页相关 Vitest 与 E2E，确认筛选结果、状态计数与 URL 状态一致。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/search-filters.test.tsx src/lib/instantsearch-routing.test.ts
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright bunx playwright test web/e2e/tests/search.spec.ts --grep "搜索流程|筛选|URL|后退|前进"
```

## Success Criteria

- URL 成为 `query`、`page`、`file_category` 的唯一真相源。
- 分类筛选走服务端 refinement，不再做本地 `items.filter(...)` 裁剪。
- 非法 URL 分类值可回退默认分类且不抛异常。
- task-005 新增测试通过。
