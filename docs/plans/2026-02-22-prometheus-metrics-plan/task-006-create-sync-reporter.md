# Task 006: Create SyncReporter interface and implementation

**depends-on**: task-003

## Description

Create `internal/metrics/sync_reporter.go` with the `SyncReporter` interface and `PrometheusSyncReporter` implementation that translates sync events into Prometheus metrics.

## Execution Context

**Task Number**: 006 of 012
**Phase**: Core Features
**Prerequisites**: SyncMetrics available

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "同步启动时 running gauge 变为 1", "全量同步完成后指标正确更新", "同步被取消时状态为 cancelled", "同步失败时状态为 error"

## Files to Modify/Create

- Create: `internal/metrics/sync_reporter.go`
- Create: `internal/metrics/sync_reporter_test.go`

## Steps

### Step 1: Write test (Red)

Create `internal/metrics/sync_reporter_test.go`. Use a fresh registry and SyncMetrics per test.

Test cases:
1. **ReportSyncStarted** — call with `SyncModeFull`, verify `npan_sync_running` gauge is 1
2. **ReportSyncFinished (done)** — call with status "done", mode "full", stats showing 500 files indexed, 2 failed. Verify:
   - `npan_sync_running` is 0
   - `npan_sync_tasks_total{mode="full",status="done"}` is 1
   - `npan_sync_files_indexed_total{mode="full"}` is 500
   - `npan_sync_files_failed_total{mode="full"}` is 2
   - `npan_sync_duration_seconds` has an observation
3. **ReportSyncFinished (cancelled)** — verify `npan_sync_tasks_total{...,status="cancelled"}` increments
4. **ReportSyncFinished (error)** — verify `npan_sync_tasks_total{...,status="error"}` increments
5. **Incremental stats** — call with IncrStats containing upserted/deleted/skipped values, verify `npan_sync_incremental_changes_total` labels

**Verification**: `go test ./internal/metrics/... -run TestSyncReporter` → MUST FAIL

### Step 2: Implement (Green)

Create `internal/metrics/sync_reporter.go`. Define:

- `SyncEvent` struct with fields: Mode (`models.SyncMode`), Status (string), Duration (`time.Duration`), Stats (`models.CrawlStats`), IncrStats (`*models.IncrementalSyncStats`)
- `SyncReporter` interface with `ReportSyncStarted(mode models.SyncMode)` and `ReportSyncFinished(event SyncEvent)`
- `PrometheusSyncReporter` struct holding `*SyncMetrics`
- `NewPrometheusSyncReporter(m *SyncMetrics) *PrometheusSyncReporter`

`ReportSyncStarted`: set Running gauge to 1.

`ReportSyncFinished`: set Running to 0, increment TasksTotal with mode+status, observe DurationSeconds, add FilesIndexedTotal and FilesFailedTotal from CrawlStats. If IncrStats is non-nil, add IncrementalChangesTotal for each operation type.

**Verification**: `go test ./internal/metrics/... -run TestSyncReporter` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestSyncReporter -v
```

## Success Criteria

- All 5 test cases pass
- SyncReporter correctly translates events to Prometheus metrics
- Interface is nil-safe (callers check for nil before calling)
