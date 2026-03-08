# Task 004: [IMPL] InstantSearch 结果列表与高亮 (GREEN)

**depends-on**: task-003-file-category-index-impl.md, task-004-instantsearch-results-test.md

## Description

实现基于官方 InstantSearch hooks 的结果列表、高亮适配与 InfiniteHits 驱动的分页累积，同时保留现有页面视觉壳层与 `FileCard` 资产。

## Execution Context

**Task Number**: 008 of 013
**Phase**: Search Migration
**Prerequisites**: `task-003-file-category-index-impl.md` 与 `task-004-instantsearch-results-test.md` 已完成

## BDD Scenario

```gherkin
Scenario: 输入关键字后触发直连 Meilisearch 搜索
  Given 搜索页已成功初始化 InstantSearch
  When 用户输入 "report" 并提交搜索
  Then 浏览器应向 Meilisearch 发起搜索请求
  And 结果列表应展示 hits
  And 状态文案应基于 InstantSearch 返回的结果数量更新

Scenario: InfiniteHits 驱动无限滚动
  Given 当前查询已有第一页结果且存在下一页
  When 用户滚动到结果列表底部
  Then 前端应通过 InfiniteHits 继续加载下一批 hits
  And 已加载结果应继续可见

Scenario: 命中结果名称显示高亮
  Given Meilisearch 返回带有 _formatted.name 的 hits
  When 结果卡片渲染文件名称
  Then 页面应展示高亮后的名称
  And 未命中高亮的结果应展示原始名称
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/file-card.tsx`
- Create: `web/src/components/search-results.tsx`
- Create: `web/src/lib/meili-hit-adapter.ts`
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/components/file-card.test.tsx`
- Create: `web/src/components/search-results.test.tsx`
- Create: `web/src/lib/meili-hit-adapter.test.ts`

## Steps

### Step 1: Introduce Hit Adapter

- 建立 Meilisearch hit 到现有 UI 结构的映射层，统一名称、高亮、时间、大小与下载所需标识字段。
- 将高亮处理收敛在适配层或结果渲染边界，避免页面内散落 HTML 处理逻辑。

### Step 2: Replace Legacy Results Ownership

- 在 InstantSearch enabled 分支中，用官方 hooks 接管结果列表、计数、空态与加载态。
- 移除旧 `useInfiniteQuery + mergePages` 在新链路中的结果拥有权，避免双状态并存。

### Step 3: Implement InfiniteHits Rendering

- 用官方 InfiniteHits 能力承载下一页加载与结果累积。
- 保留现有结果卡片、空态、骨架屏与状态文案的视觉壳层。

### Step 4: Verify Green

- 运行 task-004 新增测试并确认通过。
- 回归搜索页与结果卡片相关现有测试，确认高亮与无限滚动接入后无回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/file-card.test.tsx src/components/search-results.test.tsx src/lib/meili-hit-adapter.test.ts
cd web && bun vitest run
```

## Success Criteria

- 搜索结果已由 InstantSearch hits 驱动，而非旧 `AppSearch` 列表。
- InfiniteHits 可累积展示下一页结果。
- 高亮名称只在 `name` 字段渲染，并可回退原始名称。
- task-004 新增测试通过。
