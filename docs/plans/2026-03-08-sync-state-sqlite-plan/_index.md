# 同步状态 SQLite 迁移实施计划

> **For Claude:** REQUIRED SUB-SKILL: 使用 Skill 工具加载 `superpowers:executing-plans` 执行本计划。

## Goal

把当前同步状态持久化从多份 JSON 文件迁移到单一 SQLite 状态库，修复状态恢复不可靠的问题，并保持 Admin Connect API、CLI 与现有进度模型语义不变。

## Architecture

本次实施分四层推进：

1. 先补 SQLite 状态库与 legacy JSON 惰性导入的失败测试，锁定主状态源、迁移与并发保存语义。
2. 再把 `SyncManager` 从 JSON store 硬编码切到可注入的 progress/sync-state/checkpoint 抽象。
3. 然后补齐 checkpoint 恢复/清理语义，以及 server/CLI wiring，确保对外行为保持兼容。
4. 最后执行完整回归与运维文档收口，证明 SQLite 路径在现有构建与测试链路下可稳定工作。

## Tech Stack

- Go 1.25 + `database/sql`
- SQLite（推荐 `modernc.org/sqlite`，兼容 `CGO_ENABLED=0`）
- Connect-RPC / Echo / Cobra CLI
- Vitest / Playwright / Docker Compose CI（全链路回归）

## Constraints

- 必须严格执行 BDD：先 Red，再 Green。
- 单测必须使用 test doubles，隔离数据库文件之外的外部依赖：
  - Npan API 使用 stub/fake。
  - Meilisearch 使用 stub index manager。
  - 不依赖真实网络。
- SQLite 必须兼容当前 `Dockerfile` 的 `CGO_ENABLED=0` 构建。
- 不修改 `models.SyncProgressState`、`models.SyncState`、`models.CrawlCheckpoint` 的外部语义。
- 不删除 legacy JSON 文件；本轮只做非破坏式惰性导入。

## Design Support

- [Design Index](../2026-03-08-sync-state-sqlite-design/_index.md)
- [BDD Specs](../2026-03-08-sync-state-sqlite-design/bdd-specs.md)
- [Architecture](../2026-03-08-sync-state-sqlite-design/architecture.md)
- [Best Practices](../2026-03-08-sync-state-sqlite-design/best-practices.md)

## Execution Plan

- [Task 001: SQLite 状态库与 legacy 导入测试 (RED)](./task-001-sqlite-state-store-test.md)
- [Task 001: SQLite 状态库与 legacy 导入实现 (GREEN)](./task-001-sqlite-state-store-impl.md)
- [Task 002: SyncManager 状态抽象与游标语义测试 (RED)](./task-002-sync-manager-state-test.md)
- [Task 002: SyncManager 状态抽象与游标语义实现 (GREEN)](./task-002-sync-manager-state-impl.md)
- [Task 003: SQLite checkpoint 恢复与清理测试 (RED)](./task-003-checkpoint-lifecycle-test.md)
- [Task 003: SQLite checkpoint 恢复与清理实现 (GREEN)](./task-003-checkpoint-lifecycle-impl.md)
- [Task 004: Admin/CLI SQLite 兼容性测试 (RED)](./task-004-admin-cli-sqlite-test.md)
- [Task 004: Admin/CLI SQLite 兼容性实现 (GREEN)](./task-004-admin-cli-sqlite-impl.md)
- [Task 005: 全链路验证与运行文档收口](./task-005-verification-and-runbook.md)

## Commit Boundaries

- Boundary A: `task-001`（SQLite 状态库、schema、惰性导入）
- Boundary B: `task-002` ~ `task-003`（SyncManager 抽象、checkpoint 生命周期）
- Boundary C: `task-004` ~ `task-005`（server/CLI wiring、文档、完整验证）

---

## Execution Handoff

计划已保存到 `docs/plans/2026-03-08-sync-state-sqlite-plan/`。

建议执行方式：

1. 使用 `superpowers:executing-plans` 编排执行（推荐）
2. 使用 `superpowers:agent-team-driven-development` 并行执行
3. 当前会话串行执行（较慢）
