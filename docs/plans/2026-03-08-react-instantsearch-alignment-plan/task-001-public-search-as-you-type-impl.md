# Task 001: [IMPL] public 输入即搜与立即提交 (GREEN)

**depends-on**: task-001-public-search-as-you-type-test.md

## Description

修正 public 搜索输入链路，让 public InstantSearch 恢复为 search-as-you-type，同时保留 Enter / 搜索按钮的立即触发能力，并避免重新引入双状态拥有权。

## Execution Context

**Task Number**: 002 of 007
**Phase**: Input Behavior Alignment
**Prerequisites**: `task-001-public-search-as-you-type-test.md` 已完成并稳定失败

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

- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/search-input.tsx`（仅当需要补充立即触发或无障碍行为时）
- Modify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Restore Official Input Semantics

- 调整 public 分支的输入处理，使输入变化在 debounce 后驱动搜索，而不是只更新本地值。
- 保持 `InstantSearch` 仍是 query / routing 的唯一真相源，不把 public 结果重新交还给本地状态机。

### Step 2: Preserve Immediate Submit Behavior

- 保留 Enter 与“搜索”按钮作为立即触发当前查询的加速路径。
- 确保立即触发会正确处理待执行 debounce，避免重复请求或错位状态。

### Step 3: Verify Green

- 运行 task-001 新增测试并确认通过。
- 回归搜索页现有 public / legacy 分支测试，确保 bootstrap、clear、routing 恢复与 fallback 行为不回退。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
cd web && bun vitest run
```

## Success Criteria

- public 搜索恢复为 search-as-you-type。
- Enter / 搜索按钮仍可立即触发当前查询。
- public 分支未重新引入本地搜索状态机。
- task-001 新增测试通过。
