# Task 006: 实现 Zod Schema（搜索/下载/错误响应）

**depends-on**: task-005

## Description

实现搜索响应、下载链接响应、错误响应的 Zod schema，使 Task 005 的测试通过。

## Execution Context

**Task Number**: 006 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 005 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - 所有 API 响应 Zod Schema 校验场景

## Files to Modify/Create

- Create: `cli/src/schemas/search.ts` — IndexDocumentSchema, SearchResponseSchema 及类型导出
- Create: `cli/src/schemas/download.ts` — DownloadURLResponseSchema 及类型导出
- Create: `cli/src/schemas/error.ts` — ErrorResponseSchema 及类型导出

## Steps

### Step 1: Implement SearchResponseSchema

- 定义 IndexDocumentSchema（字段匹配 Go 后端 models.IndexDocument）
- 定义 SearchResponseSchema（items + total）
- 使用 `z.infer<>` 导出 TypeScript 类型
- highlighted_name 字段使用 `.optional().default("")`

### Step 2: Implement DownloadURLResponseSchema

- file_id: z.number()
- download_url: z.string().min(1)

### Step 3: Implement ErrorResponseSchema

- code: z.string()
- message: z.string()
- request_id: z.string().optional()

### Step 4: Verify tests PASS (Green)

- 运行 Task 005 的测试，全部通过

## Verification Commands

```bash
cd cli && npx vitest run src/schemas/
# Expected: ALL PASS (Green)
```

## Success Criteria

- Task 005 的所有测试通过
- 类型从 schema 自动推导（`z.infer<>`）
- `npx tsc --noEmit` 通过
