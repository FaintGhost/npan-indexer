# Task 011: Integration wiring in main.go

**depends-on**: task-002, task-004, task-006, task-008, task-010

## Description

在 `cmd/server/main.go` 中组装所有新组件：CachedQueryService 包装 QueryService，创建 SearchActivityTracker，将 tracker 传入 CachedQueryService 和 RequestLimiter，将 CachedQueryService 传入 NewHandlers。

## Execution Context

**Task Number**: 011 of 012
**Phase**: Integration
**Prerequisites**: 所有功能实现任务已完成

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: 综合所有 Scenario

## Files to Modify/Create

- Modify: `cmd/server/main.go` — 组装组件
- Modify: `internal/search/cached_query_service.go` — 如需要，添加 tracker 集成到 CachedQueryService
- Modify: `internal/service/sync_manager.go` — 添加 ActivityChecker 支持，使 limiter 可使用 tracker

## Steps

### Step 1: Create tracker in main.go

- 在 `main.go` 中创建 `tracker := search.NewSearchActivityTracker(5)` （5 秒活跃窗口）

### Step 2: Create CachedQueryService in main.go

- 创建 `cachedService := search.NewCachedQueryService(queryService, 256, 30*time.Second)`
- 如果 CachedQueryService 需要 tracker，传入：`search.NewCachedQueryService(queryService, 256, 30*time.Second, tracker)`

### Step 3: Pass CachedQueryService to handlers

- 修改 `handlers := httpx.NewHandlers(cfg, cachedService, syncManager)`
- 由于 Task 006 已将参数类型改为 `search.Searcher`，此处直接传入 CachedQueryService

### Step 4: Pass tracker to SyncManager

- 修改 `SyncManagerArgs` 添加 `ActivityChecker` 字段
- 在 SyncManager 的 `run()` 方法中，创建 limiter 后设置 checker：`limiter.SetActivityChecker(m.activityChecker)`
- 在 main.go 中传入 tracker：`SyncManagerArgs{..., ActivityChecker: tracker}`

### Step 5: Verify Integration

- 全项目编译成功
- 运行全部测试
- **Verification**: `go build ./... && go test ./... -v`

## Verification Commands

```bash
go build ./...
go test ./... -v
```

## Success Criteria

- 全项目编译成功
- 所有测试通过
- main.go 正确组装所有组件
