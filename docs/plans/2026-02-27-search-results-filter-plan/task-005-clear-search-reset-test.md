# Task 005: [TEST] 清空搜索与筛选状态联动 (RED)

**depends-on**: task-004-filter-switch-sync-impl.md

## Description

新增“清空搜索”联动行为失败测试，验证输入、结果区、URL 参数在清空动作后的重置一致性。

## Execution Context

**Task Number**: 009 of 012  
**Phase**: Integration  
**Prerequisites**: 筛选切换与 URL 同步已可运行

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

- Modify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Verify Scenario
- 确认清空联动场景存在且验收目标清晰。

### Step 2: Implement Test (Red)
- 新增从“有 query + 有 ext”到清空后的断言。
- 验证输入清空、视图回初始、URL 参数归一化。

### Step 3: Verify Red Failure
- 执行目标测试并确认失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 清空联动测试处于 Red 且失败原因明确。
