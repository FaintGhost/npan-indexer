# Sync Verification Implementation Plan

## Goal

实现四层验证体系，确保全量同步过程中所有文件被正确索引，并提供多维度的交叉验证。

## Constraints

- 向后兼容：旧进度 JSON 加载不报错
- 不改变外部 API 路由/契约
- 改动范围限制在 `indexer`、`service`、`search`、`models` 包及前端 `sync-*` 组件
- 所有新逻辑必须有测试覆盖

## Architecture Reference

See [design/_index.md](../2026-02-21-sync-verification-design/_index.md)

## Execution Plan

### Phase 1: Model Layer (基础数据结构)

- [Task 001: Add new fields to CrawlStats and SyncProgressState models](./task-001-models.md)

### Phase 2: Retry Infrastructure (L3 前置)

- [Task 002: Test isRetriable for MeiliSearch errors](./task-002-test-isretriable-meili.md)
- [Task 003: Implement isRetriable MeiliSearch support](./task-003-impl-isretriable-meili.md)
- [Task 004: Test WithRetryVoid](./task-004-test-withretryvoid.md)
- [Task 005: Implement WithRetryVoid](./task-005-impl-withretryvoid.md)

### Phase 3: Full Crawl Counter Fix (L1 + L2 + L3)

- [Task 006: Test accurate counters and skip behavior in RunFullCrawl](./task-006-test-fullcrawl-counters.md)
- [Task 007: Implement counter fix, discovery tracking, and upsert retry+skip in RunFullCrawl](./task-007-impl-fullcrawl-counters.md)

### Phase 4: Post-Sync Reconciliation (L4)

- [Task 008: Test DocumentCount method](./task-008-test-documentcount.md)
- [Task 009: Implement DocumentCount method](./task-009-impl-documentcount.md)
- [Task 010: Test reconciliation logic](./task-010-test-reconciliation.md)
- [Task 011: Implement reconciliation in sync_manager](./task-011-impl-reconciliation.md)

### Phase 5: Frontend

- [Task 012: Update frontend schemas and tests](./task-012-frontend-schemas.md)
- [Task 013: Update sync progress display with verification UI](./task-013-frontend-display.md)

### Phase 6: Integration

- [Task 014: Build verification and final commit](./task-014-build-verify.md)
