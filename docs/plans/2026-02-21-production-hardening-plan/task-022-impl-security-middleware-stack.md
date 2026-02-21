# Task 022: Implement security middleware stack

**depends-on**: task-001

## Description

Implement three simple security middlewares in one task: secure headers, CORS configuration, and body size limit. These are straightforward middleware configurations that don't require individual Red-Green cycles.

## Execution Context

**Task Number**: 022 of 032
**Phase**: Security Middleware
**Prerequisites**: Error response types (task-001) for CORS error responses

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 3 — "超大请求体返回 413"
**Additional ref**: architecture.md Section 2 (middleware stack), best-practices.md #1 (secure headers), #11 (CORS), #12 (body limit)

## Files to Modify/Create

- Create: `internal/httpx/middleware_security.go`
- Create: `internal/httpx/middleware_security_test.go`

## Steps

### Step 1: Implement secure headers middleware

- Create `internal/httpx/middleware_security.go`
- Implement `func SecureHeaders() echo.MiddlewareFunc` that sets:
  - `X-Content-Type-Options: nosniff`
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `X-Frame-Options: DENY`
  - `Permissions-Policy: camera=(), microphone=(), geolocation=()`
- For HTML page endpoints (/app), also set CSP header

### Step 2: Configure CORS

- In the same file, implement a CORS config factory function:
  - `func CORSConfig(allowedOrigins []string) middleware.CORSConfig`
  - Allow methods: GET, POST, OPTIONS
  - Allow headers: Authorization, X-API-Key, Content-Type
  - MaxAge: 3600
  - Origins configurable via environment variable `CORS_ALLOWED_ORIGINS`

### Step 3: Test body limit

- In `middleware_security_test.go`:
  - `TestBodyLimit_OversizedRequest_Returns413` — send a POST with body > 1MB; verify 413 status
  - `TestSecureHeaders_AreSet` — verify all security headers are present in response
  - `TestCORS_AllowedOrigin` — verify CORS headers for allowed origin
  - `TestCORS_DisallowedOrigin` — verify no CORS headers for unknown origin

### Step 4: Implement and verify

- Body limit uses Echo's built-in `middleware.BodyLimit("1MB")`
- Run tests: `go test ./internal/httpx/ -run TestBodyLimit -v`
- Run tests: `go test ./internal/httpx/ -run TestSecureHeaders -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run "TestBodyLimit|TestSecureHeaders|TestCORS" -v
go test ./... -count=1
```

## Success Criteria

- All 4 security headers set on every response
- CSP set on HTML page responses
- Body > 1MB returns 413
- CORS only allows configured origins
- All tests pass
