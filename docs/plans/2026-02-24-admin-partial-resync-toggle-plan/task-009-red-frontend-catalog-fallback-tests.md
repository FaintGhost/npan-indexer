# Task 009: RED frontend catalog fallback render tests

**depends-on**: task-004-green-backend-catalog-preserve-impl.md

## Description

先写失败测试约束渲染兼容：优先使用 catalog 字段，缺失时回退 rootProgress。

## Execution Context

**Task Number**: 009 of 011  
**Phase**: Testing (Red)  
**Prerequisites**: 后端 progress 契约已扩展（Task 004）

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `后端未返回 catalog 字段时前端回退到 rootProgress`

## Files to Modify/Create

- Modify: `web/src/components/sync-progress-display.test.tsx`
- Modify: `web/src/lib/sync-schemas.test.ts`（若 schema 增加 catalog 字段）

## Steps

### Step 1: Verify Scenario

- 明确渲染优先级：catalog 优先，rootProgress 回退。

### Step 2: Implement Test (Red)

- 用例 A：响应含 catalog 字段，断言按 catalog 渲染。
- 用例 B：响应无 catalog 字段，断言回退到 rootProgress 渲染。

### Step 3: Verify Failure

- 在未实现兼容渲染前确认失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/sync-progress-display.test.tsx src/lib/sync-schemas.test.ts
```

## Success Criteria

- 红测稳定失败并明确指出兼容逻辑缺失。
