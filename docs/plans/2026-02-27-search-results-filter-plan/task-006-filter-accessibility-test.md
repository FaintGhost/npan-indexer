# Task 006: [TEST] 筛选控件可访问性语义 (RED)

**depends-on**: (none)

## Description

补充筛选控件可访问性失败测试，覆盖语义角色与选中态可感知要求。

## Execution Context

**Task Number**: 011 of 012  
**Phase**: Refinement  
**Prerequisites**: 可访问性测试套件可运行

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

- Modify: `web/src/tests/accessibility.test.tsx`
- (Optional) Modify: `web/src/components/search-page.test.tsx`

## Steps

### Step 1: Verify Scenario
- 核对可访问性场景验收语义（角色、选中态）。

### Step 2: Implement Test (Red)
- 新增断言：筛选组存在且具备 radiogroup/radio 语义。
- 新增断言：默认选中项与切换后选中项 aria 状态正确。

### Step 3: Verify Red Failure
- 执行可访问性测试并确认新增用例失败。

## Verification Commands

```bash
cd web && bun vitest run src/tests/accessibility.test.tsx
```

## Success Criteria

- 可访问性新增用例稳定 Red。
- 失败信息可指向语义缺失点。
