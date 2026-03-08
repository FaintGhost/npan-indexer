# Task 003: [IMPL] public 默认过滤与 refinement 叠加 (GREEN)

**depends-on**: task-003-public-default-filters-test.md

## Description

在 public 搜索请求层统一注入 `type=file`、`is_deleted=false`、`in_trash=false` 默认过滤，并确保 `file_category` refinement 只在该基线上叠加，不把过滤逻辑退回结果渲染层。

## Execution Context

**Task Number**: 006 of 007
**Phase**: Filter Baseline Alignment
**Prerequisites**: `task-003-public-default-filters-test.md` 已完成并稳定失败

## BDD Scenario

```gherkin
Scenario: public 搜索始终带公开默认过滤
  Given 索引中同时存在 file、folder、in_trash=true 和 is_deleted=true 的匹配文档
  When 用户搜索 "report"
  Then 发往 Meilisearch 的请求应始终包含 type=file、in_trash=false 和 is_deleted=false
  And 返回结果中不应出现 folder、回收站或已删除文档
  And 结果总数应基于默认过滤后的结果

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

- Modify: `web/src/lib/meili-search-client.ts`
- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/search-filters.tsx`（仅当 refinement 组合方式需要轻微调整时）
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/components/search-filters.test.tsx`
- Modify: `web/src/lib/meili-search-client.test.ts`

## Steps

### Step 1: Inject Default Filter Baseline

- 通过 public 搜索请求层或官方配置入口统一注入 `type=file`、`is_deleted=false`、`in_trash=false`。
- 保证默认过滤对命中列表、总数、空态与 refinement 计数同时生效。

### Step 2: Compose User Refinement on Top

- 确保 `file_category` refinement 作为用户态过滤继续生效，并与默认过滤做叠加而不是相互覆盖。
- 删除或避免任何结果渲染层的本地二次分类裁剪。

### Step 3: Verify Green

- 运行 task-003 新增测试并确认通过。
- 回归搜索页现有 public / legacy 分支测试，确保搜索、清空、routing 与结果展示不回退。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/search-filters.test.tsx src/lib/meili-search-client.test.ts
cd web && bun vitest run
```

## Success Criteria

- public 请求始终带默认过滤基线。
- `file_category` refinement 会叠加在默认过滤之上，而不是替代默认过滤。
- 页面不再依赖结果渲染层本地裁剪来隐藏 folder / deleted / trash 文档。
