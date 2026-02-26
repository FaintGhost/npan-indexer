# Search Results Filter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: 使用 Skill 工具加载 `superpowers:executing-plans` 执行本计划。

## Goal

为搜索结果页交付“前端扩展名单选筛选 + URL 参数持久化”，确保不修改后端契约且可通过现有测试体系验证。

## Architecture

实现保持 Connect 查询链路不变，仅在前端渲染层新增分类过滤与 URL 状态同步。结果流为：请求分页数据 -> 去重合并 -> 按 `ext` 过滤 -> 渲染计数/列表/空态。筛选状态采用 URL 作为可恢复来源，非法值回退默认值。

## Tech Stack

- React + TanStack Router
- Connect Query
- Vitest + React Testing Library + MSW
- Bun（测试执行）

## Constraints

- 不改 `proto/` 与后端 `internal/` 代码
- 不引入新的后端请求字段
- 任务必须严格 Red -> Green
- 单测必须使用测试替身隔离外部依赖（MSW 拦截网络）

## Design Support

- [Design Index](../2026-02-27-search-results-filter-design/_index.md)
- [BDD Specs](../2026-02-27-search-results-filter-design/bdd-specs.md)
- [Architecture](../2026-02-27-search-results-filter-design/architecture.md)
- [Best Practices](../2026-02-27-search-results-filter-design/best-practices.md)

## Execution Plan

- [Task 001: URL 筛选状态测试 (RED)](./task-001-url-filter-state-test.md)
- [Task 001: URL 筛选状态实现 (GREEN)](./task-001-url-filter-state-impl.md)
- [Task 002: 扩展名分类规则测试 (RED)](./task-002-file-category-test.md)
- [Task 002: 扩展名分类规则实现 (GREEN)](./task-002-file-category-impl.md)
- [Task 003: 过滤流水线与计数测试 (RED)](./task-003-results-filtering-test.md)
- [Task 003: 过滤流水线与计数实现 (GREEN)](./task-003-results-filtering-impl.md)
- [Task 004: 筛选切换与请求契约测试 (RED)](./task-004-filter-switch-sync-test.md)
- [Task 004: 筛选切换与请求契约实现 (GREEN)](./task-004-filter-switch-sync-impl.md)
- [Task 005: 清空搜索联动测试 (RED)](./task-005-clear-search-reset-test.md)
- [Task 005: 清空搜索联动实现 (GREEN)](./task-005-clear-search-reset-impl.md)
- [Task 006: 筛选可访问性测试 (RED)](./task-006-filter-accessibility-test.md)
- [Task 006: 筛选可访问性实现 (GREEN)](./task-006-filter-accessibility-impl.md)

## Commit Boundaries

- Boundary A: `task-001` ~ `task-002`（URL 状态 + 分类规则基础能力）
- Boundary B: `task-003` ~ `task-004`（结果过滤与请求契约回归）
- Boundary C: `task-005` ~ `task-006`（清空联动与可访问性）

---

## Execution Handoff

计划已保存到 `docs/plans/2026-02-27-search-results-filter-plan/`。

建议执行方式：

1. 使用 `superpowers:executing-plans` 编排执行（推荐）
2. 使用 `superpowers:agent-team-driven-development` 并行执行
3. 当前会话串行执行（较慢）
