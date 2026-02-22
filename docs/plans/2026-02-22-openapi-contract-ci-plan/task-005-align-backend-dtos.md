# Task 005: Align Backend DTOs

**depends-on**: Task 002

**ref**: BDD Scenario "从 spec 生成的 Go types 与现有 handler 响应结构匹配"

## Description

让后端 handler 使用生成的 types 做响应序列化，确保 handler 输出与 OpenAPI spec 一致。

## What to do

1. 对比 `api/types.gen.go` 中生成的 types 与 `internal/httpx/dto.go` 中手写的 types：
   - 如果字段名和 JSON tags 完全一致，直接在 handler 中使用生成的 type
   - 如果有差异，需要在 handler 中做映射

2. 对 `GetFullSyncProgress` handler 特别处理：
   - 当前直接返回 `models.SyncProgressState`（包含 `meiliHost`、`checkpointTemplate` 等内部字段）
   - 应改为使用 DTO 映射（已有 `toSyncProgressResponse` 但未使用），或使用生成的 response type

3. 运行现有 Go 测试确认无回归

## Files to modify

- `internal/httpx/handlers.go` — `GetFullSyncProgress` 使用 DTO 映射
- `internal/httpx/dto.go` — 可能需要更新以对齐生成的 types，或标记为 deprecated

## Verification

- `go test ./...` 全部通过
- `GetFullSyncProgress` 的 JSON 响应不包含 `meiliHost`、`meiliIndex`、`checkpointTemplate` 等内部字段
- 响应 JSON 结构与 `api/openapi.yaml` 中定义的 `SyncProgressState` schema 一致
