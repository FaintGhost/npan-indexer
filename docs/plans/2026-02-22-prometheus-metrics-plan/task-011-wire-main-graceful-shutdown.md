# Task 011: Wire metrics in main.go with graceful shutdown

**depends-on**: task-005, task-006, task-007, task-008, task-009, task-010

## Description

Wire all metrics components together in `cmd/server/main.go`: create the registry, instantiate metrics, wrap services with instrumenters, start the metrics server on a separate port, and implement dual-server graceful shutdown.

## Execution Context

**Task Number**: 011 of 012
**Phase**: Integration
**Prerequisites**: All metrics components and integrations complete

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "Metrics 端点在独立端口上可用", "Metrics 端点不在主业务端口上暴露", "MetricsAddr 为空时禁用 metrics", "主服务器先于 metrics 服务器关闭", "关闭有超时保护"

## Files to Modify/Create

- Modify: `cmd/server/main.go`

## Steps

### Step 1: Create metrics infrastructure

In `main()`, after config load and validation:

1. Create registry: `promReg := metrics.NewRegistry()`
2. Create metric sets: `syncMetrics := metrics.NewSyncMetrics(promReg)` and `searchMetrics := metrics.NewSearchMetrics(promReg)`

### Step 2: Wrap services with instrumenters

Replace the direct service construction with instrumented versions:

1. After creating `meiliIndex`, wrap it: `instrMeili := metrics.NewInstrumentedMeiliIndex(meiliIndex, searchMetrics)`
2. Pass `instrMeili` to `search.NewQueryService()` instead of `meiliIndex` (now accepts `IndexOperator`)
3. After creating `cachedService`, wrap it: `instrSearch := metrics.NewInstrumentedSearchService(cachedService, cachedService, searchMetrics)` (cachedService implements both `Searcher` and `CacheLenner`)
4. Pass `instrSearch` to `httpx.NewHandlers()` instead of `cachedService`
5. Create sync reporter: `syncReporter := metrics.NewPrometheusSyncReporter(syncMetrics)` and pass to `SyncManagerArgs.MetricsReporter`

### Step 3: Pass registry to Echo server

Update `httpx.NewServer` call to pass `promReg` as the new parameter.

### Step 4: Start metrics server conditionally

If `cfg.MetricsAddr != ""`:
1. Create metrics server: `metricsServer := metrics.NewMetricsServer(cfg.MetricsAddr, promReg)`
2. Start in a goroutine (same pattern as main server)
3. Log: `"指标服务启动"` with addr

### Step 5: Implement dual graceful shutdown

Replace the existing single-server shutdown with:
1. On signal received, log shutdown message
2. Create shutdown context with 15s timeout
3. Shut down main HTTP server first (waits for in-flight requests)
4. If metrics server exists, create a 5s timeout context and shut it down
5. Both use the same error logging pattern

### Step 6: Verify

**Verification**: `go build ./...` compiles. Manual test: start server, curl `http://localhost:9091/metrics`, verify response contains `npan_` prefixed metrics and `go_goroutines`.

## Verification Commands

```bash
go build ./...
# Manual integration test:
# METRICS_ADDR=:9091 go run ./cmd/server
# curl http://localhost:9091/metrics | grep npan_
# curl http://localhost:1323/metrics  # should 404
```

## Success Criteria

- Metrics server starts on configured port
- `/metrics` on metrics port returns all registered metrics
- `/metrics` on main port returns 404
- Setting `METRICS_ADDR=""` disables metrics server
- Graceful shutdown closes main server first, then metrics server
- All services use instrumented wrappers
