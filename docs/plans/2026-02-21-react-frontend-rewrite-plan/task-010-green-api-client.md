# Task 010: 实现 API 客户端封装

**depends-on**: task-009

## Description

实现通用 fetchAPI 函数和 ApiError 类，使 Task 009 测试通过。

## Execution Context

**Task Number**: 010 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 009 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - API 响应 Zod Schema 校验（所有）

## Files to Modify/Create

- Create: `cli/src/lib/api-client.ts`

## Steps

### Step 1: Implement ApiError class

- 包含 status、message、code 属性
- 继承自 Error

### Step 2: Implement fetchAPI function

- 接收 URL、Zod schema、可选 RequestInit
- fetch → 检查 response.ok → JSON parse → schema.parse
- 非 ok 时解析 ErrorResponseSchema 获取友好消息
- ZodError 时转换为友好错误消息（不泄露 schema 细节）
- 支持 AbortSignal

### Step 3: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/lib/api-client.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 009 所有测试通过
- fetchAPI 是泛型函数，返回类型由 schema 推导
