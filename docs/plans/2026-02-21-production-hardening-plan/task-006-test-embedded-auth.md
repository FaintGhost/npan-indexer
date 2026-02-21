# Task 006: Test embedded auth middleware

**depends-on**: (none)

## Description

Write tests for the embedded auth middleware. This middleware is applied to `/api/v1/app/*` routes and automatically sets context values `auth_mode=embedded` and `allow_config_fallback=true`, enabling the handler to use server-side credentials without requiring an API key from the client.

## Execution Context

**Task Number**: 006 of 032
**Phase**: Authentication
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 1 — "公开端点不需要认证" (specifically /api/v1/app/* endpoints)

## Files to Modify/Create

- Create: `internal/httpx/middleware_auth_test.go` (append to file from task-004, or separate section)

## Steps

### Step 1: Verify Scenario

- Confirm Scenario Outline "公开端点不需要认证" exists with /api/v1/app/* examples

### Step 2: Implement Tests (Red)

- Add tests to `internal/httpx/middleware_auth_test.go`:
  - `TestEmbeddedAuth_SetsAuthMode` — middleware sets `auth_mode` to "embedded" in context
  - `TestEmbeddedAuth_SetsConfigFallback` — middleware sets `allow_config_fallback` to true in context
  - `TestEmbeddedAuth_CallsNextHandler` — middleware calls the next handler in the chain
  - `TestEmbeddedAuth_NoAPIKeyRequired` — request passes through without any auth headers
- Tests should create the middleware, wrap a handler that reads context values, and verify
- **Verification**: Tests should FAIL since `EmbeddedAuth` doesn't exist

## Verification Commands

```bash
go test ./internal/httpx/ -run TestEmbeddedAuth -v
```

## Success Criteria

- Tests verify context values are correctly set
- Tests confirm no authentication is required
- Tests fail in Red state
