# Task 003: Create SyncMetrics definitions

**depends-on**: task-002

## Description

Create `internal/metrics/sync_metrics.go` with the `SyncMetrics` struct holding all sync-related Prometheus metrics, and a constructor that registers them.

## Execution Context

**Task Number**: 003 of 012
**Phase**: Foundation
**Prerequisites**: metrics package exists with NewRegistry

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "同步启动时 running gauge 变为 1", "全量同步完成后指标正确更新", "增量同步完成后指标正确更新"

## Files to Modify/Create

- Create: `internal/metrics/sync_metrics.go`
- Create: `internal/metrics/sync_metrics_test.go`

## Steps

### Step 1: Write test (Red)

Create `internal/metrics/sync_metrics_test.go`. Use a fresh `prometheus.NewRegistry()` per test. Test that `NewSyncMetrics(reg)`:
- Successfully registers without panic
- `SyncMetrics.Running` gauge can be set to 1 and read back via `testutil.ToFloat64`
- `SyncMetrics.TasksTotal` counter vec can be incremented with labels `mode="full", status="done"` and read back
- `SyncMetrics.DurationSeconds` histogram vec can observe a value with label `mode="full"`
- `SyncMetrics.FilesIndexedTotal` counter vec works with label `mode="full"`
- `SyncMetrics.FilesFailedTotal` counter vec works with label `mode="full"`
- `SyncMetrics.IncrementalChangesTotal` counter vec works with label `op="upsert"`

**Verification**: `go test ./internal/metrics/... -run TestSyncMetrics` → MUST FAIL

### Step 2: Implement (Green)

Create `internal/metrics/sync_metrics.go`. Define the `SyncMetrics` struct with fields as listed in the architecture doc (see `../2026-02-22-prometheus-metrics-design/architecture.md` section 2). Implement `NewSyncMetrics(reg prometheus.Registerer) *SyncMetrics` that creates and registers all metrics.

Metric names and types:
- `npan_sync_tasks_total` — CounterVec, labels: `mode`, `status`
- `npan_sync_duration_seconds` — HistogramVec, labels: `mode`, buckets: `{1, 5, 30, 60, 300, 600, 1800}`
- `npan_sync_files_indexed_total` — CounterVec, labels: `mode`
- `npan_sync_files_failed_total` — CounterVec, labels: `mode`
- `npan_sync_running` — Gauge
- `npan_sync_incremental_changes_total` — CounterVec, labels: `op`

**Verification**: `go test ./internal/metrics/... -run TestSyncMetrics` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestSyncMetrics -v
```

## Success Criteria

- All 6 sync metrics registered and usable
- Tests pass with independent registry
