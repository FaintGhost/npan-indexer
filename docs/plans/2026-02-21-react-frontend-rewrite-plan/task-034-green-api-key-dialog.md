# Task 034: 实现 API Key 输入对话框组件

**depends-on**: task-033

## Description

实现 ApiKeyDialog 组件，使 Task 033 测试通过。

## Execution Context

**Task Number**: 034 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 033 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 5 - API Key 输入对话框场景

## Files to Modify/Create

- Create: `cli/src/components/api-key-dialog.tsx`

## Steps

### Step 1: Implement ApiKeyDialog

- Props: open, onSubmit(key), error, loading
- 使用 HTML `<dialog>` 元素或 overlay 模式
- password 类型输入框
- 确认按钮（loading 时 disabled）
- 错误消息显示
- 空输入客户端验证
- focus trap: 打开时聚焦输入框

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/api-key-dialog.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 033 所有测试通过
