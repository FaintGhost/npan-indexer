# Task 003: [TEST] public 默认过滤与 refinement 叠加 (RED)

## Description

为 public 搜索默认过滤与 `file_category` refinement 叠加补充失败测试，锁定系统默认过滤必须在请求层生效，避免当前 public 搜索泄漏 folder / deleted / trash 文档，或把过滤逻辑退回结果渲染层。该任务不实现生产代码。

## Execution Context

**Task Number**: 005 of 007
**Phase**: Filter Baseline Alignment
**Prerequisites**: 已存在 public 搜索结果渲染链路、`file_category` refinement UI 与请求捕获能力

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

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/components/search-filters.test.tsx`
- Modify: `web/src/lib/meili-search-client.test.ts`
- Modify: `web/src/tests/mocks/server.ts`
- Modify: `web/src/tests/test-providers.tsx`（仅在现有挂载能力不足时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只覆盖请求层默认过滤与 refinement 叠加，不涉及排序或更深的相关性调优。
- 明确失败信号应直接指向“public 请求未注入默认过滤”或“refinement 未与默认过滤组合”。

### Step 2: Implement Test (Red)

- 为 public 搜索补“默认过滤始终存在”的失败断言，稳定捕获请求中的 filter 参数。
- 增加 `file_category` refinement 用例，断言它叠加在默认过滤之上，而不是替代默认过滤。
- 增加结果计数断言，确认总数来自请求层过滤后的 hits，而非结果渲染层本地裁剪。
- 使用模块替身 / MSW / 请求捕获隔离网络，不依赖真实 Meilisearch。

### Step 3: Verify Red Failure

- 运行目标前端测试并确认新增用例失败。
- 失败原因必须指向“当前 public 请求缺少默认过滤基线或叠加方式错误”，而不是测试数据准备不足。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/search-filters.test.tsx src/lib/meili-search-client.test.ts
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 测试可稳定断言默认过滤与 `file_category` refinement 的组合请求。
- 失败信息明确表明当前 public 请求层尚未建立默认过滤基线。
