# Task 008: Create search service instrumenter

**depends-on**: task-004

## Description

Create an `InstrumentedSearchService` decorator that wraps `search.Searcher` and records cache hit/miss counts and cache size. Also add a `Len() int` method to `CachedQueryService` to expose current cache size.

## Execution Context

**Task Number**: 008 of 012
**Phase**: Core Features
**Prerequisites**: SearchMetrics available, CachedQueryService understood

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "首次搜索记录为缓存未命中", "相同参数的二次搜索记录为缓存命中"

## Files to Modify/Create

- Modify: `internal/search/cached_query_service.go` (add `Len()` method)
- Create: `internal/metrics/search_instrumenter.go`
- Create: `internal/metrics/search_instrumenter_test.go`

## Steps

### Step 1: Add Len() to CachedQueryService

Add a `Len() int` method to `CachedQueryService` that returns `s.cache.Len()`. This exposes the current number of cached entries.

**Verification**: `go build ./...` compiles

### Step 2: Write test (Red)

Create `internal/metrics/search_instrumenter_test.go`. Create a mock `search.Searcher` that returns configurable results. The mock should also implement a `Len() int` method (via embedding or separate interface).

Test cases using fresh registry + SearchMetrics:
1. **Cache miss** — First query with the mock (mock inner is called), verify `npan_search_queries_total{result="miss"}` is 1, `{result="hit"}` is 0
2. **Cache hit** — Second identical query (use a real `CachedQueryService` wrapping the mock to get real caching behavior), verify `npan_search_queries_total{result="hit"}` is 1
3. **Cache size updated** — After a query, verify `npan_search_cache_size` is 1

Note: the InstrumentedSearchService needs to detect cache hits vs misses. The approach: track whether the inner `Searcher` (the actual `CachedQueryService`) called through to its own inner service. Alternative simpler approach: the decorator wraps a `CachedQueryService` directly and compares cache `Len()` before and after the query — if Len increased, it was a miss.

**Verification**: `go test ./internal/metrics/... -run TestInstrumentedSearchService` → MUST FAIL

### Step 3: Implement (Green)

Create `internal/metrics/search_instrumenter.go`. Define:

- `CacheLenner` interface with `Len() int` (to detect cache state)
- `InstrumentedSearchService` struct holding `inner search.Searcher`, `lenner CacheLenner`, `metrics *SearchMetrics`
- Constructor: `NewInstrumentedSearchService(inner search.Searcher, lenner CacheLenner, m *SearchMetrics) *InstrumentedSearchService`
- `Query()`: record cache Len before, call inner.Query(), record Len after. If Len increased → miss, else → hit. Update `QueriesTotal` and `CacheSize`.
- `Ping()`: delegate to inner

The `InstrumentedSearchService` implements `search.Searcher` interface.

**Verification**: `go test ./internal/metrics/... -run TestInstrumentedSearchService` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestInstrumentedSearchService -v
go build ./...
```

## Success Criteria

- Cache hit/miss detection works correctly
- Cache size gauge updates after each query
- Decorator transparently wraps Searcher interface
