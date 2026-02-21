# Task 002: Create metrics registry with Go runtime collectors

**depends-on**: task-001

## Description

Create `internal/metrics/registry.go` with a factory function that returns a custom `*prometheus.Registry` with Go runtime and process collectors registered.

## Execution Context

**Task Number**: 002 of 012
**Phase**: Foundation
**Prerequisites**: prometheus/client_golang dependency available

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenario**: "标准 Go 指标可在 metrics 端点获取"

## Files to Modify/Create

- Create: `internal/metrics/registry.go`
- Create: `internal/metrics/registry_test.go`

## Steps

### Step 1: Write test (Red)

Create `internal/metrics/registry_test.go`. Test that `NewRegistry()`:
- Returns a non-nil `*prometheus.Registry`
- When gathered, contains metric families for `go_goroutines`, `go_memstats_alloc_bytes`, `process_cpu_seconds_total`

Use `registry.Gather()` to collect all metric families and check the names are present.

**Verification**: `go test ./internal/metrics/... -run TestNewRegistry` → MUST FAIL

### Step 2: Implement (Green)

Create `internal/metrics/registry.go` with package `metrics`. Implement `NewRegistry() *prometheus.Registry` that:
- Creates `prometheus.NewRegistry()`
- Registers `collectors.NewGoCollector()` and `collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})`
- Returns the registry

**Verification**: `go test ./internal/metrics/... -run TestNewRegistry` → MUST PASS

## Verification Commands

```bash
go test ./internal/metrics/... -run TestNewRegistry -v
```

## Success Criteria

- `NewRegistry()` returns a registry that exposes Go runtime and process metrics
- Test passes
