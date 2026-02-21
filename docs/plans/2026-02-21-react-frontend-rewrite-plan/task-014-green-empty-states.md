# Task 014: 实现骨架屏与空状态组件

**depends-on**: task-013

## Description

实现 InitialState、NoResultsState、ErrorState、SkeletonCard 组件，使 Task 013 测试通过。视觉样式匹配现有 HTML。

## Execution Context

**Task Number**: 014 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 013 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 初始空状态; 无结果; 错误状态; Feature 3 - 骨架屏

## Files to Modify/Create

- Create: `cli/src/components/empty-state.tsx` — InitialState, NoResultsState, ErrorState
- Create: `cli/src/components/skeleton-card.tsx` — SkeletonCard

## Steps

### Step 1: Implement empty state components — 参考现有 HTML UIStates 常量的样式

### Step 2: Implement SkeletonCard — 参考现有 HTML renderSkeleton 函数的样式

### Step 3: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/empty-state.test.tsx src/components/skeleton-card.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 013 所有测试通过
- 组件使用 Tailwind 类与现有 HTML 样式一致
