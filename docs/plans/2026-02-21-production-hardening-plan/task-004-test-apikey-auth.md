# Task 004: Test API Key auth middleware

**depends-on**: (none)

## Description

Write tests for the API Key authentication middleware. The middleware should validate requests using `X-API-Key` header or `Authorization: Bearer <key>` header. It should use constant-time comparison to prevent timing attacks. Missing or invalid keys return 401 with `ErrorResponse` format.

## Execution Context

**Task Number**: 004 of 032
**Phase**: Authentication
**Prerequisites**: None — test task, can start immediately

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- "未携带 API Key 访问管理端点返回 401"
- "携带错误 API Key 访问管理端点返回 401"
- "通过 X-API-Key Header 认证成功"
- "通过 Bearer Token 认证成功"

## Files to Modify/Create

- Create: `internal/httpx/middleware_auth_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 4 scenarios exist in BDD specs Feature 1

### Step 2: Implement Tests (Red)

- Create `internal/httpx/middleware_auth_test.go` with:
  - `TestAPIKeyAuth_NoKey_Returns401` — request without any auth header returns 401 JSON with code "UNAUTHORIZED"
  - `TestAPIKeyAuth_WrongKey_Returns401` — request with wrong X-API-Key returns 401
  - `TestAPIKeyAuth_ValidXAPIKey_Passes` — request with correct X-API-Key header calls next handler (returns 200)
  - `TestAPIKeyAuth_ValidBearerToken_Passes` — request with `Authorization: Bearer <correct-key>` calls next handler
  - `TestAPIKeyAuth_EmptyBearerToken_Returns401` — `Authorization: Bearer ` (empty) returns 401
  - `TestAPIKeyAuth_ResponseFormat` — verify 401 response body contains `code` and `message` fields, does NOT contain `stack`, `path`, `config`
- Tests should create the middleware with a known key, wrap a simple handler, and use `httptest`
- **Verification**: Tests should FAIL since `APIKeyAuth` function doesn't exist yet

## Verification Commands

```bash
go test ./internal/httpx/ -run TestAPIKeyAuth -v
```

## Success Criteria

- Tests cover all 4 BDD scenarios for API Key auth
- Tests verify constant-time comparison indirectly (same response format for wrong key)
- Tests verify response body format
- Tests fail with compilation error (Red state)
