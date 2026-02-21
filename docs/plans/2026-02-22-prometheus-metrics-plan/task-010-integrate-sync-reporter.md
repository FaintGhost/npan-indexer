# Task 010: Integrate SyncReporter into SyncManager

**depends-on**: task-006

## Description

Add `SyncReporter` to `SyncManager` so sync start/finish events are reported to Prometheus metrics.

## Execution Context

**Task Number**: 010 of 012
**Phase**: Integration
**Prerequisites**: SyncReporter interface defined

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "同步启动时 running gauge 变为 1", "全量同步完成后指标正确更新", "同步被取消时状态为 cancelled", "同步失败时状态为 error"

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go`

## Steps

### Step 1: Add SyncReporter to SyncManager

1. Add `metricsReporter metrics.SyncReporter` field to `SyncManager` struct (can be nil)
2. Add `MetricsReporter metrics.SyncReporter` field to `SyncManagerArgs` struct
3. In `NewSyncManager`, assign `metricsReporter: args.MetricsReporter`

Note: Import `npan/internal/metrics` package. The `SyncReporter` is an interface, so nil means disabled.

### Step 2: Instrument Start method

In `SyncManager.Start()`, after setting `m.running = true` and before `go func()`:
- If `m.metricsReporter != nil`, call `m.metricsReporter.ReportSyncStarted(request.Mode)`

### Step 3: Instrument run method completion

In the `run()` method, the final status is determined at the end. Add metrics reporting at the three exit points:

1. **Full sync path** — In `run()`, after the final status is set (done/error/cancelled), before return:
   - Build a `metrics.SyncEvent` with the effective mode, final status, duration since start, aggregate stats
   - Call `m.metricsReporter.ReportSyncFinished(event)` (if non-nil)

2. **Incremental path** — In `runIncrementalPath()`, after status is set (done/error/cancelled):
   - Build `SyncEvent` with incremental mode, status, duration, and IncrementalStats
   - Call reporter

Guard all calls with `if m.metricsReporter != nil` to keep it zero-invasive.

### Step 4: Verify compilation

**Verification**: `go build ./...` compiles. Existing behavior is unchanged when `MetricsReporter` is nil (which it currently is in main.go).

## Verification Commands

```bash
go build ./...
go test ./internal/service/... -v
```

## Success Criteria

- `SyncManagerArgs` has `MetricsReporter` field
- `SyncManager.Start()` calls `ReportSyncStarted` when reporter is non-nil
- `SyncManager.run()` and `runIncrementalPath()` call `ReportSyncFinished` at all exit points
- All guard checks (`if m.metricsReporter != nil`) are in place
- Existing tests still pass
