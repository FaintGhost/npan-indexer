# Task 025: Test health check endpoints

**depends-on**: (none)

## Description

Write tests for the health check endpoints. `/healthz` should always return 200 (liveness probe). `/readyz` should return 200 when Meilisearch is reachable and 503 when it's not (readiness probe).

## Execution Context

**Task Number**: 025 of 032
**Phase**: Health & Ops
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- "healthz 始终返回 200"
- "readyz 在 Meilisearch 可用时返回 200"
- "readyz 在 Meilisearch 不可达时返回 503"

## Files to Modify/Create

- Create: `internal/httpx/health_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 3 health check scenarios exist in Feature 6

### Step 2: Implement Tests (Red)

- Create `internal/httpx/health_test.go` with:
  - `TestHealthz_AlwaysReturns200` — GET /healthz returns 200 with JSON `{"status": "ok"}`
  - `TestReadyz_MeiliAvailable_Returns200` — with mock query service returning nil from Ping(), GET /readyz returns 200 with `{"status": "ready"}`
  - `TestReadyz_MeiliUnavailable_Returns503` — with mock query service returning error from Ping(), GET /readyz returns 503 with `{"status": "not_ready"}`
- Tests need a mock/stub for the query service's `Ping()` method
- If QueryService doesn't have a `Ping()` method yet, the test should define the expected interface
- **Verification**: Tests FAIL since `Readyz` handler and `Ping` method don't exist

## Verification Commands

```bash
go test ./internal/httpx/ -run "TestHealthz|TestReadyz" -v
```

## Success Criteria

- All 3 BDD scenarios covered
- Readyz tests use test doubles for Meilisearch connectivity
- Response bodies match expected JSON structure
