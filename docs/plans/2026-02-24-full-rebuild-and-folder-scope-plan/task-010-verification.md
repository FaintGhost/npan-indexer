# Task 010: Verification and regression checks

**depends-on**: task-003-green-backend-checkpoint-reset-impl.md, task-005-green-backend-folder-info-estimate-and-warnings.md, task-007-green-frontend-folder-scope-impl.md, task-009-green-frontend-estimate-warning-impl.md

## Scenario Reference

- 覆盖本计划全部场景

## Objective

在完成实现后给出可重复的证据，证明 OpenAPI 契约、后端逻辑和前端展示一致。

## Files

- `api/openapi.yaml`
- `api/types.gen.go`
- `web/src/api/generated/types.gen.ts`
- `web/src/api/generated/zod.gen.ts`
- 相关改动源码与测试文件

## Tasks

1. 执行 OpenAPI 生成校验，确认契约与生成文件一致。
2. 执行后端目标测试集（service/npan/indexer 相关）。
3. 执行前端目标测试集（hook/component 相关）。
4. 最后执行一次聚合检查（`make generate-check` + 关键测试）。

## Verification

- `make generate-check`
- `go test ./internal/service/... ./internal/npan/...`
- `cd web && bun vitest run src/hooks/use-sync-progress.test.ts src/components/admin-page.test.tsx src/components/sync-progress-display.test.tsx`
