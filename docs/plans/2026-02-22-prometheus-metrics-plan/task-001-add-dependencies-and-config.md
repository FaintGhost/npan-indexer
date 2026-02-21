# Task 001: Add dependencies and config

**depends-on**: (none)

## Description

Add Prometheus and echo-contrib Go dependencies, and add `MetricsAddr` config field.

## Execution Context

**Task Number**: 001 of 012
**Phase**: Setup
**Prerequisites**: Go module initialized, project compiles

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenario**: "MetricsAddr 为空时禁用 metrics" (config part)

## Files to Modify/Create

- Modify: `go.mod` (add dependencies)
- Modify: `internal/config/config.go` (add MetricsAddr field)
- Modify: `internal/config/validate.go` (add MetricsAddr to LogValue)

## Steps

### Step 1: Add Go dependencies

Run `go get` to add:
- `github.com/prometheus/client_golang`
- `github.com/labstack/echo-contrib`

### Step 2: Add MetricsAddr to Config struct

Add `MetricsAddr string` field to the `Config` struct in `internal/config/config.go`.

In the `Load()` function, add:
```
MetricsAddr: readString("METRICS_ADDR", ":9091"),
```

### Step 3: Update LogValue

In `internal/config/validate.go`, add `slog.String("MetricsAddr", c.MetricsAddr)` to the `LogValue()` method.

### Step 4: Verify compilation

**Verification**: `go build ./...` compiles successfully.

## Verification Commands

```bash
go build ./...
```

## Success Criteria

- `prometheus/client_golang` and `echo-contrib` appear in `go.mod`
- `Config.MetricsAddr` is populated from `METRICS_ADDR` env var with default `:9091`
- Project compiles without errors
