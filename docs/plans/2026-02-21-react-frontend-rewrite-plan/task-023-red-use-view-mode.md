# Task 023: 测试 useViewMode Hook（Hero/Docked 切换 + View Transition）

**depends-on**: task-004

## Description

为 useViewMode 自定义 Hook 创建失败测试用例。测试 Hero/Docked 模式切换和 View Transition API 封装。

## Execution Context

**Task Number**: 023 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 3 - 搜索触发 Hero → Docked 过渡; 清空搜索触发 Docked → Hero 过渡; 浏览器不支持 View Transition 时直接切换

## Files to Modify/Create

- Create: `cli/src/hooks/use-view-mode.test.ts`

## Steps

### Step 1: Test initial mode is Hero

### Step 2: Test setDocked(true) switches to Docked

### Step 3: Test setDocked(false) switches back to Hero

### Step 4: Test View Transition is called when supported

- Mock document.startViewTransition → 验证被调用

### Step 5: Test fallback when View Transition not supported

- 不提供 document.startViewTransition → 切换仍然生效（无动画）

### Step 6: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-view-mode.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖模式切换和 View Transition 降级
