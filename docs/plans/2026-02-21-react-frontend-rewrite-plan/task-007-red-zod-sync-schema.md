# Task 007: 测试 Zod Schema（同步进度）

**depends-on**: task-004

## Description

为同步进度 API 响应创建 Zod schema 的失败测试用例。

## Execution Context

**Task Number**: 007 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - 同步进度 API 响应通过 schema 校验

## Files to Modify/Create

- Create: `cli/src/schemas/sync.test.ts`

## Steps

### Step 1: Create sync progress schema test

- 测试有效的 SyncProgress 响应能被成功解析
- 测试 status 字段只接受 "running"、"done"、"error"、"cancelled"
- 测试 aggregateStats 中 filesIndexed/pagesFetched/failedRequests 为数字
- 测试 rootProgress 为 Record<string, RootProgress> 结构
- 测试 activeRoot 为 nullable optional
- 测试 estimatedTotalDocs 为 nullable optional
- 测试 lastError 缺省时默认为空字符串

### Step 2: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/schemas/sync.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试因 schema 模块不存在而失败
- 覆盖了同步进度的所有关键字段验证
