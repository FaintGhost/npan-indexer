# Task 001: 锁定 refinement token 回归测试

## Description

为 public 搜索的 `file_category` 筛选补充一组能体现真实 InstantSearch token 语义的测试，先让当前实现以明确、可解释的方式失败。测试要证明筛选按钮不能直接把固定的 `doc/image/video/...` 原始值传给 `refine()`，而必须使用 `useRefinementList().items` 提供的真实 refinement token。

## Execution Context

**Task Number**: 1 of 3
**Phase**: Testing
**Prerequisites**: 已确认问题只出现在 public InstantSearch 筛选链路，legacy 本地筛选链路无需改动

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

- Modify: `web/src/components/search-filters.test.tsx`
- Verify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Verify Scenario
- 确认计划中的场景对应 design 里的 Scenario 5
- 明确本任务只负责 Red 测试，不修改生产实现

### Step 2: Implement Test (Red)
- 在 `web/src/components/search-filters.test.tsx` 中把 `useRefinementList` mock 扩展为同时返回 `items` 与 `refine`
- 为固定 UI 筛选值和真实 refinement token 构造不同值的用例，例如 UI 仍选择 `doc`，但真实传给 `refine()` 的 token 来自 `items[].value`
- 让测试覆盖两类行为：
  - 从“全部”切换到某分类时，必须调用真实 token
  - 从已选分类切换到另一分类时，必须先用旧 token 取消，再用新 token 应用
- 如有必要，在 `web/src/components/search-page.test.tsx` 中补一个更贴近 public 搜索请求层的断言，证明请求里出现的是正确 refinement，而不是依赖渲染层本地过滤
- **Verification**: 运行针对 `search-filters` / `search-page` 的最小测试命令，并确认新增断言先失败且失败原因指向 token 误用，而不是 mock 配置错误

### Step 3: Handoff for Green
- 记录失败信息，明确实现任务需要从 `useRefinementList().items` 建立 UI 值到 refinement token 的映射

## Verification Commands

```bash
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun vitest run src/components/search-filters.test.tsx src/components/search-page.test.tsx
```

## Success Criteria

- 新测试能稳定复现当前 bug
- 失败原因直接指向 `SearchFilters` 对 refinement token 的使用错误
- 未引入与本场景无关的测试改动
