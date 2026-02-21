# Task 036: 实现 useSyncProgress Hook

**depends-on**: task-035

## Description

实现 useSyncProgress Hook，使 Task 035 测试通过。

## Execution Context

**Task Number**: 036 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 035 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 6 - 同步管理所有场景

## Files to Modify/Create

- Create: `cli/src/hooks/use-sync-progress.ts`

## Steps

### Step 1: Implement useSyncProgress

- 接收 apiHeaders (from useAdminAuth)
- 状态: progress (SyncProgress | null), loading, error
- fetchProgress(): GET /api/v1/admin/sync/full/progress
- startSync(): POST /api/v1/admin/sync/full
- cancelSync(): POST /api/v1/admin/sync/full/cancel
- 轮询逻辑: status="running" 时每 3 秒 fetchProgress
- 终止条件: status in ["done", "error", "cancelled"]
- useEffect cleanup 清除 interval

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-sync-progress.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 035 所有测试通过
