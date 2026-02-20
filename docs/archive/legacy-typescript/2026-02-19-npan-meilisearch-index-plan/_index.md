# Npan 外部索引与下载代理实施计划

> **执行要求**：实现阶段请使用 `executing-plans` 技能按任务逐项执行。

**Goal:** 建立一套独立于平台弱搜索的检索能力，支持全量/增量同步到 Meilisearch，并按需获取真实下载 URL。  
**Architecture:** 通过云盘 OpenAPI 拉取目录树和文件元数据，写入 Meilisearch 作为检索主数据源；下载场景实时调用云盘下载接口换取临时 `download_url`。同步任务使用统一限流、重试和 checkpoint 来控制限流风险与恢复能力。  
**Tech Stack:** Bun + TypeScript、Meilisearch、JSON/SQLite 状态存储、Bun Test。  

**Design Support:**
- [BDD 规格](./bdd-specs.md)

## 约束与边界

- 不改造云盘平台内部搜索，仅构建外部索引层。
- 不在索引中持久化临时下载链接。
- 所有外部依赖测试必须使用 test doubles（网络/API/Meili）。
- 先红后绿：每个场景优先创建失败测试，再实现通过。

## Execution Plan

- [Task 001: 场景 1 红测（全量遍历与限流）](./task-001-red-full-crawl-rate-limit.md)
- [Task 002: 场景 1 绿测（全量遍历与限流实现）](./task-002-green-full-crawl-rate-limit.md)
- [Task 003: 场景 2 红测（重试与断点续跑）](./task-003-red-retry-checkpoint.md)
- [Task 004: 场景 2 绿测（重试与断点续跑实现）](./task-004-green-retry-checkpoint.md)
- [Task 005: 场景 3 红测（Meili 文档结构与可检索性）](./task-005-red-meili-schema-searchability.md)
- [Task 006: 场景 3 绿测（Meili 写入与索引设置实现）](./task-006-green-meili-schema-searchability.md)
- [Task 007: 场景 4 红测（增量与删除同步）](./task-007-red-incremental-delete-sync.md)
- [Task 008: 场景 4 绿测（增量与删除同步实现）](./task-008-green-incremental-delete-sync.md)
- [Task 009: 场景 5 红测（Meili 查询接口）](./task-009-red-meili-query-api.md)
- [Task 010: 场景 5 绿测（Meili 查询实现）](./task-010-green-meili-query-api.md)
- [Task 011: 场景 6 红测（下载 URL 代理）](./task-011-red-download-url-proxy.md)
- [Task 012: 场景 6 绿测（下载 URL 代理实现）](./task-012-green-download-url-proxy.md)
- [Task 013: 联调与压测验证（限流阈值与稳定性）](./task-013-integration-load-verification.md)

---

## Execution Handoff

计划已保存到 `docs/plans/2026-02-19-npan-meilisearch-index-plan/`。  
可选执行方式：

1. Orchestrated Execution（推荐）：使用 `executing-plans` 技能执行全计划。  
2. Direct Agent Team：使用 `agent-team-driven-development` 技能并行执行独立任务。  
3. Manual / Serial：按任务文件顺序在当前会话逐个执行。
