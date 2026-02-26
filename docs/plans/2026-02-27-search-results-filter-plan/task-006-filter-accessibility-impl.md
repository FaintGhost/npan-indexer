# Task 006: [IMPL] 筛选控件可访问性语义 (GREEN)

**depends-on**: task-006-filter-accessibility-test.md

## Description

为筛选控件补齐可访问性语义与选中态表达，确保读屏与键盘操作符合单选组模式。

## Execution Context

**Task Number**: 012 of 012  
**Phase**: Refinement  
**Prerequisites**: `task-006-filter-accessibility-test.md` 已 Red

## BDD Scenario

```gherkin
Scenario: 筛选控件具备可访问语义
  Given 搜索页已渲染筛选控件
  When 用户通过键盘导航筛选项
  Then 控件应具备 radiogroup/radio 语义
  And 当前选中项应有明确 aria 状态
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/search-page.test.tsx`（如需补充键盘行为断言）

## Steps

### Step 1: Implement Logic (Green)
- 为筛选控件提供 radiogroup/radio 语义结构。
- 确保选中项具有明确状态表达并可通过键盘切换。

### Step 2: Verify Green
- 运行可访问性测试并确认通过。

### Step 3: Final Regression
- 运行本计划相关测试集，确保功能与可访问性均无回归。

## Verification Commands

```bash
cd web && bun vitest run src/tests/accessibility.test.tsx
cd web && bun vitest run src/components/search-page.test.tsx src/lib/file-category.test.ts
```

## Success Criteria

- 可访问性测试从 Red 变 Green。
- 计划内测试全绿，符合 BDD 验收。
