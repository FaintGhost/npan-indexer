# Task 003: [TEST] SQLite checkpoint 恢复与清理 (RED)

**depends-on**: (none)

## Description

先用失败测试锁定 SQLite checkpoint 在 `resume_progress`、`force_rebuild` 与 crawl 完成后的生命周期语义，避免迁移后只保存进度而漏掉真正的恢复与清理逻辑。

## Execution Context

**Task Number**: 003 of 009
**Phase**: Core Features
**Prerequisites**: 已阅读 `internal/service/sync_manager.go`、`internal/indexer/full_crawl.go`、`internal/service/sync_manager_routing_test.go`

## BDD Scenario

```gherkin
Scenario: resume=true 时应从 SQLite checkpoint 恢复 crawl 队列
  Given 某个根目录在 SQLite 中已有未完成的 CrawlCheckpoint
  And 用户以 resume_progress=true 启动全量同步
  When SyncManager 启动该根目录的 crawl
  Then crawler 应从已有 checkpoint 队列恢复
  And 进度中的根目录状态应继续累加而不是从零开始

Scenario: force_rebuild 或 resume=false 时应清除 SQLite checkpoint
  Given 某个根目录在 SQLite 中已有旧的 CrawlCheckpoint
  When 用户以 force_rebuild=true 或 resume_progress=false 启动全量同步
  Then SyncManager 应在 crawl 前清除该根目录的 SQLite checkpoint
  And crawler 应从根目录重新开始遍历

Scenario: crawl 完成后应清理对应的 SQLite checkpoint
  Given 某个根目录在同步过程中持续写入 SQLite checkpoint
  When 该根目录全量 crawl 成功结束
  Then 对应 checkpoint 记录应被清理或标记为空状态
  And 下次 resume 不应恢复到已经完成的旧队列
```

**Spec Source**: `../2026-03-08-sync-state-sqlite-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/service/sync_manager_routing_test.go`
- Modify: `internal/storage/sqlite_store_test.go`
- Modify: `internal/indexer/full_crawl_test.go`（如需锁定完成后 clear 语义）

## Steps

### Step 1: Verify Scenario

- 确认 BDD 场景覆盖 resume、clear-before-run、clear-after-success 三类 checkpoint 生命周期语义。

### Step 2: Implement Test (Red)

- 基于临时 SQLite DB 新增失败测试，覆盖：
  - resume=true 时从 SQLite checkpoint 恢复旧队列
  - force_rebuild / resume=false 启动前会清空 checkpoint
  - full crawl success 后 checkpoint 被清理，后续 resume 不再恢复旧数据
- 使用 fake API 与 fake index 隔离外部依赖。

### Step 3: Verify Red Failure

- 运行目标测试并确认失败。
- 失败应指向“checkpoint factory 未接入 / success 后未 clear / clear 时机错误”。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/service ./internal/storage ./internal/indexer -run 'Checkpoint|Resume|ForceRebuild|SQLite' -count=1
```

## Success Criteria

- checkpoint 生命周期相关测试稳定处于 Red。
- 失败直指 SQLite checkpoint 恢复或清理语义缺失。
- 测试不依赖真实外部服务。
