# Connect-RPC Full Migration Implementation Plan

> **For Codex:** 使用 `executing-plans` 思路按批次执行；本轮先完成 Schema 全量迁移 + Health Connect 接入。

## Goal

逐步将当前对外 REST 契约迁移到 `buf` 管理的 `.proto`，并明确使用：

- Go 服务端/客户端生成：`connect-go`
- TypeScript 生成：`connect-es`（并配合 `protobuf-es` 生成消息类型）

迁移期间保持现有 REST 路由可用，采用“并行共存、分域切换”的策略。

## Constraints

- 渐进迁移：先 schema 全量覆盖，再按域接入 handler。
- 保持兼容：REST 路由不删除，Connect 新路由并行挂载。
- 验证优先：每批次完成后执行 `buf lint`、`buf generate` 与对应测试。
- 优先复用现有 service/handler 逻辑，避免重复实现业务逻辑。

## BDD Scenarios

- Scenario 1: 全部现有公开 API 都能在 `.proto` 中找到对应 RPC 与消息定义。
- Scenario 2: `buf` 代码生成同时产出 Go Connect 代码与 TS Connect 代码。
- Scenario 3: Health 域 Connect 端点可调用，且返回值与现有 REST 语义一致（字段级）。
- Scenario 4: 引入 Connect 路由后，既有 REST 路由鉴权行为不回归。

## Execution Plan

- [Task 001: 全量 API proto 契约映射（Schema）](./task-001-full-api-proto-schema.md)
- [Task 002: Buf 生成链路配置 connect-go/connect-es](./task-002-buf-generation-connect-go-es.md)
- [Task 003: Health Connect handler 与路由挂载](./task-003-health-connect-handler-and-routing.md)
- [Task 004: Health Connect 集成测试与回归验证](./task-004-health-connect-tests-and-regression.md)
- [Task 005: App/Auth/Search 域 Connect 接入（下一批）](./task-005-app-auth-search-connect-migration.md)
- [Task 006: Admin 域 Connect 接入与切换策略（下一批）](./task-006-admin-connect-migration-and-cutover.md)

## Batch Plan

- Batch A（本轮执行）: Task 001-004
- Batch B（后续执行）: Task 005
- Batch C（后续执行）: Task 006
