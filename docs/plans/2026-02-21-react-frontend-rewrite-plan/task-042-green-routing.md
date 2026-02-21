# Task 042: 实现完整路由配置与页面集成

**depends-on**: task-041

## Description

完善 TanStack Router 路由配置，实现 404 页面、搜索参数持久化、导航链接，使 Task 041 测试通过。

## Execution Context

**Task Number**: 042 of 046
**Phase**: Integration (Green)
**Prerequisites**: Task 041 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 7 - 路由导航所有场景

## Files to Modify/Create

- Modify: `cli/src/routes/__root.tsx` — 添加导航栏 + 404 catch-all
- Modify: `cli/src/routes/index.tsx` — 添加 search params validation (Zod)
- Modify: `cli/src/main.tsx` — 确保 RouterProvider 完整配置

## Steps

### Step 1: Add 404 catch-all route

- TanStack Router 的 notFoundComponent

### Step 2: Implement search params with Zod validation

- `/app` 路由 `validateSearch: z.object({ query: z.string().optional().default("") })`
- 搜索页读取 query param 并预填充搜索框

### Step 3: Add navigation links

- 根布局中的搜索页 / 管理页导航
- 管理页中的返回搜索链接

### Step 4: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/router.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 041 所有测试通过
- URL search params 与搜索框双向同步
