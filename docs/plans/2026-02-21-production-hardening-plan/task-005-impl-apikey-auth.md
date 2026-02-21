# Task 005: Implement API Key auth middleware

**depends-on**: task-004, task-001

## Description

Implement the `APIKeyAuth` middleware function that validates API keys from `X-API-Key` header or `Authorization: Bearer` header using constant-time comparison. Returns 401 with standardized `ErrorResponse` for invalid/missing keys.

## Execution Context

**Task Number**: 005 of 032
**Phase**: Authentication
**Prerequisites**: Error response types (task-001) and auth tests (task-004) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: Feature 1, Scenarios 1-4

## Files to Modify/Create

- Create: `internal/httpx/middleware_auth.go`

## Steps

### Step 1: Implement APIKeyAuth middleware

- Create `internal/httpx/middleware_auth.go`
- Implement `func APIKeyAuth(adminKey string) echo.MiddlewareFunc`
- Extract key from `X-API-Key` header first, fallback to `Authorization: Bearer <key>` header
- Use `crypto/subtle.ConstantTimeCompare` for key comparison
- On failure, return 401 using `writeErrorResponse` with `ErrCodeUnauthorized`
- Include a helper `parseBearerHeader(header string) string` that extracts the token from "Bearer <token>"

### Step 2: Verify (Green)

- Run tests from task-004
- **Verification**: `go test ./internal/httpx/ -run TestAPIKeyAuth -v` — all should pass

## Verification Commands

```bash
go test ./internal/httpx/ -run TestAPIKeyAuth -v
go vet ./internal/httpx/
```

## Success Criteria

- All tests from task-004 pass
- Uses `crypto/subtle.ConstantTimeCompare` (not `==`)
- Supports both X-API-Key and Bearer token auth
- Returns standardized ErrorResponse on failure
