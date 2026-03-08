# React InstantSearch 纠偏实施计划

> **For Claude:** REQUIRED SUB-SKILL: 使用 Skill 工具加载 `superpowers:executing-plans` 执行本计划。

## Goal

在保留现有 React InstantSearch 架构与 legacy `AppSearch` fallback 的前提下，把 public 搜索纠偏到“官方行为 + 关键对齐”：恢复 search-as-you-type，补齐最关键的 legacy 搜索语义（query 预处理、默认过滤），并建立结果对比与回滚门槛。

## Architecture

本次实施只做最小纠偏，不推翻现有 public 搜索栈：

1. 在 public 输入链路中恢复“输入变化驱动搜索”，同时保留 Enter / 搜索按钮的立即触发能力。
2. 在 Meilisearch 请求发出前补一层独立的 query adapter，对齐 legacy `preprocessQuery()` 的关键语义，但不改写用户输入框和 URL 中的原始 query。
3. 通过 InstantSearch 官方配置入口统一注入 `type=file`、`is_deleted=false`、`in_trash=false` 默认过滤，并用 focused regression + 对比门槛收口发布风险。

## Tech Stack

- React 19 + Vite + Bun
- `react-instantsearch` + `@meilisearch/instant-meilisearch`
- Vitest + React Testing Library + MSW
- Playwright
- Go / Connect / Meilisearch（作为 legacy 语义参考与 fallback 链路）

## Constraints

- 必须严格执行 BDD：先 Red，再 Green。
- 单测必须使用测试替身隔离网络与外部服务：
  - 前端测试使用 MSW / module mock / fake timers。
  - 不依赖真实 Meilisearch 服务验证输入行为与请求参数。
- 本轮不重做 public bootstrap、routing、InfiniteHits、下载链路。
- 本轮不做排序规则、ranking rules、synonyms、typo 等更深层相关性调优。
- 本轮不删除 legacy `AppSearch` fallback。
- 本轮不承诺完整复刻 legacy `All -> Last` fallback；若关键对齐后仍有阻塞级差异，再单独立项。

## Design Support

- [Design Index](../2026-03-08-react-instantsearch-alignment-design/_index.md)
- [BDD Specs](../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md)
- [Architecture](../2026-03-08-react-instantsearch-alignment-design/architecture.md)
- [Best Practices](../2026-03-08-react-instantsearch-alignment-design/best-practices.md)

## Execution Plan

- [Task 001: public 输入即搜与立即提交测试 (RED)](./task-001-public-search-as-you-type-test.md)
- [Task 001: public 输入即搜与立即提交实现 (GREEN)](./task-001-public-search-as-you-type-impl.md)
- [Task 002: query 预处理 adapter 测试 (RED)](./task-002-query-normalizer-test.md)
- [Task 002: query 预处理 adapter 实现 (GREEN)](./task-002-query-normalizer-impl.md)
- [Task 003: public 默认过滤与 refinement 叠加测试 (RED)](./task-003-public-default-filters-test.md)
- [Task 003: public 默认过滤与 refinement 叠加实现 (GREEN)](./task-003-public-default-filters-impl.md)
- [Task 004: 结果对比与回滚闸门验证](./task-004-alignment-verification-and-rollout-gate.md)

## Commit Boundaries

- Boundary A: `task-001`（纠正 public 输入语义）
- Boundary B: `task-002` ~ `task-003`（补齐 query 预处理与默认过滤）
- Boundary C: `task-004`（结果对比、E2E 与灰度/回滚门槛）

---

## Execution Handoff

计划已保存到 `docs/plans/2026-03-08-react-instantsearch-alignment-plan/`。

建议执行方式：

1. 使用 `superpowers:executing-plans` 编排执行（推荐）
2. 使用 `superpowers:agent-team-driven-development` 并行执行
3. 当前会话串行执行（较慢）
