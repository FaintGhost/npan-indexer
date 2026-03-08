# Task 001: [TEST] public 输入即搜与立即提交 (RED)

## Description

为 public InstantSearch 分支补充失败测试，锁定“输入即搜”和“Enter / 搜索按钮立即触发”两类官方交互行为，确保当前 submit-only 偏差被精确暴露。该任务不实现生产代码。

## Execution Context

**Task Number**: 001 of 007
**Phase**: Input Behavior Alignment
**Prerequisites**: 已存在 public bootstrap、`InstantSearch` provider 与 `SearchInput` 壳层

## BDD Scenario

```gherkin
Scenario: 输入时自动触发 public 搜索
  Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
  When 用户输入 "report" 并停止输入约 280ms
  Then 浏览器应在未点击搜索按钮且未按 Enter 的情况下向 Meilisearch 发起搜索请求
  And 结果列表应展示 hits
  And 状态文案应基于 InstantSearch 返回的结果数量更新
  And URL 中的 query 应与当前搜索状态一致

Scenario: Enter 或搜索按钮可立即触发当前查询
  Given 搜索页已成功初始化 InstantSearch 且启用 public 搜索
  When 用户输入 "report" 后立即按 Enter 或点击搜索按钮
  Then 浏览器应立即向 Meilisearch 发起当前查询请求
  And 不必等待输入 debounce 到期
```

**Spec Source**: `../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/tests/mocks/server.ts`
- Modify: `web/src/tests/test-providers.tsx`（仅在测试挂载能力不足时）

## Steps

### Step 1: Verify Scenario

- 明确本任务只覆盖 public 搜索输入触发语义，不触及 query preprocess 与默认过滤。
- 确认失败信号应直接指向“当前 public 搜索仍是 submit-only”。

### Step 2: Implement Test (Red)

- 在搜索页测试中为 public 模式增加“停止输入后自动触发搜索”的断言。
- 使用 fake timers / module mock / 请求捕获隔离网络，确保能稳定统计搜索请求次数与触发时机。
- 增加“按 Enter 立即搜索”和“点击按钮立即搜索”两条断言，验证它们会跳过待触发的 debounce。
- 若现有测试辅助不足，仅补最小测试支撑，不修改生产逻辑。

### Step 3: Verify Red Failure

- 运行目标前端测试并确认新增用例失败。
- 失败原因必须指向当前 public 输入行为未在 `onChange` 期间触发搜索，而不是测试环境未初始化。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 测试使用替身隔离外部依赖，不依赖真实 Meilisearch。
- 失败信息明确表明当前 public 搜索不是 search-as-you-type。
