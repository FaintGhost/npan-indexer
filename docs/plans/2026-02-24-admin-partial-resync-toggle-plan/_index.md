# Admin Partial Resync Toggle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

## Goal

实现 Admin 同步页的“拉取目录详情 + 根目录 toggle 局部补同步”工作流，修复局部同步后根目录详情被覆盖的问题，并保持现有 `/api/v1/admin/sync` 兼容。

## Architecture

方案采用“后端目录册 + 前端选择态”双层设计：

- 后端新增目录详情查询接口，支持批量 folder id 拉取与部分成功返回。
- 同步进度新增目录册语义（或兼容字段扩展），区分“本次执行 roots”与“长期展示 roots”。
- 前端将“目录输入”从同步触发改为目录详情拉取，根目录详情区提供 toggle 选择并默认勾选新拉取目录。

## Tech Stack

- Go + Echo v5（`internal/httpx`, `internal/service`, `internal/npan`）
- OpenAPI 契约（`api/openapi.yaml` + 生成代码）
- React 19 + TanStack Router + Vitest + MSW（`web/src/*`）

## Constraints

- BDD 先红后绿：每个行为先落失败测试，再实现。
- OpenAPI-first：接口变更先改 `api/openapi.yaml`，再生成。
- 单测必须隔离外部依赖，使用 test doubles（禁止真实网络）。
- 最小影响改动，避免破坏 CLI/既有 API 调用。

## Design Support

- [Design Index](../2026-02-24-admin-partial-resync-toggle-design/_index.md)
- [BDD Specs](../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md)
- [Architecture](../2026-02-24-admin-partial-resync-toggle-design/architecture.md)
- [Best Practices](../2026-02-24-admin-partial-resync-toggle-design/best-practices.md)

## Execution Plan

- [Task 001: RED backend inspect roots API tests](./task-001-red-backend-inspect-roots-tests.md)
- [Task 002: GREEN backend inspect roots API and contract](./task-002-green-backend-inspect-roots-api-and-contract.md)
- [Task 003: RED backend catalog preserve progress tests](./task-003-red-backend-catalog-preserve-tests.md)
- [Task 004: GREEN backend catalog preserve implementation](./task-004-green-backend-catalog-preserve-impl.md)
- [Task 005: RED frontend inspect decoupling and auto-select tests](./task-005-red-frontend-inspect-and-autoselect-tests.md)
- [Task 006: GREEN frontend inspect decoupling and auto-select implementation](./task-006-green-frontend-inspect-and-autoselect-impl.md)
- [Task 007: RED frontend running lock and force-rebuild guard tests](./task-007-red-frontend-running-lock-and-guard-tests.md)
- [Task 008: GREEN frontend running lock and force-rebuild guard implementation](./task-008-green-frontend-running-lock-and-guard-impl.md)
- [Task 009: RED frontend catalog fallback render tests](./task-009-red-frontend-catalog-fallback-tests.md)
- [Task 010: GREEN frontend catalog fallback render implementation](./task-010-green-frontend-catalog-fallback-impl.md)
- [Task 011: End-to-end verification and regression suite](./task-011-verification-and-regression.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-24-admin-partial-resync-toggle-plan/`.

Execution options:

1. Orchestrated Execution (Recommended): use `executing-plans` skill with this plan folder.
2. Direct Agent Team: use `agent-team-driven-development` skill for parallel execution.
3. Manual Serial Execution: 按任务顺序逐个执行。
