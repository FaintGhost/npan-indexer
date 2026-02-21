# Task 005: Create metrics server

**depends-on**: task-002

## Description

Create `internal/metrics/server.go` with a factory function for a standalone HTTP server that serves `/metrics` from a custom Prometheus gatherer.

## Execution Context

**Task Number**: 005 of 012
**Phase**: Foundation
**Prerequisites**: NewRegistry() available

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "Metrics 端点在独立端口上可用", "关闭有超时保护"

## Files to Modify/Create

- Create: `internal/metrics/server.go`
- Create: `internal/metrics/server_test.go`

## Steps

### Step 1: Write test (Red)

Create `internal/metrics/server_test.go`. Test `NewMetricsServer`:
- Create a registry with `NewRegistry()`, create a server with `NewMetricsServer(":0", reg)` (port 0 for ephemeral)
- Start the server in a goroutine using `ListenAndServe`
- Send GET to `/metrics` endpoint
- Assert status 200, Content-Type contains "text/plain", body contains "go_goroutines"
- Send GET to a non-metrics path like `/foo` → assert 404
- Shutdown the server and verify it stops cleanly

**Verification**: `go test ./internal/metrics/... -run TestMetricsServer` → MUST FAIL

### Step 2: Implement (Green)

Create `internal/metrics/server.go`. Implement `NewMetricsServer(addr string, gatherer prometheus.Gatherer) *http.Server` that:
- Creates an `http.ServeMux`
- Registers `/metrics` with `promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})`
- Returns `*http.Server` with the mux as handler and appropriate timeouts (ReadHeaderTimeout: 5s, ReadTimeout: 10s, WriteTimeout: 10s)

**Verification**: `go test ./internal/metrics/... -run TestMetricsServer` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestMetricsServer -v
```

## Success Criteria

- `/metrics` returns 200 with Prometheus text format
- Contains Go runtime metrics
- Server can be shut down cleanly
