# Task 002: Setup Go Codegen

**depends-on**: Task 001

**ref**: BDD Scenario "从 spec 生成的 Go types 与现有 handler 响应结构匹配"

## Description

安装 oapi-codegen，配置仅生成 types 模式，从 OpenAPI spec 生成 Go types。

## What to do

1. 创建 `api/oapi-codegen.yaml` 配置文件：
   - `package: api`
   - `output: api/types.gen.go`
   - `generate.models: true`，其余（strict-server, echo-server, client）均为 false

2. 创建 `api/generate.go` 包含 `//go:generate` 指令调用 oapi-codegen

3. 运行生成命令，产出 `api/types.gen.go`

4. 验证生成的 Go types 与现有 `internal/models/models.go` 和 `internal/httpx/dto.go` 中的结构在字段名和 JSON tags 上一致

## Files to create

- `api/oapi-codegen.yaml`
- `api/generate.go`
- `api/types.gen.go`（生成产物）

## Verification

- `go build ./api/...` 编译通过
- 生成的 types 包含 `SyncProgressState`、`SearchResponse`、`IndexDocument`、`ErrorResponse` 等
- `SyncProgressState` 的 `Status` 字段使用生成的 enum 类型
- JSON tags 与 `api/openapi.yaml` 中的 property name 一致
