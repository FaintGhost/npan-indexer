# Task 010: GREEN frontend catalog fallback render implementation

**depends-on**: task-009-red-frontend-catalog-fallback-tests.md

## Description

实现 catalog 优先渲染与 rootProgress 回退，保证新旧后端响应都能正确显示根目录详情与 toggle。

## Execution Context

**Task Number**: 010 of 011  
**Phase**: Frontend (Green)  
**Prerequisites**: Task 009 红测已建立

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `后端未返回 catalog 字段时前端回退到 rootProgress`、`局部补同步后根目录详情列表仍保留历史条目`

## Files to Modify/Create

- Modify: `web/src/lib/sync-schemas.ts`
- Modify: `web/src/components/sync-progress-display.tsx`
- Modify: `web/src/components/admin-sync-page.tsx`（若选择源在页面层统一）

## Steps

### Step 1: Extend Schema

- 为 progress 增加可选 catalog 字段解析。

### Step 2: Implement Render Selection

- 根目录详情和 toggle 数据源优先使用 catalog。
- catalog 缺失时回退 rootProgress，确保兼容。

### Step 3: Verify (Green)

- 运行 Task 009 用例转绿。

## Verification Commands

```bash
cd web && bun vitest run src/components/sync-progress-display.test.tsx src/lib/sync-schemas.test.ts
```

## Success Criteria

- Task 009 红测转绿。
- 新旧响应格式均可渲染，无空白或丢数据。
