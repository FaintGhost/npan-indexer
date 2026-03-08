# Taskfile 迁移实施计划

> **For Claude:** REQUIRED SUB-SKILL: 使用 Skill 工具加载 `superpowers:executing-plans` 执行本计划。

## Goal

用 `https://taskfile.dev/` 的根级 `Taskfile.yml` 完整替换仓库根 `Makefile`，统一 guard / Go 单测 / 前端单测 / smoke / E2E 的自动化入口，并同步迁移 CI 与开发者文档。

## Architecture

本次实施不重写现有测试体系，而是把 Task 作为新的编排层。真实执行器继续保持现状：Go 测试仍走 `go test`，前端测试仍走 `web/package.json` 的 Bun scripts，smoke 仍走 `tests/smoke/smoke_test.sh`，E2E 仍走 `docker-compose.ci.yml` 中的 Playwright 容器。

实现上分 4 个特性组推进：

1. 先建立 namespaced 的核心 Task 入口与快速验证聚合任务。
2. 再补齐 smoke / E2E / 全量回归的生命周期任务，保留自动清理语义。
3. 然后把 GitHub Actions 改为通过 Task 调度，同时更新触发路径。
4. 最后迁移 README / `docs/STRUCTURE.md` 并删除根 `Makefile`，完成主入口切换。

## Tech Stack

- Taskfile.dev
- Go 1.25+
- Bun + Vitest
- Docker Compose
- Playwright
- GitHub Actions
- `rg` / `curl` / `jq`

## Constraints

- 必须严格执行 BDD：先 Red，再 Green。
- 不重写 `tests/smoke/smoke_test.sh`、`docker-compose.ci.yml`、`web/package.json` 的真实执行语义。
- 公开任务采用命名空间风格，不保留旧 Make target 作为兼容 alias。
- Task 只做编排，不承载大段 Bash 状态机。
- `verify:smoke` / `verify:e2e` 必须保留“失败也清理”的语义。
- `.github/workflows/ci.yml` 必须改为调用 `task`，并在 `Taskfile.yml` 改动时触发。
- 活跃文档只更新 `README.md` 与 `docs/STRUCTURE.md`；不追改历史设计文档与归档文档。

## Design Support

- [Design Index](../2026-03-08-taskfile-migration-design/_index.md)
- [BDD Specs](../2026-03-08-taskfile-migration-design/bdd-specs.md)
- [Architecture](../2026-03-08-taskfile-migration-design/architecture.md)
- [Best Practices](../2026-03-08-taskfile-migration-design/best-practices.md)

## Execution Plan

- [Task 001: 核心 Task 入口与快速验证测试 (RED)](./task-001-core-task-surface-test.md)
- [Task 001: 核心 Task 入口与快速验证实现 (GREEN)](./task-001-core-task-surface-impl.md)
- [Task 002: smoke / E2E 生命周期任务测试 (RED)](./task-002-lifecycle-verification-test.md)
- [Task 002: smoke / E2E 生命周期任务实现 (GREEN)](./task-002-lifecycle-verification-impl.md)
- [Task 003: GitHub Actions 迁移到 Task 测试 (RED)](./task-003-ci-workflow-migration-test.md)
- [Task 003: GitHub Actions 迁移到 Task 实现 (GREEN)](./task-003-ci-workflow-migration-impl.md)
- [Task 004: 开发者入口与文档切换测试 (RED)](./task-004-developer-entrypoints-test.md)
- [Task 004: 开发者入口与文档切换实现 (GREEN)](./task-004-developer-entrypoints-impl.md)

## Commit Boundaries

- Boundary A: `task-001`（建立核心 Task 入口、快速验证与基础前置检查）
- Boundary B: `task-002`（补齐 smoke / E2E / full regression 生命周期）
- Boundary C: `task-003`（CI workflow 切换到 Task 调度）
- Boundary D: `task-004`（README / 结构文档切换与删除 `Makefile`）

---

## Execution Handoff

计划已保存到 `docs/plans/2026-03-08-taskfile-migration-plan/`。

建议执行方式：

1. 使用 `superpowers:executing-plans` 编排执行（推荐）
2. 使用 `superpowers:agent-team-driven-development` 并行执行
3. 当前会话串行执行（较慢）
