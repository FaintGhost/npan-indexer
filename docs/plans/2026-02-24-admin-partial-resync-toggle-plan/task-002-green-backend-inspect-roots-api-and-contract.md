# Task 002: GREEN backend inspect roots API and contract

**depends-on**: task-001-red-backend-inspect-roots-tests.md

## Description

实现 Admin 目录详情拉取接口，并完成 OpenAPI 与生成代码对齐，使前端可独立拉取目录详情而不触发同步。

## Execution Context

**Task Number**: 002 of 011  
**Phase**: Backend (Green)  
**Prerequisites**: Task 001 红测已失败且断言清晰

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `拉取目录详情只更新列表，不启动同步`、`批量拉取目录详情部分成功`

## Files to Modify/Create

- Modify: `api/openapi.yaml`
- Modify: `api/types.gen.go`（生成产物）
- Modify: `internal/httpx/handlers.go`
- Modify: `internal/httpx/server.go`
- Modify: `internal/httpx/dto.go`（若需新增 DTO）
- Modify: `web/src/api/generated/types.gen.ts`（生成产物）
- Modify: `web/src/api/generated/zod.gen.ts`（生成产物）

## Steps

### Step 1: Contract First

- 在 OpenAPI 中新增 inspect roots 端点与 request/response schema。

### Step 2: Implement Backend Endpoint

- 增加 handler 与路由注册。
- 使用已存在的 `GetFolderInfo` 能力组装 `items/errors` 的部分成功响应。
- 保持鉴权与错误响应风格一致。

### Step 3: Regenerate Types

- 运行项目既有生成流程，更新 Go/TS 产物。

### Step 4: Verify (Green)

- 运行 Task 001 的用例，确认通过。
- 追加路由与基本鉴权测试（如有）。

## Verification Commands

```bash
go generate ./api/...
cd web && bun run generate
go test ./internal/httpx -run InspectRoots -count=1
go test ./internal/httpx -run Routes -count=1
```

## Success Criteria

- Task 001 红测转绿。
- 新接口契约与生成代码无漂移。
- 不影响既有 `/api/v1/admin/sync` 路由行为。
