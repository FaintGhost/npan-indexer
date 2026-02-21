# Task 004: Create SearchMetrics definitions

**depends-on**: task-002

## Description

Create `internal/metrics/search_metrics.go` with the `SearchMetrics` struct holding all search and Meilisearch-related Prometheus metrics.

## Execution Context

**Task Number**: 004 of 012
**Phase**: Foundation
**Prerequisites**: metrics package exists with NewRegistry

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "首次搜索记录为缓存未命中", "Meilisearch 操作耗时被记录", "文档总量 gauge 反映真实值"

## Files to Modify/Create

- Create: `internal/metrics/search_metrics.go`
- Create: `internal/metrics/search_metrics_test.go`

## Steps

### Step 1: Write test (Red)

Create `internal/metrics/search_metrics_test.go`. Use a fresh `prometheus.NewRegistry()` per test. Test that `NewSearchMetrics(reg)`:
- Registers without panic
- `QueriesTotal` counter vec works with labels `result="hit"` and `result="miss"`
- `CacheSize` gauge can be set and read
- `MeiliDurationSeconds` histogram vec works with label `op="search"`
- `MeiliErrorsTotal` counter vec works with label `op="search"`
- `MeiliDocumentsTotal` gauge can be set
- `MeiliUpsertedDocsTotal` counter can be incremented

**Verification**: `go test ./internal/metrics/... -run TestSearchMetrics` → MUST FAIL

### Step 2: Implement (Green)

Create `internal/metrics/search_metrics.go`. Define `SearchMetrics` struct and `NewSearchMetrics(reg prometheus.Registerer) *SearchMetrics`.

Metric names and types:
- `npan_search_queries_total` — CounterVec, labels: `result`
- `npan_search_cache_size` — Gauge
- `npan_meili_operation_duration_seconds` — HistogramVec, labels: `op`, buckets: `{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5}`
- `npan_meili_operation_errors_total` — CounterVec, labels: `op`
- `npan_meili_documents_total` — Gauge
- `npan_meili_upserted_docs_total` — Counter

**Verification**: `go test ./internal/metrics/... -run TestSearchMetrics` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestSearchMetrics -v
```

## Success Criteria

- All 6 search metrics registered and usable
- Tests pass with independent registry
