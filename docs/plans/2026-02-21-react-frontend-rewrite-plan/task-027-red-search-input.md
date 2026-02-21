# Task 027: 测试搜索输入组件（输入框 + 清空 + 快捷键徽章）

**depends-on**: task-004

## Description

为 SearchInput 组件创建失败测试用例。

## Execution Context

**Task Number**: 027 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 搜索框有文字时显示清空按钮，无文字时显示快捷键提示; 点击清空按钮恢复初始状态

## Files to Modify/Create

- Create: `cli/src/components/search-input.test.tsx`

## Steps

### Step 1: Test empty state shows keyboard shortcut badge

### Step 2: Test typing shows clear button, hides badge

### Step 3: Test clear button click clears input and calls onClear

### Step 4: Test input change calls onChange

### Step 5: Test Enter key calls onSubmit

### Step 6: Test ref forwarding for focus management

### Step 7: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/search-input.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖输入框的所有交互
