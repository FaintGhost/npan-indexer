# Task 009: Implement route restructure

**depends-on**: task-008

## Description

Restructure `server.go` to implement the new route design with three groups: public (no auth), API (X-API-Key required), admin (X-API-Key required). Replace `/demo` with `/app`, remove `requireAPIAccess()` from handlers, integrate auth middleware at the group level.

## Execution Context

**Task Number**: 009 of 032
**Phase**: Route Restructure
**Prerequisites**: Route tests (task-008) must exist; auth middlewares (task-005, task-007) must be implemented

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: All Scenario Outlines from Feature 1 (auth enforcement per endpoint group)

## Files to Modify/Create

- Modify: `internal/httpx/server.go` — complete route restructure
- Modify: `internal/httpx/handlers.go` — remove `requireAPIAccess()` method and all calls to it; update `resolveAuthOptions` to read `allow_config_fallback` from context
- Modify: `internal/httpx/server_demo_test.go` — update or remove tests referencing /demo routes (replaced by /app)

## Steps

### Step 1: Restructure server.go

- Replace the current flat route registration with grouped structure:
  - Public group: `/healthz`, `/readyz`, `/app`, `/app/*`
  - App API group (`/api/v1/app`): uses `EmbeddedAuth()` middleware — `/search`, `/download-url`
  - API group (`/api/v1`): uses `APIKeyAuth(cfg.AdminAPIKey)` — `/token`, `/search/remote`, `/search/local`, `/download-url`
  - Admin group (`/api/v1/admin`): uses `APIKeyAuth(cfg.AdminAPIKey)` — `/sync/full`, `/sync/full/progress`, `/sync/full/cancel`
- Pass `cfg` to `NewServer` (it currently only receives `*Handlers` — update signature to also accept config or pass AdminAPIKey)
- Update HTML path resolution from `web/demo/` to `web/app/`

### Step 2: Remove requireAPIAccess from handlers

- Delete the `requireAPIAccess()` method from handlers.go
- Remove all `if !h.requireAPIAccess(c) { return nil }` guards from handler methods
- Update `resolveAuthOptions` to check `c.Get("allow_config_fallback")` context value (set by EmbeddedAuth middleware) instead of receiving boolean parameter

### Step 3: Update existing tests

- Update `server_demo_test.go` to test `/app` routes instead of `/demo` routes
- Update handler tests if they reference `requireAPIAccess`

### Step 4: Verify (Green)

- Run route tests from task-008
- Run all existing tests to check for regressions
- **Verification**: `go test ./internal/httpx/ -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestRoutes -v
go test ./internal/httpx/ -v
go test ./... -count=1
```

## Success Criteria

- All route tests from task-008 pass
- `requireAPIAccess()` method is completely removed
- Auth is handled purely via middleware, not in-handler checks
- `/demo` routes replaced with `/app`
- No regression in existing tests
