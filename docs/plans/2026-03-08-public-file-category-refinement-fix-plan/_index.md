# Public file_category refinement 修复计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 修复 React InstantSearch public 搜索中 `file_category` 分类筛选失效的问题，确保非“全部”分类能够正确驱动 Meilisearch refinement 并返回结果。

**Architecture:** 保持当前 InstantSearch + Meilisearch public 搜索架构不变，只修正 `SearchFilters` 对 refinement token 的使用方式，并补齐能够锁定真实 token 语义的测试。修复应继续让默认过滤基线保留在请求层，`file_category` refinement 只负责在其之上叠加。

**Tech Stack:** React 19, react-instantsearch, @meilisearch/instant-meilisearch, Vitest, Testing Library, Playwright, Bun

**Design Support:**
- [BDD Specs](../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md)
- [Architecture](../2026-03-08-react-instantsearch-alignment-design/architecture.md)

**Execution Plan:**
- [Task 001: 锁定 refinement token 回归测试](./task-001-file-category-refinement-token-test.md)
- [Task 001: 修复 refinement token 驱动实现](./task-001-file-category-refinement-token-impl.md)
- [Task 002: 验证 public URL 与结果回归闸门](./task-002-public-filter-regression-verification.md)

---

## Execution Handoff

**Plan complete and saved to `docs/plans/2026-03-08-public-file-category-refinement-fix-plan/`. Execution options:**

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
