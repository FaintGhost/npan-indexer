# React InstantSearch 直连 Meilisearch 实施计划

> **For Claude:** REQUIRED SUB-SKILL: 使用 Skill 工具加载 `superpowers:executing-plans` 执行本计划。

## Goal

为公开搜索页交付基于官方 `react-instantsearch` + `@meilisearch/instant-meilisearch` 的直连 Meilisearch 实现，同时保留 `AppSearch` 作为灰度与回滚兜底，并确保下载链路继续走 `AppService.AppDownloadURL`。

## Architecture

本次实施分三层推进：

1. 先补“公开搜索配置”后端契约与安全边界，让浏览器只拿到 dedicated public search config。
2. 再把搜索页从手工 `useInfiniteQuery + URLSearchParams + 本地 ext 过滤` 切到 InstantSearch 状态模型，并完成结果渲染、高亮、InfiniteHits、routing、refinement。
3. 最后做下载集成、E2E、文档与回滚开关收口，确保新旧双栈可切换。

## Tech Stack

- Go 1.25 + Connect-RPC + Buf
- Meilisearch + `meilisearch-go`
- React 19 + Vite + `react-instantsearch` + `@meilisearch/instant-meilisearch`
- Vitest + React Testing Library + MSW
- Playwright + Docker Compose CI

## Constraints

- 必须严格执行 BDD：先 Red，再 Green。
- 单测必须使用测试替身隔离网络与外部服务：
  - Go 测试使用 `httptest` / stub IndexManager。
  - 前端测试使用 MSW / module mock。
- 首批不删除 `AppSearch`。
- 首批不引入 tenant token，只做公共搜索 + search-only key。
- 首批不预先复刻旧后端 `preprocessQuery()` 与 `All -> Last` fallback；是否补 adapter 由验证结果决定。

## Design Support

- [Design Index](../2026-03-07-react-instantsearch-design/_index.md)
- [BDD Specs](../2026-03-07-react-instantsearch-design/bdd-specs.md)
- [Architecture](../2026-03-07-react-instantsearch-design/architecture.md)
- [Best Practices](../2026-03-07-react-instantsearch-design/best-practices.md)

## Execution Plan

- [Task 001: 公开搜索配置契约与安全边界测试 (RED)](./task-001-public-search-config-test.md)
- [Task 001: 公开搜索配置契约与安全边界实现 (GREEN)](./task-001-public-search-config-impl.md)
- [Task 002: 搜索页配置引导与回退测试 (RED)](./task-002-search-bootstrap-fallback-test.md)
- [Task 002: 搜索页配置引导与回退实现 (GREEN)](./task-002-search-bootstrap-fallback-impl.md)
- [Task 003: `file_category` 索引契约测试 (RED)](./task-003-file-category-index-test.md)
- [Task 003: `file_category` 索引契约实现 (GREEN)](./task-003-file-category-index-impl.md)
- [Task 004: InstantSearch 结果列表与高亮测试 (RED)](./task-004-instantsearch-results-test.md)
- [Task 004: InstantSearch 结果列表与高亮实现 (GREEN)](./task-004-instantsearch-results-impl.md)
- [Task 005: routing 与 refinement 分类筛选测试 (RED)](./task-005-routing-refinement-test.md)
- [Task 005: routing 与 refinement 分类筛选实现 (GREEN)](./task-005-routing-refinement-impl.md)
- [Task 006: 直连搜索结果下载集成测试 (RED)](./task-006-download-integration-test.md)
- [Task 006: 直连搜索结果下载集成实现 (GREEN)](./task-006-download-integration-impl.md)
- [Task 007: 全链路验证与发布闸门](./task-007-verification-and-rollout-gate.md)

## Commit Boundaries

- Boundary A: `task-001` ~ `task-002`（公开搜索配置与双栈引导）
- Boundary B: `task-003` ~ `task-005`（索引字段、结果渲染、routing/refinement）
- Boundary C: `task-006` ~ `task-007`（下载集成、E2E、文档与回滚验证）

---

## Execution Handoff

计划已保存到 `docs/plans/2026-03-07-react-instantsearch-plan/`。

建议执行方式：

1. 使用 `superpowers:executing-plans` 编排执行（推荐）
2. 使用 `superpowers:agent-team-driven-development` 并行执行
3. 当前会话串行执行（较慢）
