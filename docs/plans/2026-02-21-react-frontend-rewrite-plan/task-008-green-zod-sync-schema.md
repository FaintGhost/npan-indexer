# Task 008: 实现 Zod Schema（同步进度）

**depends-on**: task-007

## Description

实现同步进度 Zod schema（CrawlStatsSchema, RootProgressSchema, SyncProgressSchema），使 Task 007 测试通过。

## Execution Context

**Task Number**: 008 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 007 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - 同步进度 API 响应通过 schema 校验

## Files to Modify/Create

- Create: `cli/src/schemas/sync.ts`

## Steps

### Step 1: Implement schemas

- CrawlStatsSchema: foldersVisited, filesIndexed, pagesFetched, failedRequests, startedAt, endedAt
- RootProgressSchema: rootFolderId, status, estimatedTotalDocs (nullable optional), stats, updatedAt
- SyncProgressSchema: status (enum), startedAt, updatedAt, roots, completedRoots, activeRoot (nullable optional), aggregateStats, rootProgress (record), lastError (optional default "")

### Step 2: Export types via z.infer

### Step 3: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/schemas/sync.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 007 的所有测试通过
- 类型与 Go 后端 SyncProgressResponse DTO 对齐
