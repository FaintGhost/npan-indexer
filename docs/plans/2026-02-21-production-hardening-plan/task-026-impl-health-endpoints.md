# Task 026: Implement health check endpoints

**depends-on**: task-025

## Description

Implement the enhanced health check endpoints. Update the existing `/healthz` handler and create a new `/readyz` handler that checks Meilisearch connectivity via a `Ping()` method on the query service.

## Execution Context

**Task Number**: 026 of 032
**Phase**: Health & Ops
**Prerequisites**: Health check tests (task-025) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: Feature 6, all 3 scenarios

## Files to Modify/Create

- Modify: `internal/httpx/handlers.go` — update `Health` handler, add `Readyz` handler
- Modify: `internal/search/query_service.go` or `internal/search/meili_index.go` — add `Ping()` method

## Steps

### Step 1: Add Ping method to search layer

- Add a `Ping() error` method to the query service or meili index that checks Meilisearch connectivity
- This can use Meilisearch client's health check endpoint or a simple search query

### Step 2: Update Health handler

- Ensure `Health` returns `{"status": "ok"}` with 200 (may already work)

### Step 3: Implement Readyz handler

- Add `func (h *Handlers) Readyz(c *echo.Context) error`
- Call query service's `Ping()`
- If error: return 503 with `{"status": "not_ready", "meili": "unreachable"}`
- If ok: return 200 with `{"status": "ready"}`

### Step 4: Register routes

- Ensure `/readyz` is registered in server.go (this may already be covered by task-009)

### Step 5: Verify (Green)

- Run tests from task-025
- **Verification**: `go test ./internal/httpx/ -run "TestHealthz|TestReadyz" -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run "TestHealthz|TestReadyz" -v
go test ./internal/search/ -v
go test ./... -count=1
```

## Success Criteria

- All health check tests pass
- `/healthz` always returns 200
- `/readyz` returns 503 when Meili is unreachable
- Ping method properly checks Meilisearch connectivity
