# Task 002: [TEST] query 预处理 adapter (RED)

## Description

为 public InstantSearch 请求层补充失败测试，锁定 legacy `preprocessQuery()` 的最小对齐语义，确保当前 public 请求仍直接透传原始 query 的偏差被精确暴露。该任务不实现生产代码。

## Execution Context

**Task Number**: 003 of 007
**Phase**: Query Semantics Alignment
**Prerequisites**: 已存在 public InstantSearch 分支与请求替身能力，可稳定捕获发往 Meilisearch 的实际 query

## BDD Scenario

```gherkin
Scenario: public 搜索应对 query 应用最小 legacy 预处理
  Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
  When 用户输入带扩展名、版本号或多词组合的查询并触发搜索
  Then 发往 Meilisearch 的 query 应与 legacy preprocess 规则等价
  And 搜索框展示值仍应保留用户原始输入
  And URL 中的 query 仍应保留用户原始输入
```

**Spec Source**: `../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/lib/meili-search-client.test.ts`
- Create: `web/src/lib/search-query-normalizer.test.ts`
- Modify: `web/src/tests/mocks/server.ts`（仅当需要稳定捕获 outbound query 时）

## Steps

### Step 1: Verify Scenario

- 确认本任务只覆盖 outbound query 语义对齐，不改变输入展示、routing 拥有权与结果渲染。
- 明确失败信号应直接指向“public 请求未做 legacy 等价预处理”。

### Step 2: Implement Test (Red)

- 为扩展名、版本号、多词组合等代表性查询补失败断言。
- 在页面级测试中同时断言：
  - 发往 Meilisearch 的 query 已被改写；
  - 输入框展示值保持用户原样输入；
  - URL 中的 query 保持用户原样输入。
- 使用模块替身 / fake timers / 请求捕获隔离网络，不依赖真实 Meilisearch。

### Step 3: Verify Red Failure

- 运行目标前端测试并确认新增用例失败。
- 失败原因必须指向“当前 public 请求直接透传原始 query”，而不是测试环境未初始化。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/lib/meili-search-client.test.ts src/lib/search-query-normalizer.test.ts
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 测试能稳定捕获发往 Meilisearch 的 outbound query。
- 失败信息明确表明当前 public 请求尚未对齐 legacy `preprocessQuery()` 语义。
