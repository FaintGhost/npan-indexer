# Task 007: Implement embedded auth middleware

**depends-on**: task-006

## Description

Implement the `EmbeddedAuth` middleware function that sets context values for embedded frontend requests, allowing handlers to use server-side credentials automatically.

## Execution Context

**Task Number**: 007 of 032
**Phase**: Authentication
**Prerequisites**: Embedded auth tests (task-006) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 1 — "公开端点不需要认证"

## Files to Modify/Create

- Modify: `internal/httpx/middleware_auth.go` (add to file created in task-005)

## Steps

### Step 1: Implement EmbeddedAuth middleware

- Add `func EmbeddedAuth() echo.MiddlewareFunc` to `middleware_auth.go`
- Set `c.Set("auth_mode", "embedded")` and `c.Set("allow_config_fallback", true)`
- Call `next(c)` to continue the chain

### Step 2: Verify (Green)

- Run tests from task-006
- **Verification**: `go test ./internal/httpx/ -run TestEmbeddedAuth -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestEmbeddedAuth -v
```

## Success Criteria

- All tests from task-006 pass
- Middleware correctly sets both context values
- No authentication check is performed
