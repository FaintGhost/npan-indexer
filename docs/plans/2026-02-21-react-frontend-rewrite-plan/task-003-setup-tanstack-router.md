# Task 003: 配置 TanStack Router + 文件路由

**depends-on**: task-001

## Description

安装并配置 TanStack Router，设置文件路由约定、Vite 插件、basePath `/app`，创建根布局和基础路由骨架。

## Execution Context

**Task Number**: 003 of 046
**Phase**: Setup
**Prerequisites**: Task 001 项目已初始化

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 7 - 客户端路由（所有路由场景的基础）

## Files to Modify/Create

- Modify: `cli/vite.config.ts` — 添加 `@tanstack/router-plugin/vite` 插件（在 react 之前）
- Create: `cli/src/routes/__root.tsx` — 根布局组件
- Create: `cli/src/routes/index.tsx` — 搜索页路由骨架
- Create: `cli/src/routes/admin.tsx` — 管理页路由骨架（懒加载）
- Modify: `cli/src/main.tsx` — 创建 RouterProvider

## Steps

### Step 1: Install TanStack Router

- Install `@tanstack/react-router` 和 `@tanstack/router-plugin`
- 在 `vite.config.ts` 中添加 `tanstackRouter()` 插件，确保在 `react()` 之前
- 配置 `autoCodeSplitting: true`

### Step 2: Create root layout

- 在 `__root.tsx` 中使用 `createRootRoute` 定义根布局
- 根布局包含 `<Outlet />` 用于渲染子路由
- 可选添加导航链接（搜索页 / 管理页）

### Step 3: Create route skeletons

- `index.tsx`：使用 `createFileRoute('/')` 创建搜索页路由，组件暂时返回 placeholder
- `admin.tsx`：使用 `createFileRoute('/admin')` 创建管理页路由

### Step 4: Configure router with basePath

- 在 `main.tsx` 中 `createRouter({ routeTree, basepath: '/app' })`
- 使用 `<RouterProvider router={router} />` 渲染

### Step 5: Verify

- 访问 `/app` 渲染搜索页 placeholder
- 访问 `/app/admin` 渲染管理页 placeholder
- `npx tsc --noEmit` 通过（类型安全路由）

## Verification Commands

```bash
cd cli && npm run dev &
sleep 3
curl -s http://localhost:5173/app | grep -c "placeholder"
curl -s http://localhost:5173/app/admin | grep -c "placeholder"
npx tsc --noEmit
```

## Success Criteria

- 文件路由自动生成 `routeTree.gen.ts`
- `/app` 和 `/app/admin` 分别渲染对应页面
- TypeScript 编译通过，路由参数有类型推导
- `autoCodeSplitting` 配置生效
