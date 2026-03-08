# Task 002: 验证 public URL 与结果回归闸门

**depends-on**: task-001-file-category-refinement-token-impl.md

## Description

在实现修复后执行聚焦验证，确认 public 搜索的文件分类筛选已恢复，并且没有破坏既有的 search-as-you-type、默认过滤基线、URL 恢复和 legacy fallback 闸门。该任务只负责验证与结果记录，不再扩展实现范围。

## Execution Context

**Task Number**: 3 of 3
**Phase**: Testing
**Prerequisites**: `SearchFilters` 已切换到真实 refinement token 驱动，相关单测已转绿

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

- Modify: `tasks/todo.md`
- Verify: `web/src/components/search-filters.test.tsx`
- Verify: `web/src/components/search-page.test.tsx`
- Verify: `web/e2e/tests/search.spec.ts`

## Steps

### Step 1: Focused Verification
- 运行前端聚焦测试，覆盖：
  - `SearchFilters` token 驱动
  - `SearchPage` 中 public 搜索 URL / refinement / 结果数量
- 如环境允许，运行 public 搜索相关 Playwright 用例，确认“图片/视频/文档”等筛选在浏览器端可见恢复

### Step 2: Full Regression Gate
- 运行前端全量 Vitest
- 若本轮修改影响到容器链路或浏览器 E2E，再补跑搜索相关 E2E 或按项目最小回归链执行

### Step 3: Record Review
- 在 `tasks/todo.md` 中补充本轮 Review：
  - 根因
  - 改动文件
  - 验证命令与结果
  - 是否发现新的阻塞级差异

## Verification Commands

```bash
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun vitest run src/components/search-filters.test.tsx src/components/search-page.test.tsx
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun vitest run
cd /var/tmp/vibe-kanban/worktrees/caa8-meilisearch-inst/npan-indexer/web && bun playwright test --grep search
```

## Success Criteria

- public 文件分类筛选回归已被自动化验证覆盖
- 未出现默认过滤、URL 恢复、legacy fallback 的新增回归
- `tasks/todo.md` 已记录本轮 Review，方便后续追踪
