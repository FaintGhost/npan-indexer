# Task 012: Update Dockerfile and docker-compose

**depends-on**: task-011

## Description

Update infrastructure files to expose the metrics port.

## Execution Context

**Task Number**: 012 of 012
**Phase**: Refinement
**Prerequisites**: Metrics server wired in main.go

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenario**: "Metrics 端点在独立端口上可用" (infrastructure enablement)

## Files to Modify/Create

- Modify: `Dockerfile`
- Modify: `docker-compose.yml`

## Steps

### Step 1: Update Dockerfile

Find the existing `EXPOSE` line and add port 9091. The result should be:
```
EXPOSE 1323 9091
```

### Step 2: Update docker-compose.yml

In the npan service:
1. Add `"9091:9091"` to the `ports` section
2. Add `METRICS_ADDR=:9091` to the `environment` section

### Step 3: Verify

**Verification**: `docker compose config` validates without errors (if docker compose is available). Otherwise, manual review of YAML syntax.

## Verification Commands

```bash
# Verify Dockerfile syntax
docker build --check . 2>/dev/null || echo "Docker not available, skip"

# Verify docker-compose syntax
docker compose config 2>/dev/null || echo "Docker compose not available, skip"

# Final: run full test suite
go test ./...
```

## Success Criteria

- Dockerfile exposes both ports
- docker-compose.yml maps metrics port and sets environment variable
- Full test suite passes: `go test ./...`
