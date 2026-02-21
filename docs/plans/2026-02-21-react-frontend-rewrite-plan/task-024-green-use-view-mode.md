# Task 024: 实现 useViewMode Hook

**depends-on**: task-023

## Description

实现 useViewMode Hook，使 Task 023 测试通过。

## Execution Context

**Task Number**: 024 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 023 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 3 - Hero/Docked 过渡所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-view-mode.ts`

## Steps

### Step 1: Implement useViewMode

- 状态：isDocked (boolean)
- setDocked(docked): 通过 View Transition API 切换模式
- runViewTransition(callback): 封装 document.startViewTransition，不支持时直接执行 callback
- 返回 { isDocked, setDocked, runViewTransition }

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-view-mode.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 023 所有测试通过
