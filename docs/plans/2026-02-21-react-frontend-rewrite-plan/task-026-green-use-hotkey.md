# Task 026: 实现 useHotkey Hook

**depends-on**: task-025

## Description

实现 useHotkey Hook，使 Task 025 测试通过。

## Execution Context

**Task Number**: 026 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 025 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 4 - 键盘快捷键所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-hotkey.ts`

## Steps

### Step 1: Implement useHotkey

- 接收 key 和 callback 参数
- 检测 navigator.platform 判断 Mac/非 Mac
- useEffect 注册 keydown listener
- 匹配 metaKey (Mac) 或 ctrlKey (非 Mac) + 指定 key
- 调用 preventDefault + callback
- cleanup 移除 listener

### Step 2: 导出 isMac 工具函数

- 用于快捷键徽章文字判断

### Step 3: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-hotkey.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 025 所有测试通过
