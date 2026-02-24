# Task 008: GREEN frontend running lock and force-rebuild guard implementation

**depends-on**: task-007-red-frontend-running-lock-and-guard-tests.md

## Description

实现运行态交互锁与 force_rebuild 互斥规则，防止高风险误操作。

## Execution Context

**Task Number**: 008 of 011  
**Phase**: Frontend (Green)  
**Prerequisites**: Task 007 红测已完成

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `运行中禁止修改目录选择和拉取目录详情`、`运行中禁止重复发起局部补同步`、`force_rebuild 与局部补同步互斥`、`全量全库时允许 force_rebuild`

## Files to Modify/Create

- Modify: `web/src/components/admin-sync-page.tsx`
- Modify: `web/src/hooks/use-sync-progress.ts`（若需提交参数防线）

## Steps

### Step 1: Implement Running Lock

- 运行态统一禁用 toggle、inspect、模式切换和启动按钮。

### Step 2: Implement Guardrail

- 若存在局部勾选范围且 force_rebuild=true，则阻止提交并提示。
- 全量全库（无 scoped selection）保持 force_rebuild 可用。

### Step 3: Verify (Green)

- 运行 Task 007 用例转绿。

## Verification Commands

```bash
cd web && bun vitest run src/components/admin-page.test.tsx
```

## Success Criteria

- Task 007 红测转绿。
- 风险组合被前端防线阻断，正常全量强制重建不受影响。
