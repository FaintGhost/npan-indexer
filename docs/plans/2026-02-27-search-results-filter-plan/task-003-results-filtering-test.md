# Task 003: [TEST] 去重后过滤、计数与空态一致性 (RED)

**depends-on**: task-002-file-category-impl.md

## Description

在搜索页测试中新增“去重后再过滤”的失败用例，验证列表、计数和空态在筛选后保持一致。

## Execution Context

**Task Number**: 005 of 012  
**Phase**: Core Features  
**Prerequisites**: 分类模块可用，搜索页基础测试可运行

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

- Modify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Verify Scenario
- 确认场景描述与设计文档一致。

### Step 2: Implement Test (Red)
- 构造多页+重复 `source_id` 测试数据（通过 MSW 替身）。
- 新增断言：筛选后列表长度、计数文案、空态展示应一致。

### Step 3: Verify Red Failure
- 运行目标测试并确认新增断言失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 新增过滤一致性测试稳定失败（Red）。
- 失败原因是行为不符合预期而非环境问题。
