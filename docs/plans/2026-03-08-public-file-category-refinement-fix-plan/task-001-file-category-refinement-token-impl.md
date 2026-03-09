# Task 001: 修复 refinement token 驱动实现

**depends-on**: task-001-file-category-refinement-token-test.md

## Description

修复 `SearchFilters` 组件，使其不再把固定 UI 分类值直接传给 `refine()`，而是改为使用 `useRefinementList().items` 提供的真实 refinement token。实现后，public 搜索在切换“文档/图片/视频/压缩包/其他”时应正确驱动 Meilisearch refinement，并保持 URL、已选态与结果数量一致。

## Execution Context

**Task Number**: 2 of 3
**Phase**: Core Features
**Prerequisites**: Red 测试已稳定失败并证明 token 误用是根因

## BDD Scenario

```gherkin
Scenario: file_category refinement 应叠加在默认过滤之上
  Given public 搜索默认过滤已经生效
  And 索引文档包含 file_category 字段并配置为 filterable
  When 用户选择 "文档" 分类筛选
  Then 搜索请求应同时携带默认过滤和 file_category refinement
  And 结果总数应与筛选后的命中数一致
  And 页面不应在结果渲染层再次做本地分类裁剪
```

**Spec Source**: `../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-filters.tsx`
- Modify: `web/src/lib/file-category.ts`
- Verify: `web/src/components/search-filters.test.tsx`
- Verify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Re-read Current Pattern
- 重新核对 `SearchFilters` 当前如何从 `useCurrentRefinements` 推导 active filter，以及如何调用 `refine()`
- 明确 legacy 分支位于 `web/src/routes/index.lazy.tsx`，本任务不能误改 legacy 本地筛选逻辑

### Step 2: Implement Logic (Green)
- 在 `SearchFilters` 中读取 `useRefinementList().items`
- 建立“固定 UI 分类值”到“当前 refinement item token”的稳定映射，避免直接把 UI 值传给 `refine()`
- 更新已选态推导逻辑，确保能够从当前 refinement 状态稳定恢复到 `SearchFilter` 枚举值
- 保持“全部”语义为移除当前 `file_category` refinement，而不是写入额外默认过滤
- 保持默认过滤基线继续由请求适配层负责，不在结果列表层新增本地裁剪

### Step 3: Verify & Refactor
- 先运行 Red 测试，确认它们转绿
- 再运行相关 public 搜索测试，确认 URL、请求断言与结果断言无回归
- 如出现类型收窄问题，使用显式映射/守卫解决，不用 `any` 或断言兜底

## Verification Commands

```bash
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun vitest run src/components/search-filters.test.tsx src/components/search-page.test.tsx
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun vitest run
```

## Success Criteria

- 选择任一非“全部”分类时，`refine()` 使用真实 token 而不是固定 UI 值
- public 搜索结果数量与命中内容按分类筛选正确变化
- URL 中 `file_category`、按钮选中态、请求 refinement 三者一致
- 前端全量 Vitest 通过
