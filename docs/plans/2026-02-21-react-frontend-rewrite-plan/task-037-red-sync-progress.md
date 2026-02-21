# Task 037: 测试同步进度展示组件

**depends-on**: task-004

## Description

为 SyncProgress 展示组件创建失败测试用例。

## Execution Context

**Task Number**: 037 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 6 - 进度信息正确展示; 根据 estimatedTotalDocs 计算进度百分比; 无 estimatedTotalDocs 时不显示百分比

## Files to Modify/Create

- Create: `cli/src/components/sync-progress.test.tsx`

## Steps

### Step 1: Test running status display — 显示"运行中"标签

### Step 2: Test roots progress — 显示 "1 / 2"

### Step 3: Test active root display

### Step 4: Test aggregate stats — filesIndexed, pagesFetched, failedRequests

### Step 5: Test percentage with estimatedTotalDocs

### Step 6: Test no percentage when estimatedTotalDocs is null

### Step 7: Test done status display

### Step 8: Test error status with lastError

### Step 9: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/sync-progress.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖进度展示的所有变体
