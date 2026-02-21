# Task 033: 测试 API Key 输入对话框组件

**depends-on**: task-004

## Description

为 ApiKeyDialog 组件创建失败测试用例。

## Execution Context

**Task Number**: 033 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 - 首次访问弹出输入对话框; 输入有效 Key 后关闭; 无效 Key 显示错误; 空输入显示验证错误

## Files to Modify/Create

- Create: `cli/src/components/api-key-dialog.test.tsx`

## Steps

### Step 1: Test dialog renders with password input and confirm button

### Step 2: Test empty submit shows validation error

### Step 3: Test valid submit calls onSubmit with key value

### Step 4: Test error message display

### Step 5: Test dialog has focus trap (Escape 关闭)

### Step 6: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/api-key-dialog.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖对话框所有交互
