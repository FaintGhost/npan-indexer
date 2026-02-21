# Task 007: Extract IndexOperator interface and create MeiliIndex instrumenter

**depends-on**: task-004

## Description

Extract an `IndexOperator` interface from `MeiliIndex`'s public methods, then create an `InstrumentedMeiliIndex` decorator in the metrics package that wraps `IndexOperator` and records operation duration and errors.

## Execution Context

**Task Number**: 007 of 012
**Phase**: Core Features
**Prerequisites**: SearchMetrics available, MeiliIndex implementation understood

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "Meilisearch 操作耗时被记录", "Meilisearch 操作错误被计数", "文档总量 gauge 反映真实值"

## Files to Modify/Create

- Modify: `internal/search/meili_index.go` (extract IndexOperator interface)
- Create: `internal/metrics/meili_instrumenter.go`
- Create: `internal/metrics/meili_instrumenter_test.go`

## Steps

### Step 1: Extract IndexOperator interface

In `internal/search/meili_index.go`, add an `IndexOperator` interface that covers the public methods of `MeiliIndex`:
- `EnsureSettings(ctx context.Context) error`
- `UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error`
- `DeleteDocuments(ctx context.Context, docIDs []string) error`
- `Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error)`
- `Ping() error`
- `DocumentCount(ctx context.Context) (int64, error)`

`MeiliIndex` already satisfies this interface — no changes to its methods.

**Verification**: `go build ./...` compiles

### Step 2: Write test (Red)

Create `internal/metrics/meili_instrumenter_test.go`. Create a mock implementing `search.IndexOperator` (all methods return configurable results/errors).

Test cases using fresh registry + SearchMetrics:
1. **UpsertDocuments success** — call with 10 docs, verify `npan_meili_operation_duration_seconds{op="upsert"}` has 1 observation, `npan_meili_upserted_docs_total` is 10, no error counter increment
2. **UpsertDocuments error** — mock returns error, verify `npan_meili_operation_errors_total{op="upsert"}` is 1, `npan_meili_upserted_docs_total` is 0
3. **Search success** — verify `npan_meili_operation_duration_seconds{op="search"}` has observation
4. **Search error** — verify error counter increments
5. **DocumentCount** — mock returns 1000, verify `npan_meili_documents_total` gauge is 1000
6. **DeleteDocuments** — verify duration and error metrics for op="delete"

**Verification**: `go test ./internal/metrics/... -run TestInstrumentedMeiliIndex` → MUST FAIL

### Step 3: Implement (Green)

Create `internal/metrics/meili_instrumenter.go`. Implement `InstrumentedMeiliIndex` struct that:
- Holds `inner search.IndexOperator` and `metrics *SearchMetrics`
- Constructor: `NewInstrumentedMeiliIndex(inner search.IndexOperator, m *SearchMetrics) *InstrumentedMeiliIndex`
- Implements `search.IndexOperator` interface
- Each method: records start time, calls inner, observes duration in `MeiliDurationSeconds{op=...}`, on error increments `MeiliErrorsTotal{op=...}`, on success updates relevant counters/gauges

**Verification**: `go test ./internal/metrics/... -run TestInstrumentedMeiliIndex` → MUST PASS

### Step 4: Update QueryService to accept IndexOperator

Modify `internal/search/query_service.go`: change `QueryService.index` field type from `*MeiliIndex` to `IndexOperator`. Update `NewQueryService` parameter type accordingly. This allows passing either `*MeiliIndex` or `*InstrumentedMeiliIndex`.

**Verification**: `go build ./...` compiles

## Verification Commands

```bash
go test ./internal/metrics/... -run TestInstrumentedMeiliIndex -v
go build ./...
```

## Success Criteria

- `IndexOperator` interface extracted in search package
- `InstrumentedMeiliIndex` correctly instruments all operations
- `QueryService` accepts `IndexOperator` interface
- All tests pass, project compiles
