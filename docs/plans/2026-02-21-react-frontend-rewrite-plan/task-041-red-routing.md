# Task 041: 测试路由导航（搜索页、管理页、404、search params）

**depends-on**: task-003, task-004, task-030, task-040

## Description

为路由导航创建集成测试。

## Execution Context

**Task Number**: 041 of 046
**Phase**: Integration (Red)
**Prerequisites**: 搜索页和管理页已实现，TanStack Router 已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 7 - 访问 /app 渲染搜索页; 访问 /app/admin 渲染管理页; 访问未定义路由显示 404; 搜索页导航到管理页; 管理页导航回搜索页; 搜索关键词通过 URL search params 持久化

## Files to Modify/Create

- Create: `cli/src/router.test.tsx`

## Steps

### Step 1: Test /app renders SearchPage

### Step 2: Test /app/admin renders AdminPage

### Step 3: Test /app/unknown renders 404

### Step 4: Test navigation between pages (Link click)

### Step 5: Test search params — /app?query=MX40 → 搜索框预填充

### Step 6: Test search updates URL params

### Step 7: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/router.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖所有路由场景
