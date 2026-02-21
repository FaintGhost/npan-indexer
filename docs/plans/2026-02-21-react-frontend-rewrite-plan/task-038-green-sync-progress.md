# Task 038: 实现同步进度展示组件

**depends-on**: task-037

## Description

实现 SyncProgress 展示组件，使 Task 037 测试通过。

## Execution Context

**Task Number**: 038 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 037 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 6 - 进度展示场景

## Files to Modify/Create

- Create: `cli/src/components/sync-progress.tsx`

## Steps

### Step 1: Implement SyncProgress component

- Props: SyncProgress 类型数据
- 状态标签（running=蓝色运转, done=绿色完成, error=红色出错, cancelled=灰色已取消）
- 根目录进度: completedRoots.length / roots.length
- 活跃根目录 ID
- 聚合统计: filesIndexed, pagesFetched, failedRequests, foldersVisited
- 百分比进度条: 有 estimatedTotalDocs 时计算 filesIndexed/estimatedTotalDocs
- 错误信息: lastError 不为空时显示

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/sync-progress.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 037 所有测试通过
