# Task 005: RED frontend timestamp fallback tests

**depends-on**: task-002-green-proto-add-timestamp-sidecar-fields

## Description

先编写前端失败测试，覆盖“优先读取 `*_ts`，缺失时回退旧 `int64`”的消费行为。

## Execution Context

**Task Number**: 005 of 007  
**Phase**: Testing (Red)  
**Prerequisites**: Task 002 已完成（前端类型可见新字段）

## BDD Scenario Reference

**Spec**: `../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md`  
**Scenario**: `前端优先读取 Timestamp 并正确展示`  
**Scenario**: `旧字段回退路径持续可用`

## Files to Modify/Create

- Modify: `web/src/hooks/use-sync-progress.test.ts`
- Modify: `web/src/components/sync-progress-display.test.tsx`

## Steps

### Step 1: Verify Scenario

- 明确两类输入：仅新字段、仅旧字段。

### Step 2: Implement Test (Red)

- 为 hook/组件新增双输入测试：
  - 仅 `*_ts` 时应正确渲染；
  - 无 `*_ts` 时应回退旧字段并保持现有展示。
- 当前实现下应至少有一类失败。

### Step 3: Verify Failure

- 运行测试确认 Red 状态。

## Verification Commands

```bash
cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/sync-progress-display.test.tsx
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 失败原因聚焦在消费适配缺口。
