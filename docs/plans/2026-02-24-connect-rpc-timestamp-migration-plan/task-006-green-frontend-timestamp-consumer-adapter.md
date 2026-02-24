# Task 006: GREEN frontend timestamp consumer adapter

**depends-on**: task-005-red-frontend-timestamp-fallback-tests

## Description

实现前端时间消费适配器，统一处理 `Timestamp | int64` 输入并驱动页面正确展示。

## Execution Context

**Task Number**: 006 of 007  
**Phase**: Implementation (Green)  
**Prerequisites**: Task 005 已失败并定位到消费缺口

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `前端优先读取 Timestamp 并正确展示`  
**Scenario**: `旧字段回退路径持续可用`

## Files to Modify/Create

- Modify: `web/src/hooks/use-sync-progress.ts`
- Modify: `web/src/components/sync-progress-display.tsx`
- Modify: `web/src/lib/*`（若需新增通用时间适配器）

## Steps

### Step 1: Implement Adapter

- 新增统一时间解析逻辑：
  - 优先 `*_ts`
  - 回退旧 `int64`
- 避免在多个组件重复解析逻辑。

### Step 2: Verify Green

- 运行 Task 005 的测试，确认转绿。

### Step 3: Frontend Regression

- 运行相关前端测试，确认现有进度展示未回归。

## Verification Commands

```bash
cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/sync-progress-display.test.tsx src/components/admin-page.test.tsx
```

## Success Criteria

- Task 005 转绿。
- 管理页进度展示行为无回归。
