# OpenAPI Contract + CI Smoke Test Implementation Plan

## Goal

为 npan 项目建立 spec-first 的 API 契约体系和 CI 冒烟测试，消除前后端契约不一致的问题。

## Design Reference

[Design Document](../2026-02-22-openapi-contract-ci-design/_index.md)

## Constraints

- oapi-codegen 不支持 Echo v5 server 接口生成，仅生成 types
- 前端已使用 Zod v3，codegen 需要兼容
- 渐进式替换：先生成代码并验证，再逐步替换手写 schema
- 现有测试必须持续通过

## Execution Plan

### Phase 1: 基础设施（顺序执行）

- [Task 001: Write OpenAPI spec](./task-001-write-openapi-spec.md) — 编写 API spec，所有后续任务的基础
- [Task 002: Setup Go codegen](./task-002-setup-go-codegen.md) — 安装 oapi-codegen，生成 Go types
- [Task 003: Setup TS codegen](./task-003-setup-ts-codegen.md) — 安装 @hey-api/openapi-ts，生成 Zod schemas

### Phase 2: 代码适配（可并行）

- [Task 004: Replace frontend schemas](./task-004-replace-frontend-schemas.md) — 用生成的 Zod schemas 替换手写版本
- [Task 005: Align backend DTOs](./task-005-align-backend-dtos.md) — 后端 handler 使用生成的 types

### Phase 3: 构建工作流

- [Task 006: Create Makefile](./task-006-create-makefile.md) — 统一的 generate/check/test 入口

### Phase 4: CI Pipeline（可并行）

- [Task 007: Create CI compose and smoke test](./task-007-ci-compose-smoke-test.md) — docker-compose.ci.yml + smoke_test.sh
- [Task 008: Create GitHub Actions workflow](./task-008-github-actions.md) — .github/workflows/ci.yml

## Commit Boundaries

1. Tasks 001-003: `feat: add OpenAPI spec and codegen infrastructure`
2. Tasks 004-005: `refactor: replace hand-written schemas with generated code`
3. Task 006: `build: add Makefile with generate and check targets`
4. Tasks 007-008: `ci: add Docker Compose smoke tests and GitHub Actions`
