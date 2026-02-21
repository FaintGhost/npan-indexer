# Task 009: 测试 API 客户端封装

**depends-on**: task-006, task-004

## Description

为通用 API 客户端函数创建失败测试用例。测试 fetchAPI 函数的 Zod 解析集成、错误处理、AbortController 支持。

## Execution Context

**Task Number**: 009 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 006 Zod schema 已实现，Task 004 MSW 已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - 搜索 API 响应结构异常时优雅降级; Feature 1 - 新搜索取消进行中的旧请求

## Files to Modify/Create

- Create: `cli/src/lib/api-client.test.ts`

## Steps

### Step 1: Test successful fetch with schema validation

- MSW 返回有效搜索响应 → fetchAPI 成功解析返回数据

### Step 2: Test HTTP error handling

- MSW 返回 500 → fetchAPI 抛出 ApiError，包含 status 和 message

### Step 3: Test schema validation failure

- MSW 返回格式异常的 JSON → fetchAPI 抛出友好错误（不是原始 ZodError）

### Step 4: Test abort signal support

- 传入已中止的 AbortController.signal → fetchAPI 抛出 AbortError

### Step 5: Test admin API with X-API-Key header

- fetchAPI 支持传入自定义 headers（用于 admin 端点）

### Step 6: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/lib/api-client.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖成功请求、HTTP 错误、schema 错误、中止请求、自定义 header
