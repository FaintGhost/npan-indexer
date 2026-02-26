# Task 005: [IMPL] 清空搜索与筛选状态联动 (GREEN)

**depends-on**: task-005-clear-search-reset-test.md

## Description

实现清空搜索动作对 query、结果区和 URL 参数的一致化重置，满足“返回初始视图”的交互预期。

## Execution Context

**Task Number**: 010 of 012  
**Phase**: Integration  
**Prerequisites**: `task-005-clear-search-reset-test.md` 已 Red

## BDD Scenario

```gherkin
Scenario: 清空搜索后回到初始视图
  Given 用户已有 query 与 ext 筛选状态
  When 用户点击清空搜索
  Then 输入框应清空
  And 结果区应回到初始状态
  And URL 中搜索相关参数应被清理或归一化
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`

## Steps

### Step 1: Implement Logic (Green)
- 在清空事件处理中补齐 URL 参数重置策略。
- 确保清空后页面状态与初始态语义一致。

### Step 2: Verify Green
- 运行 task-005 测试并通过。

### Step 3: Regression Check
- 运行搜索页相关测试，确认清空按钮已有行为无回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 清空联动场景从 Red 变 Green。
- URL 状态与 UI 状态保持一致。
