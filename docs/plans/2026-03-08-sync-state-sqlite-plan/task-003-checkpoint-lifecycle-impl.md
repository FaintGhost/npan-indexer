# Task 003: [IMPL] SQLite checkpoint 恢复与清理 (GREEN)

**depends-on**: task-003-checkpoint-lifecycle-test, task-001-sqlite-state-store-impl, task-002-sync-manager-state-impl

## Description

把 checkpoint 生命周期完整切到 SQLite，包括 resume 恢复、force rebuild / resume=false 清理，以及 crawl 成功后的 checkpoint 清空，确保断点恢复语义与迁移前一致。

## Execution Context

**Task Number**: 003 of 009
**Phase**: Core Features
**Prerequisites**: `task-003-checkpoint-lifecycle-test` 已稳定处于 Red；`task-001` 与 `task-002` 已提供 SQLite store 与可注入 checkpoint factory

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

- Modify: `internal/service/sync_manager.go`
- Modify: `internal/storage/sqlite_store.go`
- Modify: `internal/storage/sqlite_store_test.go`
- Modify: `internal/service/sync_manager_routing_test.go`
- Modify: `internal/indexer/full_crawl.go`
- Modify: `internal/indexer/full_crawl_test.go`

## Steps

### Step 1: Verify Scenario

- 确认本任务只聚焦 checkpoint 生命周期，不扩展到 Admin/CLI 协议与文档。

### Step 2: Implement Logic (Green)

- 将 checkpoint 读取、保存、清理完整接到 SQLite checkpoint store factory。
- 保持 `buildCheckpointFilePath(...)` 生成的逻辑 key 语义，避免破坏 progress 中已有字段。
- 确认 force rebuild、resume=false、resume=true、crawl success 的 checkpoint 行为与 BDD 一致。
- 如有必要，补齐 `FullCrawlDeps.CheckpointStore.Clear()` 成功终态的测试与实现衔接。

### Step 3: Verify Pass

- 运行 `task-003` 的 checkpoint 生命周期测试并确认通过。
- 验证完成后的 checkpoint 不会在下次 resume 中误恢复。

### Step 4: Refactor & Safety Check

- 清理任何残留的 JSON checkpoint 直连路径。
- 确认服务层与 indexer 层对 checkpoint 语义的一致性。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/service ./internal/storage ./internal/indexer -run 'Checkpoint|Resume|ForceRebuild|SQLite' -count=1
GOCACHE=/tmp/go-build go test ./internal/service ./internal/indexer -count=1
```

## Success Criteria

- checkpoint 生命周期测试全部转绿。
- resume / force_rebuild / resume=false / success-clear 语义保持正确。
- 不再残留运行时 JSON checkpoint 主路径。
