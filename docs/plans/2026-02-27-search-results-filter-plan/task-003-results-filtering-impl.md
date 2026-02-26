# Task 003: [IMPL] 去重后过滤、计数与空态一致性 (GREEN)

**depends-on**: task-003-results-filtering-test.md

## Description

在搜索页接入过滤流水线，使去重结果按筛选条件生成展示结果，并统一驱动计数、空态和状态文案。

## Execution Context

**Task Number**: 006 of 012  
**Phase**: Core Features  
**Prerequisites**: `task-003-results-filtering-test.md` 已 Red

## BDD Scenario

```gherkin
Scenario: 去重后再筛选保持数量一致性
  Given 两页结果中存在相同 source_id 的重复项
  When 页面合并分页并应用任一筛选
  Then 重复项应只显示一次
  And 计数文案应与实际显示条目一致

Scenario: 筛选为空时展示可理解的空态
  Given 搜索结果总量大于 0
  And 当前筛选下没有匹配条目
  When 页面渲染结果区
  Then 应展示筛选后空态提示
  And 不应展示错误态
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`

## Steps

### Step 1: Implement Logic (Green)
- 在 `items` 基础上派生 `filteredItems`。
- 将列表渲染与计数、空态、状态文案切换为以 `filteredItems` 为基准。
- 保持分页与请求逻辑不变。

### Step 2: Verify Green
- 运行 task-003 新增测试并通过。

### Step 3: Regression Check
- 运行搜索页现有测试，确保初始态、有结果、错误态等无回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 去重后过滤行为正确。
- 计数、空态、列表展示一致。
