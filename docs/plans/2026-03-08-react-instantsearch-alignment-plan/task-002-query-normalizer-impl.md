# Task 002: [IMPL] query 预处理 adapter (GREEN)

**depends-on**: task-002-query-normalizer-test.md

## Description

在 public 搜索请求层增加独立 query adapter，对齐 legacy `preprocessQuery()` 的最小语义，同时保持 InstantSearch 继续拥有输入显示与 URL 状态，不把兼容逻辑散落回组件事件处理函数。

## Execution Context

**Task Number**: 004 of 007
**Phase**: Query Semantics Alignment
**Prerequisites**: `task-002-query-normalizer-test.md` 已完成并稳定失败

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

- Modify: `web/src/lib/meili-search-client.ts`
- Create: `web/src/lib/search-query-normalizer.ts`
- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/lib/meili-search-client.test.ts`
- Modify: `web/src/lib/search-query-normalizer.test.ts`

## Steps

### Step 1: Add Minimal Query Adapter

- 将 legacy `preprocessQuery()` 的关键语义抽成独立、可测的前端 adapter。
- adapter 仅改写 outbound query，不直接改写输入框展示值、URL state 或结果拥有权。

### Step 2: Attach Adapter at Request Layer

- 将 public 搜索链路改为在请求发出前应用 adapter，而不是在组件事件处理函数内直接改写用户输入。
- 保持 InstantSearch 仍是 query / routing 的唯一真相源。

### Step 3: Verify Green

- 运行 task-002 新增测试并确认通过。
- 回归搜索页现有 public / legacy 分支测试，确保输入框显示、URL 恢复与结果渲染不回退。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/lib/meili-search-client.test.ts src/lib/search-query-normalizer.test.ts
cd web && bun vitest run
```

## Success Criteria

- public 请求会对 outbound query 应用 legacy 等价预处理。
- 输入框展示值与 URL 中的 query 保持用户原始输入。
- query 兼容逻辑保持在独立 adapter / 请求层，不重新散落进组件输入处理。
