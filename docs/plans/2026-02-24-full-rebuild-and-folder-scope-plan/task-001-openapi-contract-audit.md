# Task 001: OpenAPI contract audit and generation guard

**depends-on**: none

## Scenario Reference

- Feature: Folder-Scoped Full Sync From Admin UI
- Scenario: 指定单个目录 ID 启动范围索引
- Scenario: 输入多个目录 ID（逗号分隔）

## Objective

在实现前确认契约边界，保证前后端字段以 `api/openapi.yaml` 为准，避免手写字段漂移。

## Files

- `api/openapi.yaml`
- `api/types.gen.go`（生成产物）
- `web/src/api/generated/types.gen.ts`（生成产物）
- `web/src/api/generated/zod.gen.ts`（生成产物）

## Tasks

1. 审核 `SyncStartRequest` 与 `SyncProgressState` 所需字段是否已覆盖：
  - `root_folder_ids`
  - `include_departments`
  - `verification.warnings`
  - `rootProgress[*].estimatedTotalDocs`
2. 如发现契约缺失，先更新 `api/openapi.yaml` 再生成类型。
3. 执行类型生成并确认生成产物与契约一致。

## Verification

- `make generate`
- `make generate-check`
- 若有 schema 变更，确认后端 DTO 与前端 zod 类型均编译通过
