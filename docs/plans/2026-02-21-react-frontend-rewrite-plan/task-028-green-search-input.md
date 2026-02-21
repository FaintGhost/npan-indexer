# Task 028: 实现搜索输入组件

**depends-on**: task-027

## Description

实现 SearchInput 组件，使 Task 027 测试通过。

## Execution Context

**Task Number**: 028 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 027 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 清空输入; Feature 4 - 快捷键徽章

## Files to Modify/Create

- Create: `cli/src/components/search-input.tsx`

## Steps

### Step 1: Implement SearchInput

- Props: value, onChange, onSubmit, onClear, ref (React 19 — 直接 prop)
- 搜索图标（左侧）
- 清空按钮（右侧，仅 value 非空时显示）
- 快捷键徽章（右侧，仅 value 为空时显示，根据 isMac 显示 ⌘K / Ctrl K）
- Enter 键触发 onSubmit
- 清空按钮触发 onClear
- 样式匹配现有 HTML

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/search-input.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 027 所有测试通过
- 使用 React 19 ref-as-prop 模式（不用 forwardRef）
