# Connect-RPC Timestamp Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

## Goal

在保持兼容的前提下，为 Connect 进度响应引入 `google.protobuf.Timestamp` sidecar 字段，并完成前后端消费适配与回归验证。

## Architecture

采用“双字段过渡”：

- proto 新增 `*_ts` 字段（`Timestamp`）
- 后端输出同时填充新旧字段
- 前端优先新字段，缺失回退旧字段

## Tech Stack

- Protobuf/Buf
- connect-go / connect-es
- Go backend tests + Vitest frontend tests

## Constraints

- BDD：Red -> Green
- 不移除旧 `int64` 字段
- 不修改持久化结构

## Design Support

- [Design Index](../2026-02-24-connect-rpc-timestamp-migration-design/_index.md)
- [BDD Specs](../2026-02-24-connect-rpc-timestamp-migration-design/bdd-specs.md)
- [Architecture](../2026-02-24-connect-rpc-timestamp-migration-design/architecture.md)
- [Best Practices](../2026-02-24-connect-rpc-timestamp-migration-design/best-practices.md)

## Execution Plan

- [Task 001: RED proto descriptor timestamp fields](./task-001-red-proto-descriptor-timestamp-fields.md)
- [Task 002: GREEN proto add timestamp sidecar fields](./task-002-green-proto-add-timestamp-sidecar-fields.md)
- [Task 003: RED backend connect progress timestamp tests](./task-003-red-backend-connect-progress-timestamp-tests.md)
- [Task 004: GREEN backend progress timestamp mapping](./task-004-green-backend-progress-timestamp-mapping.md)
- [Task 005: RED frontend timestamp fallback tests](./task-005-red-frontend-timestamp-fallback-tests.md)
- [Task 006: GREEN frontend timestamp consumer adapter](./task-006-green-frontend-timestamp-consumer-adapter.md)
- [Task 007: verification and compatibility gate](./task-007-verification-and-compatibility-gate.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-24-connect-rpc-timestamp-migration-plan/`.

Execution options:

1. Orchestrated Execution (Recommended): use `executing-plans` skill with this plan folder.
2. Direct Agent Team: use `agent-team-driven-development` skill for parallel execution.
3. Manual Serial Execution: 按任务顺序逐个执行。
