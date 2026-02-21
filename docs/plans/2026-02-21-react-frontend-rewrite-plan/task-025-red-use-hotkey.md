# Task 025: 测试 useHotkey Hook（Cmd/Ctrl+K）

**depends-on**: task-004

## Description

为 useHotkey 自定义 Hook 创建失败测试用例。

## Execution Context

**Task Number**: 025 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 4 - Mac 用户按 Cmd+K 聚焦搜索框; 非 Mac 用户按 Ctrl+K 聚焦搜索框; 焦点已在搜索框时快捷键仍然正常响应

## Files to Modify/Create

- Create: `cli/src/hooks/use-hotkey.test.ts`

## Steps

### Step 1: Test Cmd+K triggers callback on Mac

- Mock navigator.platform 为 Mac
- 派发 keydown 事件 (metaKey=true, key="k")
- 回调应被调用

### Step 2: Test Ctrl+K triggers callback on non-Mac

- Mock navigator.platform 为 Win
- 派发 keydown 事件 (ctrlKey=true, key="k")

### Step 3: Test preventDefault is called

### Step 4: Test cleanup removes event listener on unmount

### Step 5: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-hotkey.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖 Mac 和非 Mac 平台
