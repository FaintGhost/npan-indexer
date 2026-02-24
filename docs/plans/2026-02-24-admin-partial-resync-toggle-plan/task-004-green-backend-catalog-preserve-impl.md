# Task 004: GREEN backend catalog preserve implementation

**depends-on**: task-003-red-backend-catalog-preserve-tests.md

## Description

实现 scoped full 下的目录册保留语义，并对外暴露可供前端渲染的兼容字段，修复“局部同步后列表被覆盖”根因。

## Execution Context

**Task Number**: 004 of 011  
**Phase**: Backend (Green)  
**Prerequisites**: Task 003 红测已建立

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `局部补同步后根目录详情列表仍保留历史条目`、`后端未返回 catalog 字段时前端回退到 rootProgress`

## Files to Modify/Create

- Modify: `api/openapi.yaml`（若新增 catalog 相关字段）
- Modify: `api/types.gen.go`（生成产物）
- Modify: `internal/models/models.go`
- Modify: `internal/httpx/dto.go`
- Modify: `internal/service/sync_manager.go`
- Modify: `web/src/api/generated/types.gen.ts`（生成产物）
- Modify: `web/src/api/generated/zod.gen.ts`（生成产物）

## Steps

### Step 1: Define Progress Semantics

- 明确并实现目录册字段或等价兼容语义：
  - 本次执行范围状态
  - 历史目录册状态

### Step 2: Implement Preserve Logic

- 在 scoped full 执行前合并 existing progress 目录册。
- 执行后仅更新本次 roots 对应目录项，不清空其他历史项。

### Step 3: Keep Backward Compatibility

- 保持旧字段可读，避免旧前端/CLI 解析失败。

### Step 4: Regenerate and Verify

- 更新 OpenAPI 生成产物并运行 Task 003 用例转绿。

## Verification Commands

```bash
go generate ./api/...
cd web && bun run generate
go test ./internal/service -run CatalogPreserve -count=1
go test ./internal/httpx -run SyncProgress -count=1
```

## Success Criteria

- Task 003 红测转绿。
- 目录册保留行为稳定，局部同步不再清空历史详情。
- 契约与生成产物一致。
