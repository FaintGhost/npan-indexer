# Task 008: Test route structure with auth enforcement

**depends-on**: task-005, task-007

## Description

Write integration-level tests that verify the complete route structure and auth enforcement. Each endpoint group must enforce the correct authentication: public endpoints (healthz, readyz, /app, /api/v1/app/*) require no auth; API endpoints (/api/v1/*) require X-API-Key; admin endpoints (/api/v1/admin/*) require X-API-Key.

## Execution Context

**Task Number**: 008 of 032
**Phase**: Route Restructure
**Prerequisites**: Both auth middlewares must be implemented (task-005, task-007)

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- Scenario Outline: "管理端点必须经过认证" (4 examples)
- Scenario Outline: "管理端点（admin 组）必须经过认证" (3 examples)
- Scenario Outline: "公开端点不需要认证" (5 examples)

## Files to Modify/Create

- Create: `internal/httpx/server_routes_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 3 Scenario Outlines exist in BDD specs

### Step 2: Implement Tests (Red)

- Create `internal/httpx/server_routes_test.go` with:
  - `TestRoutes_PublicEndpoints_NoAuthRequired` — table-driven test for: GET /healthz, GET /readyz, GET /app, GET /api/v1/app/search?q=test, GET /api/v1/app/download-url?file_id=1 — all should NOT return 401
  - `TestRoutes_APIEndpoints_RequireAuth` — table-driven test for: POST /api/v1/token, GET /api/v1/search/remote?q=test, GET /api/v1/search/local?q=test, GET /api/v1/download-url?file_id=1 — all should return 401 without API key
  - `TestRoutes_AdminEndpoints_RequireAuth` — table-driven test for: POST /api/v1/admin/sync/full, GET /api/v1/admin/sync/full/progress, POST /api/v1/admin/sync/full/cancel — all should return 401 without API key
  - `TestRoutes_APIEndpoints_WithKey_Pass` — same API endpoints with valid X-API-Key should NOT return 401
- Tests use `NewServer()` + `httptest.NewRecorder` to make real HTTP requests through the full middleware chain
- **Verification**: Tests should FAIL since new route structure doesn't exist yet

## Verification Commands

```bash
go test ./internal/httpx/ -run TestRoutes -v
```

## Success Criteria

- All BDD scenario outline examples are covered
- Tests verify auth enforcement at the route level (not just middleware unit)
- Tests use table-driven patterns for the scenario outlines
