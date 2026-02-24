# Connect-RPC Protovalidate Incremental Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

## Goal

在不引入 `Timestamp` 契约变更和包层重构的前提下，增量把高价值输入约束迁移到 `.proto` 的 `protovalidate` 规则，并验证 Connect validation interceptor 的命中与 no-op 行为。

## Architecture

本计划沿用当前 `internal/httpx` 的 Connect 集成架构，仅做 schema 层和测试层增强：

- 在 `proto/npan/v1/api.proto` 增量添加 `buf.validate` 规则；
- 继续复用现有 `NewConnectValidationInterceptor` 作为统一入口；
- 通过 Red/Green 方式先建立失败测试，再落规则与回归收口。

## Tech Stack

- Protobuf + Buf (`buf lint` / `buf generate`)
- Connect-Go (`connectrpc.com/connect`)
- Protovalidate (`buf.build/go/protovalidate`)
- Go test（`internal/httpx` 集成测试 + 全量回归）

## Constraints

- BDD 驱动：优先 Red -> Green。
- 单测隔离外部依赖：禁止真实网络调用，使用 test doubles/fake next handler。
- 最小影响改动：本批次不改 `Timestamp` 字段类型，不改路由结构。

## Design Support

- [Design Index](../2026-02-24-connect-rpc-review-alignment-design/_index.md)
- [BDD Specs](../2026-02-24-connect-rpc-review-alignment-design/bdd-specs.md)
- [Architecture](../2026-02-24-connect-rpc-review-alignment-design/architecture.md)
- [Best Practices](../2026-02-24-connect-rpc-review-alignment-design/best-practices.md)

## Execution Plan

- [Task 001: RED admin validation hit tests](./task-001-red-admin-validation-hit-tests.md)
- [Task 002: GREEN admin proto validation rules](./task-002-green-admin-proto-validation-rules.md)
- [Task 003: RED search pagination validation hit tests](./task-003-red-search-pagination-validation-hit-tests.md)
- [Task 004: GREEN search pagination proto validation rules](./task-004-green-search-pagination-proto-validation-rules.md)
- [Task 005: GREEN no-op and business guard regression](./task-005-green-noop-and-business-guard-regression.md)
- [Task 006: Verification and timestamp compatibility gate](./task-006-verification-and-timestamp-compat-gate.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-24-connect-rpc-protovalidate-plan/`.

Execution options:

1. Orchestrated Execution (Recommended): use `executing-plans` skill with this plan folder.
2. Direct Agent Team: use `agent-team-driven-development` skill for parallel execution.
3. Manual Serial Execution: 按任务顺序逐个执行。
