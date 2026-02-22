# Task 006: Create Makefile

**depends-on**: Task 002, Task 003

**ref**: BDD Scenario "Spec 变更后未重新生成代码时 CI 失败"

## Description

创建 Makefile 作为统一的构建工作流入口，提供代码生成、一致性检查、测试等目标。

## What to do

1. 创建 `Makefile`，包含以下 targets：

   - **`generate-go`**: 调用 oapi-codegen 从 spec 生成 Go types
   - **`generate-ts`**: 调用 @hey-api/openapi-ts 从 spec 生成 TS types + Zod schemas
   - **`generate`**: 同时执行 generate-go 和 generate-ts
   - **`generate-check`**: 执行 generate 后，用 `git diff --exit-code` 检查生成文件是否与已提交版本一致。不一致则报错退出
   - **`test`**: 运行 `go test ./...`
   - **`test-frontend`**: 运行 `cd web && bun run test`
   - **`smoke-test`**: 运行 `docker compose -f docker-compose.ci.yml up -d --wait` 然后执行 `tests/smoke/smoke_test.sh`，最后 cleanup

2. 添加 `.PHONY` 声明

## Files to create

- `Makefile`

## Verification

- `make generate` 成功执行，产出 `api/types.gen.go` 和 `web/src/api/generated/`
- 修改 `api/openapi.yaml` 后不执行 `make generate`，`make generate-check` 报错退出
- 执行 `make generate` 后，`make generate-check` 成功通过
