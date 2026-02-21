# Task 030: 实现搜索页面

**depends-on**: task-029

## Description

实现搜索页面组件（`/app` 路由），组合所有 hooks 和组件，使 Task 029 测试通过。

## Execution Context

**Task Number**: 030 of 046
**Phase**: Integration (Green)
**Prerequisites**: Task 029 测试已编写（Red），所有依赖组件和 hooks 已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 所有搜索场景; Feature 2 - 所有下载场景; Feature 3 - 所有 UI 过渡场景

## Files to Modify/Create

- Modify: `cli/src/routes/index.tsx` — 实现 SearchPage 组件

## Steps

### Step 1: Compose SearchPage

- 使用 useSearch 管理搜索状态
- 使用 useDownload 管理下载状态
- 使用 useViewMode 管理 Hero/Docked 切换
- 使用 useHotkey 注册 Cmd/Ctrl+K
- 使用 useIntersectionObserver 实现无限滚动
- 渲染 SearchInput, FileCard 列表, DownloadButton, EmptyStates, SkeletonCards
- 渲染状态栏和计数器

### Step 2: Wire up View Transition

- 搜索触发时 → setDocked(true)
- 清空时 → setDocked(false)

### Step 3: Implement infinite scroll sentinel

- 在列表末尾放置 sentinel 元素
- IntersectionObserver 触发时调用 loadMore

### Step 4: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/routes/index.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 029 所有集成测试通过
- 页面视觉布局匹配现有 HTML
